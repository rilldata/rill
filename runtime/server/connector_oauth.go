package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/singleflight"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// connectorOAuthTokenCacheTTL is the longest a token is cached.
	connectorOAuthTokenCacheTTL = 60 * time.Second
	// connectorOAuthTokenExpiryMargin is how long before expiry a token stops being served from the cache.
	connectorOAuthTokenExpiryMargin = time.Minute
)

// connectorOAuthCaller implements drivers.OAuthCaller for the user a request is authenticated as.
// It reads the claims lazily, so only connectors that use per-user OAuth fetch a token.
// Tokens are cached briefly to avoid an admin round trip per query.
type connectorOAuthCaller struct {
	admin  func(ctx context.Context, instanceID string) (drivers.AdminService, func(), error)
	tokens *connectorOAuthTokenCache
	flight *singleflight.Group[string, *drivers.OAuthToken]
	// probes maps an instance ID to the outcome of probing admin without a user, which never changes for an instance.
	probes sync.Map
}

var _ drivers.OAuthCaller = (*connectorOAuthCaller)(nil)

func newConnectorOAuthCaller(admin func(ctx context.Context, instanceID string) (drivers.AdminService, func(), error)) *connectorOAuthCaller {
	return &connectorOAuthCaller{
		admin:  admin,
		tokens: &connectorOAuthTokenCache{entries: make(map[string]connectorOAuthTokenCacheEntry)},
		flight: &singleflight.Group[string, *drivers.OAuthToken]{},
	}
}

// Token implements drivers.OAuthCaller.
func (c *connectorOAuthCaller) Token(ctx context.Context, instanceID, connector, configHash string, forceRefresh bool) (*drivers.OAuthToken, error) {
	// Only a user's own session gets a personal token; "View as", magic tokens, embeds and service accounts also carry a UserID.
	claims := auth.GetClaims(ctx, instanceID)
	if claims == nil || !claims.InteractiveUser || claims.UserID == "" {
		return nil, c.nonInteractiveError(ctx, instanceID, connector, configHash)
	}

	key := instanceID + "\x00" + connector + "\x00" + configHash + "\x00" + claims.UserID
	if !forceRefresh {
		if tok, ok := c.tokens.get(key); ok {
			return tok, nil
		}
	}

	// A forced refresh must not join a fetch that may return the token being replaced.
	return c.flight.Do(ctx, key+"\x00"+strconv.FormatBool(forceRefresh), func(ctx context.Context) (*drivers.OAuthToken, error) {
		admin, release, err := c.admin(ctx, instanceID)
		if err != nil {
			return nil, err
		}
		defer release()

		tok, err := admin.GetConnectorOAuthToken(ctx, connector, configHash, claims.UserID, forceRefresh)
		if err != nil {
			if errors.Is(err, drivers.ErrNotFound) {
				return nil, &drivers.OAuthAuthorizationRequiredError{Connector: connector}
			}
			return nil, err
		}
		res := &drivers.OAuthToken{UserID: claims.UserID, AccessToken: tok.AccessToken, ExpiresOn: tok.ExpiresOn}
		c.tokens.set(key, res)
		return res, nil
	})
}

// nonInteractiveError returns the error for a caller without a personal session.
func (c *connectorOAuthCaller) nonInteractiveError(ctx context.Context, instanceID, connector, configHash string) error {
	if err, ok := c.probes.Load(instanceID); ok {
		return err.(error)
	}

	admin, release, err := c.admin(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	// Probe without a user to tell Rill Developer, where connectors use their service credential, from Rill Cloud.
	// Rill Cloud rejects the empty user ID as invalid.
	// Only definitive outcomes are remembered, not transient errors.
	_, err = admin.GetConnectorOAuthToken(ctx, connector, configHash, "", false)
	switch {
	case errors.Is(err, drivers.ErrNotImplemented):
		c.probes.Store(instanceID, drivers.ErrNotImplemented)
		return drivers.ErrNotImplemented
	case err == nil || errors.Is(err, drivers.ErrNotFound) || status.Code(err) == codes.InvalidArgument:
		c.probes.Store(instanceID, drivers.ErrOAuthUserRequired)
	}
	return drivers.ErrOAuthUserRequired
}

// connectorOAuthTokenCache holds access tokens briefly.
type connectorOAuthTokenCache struct {
	mu      sync.Mutex
	entries map[string]connectorOAuthTokenCacheEntry
}

type connectorOAuthTokenCacheEntry struct {
	tok     *drivers.OAuthToken
	expires time.Time
}

func (c *connectorOAuthTokenCache) get(key string) (*drivers.OAuthToken, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || !time.Now().Before(e.expires) {
		return nil, false
	}
	return e.tok, true
}

// set caches tok for min(connectorOAuthTokenCacheTTL, time until expiry minus the margin), or not at all if that is not positive.
// A token without an expiry is cached for the full TTL. Expired entries are pruned.
func (c *connectorOAuthTokenCache) set(key string, tok *drivers.OAuthToken) {
	now := time.Now()
	ttl := connectorOAuthTokenCacheTTL
	if !tok.ExpiresOn.IsZero() {
		ttl = min(ttl, tok.ExpiresOn.Sub(now)-connectorOAuthTokenExpiryMargin)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for k, e := range c.entries {
		if !now.Before(e.expires) {
			delete(c.entries, k)
		}
	}
	if ttl <= 0 {
		delete(c.entries, key)
		return
	}
	c.entries[key] = connectorOAuthTokenCacheEntry{tok: tok, expires: now.Add(ttl)}
}

// connectorOAuthUnaryServerInterceptor attaches a drivers.OAuthCaller for the request's user.
func (s *Server) connectorOAuthUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(drivers.WithOAuthCaller(ctx, s.connectorOAuth), req)
	}
}

// connectorOAuthStreamServerInterceptor attaches a drivers.OAuthCaller for the request's user.
func (s *Server) connectorOAuthStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = drivers.WithOAuthCaller(ss.Context(), s.connectorOAuth)
		return handler(srv, wrapped)
	}
}

// connectorOAuthHTTPMiddleware attaches a drivers.OAuthCaller for the request's user.
func (s *Server) connectorOAuthHTTPMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(drivers.WithOAuthCaller(r.Context(), s.connectorOAuth)))
	})
}
