package observability

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const TracingHeader = "x-trace-id"

// Middleware is HTTP middleware that combines all observability-related middlewares.
func Middleware(serviceName string, logger *zap.Logger, next http.Handler) http.Handler {
	return TracingMiddleware(LoggingMiddleware(logger, next), serviceName)
}

// TracingMiddleware is HTTP middleware that adds tracing to the request.
func TracingMiddleware(next http.Handler, serviceName string) http.Handler {
	return otelhttp.NewHandler(next, serviceName)
}

// LoggingUnaryServerInterceptor is a gRPC unary interceptor that logs requests.
// It also recovers from panics and returns them as internal errors.
// Also sends the trace ID in response headers to correlate with logs.
func LoggingUnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	logger = logger.WithOptions(zap.AddStacktrace(zapcore.InvalidLevel)) // Disable stacktraces for error logs in this interceptor
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// Log pings at debug level
		// TODO: Change when we move to standard gRPC health checks
		lvl := zap.InfoLevel
		if info.FullMethod == "/rill.admin.v1.AdminService/Ping" || info.FullMethod == "/rill.runtime.v1.RuntimeService/Ping" {
			lvl = zap.DebugLevel
		}

		fields := []zap.Field{
			zap.String("protocol", "grpc"),
			zap.String("peer.address", GrpcPeer(ctx)),
			zap.String("grpc.component", "server"),
			zap.String("grpc.method_type", "unary"),
			zap.String("grpc.method", info.FullMethod),
			ZapCtx(ctx),
		}

		start := time.Now()

		// Add datadog trace ID to response headers
		traceID := TraceID(ctx)
		if traceID != "" {
			header := metadata.Pairs(TracingHeader, traceID)
			_ = grpc.SetHeader(ctx, header)
		}

		defer func() {
			// Recover panics and handle as internal errors
			if rerr := recover(); rerr != nil {
				stack := make([]byte, 64<<10)
				stack = stack[:runtime.Stack(stack, false)]
				err = status.Errorf(codes.Internal, "panic caught: %v", rerr)
				// Not putting stack in err to prevent leaking to clients
				fields = append(fields, zap.String("stack", string(stack)))
			}

			// Get code and log level
			code := status.Code(err)
			if err != nil {
				lvl = grpcCodeToLevel(code)
			}

			// Format err for logging. If err is a gRPC error, we want to show only the description.
			logErr := err
			if logErr != nil {
				if s, ok := status.FromError(logErr); ok {
					logErr = errors.New(s.Message())
				}
			}

			// Log finish message
			fields = append(fields,
				zap.String("grpc.code", code.String()),
				zap.Duration("duration", time.Since(start)),
				zap.Error(logErr),
			)
			logger.Log(lvl, "grpc finished call", fields...)
		}()

		// Add log fields to context
		ctx = contextWithLogFields(ctx, &fields)

		logger.Log(lvl, "grpc started call", fields...)
		return handler(ctx, req)
	}
}

// LoggingStreamServerInterceptor is a gRPC streaming interceptor that logs requests.
// It also recovers from panics and returns them as internal errors.
// Also sends the trace ID in response headers to correlate with logs.
func LoggingStreamServerInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	logger = logger.WithOptions(zap.AddStacktrace(zapcore.InvalidLevel)) // Disable stacktraces for error logs in this interceptor
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		fields := []zap.Field{
			zap.String("protocol", "grpc"),
			zap.String("peer.address", GrpcPeer(ss.Context())),
			zap.String("grpc.component", "server"),
			zap.String("grpc.method_type", "server_stream"),
			zap.String("grpc.method", info.FullMethod),
			ZapCtx(ss.Context()),
		}

		start := time.Now()

		// Add datadog trace ID to response headers
		traceID := TraceID(ss.Context())
		if traceID != "" {
			header := metadata.Pairs(TracingHeader, traceID)
			_ = grpc.SetHeader(ss.Context(), header)
		}

		defer func() {
			// Recover panics and handle as internal errors
			if rerr := recover(); rerr != nil {
				stack := make([]byte, 64<<10)
				stack = stack[:runtime.Stack(stack, false)]
				err = status.Errorf(codes.Internal, "panic caught: %v", rerr)
				// Not putting stack in err to prevent leaking to clients
				fields = append(fields, zap.String("stack", string(stack)))
			}

			// Get code and log level
			code := status.Code(err)
			lvl := grpcCodeToLevel(code)

			// Format err for logging. If err is a gRPC error, we want to show only the description.
			logErr := err
			if logErr != nil {
				if s, ok := status.FromError(logErr); ok {
					logErr = errors.New(s.Message())
				}
			}

			// Log finish message
			fields = append(fields,
				zap.String("grpc.code", code.String()),
				zap.Duration("duration", time.Since(start)),
				zap.Error(logErr),
			)
			logger.Log(lvl, "grpc finished call", fields...)
		}()

		// Add log fields to context
		wss := grpc_middleware.WrapServerStream(ss)
		wss.WrappedContext = contextWithLogFields(ss.Context(), &fields)

		logger.Info("grpc started call", fields...)
		return handler(srv, wss)
	}
}

