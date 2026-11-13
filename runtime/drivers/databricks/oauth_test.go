package databricks

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeCaller struct {
	tok *drivers.OAuthToken
	err error
}

func (f fakeCaller) Token(ctx context.Context, instanceID, connector, configHash string, forceRefresh bool) (*drivers.OAuthToken, error) {
	return f.tok, f.err
}

func openOAuthConnection(t *testing.T) *connection {
	h, err := driver{}.Open(context.Background(), "dbx", "inst", map[string]any{
		"host":                "dbc-test.cloud.databricks.com",
		"http_path":           "/sql/1.0/warehouses/abc",
		"token":               "service",
		"auth":                "oauth",
		"oauth_client_id":     "id",
		"oauth_client_secret": "secret",
	}, nil, nil, zap.NewNop())
	require.NoError(t, err)
	c := h.(*connection)
	t.Cleanup(func() { require.NoError(t, c.Close()) })
	return c
}

func TestGetDBCredentialSelection(t *testing.T) {
	c := openOAuthConnection(t)
	require.Equal(t, "https://dbc-test.cloud.databricks.com/oidc", c.OAuthConfig().Issuer)

	ctx := context.Background()
	service, err := c.getServiceDB(ctx)
	require.NoError(t, err)

	// No caller fails closed
	_, err = c.getDB(ctx)
	require.ErrorIs(t, err, drivers.ErrOAuthUserRequired)

	// Rill Developer uses the service credential
	db, err := c.getDB(drivers.WithOAuthCaller(ctx, fakeCaller{err: drivers.ErrNotImplemented}))
	require.NoError(t, err)
	require.Same(t, service, db)

	// A user without a token gets the caller's authorization error unchanged, never the service credential
	authErr := &drivers.OAuthAuthorizationRequiredError{Connector: "dbx"}
	_, err = c.getDB(drivers.WithOAuthCaller(ctx, fakeCaller{err: authErr}))
	var target *drivers.OAuthAuthorizationRequiredError
	require.ErrorAs(t, err, &target)
	require.Equal(t, "dbx", target.Connector)
	require.Equal(t, authErr.Error(), err.Error())

	// Each user gets their own pool, which follows token rotation
	a1, err := c.getDB(drivers.WithOAuthCaller(ctx, fakeCaller{tok: &drivers.OAuthToken{UserID: "a", AccessToken: "a1"}}))
	require.NoError(t, err)
	b, err := c.getDB(drivers.WithOAuthCaller(ctx, fakeCaller{tok: &drivers.OAuthToken{UserID: "b", AccessToken: "b1"}}))
	require.NoError(t, err)
	a2, err := c.getDB(drivers.WithOAuthCaller(ctx, fakeCaller{tok: &drivers.OAuthToken{UserID: "a", AccessToken: "a2"}}))
	require.NoError(t, err)
	require.NotSame(t, service, a1)
	require.NotSame(t, a1, b)
	require.Same(t, a1, a2)
	require.Equal(t, "a2", *c.userDBs["a"].token.Load())
}

func TestUserDBConcurrentOpen(t *testing.T) {
	c := openOAuthConnection(t)
	ctx := drivers.WithOAuthCaller(context.Background(), fakeCaller{tok: &drivers.OAuthToken{UserID: "a", AccessToken: "a1"}})

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.getDB(ctx)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	require.Len(t, c.userDBs, 1)
}

// refreshCaller returns user a's token and counts forced refreshes.
type refreshCaller struct {
	refreshErr error
	forced     atomic.Int32
}

func (f *refreshCaller) Token(ctx context.Context, instanceID, connector, configHash string, forceRefresh bool) (*drivers.OAuthToken, error) {
	if !forceRefresh {
		return &drivers.OAuthToken{UserID: "a", AccessToken: "a1"}, nil
	}
	f.forced.Add(1)
	if f.refreshErr != nil {
		return nil, f.refreshErr
	}
	return &drivers.OAuthToken{UserID: "a", AccessToken: "a2"}, nil
}

