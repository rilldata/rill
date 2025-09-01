package middleware

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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

func TimeoutMCPToolHandlerMiddleware(fn func(tool string) time.Duration) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			duration := fn(req.Params.Name)
			if duration == 0 {
				return next(ctx, req)
			}

			ctx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()

			return next(ctx, req)
		}
	}
}
