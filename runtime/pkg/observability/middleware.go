package observability

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
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
		zapInterceptorLogger(logger),
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
		logging.WithFieldsFromContext(tracingFieldsFromCtx),
		logging.WithCodes(errorToCode),
		logging.WithLevels(grpcCodeToLevel),
	)
}

// LoggingStreamServerInterceptor is the streaming equivalent of LoggingUnaryServerInterceptor
func LoggingStreamServerInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return logging.StreamServerInterceptor(
		zapInterceptorLogger(logger),
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
		logging.WithFieldsFromContext(tracingFieldsFromCtx),
		logging.WithCodes(errorToCode),
		logging.WithLevels(grpcCodeToLevel),
	)
}

// zapInterceptorLogger adapts zap logger to a gRPC interceptor logger.
// Source: https://github.com/grpc-ecosystem/go-grpc-middleware/blob/main/interceptors/logging/examples/zap/example_test.go
func zapInterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)
		for i := 0; i < len(fields); i += 2 {
			i := logging.Fields(fields).Iterator()
			if i.Next() {
				k, v := i.At()
				f = append(f, zap.Any(k, v))
			}
		}
		l = l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			l.Debug(msg)
		case logging.LevelInfo:
			l.Info(msg)
		case logging.LevelWarn:
			l.Warn(msg)
		case logging.LevelError:
			l.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

// tracingFieldsFromCtx picks tracing-related fields from the ctx for use in the gRPC logging interceptor.
func tracingFieldsFromCtx(ctx context.Context) logging.Fields {
	sctx := trace.SpanFromContext(ctx).SpanContext()
	if !sctx.IsValid() {
		return nil
	}
	return []any{
		"trace_id", sctx.TraceID().String(),
		"span_id", sctx.SpanID().String(),
	}
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
		return logging.LevelInfo
	case codes.Unimplemented, codes.DeadlineExceeded, codes.Aborted, codes.Unavailable:
		return logging.LevelWarn
	case codes.Internal, codes.DataLoss:
		return logging.LevelError
	default:
		return logging.LevelError
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
