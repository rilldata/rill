package databricks

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	dbsql "github.com/databricks/databricks-sql-go"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
)

const (
	authOAuth = "oauth"
	// userDBIdleTTL is how long a user's connection pool stays open without queries.
	userDBIdleTTL = 30 * time.Minute
	// userDBMaxOpenConns caps the Databricks sessions per user.
	userDBMaxOpenConns = 8
)

// defaultOAuthScopes lets Rill run SQL as the user and refresh the user's token.
var defaultOAuthScopes = []string{"sql", "offline_access"}

var _ drivers.OAuthConfigurer = (*connection)(nil)

// OAuthConfig implements drivers.OAuthConfigurer.
func (c *connection) OAuthConfig() *drivers.OAuthConfig {
	return c.oauthConfig
}

// newOAuthConfig returns the configuration for a Databricks account-level custom app connection, used against the workspace's OIDC endpoints.
func newOAuthConfig(conf *configProperties) *drivers.OAuthConfig {
	scopes := conf.OAuthScopes
	if len(scopes) == 0 {
		scopes = defaultOAuthScopes
	}
	cfg := &drivers.OAuthConfig{
		ServerURL:    "https://" + conf.Host,
		Issuer:       "https://" + conf.Host + "/oidc",
		ClientID:     conf.OAuthClientID,
		ClientSecret: conf.OAuthClientSecret,
		Scopes:       scopes,
	}
	// Changing any field invalidates existing authorizations; fields are length-prefixed so different configurations can't collide.
	h := sha256.New()
	for _, s := range append([]string{cfg.Issuer, cfg.ClientID, cfg.ClientSecret}, scopes...) {
		_, _ = fmt.Fprintf(h, "%d:%s;", len(s), s)
	}
	cfg.Hash = hex.EncodeToString(h.Sum(nil))
	return cfg
}

// getDB returns the connection pool for the context's caller.
// Without per-user OAuth, and for system work marked with drivers.WithServiceCredential, it is the service credential's pool.
// Otherwise it is the caller's own pool; it never falls back to the service credential for a caller.
func (c *connection) getDB(ctx context.Context) (*sqlx.DB, error) {
	db, _, err := c.getPool(ctx)
	return db, err
}

// getPool is like getDB, and also returns the caller's userDB, or nil for the service credential's pool.
func (c *connection) getPool(ctx context.Context) (*sqlx.DB, *userDB, error) {
	if c.oauthConfig == nil || drivers.UsesServiceCredential(ctx) {
		db, err := c.getServiceDB(ctx)
		return db, nil, err
	}
	caller := drivers.OAuthCallerFromContext(ctx)
	if caller == nil {
		return nil, nil, drivers.ErrOAuthUserRequired
	}
	tok, err := caller.Token(ctx, c.instanceID, c.name, c.oauthConfig.Hash, false)
	if err != nil {
		if errors.Is(err, drivers.ErrNotImplemented) {
			// Rill Developer: there are no users, so the service credential is used.
			db, err := c.getServiceDB(ctx)
			return db, nil, err
		}
		return nil, nil, err
	}
	u, err := c.getUserDB(tok)
	if err != nil {
		return nil, nil, err
	}
	return u.db, u, nil
}

// withRetry runs fn on the caller's pool.
// If Databricks rejected the user's token during the call, it force-refreshes the token and runs fn once more.
func (c *connection) withRetry(ctx context.Context, fn func(db *sqlx.DB) error) error {
	db, u, err := c.getPool(ctx)
	if err != nil {
		return err
	}
	err = fn(db)
	if err == nil || u == nil || !u.rejected.Swap(false) {
		return err
	}
	tok, refreshErr := drivers.OAuthCallerFromContext(ctx).Token(ctx, c.instanceID, c.name, c.oauthConfig.Hash, true)
	if refreshErr != nil {
		var authErr *drivers.OAuthAuthorizationRequiredError
		if errors.As(refreshErr, &authErr) {
			return refreshErr
		}
		return errors.Join(err, refreshErr)
	}
	u, err = c.getUserDB(tok)
	if err != nil {
		return err
	}
	return fn(u.db)
}

// getUserDB returns the user's connection pool, opening it if needed, and sets the token it authenticates with.
func (c *connection) getUserDB(tok *drivers.OAuthToken) (*userDB, error) {
	c.userDBsMu.Lock()
	defer c.userDBsMu.Unlock()

	now := time.Now()
	for id, u := range c.userDBs {
		if id != tok.UserID && now.Sub(time.Unix(0, u.lastUsed.Load())) > userDBIdleTTL {
			_ = u.db.Close()
			delete(c.userDBs, id)
		}
	}

	u, ok := c.userDBs[tok.UserID]
	if !ok {
		u = &userDB{}
		connector, err := dbsql.NewConnector(
			dbsql.WithServerHostname(c.config.Host),
			dbsql.WithPort(443),
			dbsql.WithHTTPPath(c.config.HTTPPath),
			dbsql.WithInitialNamespace(c.config.Catalog, c.config.Schema),
			dbsql.WithSessionParams(map[string]string{"timezone": "UTC"}),
			// The authenticator runs on every HTTP request, so the pool keeps working as the user's access token rotates.
			dbsql.WithAuthenticator(u),
			dbsql.WithTransport(u),
		)
		if err != nil {
			return nil, err
		}
		u.db = sqlx.NewDb(sql.OpenDB(connector), "databricks")
		u.db.SetMaxOpenConns(userDBMaxOpenConns)
		u.db.SetConnMaxIdleTime(userDBIdleTTL)
		c.userDBs[tok.UserID] = u
	}
	u.token.Store(&tok.AccessToken)
	u.lastUsed.Store(now.UnixNano())
	return u, nil
}

// userDB is a connection pool that authenticates as one user.
type userDB struct {
	db    *sqlx.DB
	token atomic.Pointer[string]
	// rejected is set when Databricks responds 401 to the token.
	rejected atomic.Bool
	lastUsed atomic.Int64
}

// Authenticate implements auth.Authenticator from databricks-sql-go.
func (u *userDB) Authenticate(r *http.Request) error {
	tok := u.token.Load()
	if tok == nil || *tok == "" {
		return errors.New("databricks: no OAuth token for the user")
	}
	r.Header.Set("Authorization", "Bearer "+*tok)
	return nil
}

// RoundTrip implements http.RoundTripper; it carries the pool's requests, including those of the Thrift client.
func (u *userDB) RoundTrip(r *http.Request) (*http.Response, error) {
	res, err := http.DefaultTransport.RoundTrip(r)
	if err == nil && res.StatusCode == http.StatusUnauthorized {
		u.rejected.Store(true)
	}
	return res, err
}
