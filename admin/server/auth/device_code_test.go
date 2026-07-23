package auth_test

import (
	"context"
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

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	serverauth "github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/admin/testadmin"
	"github.com/stretchr/testify/require"
)

const oauth004DeviceGrantType = "urn:ietf:params:oauth:grant-type:device_code"

var (
	oauth004Sequence      atomic.Uint64
	errInjectedTokenIssue = errors.New("injected token issuance failure")
	errInjectedCodeDelete = errors.New("injected device code deletion failure")
)

type oauth004FormResult struct {
	status int
	body   string
	err    error
}

type oauth004CollisionDB struct {
	database.DB
	calls atomic.Int32
}

func (d *oauth004CollisionDB) InsertDeviceAuthCode(ctx context.Context, deviceCode, userCode, clientID string, expiresOn time.Time) (*database.DeviceAuthCode, error) {
	if d.calls.Add(1) == 1 {
		return nil, database.NewNotUniqueError("injected user code collision")
	}
	return d.DB.InsertDeviceAuthCode(ctx, deviceCode, userCode, clientID, expiresOn)
}

type oauth004IssueFailureDB struct {
	database.DB
}

func (d *oauth004IssueFailureDB) InsertUserAuthToken(context.Context, *database.InsertUserAuthTokenOptions) (*database.UserAuthToken, error) {
	return nil, errInjectedTokenIssue
}

type oauth004DeleteFailureDB struct {
	database.DB
}

func (d *oauth004DeleteFailureDB) DeleteDeviceAuthCode(context.Context, string) error {
	return errInjectedCodeDelete
}