// grpcCodeToLevel overrides the log level of various gRPC codes.
// We're currently not doing very granular error handling, so we get quite a lot of codes.Unknown errors, which we do not want to emit as error logs.
func grpcCodeToLevel(code codes.Code) zapcore.Level {
	switch code {
	case codes.OK, codes.NotFound, codes.Canceled, codes.AlreadyExists, codes.InvalidArgument, codes.Unauthenticated,
		codes.Unknown, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition, codes.OutOfRange:
		return zap.InfoLevel
	case codes.Unimplemented, codes.DeadlineExceeded, codes.Aborted, codes.Unavailable:
		return zap.WarnLevel
	case codes.Internal, codes.DataLoss:
		return zap.ErrorLevel
	default:
		return zap.ErrorLevel
	}
}

// GrpcPeer returns the client address, using the "real" IP passed by the load balancer if available.
func GrpcPeer(ctx context.Context) string {
	var addr string

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		v := md.Get("x-forwarded-for")
		if len(v) > 0 {
			addr = v[0]
		}
	}

	if addr == "" {
		p, _ := peer.FromContext(ctx)
		addr = p.Addr.String()
	}

	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		ip = addr
	}

	return ip
}

// LoggingMiddleware is a HTTP request logging middleware.
// Note: It also recovers from panics and handles them as internal errors.
// Also sends the trace ID in response headers to correlate with logs.
func LoggingMiddleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := []zap.Field{
			zap.String("protocol", r.Proto),
			zap.String("peer.address", HTTPPeer(r)),
			zap.String("http.method", r.Method),
			zap.String("http.path", r.URL.EscapedPath()),
			zap.String("http.user_agent", r.UserAgent()),
			ZapCtx(r.Context()),
		}

		start := time.Now()

		// Set datadog trace ID header in response headers
		traceID := TraceID(r.Context())
		if traceID != "" {
			w.Header().Set(TracingHeader, traceID)
		}

		ww := httputil.WrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			// Recover panics and handle as internal errors
			if err := recover(); err != nil {
				// Write status
				ww.WriteHeader(http.StatusInternalServerError)
				_, _ = ww.Write([]byte(http.StatusText(http.StatusInternalServerError)))

				// Get stacktrace
				stack := make([]byte, 64<<10)
				stack = stack[:runtime.Stack(stack, false)]

				// Add log fieds
				err := fmt.Errorf("panic caught: %v", err)
				fields = append(fields, zap.Error(err), zap.String("stack", string(stack)))
			}

			// Get status
			httpStatus := ww.Status()
			if httpStatus == 0 {
				httpStatus = 200
			}

			// Print finish message
			fields = append(fields,
				zap.Int("http.status", httpStatus),
				zap.Duration("duration", time.Since(start)),
			)
			logger.Info("http request finished", fields...)
		}()

		// Add log fields to context
		r = r.WithContext(contextWithLogFields(r.Context(), &fields))

		// Print start message
		logger.Info("http request started", fields...)

		next.ServeHTTP(ww, r)
	})
}

// HTTPPeer returns the client address, using the "real" IP passed by the load balancer if available.
func HTTPPeer(r *http.Request) string {
	addr := r.Header.Get("x-forwarded-for")
	if addr == "" {
		addr = r.RemoteAddr
	}

	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		ip = addr
	}

	return ip
}

// logFieldsContextKey is used to set and get request log fields in the context.
type logFieldsContextKey struct{}

func contextWithLogFields(ctx context.Context, fields *[]zap.Field) context.Context {
	return context.WithValue(ctx, logFieldsContextKey{}, fields)
}

func logFieldsFromContext(ctx context.Context) *[]zap.Field {
	v, ok := ctx.Value(logFieldsContextKey{}).(*[]zap.Field)
	if !ok {
		return nil
	}
	return v
}

// AddRequestAttributes sets attributes on the current trace span, the finish log of the current request, and the activity track
func AddRequestAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	// Set attributes on the span
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)

	// Add attributes in request log fields
	fields := logFieldsFromContext(ctx)
	if fields != nil {
		for _, attr := range attrs {
			*fields = append(*fields, zap.Any(string(attr.Key), attr.Value.AsInterface()))
		}
	}

	// Add attributes for emitted events
	activity.SetAttributes(ctx, attrs...)
}

// MCPMiddleware is a middleware for MCP servers that adds MCP-related observability attributes.
// It is expected to run on an MCP server that has already been wrapped with observability.Middleware(...).
func MCPMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (result mcp.Result, err error) {
			// Add method, session, tool
			AddRequestAttributes(
				ctx,
				attribute.String("mcp.method", method),
				attribute.String("mcp.session.id", req.GetSession().ID()),
			)
			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				AddRequestAttributes(ctx, attribute.String("mcp.tool", ctr.Params.Name))
			}

			// Process and add error attributes
			res, err := next(ctx, method, req)
			if err != nil {
				AddRequestAttributes(ctx, attribute.String("mcp.error", err.Error()))
			} else if res != nil {
				ctr, ok := res.(*mcp.CallToolResult)
				if ok && ctr.IsError {
					var msg string
					if len(ctr.Content) > 0 {
						txt, ok := ctr.Content[0].(*mcp.TextContent)
						if ok {
							msg = txt.Text
						} else {
							msg = fmt.Sprintf("unknown error with content type %T", ctr.Content[0])
						}
					} else {
						msg = "unknown error with no content"
					}

					AddRequestAttributes(ctx, attribute.String("mcp.error", msg))
				}
			}
			return res, err
		}
	}
}
