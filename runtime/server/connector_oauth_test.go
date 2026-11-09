package server

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConnectorOAuthCaller(t *testing.T) {
	tests := []struct {
		name      string
		claims    *runtime.SecurityClaims
		adminTok  string
		adminErr  error
		wantTok   *drivers.OAuthToken
		wantErr   error
		wantAuthz bool
	}{
		{
			name:     "interactive user",
			claims:   &runtime.SecurityClaims{UserID: "u", InteractiveUser: true},
			adminTok: "t",
			wantTok:  &drivers.OAuthToken{UserID: "u", AccessToken: "t", ExpiresOn: fakeOAuthExpiresOn},
		},
		{
			name:      "interactive user without authorization",
			claims:    &runtime.SecurityClaims{UserID: "u", InteractiveUser: true},
			adminErr:  drivers.ErrNotFound,
			wantAuthz: true,
		},
		{
			name:     "view as",
			claims:   &runtime.SecurityClaims{UserID: "mocked", InteractiveUser: false},
			adminTok: "t",
			wantErr:  drivers.ErrOAuthUserRequired,
		},
		{
			name:     "service",
			claims:   &runtime.SecurityClaims{UserID: "svc", SkipChecks: true},
			adminTok: "t",
			wantErr:  drivers.ErrOAuthUserRequired,
		},
		{
			name:     "rill developer",
			claims:   &runtime.SecurityClaims{},
			adminErr: drivers.ErrNotImplemented,
			wantErr:  drivers.ErrNotImplemented,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			admin := &fakeOAuthAdmin{tok: tt.adminTok, err: tt.adminErr}
			c := newFakeOAuthCaller(admin)
			ctx := auth.WithClaims(context.Background(), tt.claims)

			tok, err := c.Token(ctx, "inst", "dbx", "hash", false)
			switch {
			case tt.wantAuthz:
				var authzErr *drivers.OAuthAuthorizationRequiredError
				require.ErrorAs(t, err, &authzErr)
				require.Equal(t, "dbx", authzErr.Connector)
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			default:
				require.NoError(t, err)
				require.Equal(t, tt.wantTok, tok)
			}

			if tt.claims.UserID != "" && !tt.claims.InteractiveUser {
				require.NotContains(t, admin.userIDs, tt.claims.UserID)
			}
		})
	}
}

func TestMapGRPCErrorOAuth(t *testing.T) {
	err := mapGRPCError(fmt.Errorf("query failed: %w", &drivers.OAuthAuthorizationRequiredError{Connector: "dbx"}))
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
	require.Equal(t, `sign in to connector "dbx" to view this data`, st.Message())

	err = mapGRPCError(fmt.Errorf("query failed: %w", drivers.ErrOAuthUserRequired))
	st, ok = status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
	require.Equal(t, drivers.ErrOAuthUserRequired.Error(), st.Message())
}

