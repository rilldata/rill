package jwt

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type jwtContextKey struct{}

// JWT secret for signing and verifying tokens
var jwtSecret = []byte("secret")

// Claims represents the payload of a JWT
type Claims struct {
	InstanceID string `json:"instance_id"`
	jwt.StandardClaims
}

// UnaryServerInterceptor is a middleware for parsing and extracting JWT from metadata
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	newCtx, err := parseJWTAndPopulateContextWithClaims(ctx)
	if err != nil {
		return nil, err
	}
	return handler(newCtx, req)
}

func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	newCtx, err := parseJWTAndPopulateContextWithClaims(ss.Context())
	if err != nil {
		return err
	}
	wrapped := grpc_middleware.WrapServerStream(ss)
	wrapped.WrappedContext = newCtx

	return handler(srv, wrapped)
}

func parseJWTAndPopulateContextWithClaims(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get metadata from context")
	}

	// Get authorization header
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

	token, err := jwt.ParseWithClaims(bearerToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Errorf(codes.InvalidArgument, "unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !token.Valid {
		return nil, status.Error(codes.PermissionDenied, "invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to parse JWT claims")
	}

	// Add JWT payload to the request context
	ctx = context.WithValue(ctx, jwtContextKey{}, claims)

	return ctx, nil
}

// GetJWTFromContext is a helper function for getting JWT payload from context
func GetJWTFromContext(ctx context.Context) *Claims {
	jwtClaims, ok := ctx.Value(jwtContextKey{}).(*Claims)
	if !ok {
		return nil
	}

	return jwtClaims
}
