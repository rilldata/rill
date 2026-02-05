package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Annotator is a gRPC-gateway annotator that moves access tokens in HTTP cookies to the "authorization" gRPC metadata.
func (a *Authenticator) Annotator(ctx context.Context, r *http.Request) metadata.MD {
	// Get auth cookie
	sess := a.cookies.Get(r, cookieName)
	// Get access token from cookie and pretend it's a bearer token
	token, ok := sess.Values[cookieFieldAccessToken].(string)
	if ok && token != "" {
		return metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
	}

	return metadata.Pairs()
}

// UnaryServerInterceptor is a middleware for setting claims on runtime server requests.
// It authenticates the user and acquires the claims using the bearer token in the "authorization" request metadata field.
// If no bearer token is found, it will still succeed, setting anonClaims on the request.
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

// CookieToAuthHeader is a middleware that reads the access token from the cookie and sets it in the "Authorization" header
// only if the Authorization header isn't already present.
func (a *Authenticator) CookieToAuthHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			sess := a.cookies.Get(r, cookieName)
			authToken, ok := sess.Values[cookieFieldAccessToken].(string)
			if ok && authToken != "" {
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// HTTPMiddleware is a HTTP middleware variant of UnaryServerInterceptor.
// It additionally supports reading access tokens from cookies.
// It should be used for non-gRPC HTTP endpoints (CookieAuthAnnotator takes care of handling cookies in gRPC-gateway requests).
func (a *Authenticator) HTTPMiddleware(next http.Handler) http.Handler {
	return a.httpMiddleware(next, false)
}

// HTTPMiddlewareLenient is a lenient variant of HTTPMiddleware.
// If the authoriztion header is malformed or invalid, it will still succeed, setting anonClaims on the request.
func (a *Authenticator) HTTPMiddlewareLenient(next http.Handler) http.Handler {
	return a.httpMiddleware(next, true)
}

// CookieRefreshMiddleware is a middleware that refreshes the auth cookie.
// This enables us to do rolling cookie refreshes so we can have a relatively short cookie max age.
// Note that it does not update the auth token encrypted inside the cookie.
func (a *Authenticator) CookieRefreshMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := a.cookies.Get(r, cookieName)
		if authToken, ok := sess.Values[cookieFieldAccessToken].(string); ok && authToken != "" {
			// Re-save the cookie to refresh its expiration
			if err := sess.Save(r, w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// httpMiddleware is the actual implementation of HTTPMiddleware and HTTPMiddlewareLenient.
func (a *Authenticator) httpMiddleware(next http.Handler, lenient bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			newCtx, err := a.parseClaimsFromBearer(r.Context(), authHeader)
			if err != nil {
				// In lenient mode, we set anonClaims.
				if lenient {
					newCtx := context.WithValue(r.Context(), claimsContextKey{}, anonClaims{})
					next.ServeHTTP(w, r.WithContext(newCtx))
					return
				}

				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(newCtx))
			return
		}

		// There was no authorization header. Try reading an access token from cookies.
		sess := a.cookies.Get(r, cookieName)
		authToken, ok := sess.Values[cookieFieldAccessToken].(string)
		if ok && authToken != "" {
			newCtx, err := a.parseClaimsFromToken(r.Context(), authToken)
			if err != nil {
				// In lenient mode, we set anonClaims.
				if lenient {
					newCtx := context.WithValue(r.Context(), claimsContextKey{}, anonClaims{})
					next.ServeHTTP(w, r.WithContext(newCtx))
					return
				}

				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(newCtx))
			return
		}

		// No token was found. Set anonClaims.
		newCtx := context.WithValue(r.Context(), claimsContextKey{}, anonClaims{})
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
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
	// Validate token against database
	validated, err := a.admin.ValidateAuthToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Set claims
	claims := newAuthTokenClaims(validated, a.admin)
	ctx = context.WithValue(ctx, claimsContextKey{}, claims)

	// Set user ID in span
	if claims.OwnerType() == OwnerTypeUser {
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(semconv.EnduserID(claims.OwnerID()))
	}

	return ctx, nil
}
