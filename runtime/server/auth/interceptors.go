package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// claimsContextKey is used to set and get Claims on a request context.
type claimsContextKey struct{}

// GetClaims retrieves Claims from a request context.
// It should only be used in handlers intercepted by UnaryServerInterceptor or StreamServerInterceptor.
func GetClaims(ctx context.Context) Claims {
	claims, ok := ctx.Value(claimsContextKey{}).(Claims)
	if !ok {
		return nil
	}

	return claims
}

// UnaryServerInterceptor is a middleware for setting claims on runtime server requests.
// The assigned claims can be retrieved using GetClaims. If the interceptor succeeds, a Claims value is guaranteed to be set on the ctx.
// The claim parsing logic is as follows
// - When aud is nil, auth is considered disabled. We set a Claims that allows all actions (openClaims).
// - When aud is not nil, we set a Claims based on a JWT set as a bearer token in the authorization header (jwtClaims).
// - When aud is not nil and no authoriation header is passed, we set a Claims that denies any action (anonClaims).
func UnaryServerInterceptor(aud *Audience) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		authHeader := metautils.ExtractIncoming(ctx).Get("authorization")
		newCtx, err := parseClaims(ctx, aud, authHeader)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		return handler(newCtx, req)
	}
}

// StreamServerInterceptor is the streaming variant of UnaryServerInterceptor.
func StreamServerInterceptor(aud *Audience) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		authHeader := metautils.ExtractIncoming(ss.Context()).Get("authorization")
		newCtx, err := parseClaims(ss.Context(), aud, authHeader)
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}

		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = newCtx

		return handler(srv, wrapped)
	}
}

// HTTPMiddleware is a HTTP middleware variant of UnaryServerInterceptor.
// It should be used for non-gRPC HTTP endpoints.
func HTTPMiddleware(aud *Audience, next gateway.HandlerFunc) gateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		authHeader := r.Header.Get("Authorization")
		newCtx, err := parseClaims(r.Context(), aud, authHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		next(w, r.WithContext(newCtx), pathParams)
	}
}

func parseClaims(ctx context.Context, aud *Audience, authorizationHeader string) (context.Context, error) {
	// When aud == nil, it means auth is disabled. We set openClaims.
	if aud == nil {
		return context.WithValue(ctx, claimsContextKey{}, openClaims{}), nil
	}

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

	// Parse, validate and set claims from JWT
	claims, err := aud.ParseAndValidate(bearerToken)
	if err != nil {
		return nil, err
	}

	// Set subject in span
	subject := claims.Subject()
	if subject != "" {
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(semconv.EnduserID(subject))
	}

	ctx = context.WithValue(ctx, claimsContextKey{}, claims)
	return ctx, nil
}

// WithOpen wraps a context with open claims. It's used for testing.
// NOTE: We should remove this when the server tests support interceptors.
func WithOpen(ctx context.Context) context.Context {
	return context.WithValue(ctx, claimsContextKey{}, openClaims{})
}
