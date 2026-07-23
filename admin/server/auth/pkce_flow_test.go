package auth_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/oauth"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/stretchr/testify/require"
)

const (
	oauth002AuthorizationCodeGrant = "authorization_code"
	oauth002RefreshTokenGrant      = "refresh_token"
	oauth002RedirectURI            = "https://client.example/callback"
	oauth002Verifier               = "oauth002-verifier-with-enough-entropy-for-pkce"
)

var (
	oauth002Sequence              atomic.Uint64
	oauth002ErrAccessPersistence  = errors.New("injected access token persistence failure")
	oauth002ErrRefreshPersistence = errors.New("injected refresh token persistence failure")
)

type oauth002FormResult struct {
	status int
	body   string
	err    error
}

type oauth002TokenFailureDB struct {
	database.DB
	failRefresh bool
}

func (d *oauth002TokenFailureDB) InsertUserAuthToken(ctx context.Context, opts *database.InsertUserAuthTokenOptions) (*database.UserAuthToken, error) {
	if opts.Refresh == d.failRefresh {
		if d.failRefresh {
			return nil, oauth002ErrRefreshPersistence
		}
		return nil, oauth002ErrAccessPersistence
	}
	return d.DB.InsertUserAuthToken(ctx, opts)
}

func TestOAUTH002AuthorizationCodeExchange(t *testing.T) {
	// A real-Postgres exchange must bind every PKCE input, consume once, and persist both tokens atomically.
	fix := testadmin.New(t)
	tokenURL := fix.ExternalURL() + "/auth/oauth/token"

	t.Run("binds client redirect verifier and expiry before consumption", func(t *testing.T) {
		tests := []struct {
			name       string
			changeForm func(url.Values)
			expiresOn  time.Time
			wantBody   string
			retryable  bool
		}{
			{
				name: "wrong client",
				changeForm: func(values url.Values) {
					values.Set("client_id", database.AuthClientIDRillWeb)
				},
				expiresOn: time.Now().Add(time.Minute),
				wantBody:  "invalid client ID",
				retryable: true,
			},
			{
				name: "wrong redirect",
				changeForm: func(values url.Values) {
					values.Set("redirect_uri", "https://client.example/wrong")
				},
				expiresOn: time.Now().Add(time.Minute),
				wantBody:  "invalid redirect URI",
				retryable: true,
			},
			{
				name: "wrong verifier",
				changeForm: func(values url.Values) {
					values.Set("code_verifier", "wrong-verifier")
				},
				expiresOn: time.Now().Add(time.Minute),
				wantBody:  "invalid code verifier",
				retryable: true,
			},
			{
				name:       "expired code",
				changeForm: func(url.Values) {},
				expiresOn:  time.Now().Add(-time.Minute),
				wantBody:   "authorization code has expired",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				user, _ := fix.NewUser(t)
				client := oauth002NewClient(t, fix.Admin.DB, false)
				code := oauth002StoreCode(t, fix.Admin.DB, user.ID, client.ID, tt.expiresOn)
				before := oauth002UserTokenCount(t, fix.Admin.DB, user.ID)

				values := oauth002ValidExchangeForm(code, client.ID)
				tt.changeForm(values)
				res := oauth002PostForm(t.Context(), tokenURL, values)

				require.NoError(t, res.err)
				require.Equal(t, http.StatusBadRequest, res.status, res.body)
				require.Contains(t, res.body, tt.wantBody)
				require.Equal(t, before, oauth002UserTokenCount(t, fix.Admin.DB, user.ID), "an invalid exchange must mint nothing")
				_, err := fix.Admin.DB.FindAuthorizationCode(t.Context(), code)
				require.NoError(t, err, "validation failures do not consume the authorization code")

				if tt.retryable {
					res = oauth002PostForm(t.Context(), tokenURL, oauth002ValidExchangeForm(code, client.ID))
					require.NoError(t, res.err)
					require.Equal(t, http.StatusOK, res.status, res.body)
					require.Equal(t, before+1, oauth002UserTokenCount(t, fix.Admin.DB, user.ID))
					_, err = fix.Admin.DB.FindAuthorizationCode(t.Context(), code)
					require.ErrorIs(t, err, database.ErrNotFound, "a valid exchange consumes the code")
				}
			})
		}
	})

	t.Run("two concurrent valid exchanges have one winner", func(t *testing.T) {
		user, _ := fix.NewUser(t)
		client := oauth002NewClient(t, fix.Admin.DB, false)
		code := oauth002StoreCode(t, fix.Admin.DB, user.ID, client.ID, time.Now().Add(time.Minute))
		before := oauth002UserTokenCount(t, fix.Admin.DB, user.ID)

		// Both requests may validate the same row, but the transactional delete is
		// the single-use claim, so only its winner can persist a token.
		start := make(chan struct{})
		results := make(chan oauth002FormResult, 2)
		var group sync.WaitGroup
		for range 2 {
			group.Add(1)
			go func() {
				defer group.Done()
				<-start
				results <- oauth002PostForm(context.Background(), tokenURL, oauth002ValidExchangeForm(code, client.ID))
			}()
		}
		close(start)
		group.Wait()
		close(results)

		successes := 0
		failures := 0
		for res := range results {
			require.NoError(t, res.err)
			switch res.status {
			case http.StatusOK:
				successes++
			case http.StatusBadRequest:
				failures++
				require.Contains(t, res.body, "no such authorization code found")
			default:
				require.Failf(t, "unexpected exchange response", "status=%d body=%s", res.status, res.body)
			}
		}
		require.Equal(t, 1, successes)
		require.Equal(t, 1, failures)
		require.Equal(t, before+1, oauth002UserTokenCount(t, fix.Admin.DB, user.ID))
		require.Len(t, oauth002ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 1)
		_, err := fix.Admin.DB.FindAuthorizationCode(t.Context(), code)
		require.ErrorIs(t, err, database.ErrNotFound)

		replay := oauth002PostForm(t.Context(), tokenURL, oauth002ValidExchangeForm(code, client.ID))
		require.NoError(t, replay.err)
		require.Equal(t, http.StatusBadRequest, replay.status, replay.body)
		require.Equal(t, before+1, oauth002UserTokenCount(t, fix.Admin.DB, user.ID), "replay must not mint another token")
	})

	t.Run("access token persistence failure rolls back code consumption", func(t *testing.T) {
		user, _ := fix.NewUser(t)
		client := oauth002NewClient(t, fix.Admin.DB, false)
		code := oauth002StoreCode(t, fix.Admin.DB, user.ID, client.ID, time.Now().Add(time.Minute))
		before := oauth002UserTokenCount(t, fix.Admin.DB, user.ID)

		oauth002WithDB(t, fix, &oauth002TokenFailureDB{DB: fix.Admin.DB}, func() {
			res := oauth002PostForm(t.Context(), tokenURL, oauth002ValidExchangeForm(code, client.ID))
			require.NoError(t, res.err)
			require.Equal(t, http.StatusInternalServerError, res.status, res.body)
			require.Contains(t, res.body, oauth002ErrAccessPersistence.Error())
		})

		require.Equal(t, before, oauth002UserTokenCount(t, fix.Admin.DB, user.ID))
		require.Empty(t, oauth002ClientTokens(t, fix.Admin.DB, user.ID, client.ID), "failed persistence must leave no access token")
		_, err := fix.Admin.DB.FindAuthorizationCode(t.Context(), code)
		require.NoError(t, err, "rolled-back consumption keeps the code retryable")

		res := oauth002PostForm(t.Context(), tokenURL, oauth002ValidExchangeForm(code, client.ID))
		require.NoError(t, res.err)
		require.Equal(t, http.StatusOK, res.status, res.body)
		require.Len(t, oauth002ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 1)
	})

	t.Run("refresh token persistence failure leaves no orphan access token", func(t *testing.T) {
		user, _ := fix.NewUser(t)
		client := oauth002NewClient(t, fix.Admin.DB, true)
		code := oauth002StoreCode(t, fix.Admin.DB, user.ID, client.ID, time.Now().Add(time.Minute))
		before := oauth002UserTokenCount(t, fix.Admin.DB, user.ID)

		// The access insert happens first; failing the refresh insert proves that
		// both token rows and code consumption share one rollback boundary.
		oauth002WithDB(t, fix, &oauth002TokenFailureDB{DB: fix.Admin.DB, failRefresh: true}, func() {
			res := oauth002PostForm(t.Context(), tokenURL, oauth002ValidExchangeForm(code, client.ID))
			require.NoError(t, res.err)
			require.Equal(t, http.StatusInternalServerError, res.status, res.body)
			require.Contains(t, res.body, oauth002ErrRefreshPersistence.Error())
		})

		require.Equal(t, before, oauth002UserTokenCount(t, fix.Admin.DB, user.ID))
		require.Empty(t, oauth002ClientTokens(t, fix.Admin.DB, user.ID, client.ID), "rollback must remove the provisional access token")
		_, err := fix.Admin.DB.FindAuthorizationCode(t.Context(), code)
		require.NoError(t, err, "failed all-or-nothing issuance keeps the code retryable")

		res := oauth002PostForm(t.Context(), tokenURL, oauth002ValidExchangeForm(code, client.ID))
		require.NoError(t, res.err)
		require.Equal(t, http.StatusOK, res.status, res.body)
		var payload oauth.TokenResponse
		require.NoError(t, json.Unmarshal([]byte(res.body), &payload))
		require.NotEmpty(t, payload.AccessToken)
		require.NotEmpty(t, payload.RefreshToken)

		tokens := oauth002ClientTokens(t, fix.Admin.DB, user.ID, client.ID)
		require.Len(t, tokens, 2)
		refreshTokens := 0
		for _, token := range tokens {
			if token.Refresh {
				refreshTokens++
			}
		}
		require.Equal(t, 1, refreshTokens)
		_, err = fix.Admin.DB.FindAuthorizationCode(t.Context(), code)
		require.ErrorIs(t, err, database.ErrNotFound)
	})
}

