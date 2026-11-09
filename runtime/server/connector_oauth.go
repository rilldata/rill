package server

import (
	"context"
	"errors"
	"net/http"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc"
)

// connectorOAuthCaller implements drivers.OAuthCaller for the user a request is authenticated as.
// It reads the claims lazily, so only connectors that use per-user OAuth fetch a token.
type connectorOAuthCaller struct {
	admin func(ctx context.Context, instanceID string) (drivers.AdminService, func(), error)
}

var _ drivers.OAuthCaller = connectorOAuthCaller{}

// Token implements drivers.OAuthCaller.
func (c connectorOAuthCaller) Token(ctx context.Context, instanceID, connector, configHash string, forceRefresh bool) (*drivers.OAuthToken, error) {
	admin, release, err := c.admin(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	// Only a user's own session gets a personal token; "View as", magic tokens, embeds and service accounts also carry a UserID.
	claims := auth.GetClaims(ctx, instanceID)
	if claims == nil || !claims.InteractiveUser || claims.UserID == "" {
		// Probe without a user to tell Rill Developer, where connectors use their service credential, from Rill Cloud.
		_, err := admin.GetConnectorOAuthToken(ctx, connector, configHash, "", false)
		if errors.Is(err, drivers.ErrNotImplemented) {
			return nil, err
		}
		return nil, drivers.ErrOAuthUserRequired
	}

	tok, err := admin.GetConnectorOAuthToken(ctx, connector, configHash, claims.UserID, forceRefresh)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return nil, &drivers.OAuthAuthorizationRequiredError{Connector: connector}
		}
		return nil, err
	}
	return &drivers.OAuthToken{UserID: claims.UserID, AccessToken: tok.AccessToken, ExpiresOn: tok.ExpiresOn}, nil
}

// connectorOAuthUnaryServerInterceptor attaches a drivers.OAuthCaller for the request's user.
func (s *Server) connectorOAuthUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	caller := connectorOAuthCaller{admin: s.runtime.Admin}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(drivers.WithOAuthCaller(ctx, caller), req)
	}
}

// connectorOAuthStreamServerInterceptor attaches a drivers.OAuthCaller for the request's user.
func (s *Server) connectorOAuthStreamServerInterceptor() grpc.StreamServerInterceptor {
	caller := connectorOAuthCaller{admin: s.runtime.Admin}
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = drivers.WithOAuthCaller(ss.Context(), caller)
		return handler(srv, wrapped)
	}
}

// connectorOAuthHTTPMiddleware attaches a drivers.OAuthCaller for the request's user.
func (s *Server) connectorOAuthHTTPMiddleware(h http.Handler) http.Handler {
	caller := connectorOAuthCaller{admin: s.runtime.Admin}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(drivers.WithOAuthCaller(r.Context(), caller)))
	})
}
