package observability

import (
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	// TODO: Add recovery middleware
	return next
}

func RecoveryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor()
}

func RecoveryStreamServerInterceptor() grpc.StreamServerInterceptor {
	return recovery.StreamServerInterceptor()
}
