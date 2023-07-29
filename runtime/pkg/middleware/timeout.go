package middleware

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

func TimeoutStreamServerInterceptor(fn func(method string) time.Duration) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		duration := fn(info.FullMethod)

		if duration != 0 {
			ctx, cancel := context.WithTimeout(ss.Context(), duration)
			defer cancel()

			wss := grpc_middleware.WrapServerStream(ss)
			wss.WrappedContext = ctx
			return handler(srv, wss)
		}

		return handler(srv, ss)
	}
}

func TimeoutUnaryServerInterceptor(fn func(method string) time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		duration := fn(info.FullMethod)

		if duration != 0 {
			ctx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()
			return handler(ctx, req)
		}

		return handler(ctx, req)
	}
}
