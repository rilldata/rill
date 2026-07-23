package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/authtoken"
	"github.com/rilldata/rill/admin/pkg/oauth"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/stretchr/testify/require"
)

const oauth003RefreshTokenGrant = "refresh_token"

var (
	oauth003Sequence              atomic.Uint64
	oauth003ErrAccessPersistence  = errors.New("injected access token persistence failure")
	oauth003ErrRefreshPersistence = errors.New("injected refresh token persistence failure")
	oauth003ErrRevokePersistence  = errors.New("injected old token revoke failure")
	oauth003ErrResponseWrite      = errors.New("injected response write failure")
)

type oauth003FormResult struct {
	status int
	body   string
	err    error
}

type oauth003FailureDB struct {
	database.DB
	failInsertRefresh *bool
	failDeleteID      string
}

func (d *oauth003FailureDB) InsertUserAuthToken(ctx context.Context, opts *database.InsertUserAuthTokenOptions) (*database.UserAuthToken, error) {
	if d.failInsertRefresh != nil && opts.Refresh == *d.failInsertRefresh {
		if opts.Refresh {
			return nil, oauth003ErrRefreshPersistence
		}
		return nil, oauth003ErrAccessPersistence
	}
	return d.DB.InsertUserAuthToken(ctx, opts)
}

func (d *oauth003FailureDB) DeleteUserAuthToken(ctx context.Context, id string) error {
	if id == d.failDeleteID {
		return oauth003ErrRevokePersistence
	}
	return d.DB.DeleteUserAuthToken(ctx, id)
}

type oauth003FailingResponseWriter struct {
	header http.Header
	writes int
}

func (w *oauth003FailingResponseWriter) Header() http.Header {
	return w.header
}

func (w *oauth003FailingResponseWriter) WriteHeader(int) {}

func (w *oauth003FailingResponseWriter) Write([]byte) (int, error) {
	w.writes++
	return 0, oauth003ErrResponseWrite
}

