package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

func TimeoutStreamServerInterceptor(fn func(service, method string) time.Duration) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		service, method, err := parseFullMethod(info.FullMethod)
		if err != nil {
			return err
		}

		duration := fn(service, method)

		ctx, cancel := context.WithTimeout(ss.Context(), duration)
		defer cancel()

		wss := grpc_middleware.WrapServerStream(ss)
		wss.WrappedContext = ctx
		return handler(srv, wss)
	}
}

func TimeoutUnaryServerInterceptor(fn func(service, method string) time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		service, method, err := parseFullMethod(info.FullMethod)
		if err != nil {
			return nil, err
		}

		duration := fn(service, method)

		ctx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()
		return handler(ctx, req)
	}
}

func parseFullMethod(fullMethod string) (string, string, error) {
	name := strings.TrimLeft(fullMethod, "/")
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid format, %s does not follow `/package.service/method`", name)
	}
	return parts[0], parts[1], nil
}
