package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RegisterEndpoints adds HTTP endpoints for auth.
// Note that these are not gRPC handlers, just regular HTTP endpoints that we mount on the gRPC-gateway.
func (a *Authenticator) RegisterEndpoints(mux *gateway.ServeMux) error {
	err := mux.HandlePath("GET", "/auth/login", a.authLogin)
	if err != nil {
		return err
	}

	err = mux.HandlePath("GET", "/auth/callback", a.authLoginCallback)
	if err != nil {
		return err
	}

	err = mux.HandlePath("GET", "/auth/logout", a.authLogout)
	if err != nil {
		return err
	}

	err = mux.HandlePath("GET", "/auth/logout/callback", a.authLogoutCallback)
	if err != nil {
		return err
	}

	return nil
}

// Annotator is a gRPC-gateway annotator that moves access tokens in HTTP cookies to the "authorization" gRPC metadata.
func (a *Authenticator) Annotator(ctx context.Context, r *http.Request) metadata.MD {
	// Get auth cookie
	sess, err := a.cookies.Get(r, authCookieName)
	if err != nil {
		return metadata.Pairs()
	}

	token, ok := sess.Values["access_token"].(string)
	if ok && token != "" {
		return metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
	}

	return metadata.Pairs()
}

// UnaryServerInterceptor is a middleware for setting claims on runtime server requests.
// The assigned claims can be retrieved using GetClaims. If the interceptor succeeds, a Claims value is guaranteed to be set on the ctx.
func (a *Authenticator) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		authHeader := metautils.ExtractIncoming(ctx).Get("authorization")
		newCtx, err := a.parseClaimsFromBearer(ctx, authHeader)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		return handler(newCtx, req)
	}
}

// StreamServerInterceptor is the streaming variant of UnaryServerInterceptor.
func (a *Authenticator) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		authHeader := metautils.ExtractIncoming(ss.Context()).Get("authorization")
		newCtx, err := a.parseClaimsFromBearer(ss.Context(), authHeader)
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}

		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = newCtx

		return handler(srv, wrapped)
	}
}

// HTTPMiddleware is a HTTP middleware variant of UnaryServerInterceptor.
// It additionally supports reading access tokens from cookies.
// It should be used for non-gRPC HTTP endpoints (CookieAuthAnnotator takes care of handling cookies in gRPC-gateway requests).
func (a *Authenticator) HTTPMiddleware(next gateway.HandlerFunc) gateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		// Handle authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			newCtx, err := a.parseClaimsFromBearer(r.Context(), authHeader)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			next(w, r.WithContext(newCtx), pathParams)
			return
		}

		// There was no authorization header. Try the cookie.
		sess, err := a.cookies.Get(r, authCookieName)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get session: %s", err), http.StatusInternalServerError)
			return
		}

		// Read access token from cookie
		authToken, ok := sess.Values["access_token"].(string)
		if ok && authToken != "" {
			newCtx, err := a.parseClaimsFromToken(r.Context(), authToken)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			next(w, r.WithContext(newCtx), pathParams)
			return
		}

		// No token was found. Set anonClaims.
		newCtx := context.WithValue(r.Context(), claimsContextKey{}, anonClaims{})
		next(w, r.WithContext(newCtx), pathParams)
	}
}

func (a *Authenticator) parseClaimsFromBearer(ctx context.Context, authorizationHeader string) (context.Context, error) {
	// If authorization header is not set, we set anonClaims.
	if authorizationHeader == "" {
		ctx = context.WithValue(ctx, claimsContextKey{}, anonClaims{})
		return ctx, nil
	}

	// Extract bearer token
	bearerToken := ""
	if len(authorizationHeader) >= 6 && strings.EqualFold(authorizationHeader[0:6], "bearer") {
		bearerToken = strings.TrimSpace(authorizationHeader[6:])
	}
	if bearerToken == "" {
		return nil, errors.New("no bearer token found in authorization header")
	}

	return a.parseClaimsFromToken(ctx, bearerToken)
}

func (a *Authenticator) parseClaimsFromToken(ctx context.Context, token string) (context.Context, error) {
	validated, err := a.admin.ValidateAuthToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Set claims
	claims := &tokenClaims{token: validated}
	ctx = context.WithValue(ctx, claimsContextKey{}, claims)
	return ctx, nil
}