func oauth002NewClient(t *testing.T, db database.DB, refresh bool) *database.AuthClient {
	t.Helper()
	grants := []string{oauth002AuthorizationCodeGrant}
	if refresh {
		grants = append(grants, oauth002RefreshTokenGrant)
	}
	client, err := db.InsertAuthClient(t.Context(), oauth002ID("client"), "offline_access", grants, []string{oauth002RedirectURI})
	require.NoError(t, err)
	return client
}

func oauth002StoreCode(t *testing.T, db database.DB, userID, clientID string, expiresOn time.Time) string {
	t.Helper()
	code := oauth002ID("code")
	_, err := db.InsertAuthorizationCode(t.Context(), code, userID, clientID, oauth002RedirectURI, oauth002Challenge(oauth002Verifier), "S256", expiresOn)
	require.NoError(t, err)
	return code
}

func oauth002ValidExchangeForm(code, clientID string) url.Values {
	return url.Values{
		"grant_type":             {oauth002AuthorizationCodeGrant},
		"code":                   {code},
		"client_id":              {clientID},
		"redirect_uri":           {oauth002RedirectURI},
		"code_verifier":          {oauth002Verifier},
		"token_response_version": {"standard"},
	}
}

func oauth002PostForm(ctx context.Context, endpoint string, values url.Values) oauth002FormResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return oauth002FormResult{err: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return oauth002FormResult{err: err}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return oauth002FormResult{status: resp.StatusCode, body: string(body), err: err}
}

func oauth002Challenge(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func oauth002UserTokenCount(t *testing.T, db database.DB, userID string) int {
	t.Helper()
	tokens, err := db.FindUserAuthTokens(t.Context(), userID, "", 100, nil)
	require.NoError(t, err)
	return len(tokens)
}

func oauth002ClientTokens(t *testing.T, db database.DB, userID, clientID string) []*database.UserAuthToken {
	t.Helper()
	tokens, err := db.FindUserAuthTokens(t.Context(), userID, "", 100, nil)
	require.NoError(t, err)
	res := make([]*database.UserAuthToken, 0, len(tokens))
	for _, token := range tokens {
		if token.AuthClientID != nil && *token.AuthClientID == clientID {
			res = append(res, token)
		}
	}
	return res
}

func oauth002WithDB(t *testing.T, fix *testadmin.Fixture, db database.DB, fn func()) {
	t.Helper()
	original := fix.Admin.DB
	fix.Admin.DB = db
	defer func() { fix.Admin.DB = original }()
	fn()
}

func oauth002ID(label string) string {
	return fmt.Sprintf("oauth002-%s-%d", label, oauth002Sequence.Add(1))
}