func TestOAUTH003RefreshTokenRotation(t *testing.T) {
	// Setup: share one real Postgres-backed admin fixture across isolated users and clients.
	fix := testadmin.New(t)
	tokenURL := fix.ExternalURL() + "/auth/oauth/token"
	// Action: each subtest exercises one rotation, replay, concurrency, or failure path.
	// Assertion: every path checks both the HTTP result and the durable token rows.

	t.Run("valid rotation is durable and single use", func(t *testing.T) {
		// Setup: issue one refresh token for a client that permits refresh grants.
		user, _ := fix.NewUser(t)
		client := oauth003NewClient(t, fix.Admin.DB)
		oldToken := oauth003IssueRefreshToken(t, fix, user.ID, client.ID, time.Hour)

		// Action: exchange the refresh token once through the public token endpoint.
		res := oauth003PostRefresh(t.Context(), tokenURL, client.ID, oldToken)

		// Assertion: the old token is gone and exactly one access/refresh pair is usable.
		require.NoError(t, res.err)
		require.Equal(t, http.StatusOK, res.status, res.body)
		payload := oauth003DecodeResponse(t, res.body)
		require.NotEmpty(t, payload.AccessToken)
		require.NotEmpty(t, payload.RefreshToken)
		require.NotEqual(t, oldToken, payload.RefreshToken)
		_, err := fix.Admin.ValidateAuthToken(t.Context(), oldToken)
		require.Error(t, err, "the committed rotation must invalidate the old token immediately")
		_, err = fix.Admin.ValidateAuthToken(t.Context(), payload.AccessToken)
		require.NoError(t, err)
		_, err = fix.Admin.ValidateAuthToken(t.Context(), payload.RefreshToken)
		require.NoError(t, err)
		oauth003RequireTokenShape(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 1, 1)
	})

	t.Run("invalid tokens mint nothing", func(t *testing.T) {
		// Setup: enumerate malformed, expired, and cached-then-revoked token states.
		tests := []struct {
			name  string
			setup func(*testing.T, *database.User, *database.AuthClient) string
		}{
			{
				name: "malformed",
				setup: func(*testing.T, *database.User, *database.AuthClient) string {
					return "not-an-auth-token"
				},
			},
			{
				name: "expired",
				setup: func(t *testing.T, user *database.User, client *database.AuthClient) string {
					return oauth003IssueRefreshToken(t, fix, user.ID, client.ID, -time.Hour)
				},
			},
			{
				name: "revoked",
				setup: func(t *testing.T, user *database.User, client *database.AuthClient) string {
					token := oauth003IssueRefreshToken(t, fix, user.ID, client.ID, time.Hour)
					_, err := fix.Admin.ValidateAuthToken(t.Context(), token)
					require.NoError(t, err, "prime the validation cache before revocation")
					require.NoError(t, fix.Admin.RevokeAuthToken(t.Context(), token))
					return token
				},
			},
		}

		// Action: submit each unusable state through the same public refresh endpoint.
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Setup: prepare a malformed, expired, or already-revoked refresh token.
				user, _ := fix.NewUser(t)
				client := oauth003NewClient(t, fix.Admin.DB)
				token := tt.setup(t, user, client)

				// Action: submit the unusable token to the refresh grant endpoint.
				res := oauth003PostRefresh(t.Context(), tokenURL, client.ID, token)

				// Assertion: authentication fails and no replacement pair is persisted.
				require.NoError(t, res.err)
				require.Equal(t, http.StatusUnauthorized, res.status, res.body)
				require.Contains(t, res.body, "invalid refresh token")
				require.Empty(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID))
			})
		}
		// Assertion: the per-case checks prove no invalid state persisted a replacement.
	})

	t.Run("simultaneous reuse has exactly one winner", func(t *testing.T) {
		// Setup: synchronize two requests that carry the same valid refresh token.
		user, _ := fix.NewUser(t)
		client := oauth003NewClient(t, fix.Admin.DB)
		oldToken := oauth003IssueRefreshToken(t, fix, user.ID, client.ID, time.Hour)
		start := make(chan struct{})
		results := make(chan oauth003FormResult, 2)
		var group sync.WaitGroup
		for range 2 {
			group.Add(1)
			go func() {
				defer group.Done()
				<-start
				results <- oauth003PostRefresh(context.Background(), tokenURL, client.ID, oldToken)
			}()
		}

		// Action: release both requests together without timing sleeps.
		close(start)
		group.Wait()
		close(results)

		// Assertion: one claim commits, one loses, and only one replacement pair remains.
		successes := 0
		failures := 0
		for res := range results {
			require.NoError(t, res.err)
			switch res.status {
			case http.StatusOK:
				successes++
			case http.StatusUnauthorized:
				failures++
				require.Contains(t, res.body, "invalid refresh token")
			default:
				require.Failf(t, "unexpected rotation response", "status=%d body=%s", res.status, res.body)
			}
		}
		require.Equal(t, 1, successes)
		require.Equal(t, 1, failures)
		oauth003RequireTokenShape(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 1, 1)
		_, err := fix.Admin.ValidateAuthToken(t.Context(), oldToken)
		require.Error(t, err)
	})

	t.Run("ancestor replay cannot fork the active family", func(t *testing.T) {
		// Setup: rotate an initial token to establish one active descendant.
		user, _ := fix.NewUser(t)
		client := oauth003NewClient(t, fix.Admin.DB)
		ancestor := oauth003IssueRefreshToken(t, fix, user.ID, client.ID, time.Hour)
		first := oauth003PostRefresh(t.Context(), tokenURL, client.ID, ancestor)
		require.NoError(t, first.err)
		require.Equal(t, http.StatusOK, first.status, first.body)
		descendant := oauth003DecodeResponse(t, first.body).RefreshToken

		// Action: replay the consumed ancestor, then rotate the legitimate descendant.
		replay := oauth003PostRefresh(t.Context(), tokenURL, client.ID, ancestor)
		second := oauth003PostRefresh(t.Context(), tokenURL, client.ID, descendant)

		// Assertion: replay mints no branch while the sole descendant advances normally.
		require.NoError(t, replay.err)
		require.Equal(t, http.StatusUnauthorized, replay.status, replay.body)
		require.NoError(t, second.err)
		require.Equal(t, http.StatusOK, second.status, second.body)
		latest := oauth003DecodeResponse(t, second.body).RefreshToken
		require.NotEqual(t, descendant, latest)
		oauth003RequireTokenShape(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 2, 1)
		descendantReplay := oauth003PostRefresh(t.Context(), tokenURL, client.ID, descendant)
		require.NoError(t, descendantReplay.err)
		require.Equal(t, http.StatusUnauthorized, descendantReplay.status, descendantReplay.body)
		oauth003RequireTokenShape(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 2, 1)
	})

	t.Run("replacement persistence failures roll back consumption", func(t *testing.T) {
		// Setup: inject failures separately for the new access and refresh token rows.
		failAccess := false
		failRefresh := true
		tests := []struct {
			name        string
			failRefresh *bool
			wantError   error
		}{
			{name: "access insert", failRefresh: &failAccess, wantError: oauth003ErrAccessPersistence},
			{name: "refresh insert", failRefresh: &failRefresh, wantError: oauth003ErrRefreshPersistence},
		}

		// Action: run each insert failure after the old token has been claimed in a transaction.
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Setup: keep one valid old token and inject a replacement-row insert failure.
				user, _ := fix.NewUser(t)
				client := oauth003NewClient(t, fix.Admin.DB)
				oldToken := oauth003IssueRefreshToken(t, fix, user.ID, client.ID, time.Hour)
				failureDB := &oauth003FailureDB{DB: fix.Admin.DB, failInsertRefresh: tt.failRefresh}

				// Action: attempt rotation while the selected replacement insert fails.
				var res oauth003FormResult
				oauth003WithDB(t, fix, failureDB, func() {
					res = oauth003PostRefresh(t.Context(), tokenURL, client.ID, oldToken)
				})

				// Assertion: the transaction leaves only the old token, which remains retryable.
				require.NoError(t, res.err)
				require.Equal(t, http.StatusInternalServerError, res.status, res.body)
				require.Contains(t, res.body, tt.wantError.Error())
				oauth003RequireTokenShape(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 0, 1)
				retry := oauth003PostRefresh(t.Context(), tokenURL, client.ID, oldToken)
				require.NoError(t, retry.err)
				require.Equal(t, http.StatusOK, retry.status, retry.body)
				oauth003RequireTokenShape(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 1, 1)
			})
		}
		// Assertion: every case retries the old token, proving rollback restored one family.
	})

	t.Run("old token revoke failure creates no replacement branch", func(t *testing.T) {
		// Setup: inject a database failure for the old token's single-use delete.
		user, _ := fix.NewUser(t)
		client := oauth003NewClient(t, fix.Admin.DB)
		oldToken := oauth003IssueRefreshToken(t, fix, user.ID, client.ID, time.Hour)
		failureDB := &oauth003FailureDB{DB: fix.Admin.DB, failDeleteID: oauth003TokenID(t, oldToken)}

		// Action: rotate while persistence refuses to revoke the old token.
		var res oauth003FormResult
		oauth003WithDB(t, fix, failureDB, func() {
			res = oauth003PostRefresh(t.Context(), tokenURL, client.ID, oldToken)
		})

		// Assertion: no replacements survive, and the unchanged old token can be retried.
		require.NoError(t, res.err)
		require.Equal(t, http.StatusInternalServerError, res.status, res.body)
		require.Contains(t, res.body, oauth003ErrRevokePersistence.Error())
		oauth003RequireTokenShape(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 0, 1)
		retry := oauth003PostRefresh(t.Context(), tokenURL, client.ID, oldToken)
		require.NoError(t, retry.err)
		require.Equal(t, http.StatusOK, retry.status, retry.body)
		oauth003RequireTokenShape(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 1, 1)
	})

	t.Run("response write failure cannot make a second family", func(t *testing.T) {
		// Setup: invoke the real route with a writer that rejects every response body.
		user, _ := fix.NewUser(t)
		client := oauth003NewClient(t, fix.Admin.DB)
		oldToken := oauth003IssueRefreshToken(t, fix, user.ID, client.ID, time.Hour)
		handler, err := fix.Server.HTTPHandler(t.Context())
		require.NoError(t, err)
		form := oauth003RefreshForm(client.ID, oldToken)
		req := httptest.NewRequest(http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		writer := &oauth003FailingResponseWriter{header: make(http.Header)}

		// Action: complete rotation while delivery of its committed response fails.
		handler.ServeHTTP(writer, req)

		// Assertion: one committed pair exists and replay cannot recover or fork it.
		require.Positive(t, writer.writes)
		oauth003RequireTokenShape(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 1, 1)
		_, err = fix.Admin.ValidateAuthToken(t.Context(), oldToken)
		require.Error(t, err)
		replay := oauth003PostRefresh(t.Context(), tokenURL, client.ID, oldToken)
		require.NoError(t, replay.err)
		require.Equal(t, http.StatusUnauthorized, replay.status, replay.body)
		oauth003RequireTokenShape(t, oauth003ClientTokens(t, fix.Admin.DB, user.ID, client.ID), 1, 1)
	})
}

