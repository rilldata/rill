package middleware

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rilldata/rill/runtime/pkg/publisher"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc"
)

func UsageStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		subject := auth.GetClaims(ss.Context()).Subject()

		newCtx := publisher.WithDims(ss.Context(),
			*publisher.String("user_id", subject),
			*publisher.String("request_method", info.FullMethod),
		)
		wss := grpc_middleware.WrapServerStream(ss)
		wss.WrappedContext = newCtx

		start := time.Now()
		defer func() {
			// Emit usage metric
			publisher.Get().Emit(newCtx, "request/time", float64(time.Since(start).Milliseconds()))
		}()

		return handler(srv, wss)
	}
}

func UsageUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		subject := auth.GetClaims(ctx).Subject()
		if subject == "" {
			subject = "anonymous"
		}

		newCtx := publisher.WithDims(ctx,
			*publisher.String("user_id", subject),
			*publisher.String("request_method", info.FullMethod),
		)

		start := time.Now()
		defer func() {
			// Emit usage metric
			publisher.Get().Emit(newCtx, "request/time", float64(time.Since(start).Milliseconds()))
		}()

		return handler(newCtx, req)
	}
}
