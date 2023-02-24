package jwt

import (
	"context"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
// The assigned claims get be retrieved using GetClaims. It will always return a Claims interface.
// If auth is enabled, the claims are assigned from the JWT. If auth is disabled, the claims will allow any action.
func UnaryServerInterceptor(aud *Audience) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, err := parseClaims(ctx, aud)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// StreamServerInterceptor is the streaming variant of UnaryServerInterceptor.
func StreamServerInterceptor(aud *Audience) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx, err := parseClaims(ss.Context(), aud)
		if err != nil {
			return err
		}

		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = newCtx

		return handler(srv, wrapped)
	}
}

func parseClaims(ctx context.Context, aud *Audience) (context.Context, error) {
	// Special case: auth is disabled if aud is nil
	if aud == nil {
		return context.WithValue(ctx, claimsContextKey{}, openClaims{}), nil
	}

	// Parse claims from "authorization" header

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get metadata from context")
	}

	authHeaders, ok := md["authorization"]
	if !ok || len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	bearerToken := ""
	for _, authHeader := range authHeaders {
		if strings.HasPrefix(authHeader, "Bearer ") {
			bearerToken = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if bearerToken == "" {
		return nil, status.Error(codes.Unauthenticated, "no bearer token found in authorization header")
	}

	claims, err := aud.ParseAndValidate(bearerToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	ctx = context.WithValue(ctx, claimsContextKey{}, claims)
	return ctx, nil
}