func TestRejectedTokenRefreshesOnce(t *testing.T) {
	c := openOAuthConnection(t)
	errRejected := errors.New("HTTP Response code: 401")
	reject := func(ctx context.Context) error {
		ctx.Value(rejectedKey{}).(*atomic.Bool).Store(true)
		return errRejected
	}
	// rejectFirst rejects its first call only.
	rejectFirst := func(calls *int) func(ctx context.Context, db *sqlx.DB) error {
		return func(ctx context.Context, db *sqlx.DB) error {
			*calls++
			if *calls == 1 {
				return reject(ctx)
			}
			return nil
		}
	}

	caller := &refreshCaller{}
	ctx := drivers.WithOAuthCaller(context.Background(), caller)
	var calls int
	require.NoError(t, c.withRetry(ctx, rejectFirst(&calls)))
	require.Equal(t, 2, calls)
	require.Equal(t, int32(1), caller.forced.Load())
	require.Equal(t, "a2", *c.userDBs["a"].token.Load())

	// Other errors aren't retried
	calls = 0
	err := c.withRetry(ctx, func(ctx context.Context, db *sqlx.DB) error {
		calls++
		return errRejected
	})
	require.ErrorIs(t, err, errRejected)
	require.Equal(t, 1, calls)
	require.Equal(t, int32(1), caller.forced.Load())

	// The refreshed token is rejected too
	calls = 0
	err = c.withRetry(ctx, func(ctx context.Context, db *sqlx.DB) error {
		calls++
		return reject(ctx)
	})
	var target *drivers.OAuthAuthorizationRequiredError
	require.ErrorAs(t, err, &target)
	require.Equal(t, "dbx", target.Connector)
	require.Equal(t, 2, calls)
	require.Equal(t, int32(2), caller.forced.Load())

	// The user must sign in again; concurrent rejected calls each get the authorization error
	caller = &refreshCaller{refreshErr: &drivers.OAuthAuthorizationRequiredError{Connector: "dbx"}}
	ctx = drivers.WithOAuthCaller(context.Background(), caller)
	var inFlight, wg sync.WaitGroup
	inFlight.Add(2)
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var calls int
			errs[i] = c.withRetry(ctx, func(ctx context.Context, db *sqlx.DB) error {
				calls++
				if calls == 1 {
					inFlight.Done()
					inFlight.Wait()
				}
				return reject(ctx)
			})
		}()
	}
	wg.Wait()
	for _, err := range errs {
		require.ErrorAs(t, err, &target)
		require.Equal(t, "dbx", target.Connector)
	}
	require.Equal(t, int32(2), caller.forced.Load())
}