func TestConnectorOAuthTokenCache(t *testing.T) {
	ctxFor := func(user string) context.Context {
		return auth.WithClaims(context.Background(), &runtime.SecurityClaims{UserID: user, InteractiveUser: true})
	}

	t.Run("concurrent calls collapse", func(t *testing.T) {
		admin := &fakeOAuthAdmin{tok: "t", delay: 50 * time.Millisecond}
		c := newFakeOAuthCaller(admin)
		var wg sync.WaitGroup
		for range 20 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				tok, err := c.Token(ctxFor("u"), "inst", "dbx", "hash", false)
				require.NoError(t, err)
				require.Equal(t, "t", tok.AccessToken)
			}()
		}
		wg.Wait()
		require.Equal(t, 1, admin.callCount())

		_, err := c.Token(ctxFor("u"), "inst", "dbx", "hash", false)
		require.NoError(t, err)
		require.Equal(t, 1, admin.callCount())
	})

	t.Run("short-lived token is not cached", func(t *testing.T) {
		admin := &fakeOAuthAdmin{tok: "t", expiresOn: time.Now().Add(30 * time.Second)}
		c := newFakeOAuthCaller(admin)
		for range 2 {
			_, err := c.Token(ctxFor("u"), "inst", "dbx", "hash", false)
			require.NoError(t, err)
		}
		require.Equal(t, 2, admin.callCount())
	})

	t.Run("force refresh", func(t *testing.T) {
		admin := &fakeOAuthAdmin{tok: "old"}
		c := newFakeOAuthCaller(admin)
		_, err := c.Token(ctxFor("u"), "inst", "dbx", "hash", false)
		require.NoError(t, err)

		admin.tok = "new"
		tok, err := c.Token(ctxFor("u"), "inst", "dbx", "hash", true)
		require.NoError(t, err)
		require.Equal(t, "new", tok.AccessToken)
		require.Equal(t, []bool{false, true}, admin.forceRefreshes)

		tok, err = c.Token(ctxFor("u"), "inst", "dbx", "hash", false)
		require.NoError(t, err)
		require.Equal(t, "new", tok.AccessToken)
		require.Equal(t, 2, admin.callCount())
	})

	t.Run("force refresh doesn't join a fetch in flight", func(t *testing.T) {
		admin := &fakeOAuthAdmin{tok: "t", delay: 100 * time.Millisecond}
		c := newFakeOAuthCaller(admin)
		done := make(chan error)
		go func() {
			_, err := c.Token(ctxFor("u"), "inst", "dbx", "hash", false)
			done <- err
		}()
		require.Eventually(t, func() bool { return admin.callCount() == 1 }, time.Second, time.Millisecond)

		_, err := c.Token(ctxFor("u"), "inst", "dbx", "hash", true)
		require.NoError(t, err)
		require.NoError(t, <-done)
		require.Equal(t, []bool{false, true}, admin.forceRefreshes)
	})

	t.Run("users are separate", func(t *testing.T) {
		admin := &fakeOAuthAdmin{tok: "t"}
		c := newFakeOAuthCaller(admin)
		for _, u := range []string{"a", "b", "a", "b"} {
			tok, err := c.Token(ctxFor(u), "inst", "dbx", "hash", false)
			require.NoError(t, err)
			require.Equal(t, u, tok.UserID)
		}
		require.Equal(t, []string{"a", "b"}, admin.userIDs)
	})

	t.Run("errors are not cached", func(t *testing.T) {
		admin := &fakeOAuthAdmin{err: drivers.ErrNotFound}
		c := newFakeOAuthCaller(admin)
		for range 2 {
			var authzErr *drivers.OAuthAuthorizationRequiredError
			_, err := c.Token(ctxFor("u"), "inst", "dbx", "hash", false)
			require.ErrorAs(t, err, &authzErr)
		}
		require.Equal(t, 2, admin.callCount())
	})

	t.Run("non-interactive probe is cached per instance", func(t *testing.T) {
		admin := &fakeOAuthAdmin{err: drivers.ErrNotImplemented}
		c := newFakeOAuthCaller(admin)
		ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{})
		for range 2 {
			_, err := c.Token(ctx, "inst", "dbx", "hash", false)
			require.ErrorIs(t, err, drivers.ErrNotImplemented)
		}
		require.Equal(t, 1, admin.callCount())
	})

	t.Run("transient probe error is not cached", func(t *testing.T) {
		admin := &fakeOAuthAdmin{err: errors.New("unavailable")}
		c := newFakeOAuthCaller(admin)
		ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{})
		for range 2 {
			_, err := c.Token(ctx, "inst", "dbx", "hash", false)
			require.ErrorIs(t, err, drivers.ErrOAuthUserRequired)
		}
		require.Equal(t, 2, admin.callCount())
	})
}

func newFakeOAuthCaller(admin *fakeOAuthAdmin) *connectorOAuthCaller {
	return newConnectorOAuthCaller(func(ctx context.Context, instanceID string) (drivers.AdminService, func(), error) {
		return admin, func() {}, nil
	})
}

var fakeOAuthExpiresOn = time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

// fakeOAuthAdmin is a drivers.AdminService that serves a fixed connector OAuth token and records the requests it received.
type fakeOAuthAdmin struct {
	drivers.AdminService
	tok       string
	err       error
	expiresOn time.Time
	delay     time.Duration

	mu             sync.Mutex
	userIDs        []string
	forceRefreshes []bool
}

func (a *fakeOAuthAdmin) callCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.forceRefreshes)
}

func (a *fakeOAuthAdmin) GetConnectorOAuthToken(ctx context.Context, connector, configHash, userID string, forceRefresh bool) (*drivers.MCPOAuthToken, error) {
	a.mu.Lock()
	if !slices.Contains(a.userIDs, userID) {
		a.userIDs = append(a.userIDs, userID)
	}
	a.forceRefreshes = append(a.forceRefreshes, forceRefresh)
	a.mu.Unlock()
	time.Sleep(a.delay)
	if a.err != nil {
		return nil, a.err
	}
	if connector != "dbx" || configHash != "hash" {
		return nil, errors.New("unexpected connector")
	}
	exp := a.expiresOn
	if exp.IsZero() {
		exp = fakeOAuthExpiresOn
	}
	return &drivers.MCPOAuthToken{AccessToken: a.tok, ExpiresOn: exp}, nil
}