func oauth003NewClient(t *testing.T, db database.DB) *database.AuthClient {
	t.Helper()
	client, err := db.InsertAuthClient(t.Context(), oauth003ID("client"), "offline_access", []string{oauth003RefreshTokenGrant}, []string{"https://client.example/callback"})
	require.NoError(t, err)
	return client
}

func oauth003IssueRefreshToken(t *testing.T, fix *testadmin.Fixture, userID, clientID string, ttl time.Duration) string {
	t.Helper()
	token, err := fix.Admin.IssueUserAuthToken(t.Context(), userID, clientID, "Refresh Token", nil, &ttl, true)
	require.NoError(t, err)
	return token.Token().String()
}

func oauth003PostRefresh(ctx context.Context, endpoint, clientID, refreshToken string) oauth003FormResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(oauth003RefreshForm(clientID, refreshToken).Encode()))
	if err != nil {
		return oauth003FormResult{err: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return oauth003FormResult{err: err}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return oauth003FormResult{status: resp.StatusCode, body: string(body), err: err}
}

func oauth003RefreshForm(clientID, refreshToken string) url.Values {
	return url.Values{
		"grant_type":    {oauth003RefreshTokenGrant},
		"client_id":     {clientID},
		"refresh_token": {refreshToken},
	}
}

func oauth003DecodeResponse(t *testing.T, body string) oauth.TokenResponse {
	t.Helper()
	var payload oauth.TokenResponse
	require.NoError(t, json.Unmarshal([]byte(body), &payload))
	return payload
}

func oauth003ClientTokens(t *testing.T, db database.DB, userID, clientID string) []*database.UserAuthToken {
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

func oauth003RequireTokenShape(t *testing.T, tokens []*database.UserAuthToken, wantAccess, wantRefresh int) {
	t.Helper()
	access := 0
	refresh := 0
	for _, token := range tokens {
		if token.Refresh {
			refresh++
		} else {
			access++
		}
	}
	require.Equal(t, wantAccess, access, "unexpected durable access-token count")
	require.Equal(t, wantRefresh, refresh, "unexpected durable refresh-token count")
	require.Len(t, tokens, wantAccess+wantRefresh)
}

func oauth003TokenID(t *testing.T, token string) string {
	t.Helper()
	parsed, err := authtoken.FromString(token)
	require.NoError(t, err)
	return parsed.ID.String()
}

func oauth003WithDB(t *testing.T, fix *testadmin.Fixture, db database.DB, fn func()) {
	t.Helper()
	original := fix.Admin.DB
	fix.Admin.DB = db
	defer func() { fix.Admin.DB = original }()
	fn()
}

func oauth003ID(label string) string {
	return fmt.Sprintf("oauth003-%s-%d", label, oauth003Sequence.Add(1))
}
