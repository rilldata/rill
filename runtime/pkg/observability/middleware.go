package observability

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/providers/zap/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Middleware is HTTP middleware that combines all observability-related middlewares.
func Middleware(serviceName string, logger *zap.Logger, next http.Handler) http.Handler {
	return TracingMiddleware(LoggingMiddleware(logger, next), serviceName)
}

// TracingUnaryServerInterceptor is a gRPC unary interceptor that adds tracing to the request.
func TracingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return otelgrpc.UnaryServerInterceptor()
}

// TracingStreamServerInterceptor is the streaming equivalent of TracingUnaryServerInterceptor
func TracingStreamServerInterceptor() grpc.StreamServerInterceptor {
	return otelgrpc.StreamServerInterceptor()
}

// TracingMiddleware is HTTP middleware that adds tracing to the request.
func TracingMiddleware(next http.Handler, serviceName string) http.Handler {
	return otelhttp.NewHandler(next, serviceName)
}

// RecovererUnaryServerInterceptor is a gRPC unary interceptor that recovers from panics and returns them as internal errors.
func RecovererUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor()
}

// RecovererStreamServerInterceptor is the streaming equivalent of RecovererUnaryServerInterceptor
func RecovererStreamServerInterceptor() grpc.StreamServerInterceptor {
	return recovery.StreamServerInterceptor()
}

// NOTE: Recoverer for HTTP is part of LoggingMiddleware

// LoggingUnaryServerInterceptor is a gRPC unary interceptor that logs requests.
func LoggingUnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(
		grpczap.InterceptorLogger(logger),
		logging.WithDecider(logFinishDecider),
		logging.WithCodes(errorToCode),
		logging.WithLevels(grpcCodeToLevel),
	)
}

// LoggingStreamServerInterceptor is the streaming equivalent of LoggingUnaryServerInterceptor
func LoggingStreamServerInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return logging.StreamServerInterceptor(
		grpczap.InterceptorLogger(logger),
		logging.WithDecider(logFinishDecider),
		logging.WithCodes(errorToCode),
		logging.WithLevels(grpcCodeToLevel),
	)
}

// logFinishDecider filters which calls to log. It logs all calls (start and finish).
func logFinishDecider(fullMethodName string, err error) logging.Decision {
	return logging.LogStartAndFinishCall
}

// errorToCode maps an error to a gRPC code for logging. It wraps the default behavior and adds handling of context errors.
func errorToCode(err error) codes.Code {
	if errors.Is(err, context.DeadlineExceeded) {
		return codes.DeadlineExceeded
	}
	if errors.Is(err, context.Canceled) {
		return codes.Canceled
	}
	return logging.DefaultErrorToCode(err)
}

// grpcCodeToLevel overrides the log level of various gRPC codes.
// We're currently not doing very granular error handling, so we get quite a lot of codes.Unknown errors, which we do not want to emit as error logs.
func grpcCodeToLevel(code codes.Code) logging.Level {
	switch code {
	case codes.OK, codes.NotFound, codes.Canceled, codes.AlreadyExists, codes.InvalidArgument, codes.Unauthenticated,
		codes.Unknown, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition, codes.OutOfRange:
		return logging.INFO
	case codes.Unimplemented, codes.DeadlineExceeded, codes.Aborted, codes.Unavailable:
		return logging.WARNING
	case codes.Internal, codes.DataLoss:
		return logging.ERROR
	default:
		return logging.ERROR
	}
}

// LoggingMiddleware is a HTTP request logging middleware.
// Note: It also recovers from panics and handles them as internal errors.
func LoggingMiddleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		fields := []zap.Field{
			zap.String("method", r.Method),
			zap.String("proto", r.Proto),
			zap.String("path", r.URL.EscapedPath()),
			zap.String("ip", ip),
			zap.String("user_agent", r.UserAgent()),
			ZapCtx(r.Context()),
		}

		start := time.Now()
		wrapped := wrappedResponseWriter{ResponseWriter: w}

		defer func() {
			// Recover panics and handle as internal errors
			if err := recover(); err != nil {
				// Write status
				w.WriteHeader(http.StatusInternalServerError)
				wrapped.status = http.StatusInternalServerError
				_, _ = w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

				// Add error field
				switch v := err.(type) {
				case error:
					fields = append(fields, zap.Error(v))
				default:
					fields = append(fields, zap.Any("error", v))
				}
			}

			// Get status
			status := wrapped.status
			if status == 0 {
				status = 200
			}

			// Print finish message
			fields = append(fields,
				zap.Int("status", status),
				zap.Duration("duration", time.Since(start)),
			)
			logger.Info("http request finished", fields...)
		}()

		// Print start message
		logger.Info("http request started", fields...)

		next.ServeHTTP(&wrapped, r)
	})
}

// wrappedResponseWriter wraps a response writer and tracks the response status code
type wrappedResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *wrappedResponseWriter) Status() int {
	return rw.status
}

func (rw *wrappedResponseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}
