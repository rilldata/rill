package observability

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
)

func TracingMiddleware(next http.Handler, serviceName string) http.Handler {
	return otelhttp.NewHandler(next, serviceName)
}

func TracingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return otelgrpc.UnaryServerInterceptor()
}

func TracingStreamServerInterceptor() grpc.StreamServerInterceptor {
	return otelgrpc.StreamServerInterceptor()
}