func TestOAUTH004DeviceCodeFlow(t *testing.T) {
	// Exercise the full real-Postgres device-code lifecycle, including collision, concurrency, and rollback boundaries.
	fix := testadmin.New(t)
	deviceAuthorizationURL := fix.ExternalURL() + "/auth/oauth/device_authorization"
	tokenURL := fix.ExternalURL() + "/auth/oauth/token"

	t.Run("requires the one supported scope and advertises polling", func(t *testing.T) {
		res := oauth004PostForm(t.Context(), deviceAuthorizationURL, url.Values{
			"client_id": {database.AuthClientIDRillCLI},
		})
		require.NoError(t, res.err)
		require.Equal(t, http.StatusBadRequest, res.status)
		require.Equal(t, "scope is required\n", res.body)

		for _, scope := range []string{"openid", "full_account openid"} {
			res = oauth004PostForm(t.Context(), deviceAuthorizationURL, url.Values{
				"client_id": {database.AuthClientIDRillCLI},
				"scope":     {scope},
			})
			require.NoError(t, res.err)
			require.Equal(t, http.StatusBadRequest, res.status)
			require.Equal(t, "invalid scope\n", res.body)
		}

		res = oauth004PostForm(t.Context(), deviceAuthorizationURL, url.Values{
			"client_id": {database.AuthClientIDRillCLI},
			"scope":     {"  full_account  "},
		})
		require.NoError(t, res.err)
		require.Equal(t, http.StatusOK, res.status, res.body)

		var payload serverauth.DeviceCodeResponse
		require.NoError(t, json.Unmarshal([]byte(res.body), &payload))
		require.NotEmpty(t, payload.DeviceCode)
		require.Len(t, strings.ReplaceAll(payload.UserCode, "-", ""), 8)
		require.Equal(t, int(admin.DeviceAuthCodeTTL.Seconds()), payload.ExpiresIn)
		// The advertised interval is the current polling contract. A slow_down
		// response is not supported because the server stores no poll history.
		require.Equal(t, 5, payload.PollingInterval)

		stored, err := fix.Admin.DB.FindDeviceAuthCodeByDeviceCode(t.Context(), payload.DeviceCode)
		require.NoError(t, err)
		require.Equal(t, database.DeviceAuthCodeStatePending, stored.ApprovalState)
	})

	t.Run("keeps pending user codes unique and retries a collision", func(t *testing.T) {
		// A confirmation code must select exactly one pending device request. The
		// real database constraint makes a duplicate deterministic for this test.
		expiresOn := time.Now().Add(time.Minute)
		_, err := fix.Admin.DB.InsertDeviceAuthCode(t.Context(), oauth004ID("collision-a"), "COLLIDE1", database.AuthClientIDRillCLI, expiresOn)
		require.NoError(t, err)
		_, err = fix.Admin.DB.InsertDeviceAuthCode(t.Context(), oauth004ID("collision-b"), "COLLIDE1", database.AuthClientIDRillCLI, expiresOn)
		require.ErrorIs(t, err, database.ErrNotUnique)

		originalDB := fix.Admin.DB
		collisionDB := &oauth004CollisionDB{DB: originalDB}
		oauth004WithDB(t, fix, collisionDB, func() {
			res := oauth004PostForm(t.Context(), deviceAuthorizationURL, url.Values{
				"client_id": {database.AuthClientIDRillCLI},
				"scope":     {"full_account"},
			})
			require.NoError(t, res.err)
			require.Equal(t, http.StatusOK, res.status, res.body)
		})
		require.EqualValues(t, 2, collisionDB.calls.Load())
	})

	t.Run("reports pending without consuming the code", func(t *testing.T) {
		code := oauth004StoreCode(t, fix.Admin.DB, database.DeviceAuthCodeStatePending, nil, time.Now().Add(time.Minute))
		res := oauth004Poll(t.Context(), tokenURL, code.DeviceCode)
		require.NoError(t, res.err)
		require.Equal(t, http.StatusUnauthorized, res.status)
		require.Equal(t, "authorization_pending\n", res.body)
		_, err := fix.Admin.DB.FindDeviceAuthCodeByDeviceCode(t.Context(), code.DeviceCode)
		require.NoError(t, err)
	})

	t.Run("rejects and consumes a denied code", func(t *testing.T) {
		user, _ := fix.NewUser(t)
		code := oauth004StoreCode(t, fix.Admin.DB, database.DeviceAuthCodeStateRejected, &user.ID, time.Now().Add(time.Minute))
		res := oauth004Poll(t.Context(), tokenURL, code.DeviceCode)
		require.NoError(t, res.err)
		require.Equal(t, http.StatusUnauthorized, res.status)
		require.Equal(t, "rejected\n", res.body)
		_, err := fix.Admin.DB.FindDeviceAuthCodeByDeviceCode(t.Context(), code.DeviceCode)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	t.Run("never issues from an expired approved code", func(t *testing.T) {
		user, _ := fix.NewUser(t)
		before := oauth004TokenCount(t, fix.Admin.DB, user.ID)
		code := oauth004StoreCode(t, fix.Admin.DB, database.DeviceAuthCodeStateApproved, &user.ID, time.Now().Add(-time.Minute))

		// Expiry wins over approval: a stale approval must not mint a token.
		res := oauth004Poll(t.Context(), tokenURL, code.DeviceCode)
		require.NoError(t, res.err)
		require.Equal(t, http.StatusUnauthorized, res.status)
		require.Equal(t, "expired_token\n", res.body)
		require.Equal(t, before, oauth004TokenCount(t, fix.Admin.DB, user.ID))
		_, err := fix.Admin.DB.FindDeviceAuthCodeByDeviceCode(t.Context(), code.DeviceCode)
		require.NoError(t, err)
	})

	t.Run("fails closed on a malformed persisted state", func(t *testing.T) {
		user, _ := fix.NewUser(t)
		before := oauth004TokenCount(t, fix.Admin.DB, user.ID)
		code := oauth004StoreCode(t, fix.Admin.DB, database.DeviceAuthCodeState(99), &user.ID, time.Now().Add(time.Minute))

		res := oauth004Poll(t.Context(), tokenURL, code.DeviceCode)
		require.NoError(t, res.err)
		require.Equal(t, http.StatusInternalServerError, res.status)
		require.Contains(t, res.body, "inconsistent state")
		require.Equal(t, before, oauth004TokenCount(t, fix.Admin.DB, user.ID))
		_, err := fix.Admin.DB.FindDeviceAuthCodeByDeviceCode(t.Context(), code.DeviceCode)
		require.NoError(t, err)
	})

	t.Run("simultaneous approved polls persist one token", func(t *testing.T) {
		user, _ := fix.NewUser(t)
		before := oauth004TokenCount(t, fix.Admin.DB, user.ID)
		code := oauth004StoreCode(t, fix.Admin.DB, database.DeviceAuthCodeStateApproved, &user.ID, time.Now().Add(time.Minute))

		const polls = 8
		start := make(chan struct{})
		results := make(chan oauth004FormResult, polls)
		var group sync.WaitGroup
		for range polls {
			group.Add(1)
			go func() {
				defer group.Done()
				<-start
				results <- oauth004Poll(context.Background(), tokenURL, code.DeviceCode)
			}()
		}
		close(start)
		group.Wait()
		close(results)

		successes := 0
		for res := range results {
			require.NoError(t, res.err)
			if res.status == http.StatusOK {
				successes++
			}
		}
		require.Equal(t, 1, successes)
		require.Equal(t, before+1, oauth004TokenCount(t, fix.Admin.DB, user.ID))
		_, err := fix.Admin.DB.FindDeviceAuthCodeByDeviceCode(t.Context(), code.DeviceCode)
		require.ErrorIs(t, err, database.ErrNotFound)
	})

	t.Run("issuance failure leaves the approved code retryable", func(t *testing.T) {
		user, _ := fix.NewUser(t)
		before := oauth004TokenCount(t, fix.Admin.DB, user.ID)
		code := oauth004StoreCode(t, fix.Admin.DB, database.DeviceAuthCodeStateApproved, &user.ID, time.Now().Add(time.Minute))

		// Issuance and deletion are one transaction: either both persist or neither does.
		oauth004WithDB(t, fix, &oauth004IssueFailureDB{DB: fix.Admin.DB}, func() {
			res := oauth004Poll(t.Context(), tokenURL, code.DeviceCode)
			require.NoError(t, res.err)
			require.Equal(t, http.StatusInternalServerError, res.status)
			require.Contains(t, res.body, errInjectedTokenIssue.Error())
		})
		require.Equal(t, before, oauth004TokenCount(t, fix.Admin.DB, user.ID))
		_, err := fix.Admin.DB.FindDeviceAuthCodeByDeviceCode(t.Context(), code.DeviceCode)
		require.NoError(t, err)

		res := oauth004Poll(t.Context(), tokenURL, code.DeviceCode)
		require.NoError(t, res.err)
		require.Equal(t, http.StatusOK, res.status, res.body)
		require.Equal(t, before+1, oauth004TokenCount(t, fix.Admin.DB, user.ID))
	})

	t.Run("deletion failure rolls token issuance back", func(t *testing.T) {
		user, _ := fix.NewUser(t)
		before := oauth004TokenCount(t, fix.Admin.DB, user.ID)
		code := oauth004StoreCode(t, fix.Admin.DB, database.DeviceAuthCodeStateApproved, &user.ID, time.Now().Add(time.Minute))

		oauth004WithDB(t, fix, &oauth004DeleteFailureDB{DB: fix.Admin.DB}, func() {
			res := oauth004Poll(t.Context(), tokenURL, code.DeviceCode)
			require.NoError(t, res.err)
			require.Equal(t, http.StatusInternalServerError, res.status)
			require.Contains(t, res.body, errInjectedCodeDelete.Error())
		})
		require.Equal(t, before, oauth004TokenCount(t, fix.Admin.DB, user.ID))
		_, err := fix.Admin.DB.FindDeviceAuthCodeByDeviceCode(t.Context(), code.DeviceCode)
		require.NoError(t, err)

		res := oauth004Poll(t.Context(), tokenURL, code.DeviceCode)
		require.NoError(t, res.err)
		require.Equal(t, http.StatusOK, res.status, res.body)
		require.Equal(t, before+1, oauth004TokenCount(t, fix.Admin.DB, user.ID))
	})
}

func oauth004PostForm(ctx context.Context, endpoint string, values url.Values) oauth004FormResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return oauth004FormResult{err: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return oauth004FormResult{err: err}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return oauth004FormResult{status: resp.StatusCode, body: string(body), err: err}
}

func oauth004Poll(ctx context.Context, endpoint, deviceCode string) oauth004FormResult {
	return oauth004PostForm(ctx, endpoint, url.Values{
		"client_id":              {database.AuthClientIDRillCLI},
		"device_code":            {deviceCode},
		"grant_type":             {oauth004DeviceGrantType},
		"token_response_version": {"standard"},
	})
}

func oauth004StoreCode(t *testing.T, db database.DB, state database.DeviceAuthCodeState, userID *string, expiry time.Time) *database.DeviceAuthCode {
	t.Helper()
	sequence := oauth004Sequence.Add(1)
	code, err := db.InsertDeviceAuthCode(
		t.Context(),
		fmt.Sprintf("oauth004-device-%d", sequence),
		fmt.Sprintf("T%07d", sequence),
		database.AuthClientIDRillCLI,
		expiry,
	)
	require.NoError(t, err)
	if state != database.DeviceAuthCodeStatePending {
		require.NotNil(t, userID)
		require.NoError(t, db.UpdateDeviceAuthCode(t.Context(), code.ID, *userID, state))
	}
	return code
}

func oauth004TokenCount(t *testing.T, db database.DB, userID string) int {
	t.Helper()
	tokens, err := db.FindUserAuthTokens(t.Context(), userID, "", 100, nil)
	require.NoError(t, err)
	return len(tokens)
}

func oauth004WithDB(t *testing.T, fix *testadmin.Fixture, db database.DB, fn func()) {
	t.Helper()
	original := fix.Admin.DB
	fix.Admin.DB = db
	defer func() { fix.Admin.DB = original }()
	fn()
}

func oauth004ID(label string) string {
	return fmt.Sprintf("oauth004-%s-%d", label, oauth004Sequence.Add(1))
}
