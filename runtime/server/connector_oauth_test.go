package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sync"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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

func TestDownloadRunsAsTokenIssuer(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"connectors/dbx.yaml": `
type: connector
driver: oauth_test_duckdb
path: ":memory:"
mode: readwrite
init_sql: ATTACH ':memory:' AS src; CREATE TABLE src.t AS SELECT 'US' AS country
database_name: src
`,
			"mv1.yaml": `
type: metrics_view
connector: dbx
table: t
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
explore:
  skip: true
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	s, err := NewServer(t.Context(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)
	admin := &fakeOAuthAdmin{tok: "t"}
	s.connectorOAuth = newFakeOAuthCaller(admin)

	download := func(claims *runtime.SecurityClaims) *httptest.ResponseRecorder {
		tkn, err := s.generateDownloadToken(&runtimev1.ExportRequest{
			InstanceId: instanceID,
			Format:     runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
			Query: &runtimev1.Query{
				Query: &runtimev1.Query_MetricsViewAggregationRequest{
					MetricsViewAggregationRequest: &runtimev1.MetricsViewAggregationRequest{
						InstanceId:  instanceID,
						MetricsView: "mv1",
						Dimensions:  []*runtimev1.MetricsViewAggregationDimension{{Name: "country"}},
						Measures:    []*runtimev1.MetricsViewAggregationMeasure{{Name: "count"}},
					},
				},
			},
		}, claims)
		require.NoError(t, err)

		// Like a browser, the request has no Authorization header.
		req := httptest.NewRequest(http.MethodGet, "/v1/download?token="+url.QueryEscape(tkn), http.NoBody)
		rec := httptest.NewRecorder()
		s.connectorOAuthHTTPMiddleware(http.HandlerFunc(s.downloadHandler)).ServeHTTP(rec, req)
		return rec
	}

	rec := download(&runtime.SecurityClaims{UserID: "u", InteractiveUser: true, SkipChecks: true})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "US")
	require.Equal(t, []string{"u"}, admin.userIDs)

	admin.userIDs = nil
	rec = download(&runtime.SecurityClaims{UserID: "u", SkipChecks: true})
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), drivers.ErrOAuthUserRequired.Error())
	require.NotContains(t, admin.userIDs, "u")
}

func init() {
	drivers.Register("oauth_test_duckdb", oauthTestDriver{drivers.Drivers["duckdb"]})
}

// oauthTestDriver wraps DuckDB to fetch the caller's token before each query, like a connector that uses per-user OAuth.
type oauthTestDriver struct {
	drivers.Driver
}

func (d oauthTestDriver) Open(ctx context.Context, connectorName, instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	h, err := d.Driver.Open(ctx, connectorName, instanceID, config, st, ac, logger)
	if err != nil {
		return nil, err
	}
	return oauthTestHandle{h}, nil
}

type oauthTestHandle struct {
	drivers.Handle
}

func (h oauthTestHandle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	olap, ok := h.Handle.AsOLAP(instanceID)
	if !ok {
		return nil, false
	}
	return oauthTestOLAP{OLAPStore: olap, instanceID: instanceID}, true
}

// AsModelExecutor makes exports read through AsOLAP.
func (h oauthTestHandle) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

type oauthTestOLAP struct {
	drivers.OLAPStore
	instanceID string
}

func (o oauthTestOLAP) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	if !drivers.UsesServiceCredential(ctx) {
		caller := drivers.OAuthCallerFromContext(ctx)
		if caller == nil {
			return nil, drivers.ErrOAuthUserRequired
		}
		if _, err := caller.Token(ctx, o.instanceID, "dbx", "hash", false); err != nil {
			return nil, err
		}
	}
	return o.OLAPStore.Query(ctx, stmt)
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
