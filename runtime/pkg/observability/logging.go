package observability

import (
	"context"
	"errors"
	"net/http"
	"time"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/providers/zap/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func LoggingStreamServerInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return logging.StreamServerInterceptor(
		grpczap.InterceptorLogger(logger),
		logging.WithCodes(errorToCode),
		logging.WithLevels(grpcCodeToLevel),
	)
}

func LoggingUnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(
		grpczap.InterceptorLogger(logger),
		logging.WithCodes(errorToCode),
		logging.WithLevels(grpcCodeToLevel),
	)
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

func LoggingMiddleware(h gateway.HandlerFunc, logger *zap.Logger) gateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				switch v := err.(type) {
				case error:
					logger.Error("Processing failure",
						ZapCtx(r.Context()),
						zap.Error(v),
						zap.String("method", r.Method),
						zap.String("path", r.URL.EscapedPath()),
						zap.String("proto", r.Proto),
						zap.String("user-agent", r.UserAgent()),
					)
				default:
					logger.Error("Unknown processing failure",
						ZapCtx(r.Context()),
						zap.String("method", r.Method),
						zap.String("path", r.URL.EscapedPath()),
						zap.String("proto", r.Proto),
						zap.String("user-agent", r.UserAgent()),
					)
				}
			}
		}()

		start := time.Now()
		wrapped := wrapResponseWriter(w)
		h(wrapped, r, pathParams)
		logger.Info(
			"Success",
			ZapCtx(r.Context()),
			zap.Int("status", wrapped.status),
			zap.String("method", r.Method),
			zap.String("path", r.URL.EscapedPath()),
			zap.Duration("duration", time.Since(start)),
		)
	}
}

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func ZapCtx(ctx context.Context) zap.Field {
	span := trace.SpanFromContext(ctx)
	sctx := span.SpanContext()
	if !sctx.IsValid() {
		return zap.Skip()
	}

	return zap.Inline(zapSpan{
		traceID: sctx.TraceID().String(),
		spanID:  sctx.SpanID().String(),
	})
}

type zapSpan struct {
	traceID string
	spanID  string
}

func (z zapSpan) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("trace_id", z.traceID)
	enc.AddString("span_id", z.spanID)
	return nil
}
