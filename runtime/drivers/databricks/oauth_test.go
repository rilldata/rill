package databricks

import (
	"context"
	"sync"
	"testing"

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

	// System work uses the service credential
	db, err := c.getDB(drivers.WithServiceCredential(ctx))
	require.NoError(t, err)
	require.Same(t, service, db)

	// No caller and no system marker fails closed
	_, err = c.getDB(ctx)
	require.ErrorIs(t, err, drivers.ErrOAuthUserRequired)

	// Work on behalf of a user inside system work fails closed without a caller
	_, err = c.getDB(drivers.WithoutServiceCredential(drivers.WithServiceCredential(ctx)))
	require.ErrorIs(t, err, drivers.ErrOAuthUserRequired)

	// Rill Developer uses the service credential
	db, err = c.getDB(drivers.WithOAuthCaller(ctx, fakeCaller{err: drivers.ErrNotImplemented}))
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