// TestRejectionTransport runs a user pool through the real Databricks driver against a server that rejects every request.
func TestRejectionTransport(t *testing.T) {
	var auths []string
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		auths = append(auths, r.Header.Get("Authorization"))
		mu.Unlock()
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	// User pools connect to port 443 of the host; route that to srv.
	base := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = base })
	http.DefaultTransport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, srv.Listener.Addr().String())
		},
	}

	c := openOAuthConnection(t)
	c.config.Host = "localhost" // Served over HTTP by databricks-sql-go
	caller := &refreshCaller{}
	ctx := drivers.WithOAuthCaller(context.Background(), caller)

	// A 401 outside withRetry flags nothing for later calls
	db, err := c.getDB(ctx)
	require.NoError(t, err)
	require.Error(t, db.PingContext(ctx))
	errOther := errors.New("other")
	err = c.withRetry(ctx, func(ctx context.Context, db *sqlx.DB) error { return errOther })
	require.ErrorIs(t, err, errOther)
	require.Equal(t, int32(0), caller.forced.Load())

	err = c.withRetry(ctx, func(ctx context.Context, db *sqlx.DB) error { return db.PingContext(ctx) })
	var target *drivers.OAuthAuthorizationRequiredError
	require.ErrorAs(t, err, &target)
	require.Equal(t, int32(1), caller.forced.Load())
	require.Equal(t, []string{"Bearer a1", "Bearer a1", "Bearer a2"}, auths)

	// Query runs as the caller and refreshes the rejected token
	auths = nil
	_, err = c.Query(ctx, &drivers.Statement{Query: "SELECT 1"})
	require.ErrorAs(t, err, &target)
	require.Equal(t, int32(2), caller.forced.Load())
	require.Equal(t, []string{"Bearer a1", "Bearer a2"}, auths)

	// Without a caller, Query fails closed
	_, err = c.Query(context.Background(), &drivers.Statement{Query: "SELECT 1"})
	require.ErrorIs(t, err, drivers.ErrOAuthUserRequired)

	// Other operations use the service credential, with or without a caller.
	// The service pool connects to c.config.DSN; route it to srv over HTTP.
	c.config.DSN = "http://token:service@" + srv.Listener.Addr().String() + "/sql/1.0/warehouses/abc"
	for _, ctx := range []context.Context{ctx, context.Background()} {
		for name, fn := range map[string]func() error{
			"Ping": func() error { return c.Ping(ctx) },
			"Exec": func() error { return c.Exec(ctx, &drivers.Statement{Query: "SELECT 1"}) },
			"ExecDryRun": func() error {
				return c.Exec(ctx, &drivers.Statement{Query: "SELECT 1", DryRun: true})
			},
			"QuerySchema": func() error {
				_, err := c.QuerySchema(ctx, "SELECT 1", nil)
				return err
			},
			"ListDatabaseSchemas": func() error {
				_, _, err := c.ListDatabaseSchemas(ctx, 0, "")
				return err
			},
			"ListTables": func() error {
				_, _, err := c.ListTables(ctx, "", "s", 0, "")
				return err
			},
			"Lookup": func() error {
				_, err := c.Lookup(ctx, "", "s", "t")
				return err
			},
			"LoadDDL": func() error { return c.LoadDDL(ctx, &drivers.OlapTable{Name: "t"}) },
		} {
			auths = nil
			err := fn()
			require.Error(t, err, name)
			require.NotErrorIs(t, err, drivers.ErrOAuthUserRequired, name)
			require.False(t, errors.As(err, &target), name)
			require.NotEmpty(t, auths, name)
			for _, a := range auths {
				require.Equal(t, "Bearer service", a, name)
			}
		}
	}
	require.Equal(t, int32(2), caller.forced.Load())
}

func TestOAuthConfigValidation(t *testing.T) {
	base := func() *configProperties {
		return &configProperties{Host: "h", HTTPPath: "/p", Token: "t", Auth: "oauth", OAuthClientID: "id", OAuthClientSecret: "s"}
	}
	require.NoError(t, base().validate())

	c := base()
	c.OAuthClientSecret = ""
	require.Error(t, c.validate())

	c = base()
	c.Auth = "pat"
	require.Error(t, c.validate())

	// The service token is still required for system work
	c = base()
	c.Token = ""
	require.Error(t, c.validate())

	c = &configProperties{DSN: "token:t@h:443/p", Auth: "oauth", OAuthClientID: "id", OAuthClientSecret: "s"}
	require.Error(t, c.validate())

	// OAuth properties without 'auth: oauth' would silently use the service token
	for _, c := range []*configProperties{
		{Host: "h", HTTPPath: "/p", Token: "t", OAuthClientID: "id"},
		{Host: "h", HTTPPath: "/p", Token: "t", OAuthClientSecret: "s"},
		{Host: "h", HTTPPath: "/p", Token: "t", OAuthScopes: []string{"sql"}},
	} {
		require.Error(t, c.validate())
	}
}

func TestOAuthConfigHash(t *testing.T) {
	base := func() *configProperties {
		return &configProperties{Host: "h", HTTPPath: "/p", Token: "t", Auth: "oauth", OAuthClientID: "id", OAuthClientSecret: "s"}
	}
	hash := newOAuthConfig(base()).Hash
	require.NotEmpty(t, hash)
	require.Equal(t, hash, newOAuthConfig(base()).Hash)

	c := base()
	c.OAuthClientSecret = "rotated"
	require.NotEqual(t, hash, newOAuthConfig(c).Hash)

	c = base()
	c.OAuthScopes = []string{"sql"}
	require.NotEqual(t, hash, newOAuthConfig(c).Hash)
}
