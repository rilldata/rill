package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/vanguard/vanguardgrpc"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/pkg/securetoken"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrForbidden = status.Error(codes.Unauthenticated, "action not allowed")

type Options struct {
	HTTPPort        int
	GRPCPort        int
	AllowedOrigins  []string
	ServePrometheus bool
	SessionKeyPairs [][]byte
	AuthEnable      bool
	AuthIssuerURL   string
	AuthAudienceURL string
	TLSCertPath     string
	TLSKeyPath      string
}

type Server struct {
	runtimev1.UnsafeRuntimeServiceServer
	runtimev1.UnsafeQueryServiceServer
	runtimev1.UnsafeConnectorServiceServer
	runtime  *runtime.Runtime
	opts     *Options
	logger   *zap.Logger
	aud      *auth.Audience
	codec    *securetoken.Codec
	limiter  ratelimit.Limiter
	activity *activity.Client
	ai       *ai.Runner
}

var (
	_ runtimev1.RuntimeServiceServer   = (*Server)(nil)
	_ runtimev1.QueryServiceServer     = (*Server)(nil)
	_ runtimev1.ConnectorServiceServer = (*Server)(nil)
)

// NewServer creates a new runtime server.
// The provided ctx is used for the lifetime of the server for background refresh of the JWKS that is used to validate auth tokens.
func NewServer(ctx context.Context, opts *Options, rt *runtime.Runtime, logger *zap.Logger, limiter ratelimit.Limiter, activityClient *activity.Client) (*Server, error) {
	// The runtime doesn't actually set cookies, but we use securecookie to encode/decode ephemeral tokens.
	// If no session key pairs are provided, we generate a random one for the duration of the process.
	var codec *securetoken.Codec
	if len(opts.SessionKeyPairs) == 0 {
		codec = securetoken.NewRandom()
	} else {
		codec = securetoken.NewCodec(opts.SessionKeyPairs)
	}

	srv := &Server{
		runtime:  rt,
		opts:     opts,
		logger:   logger,
		codec:    codec,
		limiter:  limiter,
		activity: activityClient,
		ai:       ai.NewRunner(rt, activityClient),
	}

	if opts.AuthEnable {
		aud, err := auth.OpenAudience(ctx, logger, opts.AuthIssuerURL, opts.AuthAudienceURL)
		if err != nil {
			return nil, err
		}
		srv.aud = aud
	}

	return srv, nil
}

// Close should be called when the server is done
func (s *Server) Close() error {
	// TODO: This should probably trigger a server shutdown

	if s.aud != nil {
		s.aud.Close()
	}

	return nil
}

// Ping implements RuntimeService
func (s *Server) Ping(ctx context.Context, req *runtimev1.PingRequest) (*runtimev1.PingResponse, error) {
	resp := &runtimev1.PingResponse{
		Version: s.runtime.Version().String(),
		Time:    timestamppb.New(time.Now()),
	}
	return resp, nil
}

// Starts the HTTP server.
func (s *Server) ServeHTTP(ctx context.Context, registerAdditionalHandlers func(mux *http.ServeMux), local bool) error {
	handler, err := s.HTTPHandler(ctx, registerAdditionalHandlers, local)
	if err != nil {
		return err
	}

	return graceful.ServeHTTP(ctx, handler, graceful.ServeOptions{
		Port:     s.opts.HTTPPort,
		GRPCPort: s.opts.GRPCPort,
		CertPath: s.opts.TLSCertPath,
		KeyPath:  s.opts.TLSKeyPath,
		Logger:   s.logger,
	})
}

// HTTPHandler returns a HTTP handler that serves REST and gRPC.
func (s *Server) HTTPHandler(ctx context.Context, registerAdditionalHandlers func(mux *http.ServeMux), local bool) (http.Handler, error) {
	httpMux := http.NewServeMux()

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			middleware.TimeoutStreamServerInterceptor(timeoutSelector),
			observability.LoggingStreamServerInterceptor(s.logger),
			grpc_validator.StreamServerInterceptor(),
			auth.StreamServerInterceptor(s.aud),
			middleware.ActivityStreamServerInterceptor(s.activity),
			errorMappingStreamServerInterceptor(),
			grpc_auth.StreamServerInterceptor(s.checkRateLimit),
		),
		grpc.ChainUnaryInterceptor(
			middleware.TimeoutUnaryServerInterceptor(timeoutSelector),
			observability.LoggingUnaryServerInterceptor(s.logger),
			grpc_validator.UnaryServerInterceptor(),
			auth.UnaryServerInterceptor(s.aud),
			middleware.ActivityUnaryServerInterceptor(s.activity),
			errorMappingUnaryServerInterceptor(),
			grpc_auth.UnaryServerInterceptor(s.checkRateLimit),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	runtimev1.RegisterRuntimeServiceServer(grpcServer, s)
	runtimev1.RegisterQueryServiceServer(grpcServer, s)
	runtimev1.RegisterConnectorServiceServer(grpcServer, s)

	// Add gRPC and gRPC-to-REST transcoder.
	// This will be the fallback for REST routes like `/v1/ping` and GPRC routes like `/rill.admin.v1.RuntimeService/Ping`.
	transcoder, err := vanguardgrpc.NewTranscoder(grpcServer)
	if err != nil {
		return nil, fmt.Errorf("failed to create transcoder: %w", err)
	}
	httpMux.Handle("/v1/", transcoder)
	httpMux.Handle("/rill.runtime.v1.RuntimeService/", transcoder)
	httpMux.Handle("/rill.runtime.v1.QueryService/", transcoder)
	httpMux.Handle("/rill.runtime.v1.ConnectorService/", transcoder)

	// Call callback to register additional paths
	// NOTE: This is so ugly, but not worth refactoring it properly right now.
	if registerAdditionalHandlers != nil {
		registerAdditionalHandlers(httpMux)
	}

	// Add HTTP handler for health check
	observability.MuxHandle(httpMux, "/v1/health", observability.Middleware("runtime", s.logger, httputil.Handler(s.healthCheckHandler)))

	// Add HTTP handler for query export downloads
	observability.MuxHandle(httpMux, "/v1/download", observability.Middleware("runtime", s.logger, auth.HTTPMiddleware(s.aud, http.HandlerFunc(s.downloadHandler))))

	// Add handler for dynamic APIs, i.e. APIs backed by resolvers (such as custom APIs defined in YAML).
	observability.MuxHandle(httpMux, "/v1/instances/{instance_id}/api/{name...}", observability.Middleware("runtime", s.logger, auth.HTTPMiddleware(s.aud, httputil.Handler(s.apiHandler))))

	// Add handler for combined OpenAPI spec of custom APIs
	observability.MuxHandle(httpMux, "/v1/instances/{instance_id}/api/openapi", observability.Middleware("runtime", s.logger, auth.HTTPMiddleware(s.aud, httputil.Handler(s.combinedOpenAPISpec))))

	// Serving static assets
	observability.MuxHandle(httpMux, "/v1/instances/{instance_id}/assets/{path...}", observability.Middleware("runtime", s.logger, auth.HTTPMiddleware(s.aud, httputil.Handler(s.assetsHandler))))

	// Add HTTP handler for multipart file upload
	observability.MuxHandle(httpMux, "/v1/instances/{instance_id}/files/upload/-/{path...}", observability.Middleware("runtime", s.logger, auth.HTTPMiddleware(s.aud, http.HandlerFunc(s.UploadMultipartFile))))

	// We need to manually add HTTP handlers for streaming RPCs since Vanguard can't map these to HTTP routes automatically.
	httpMux.Handle("/v1/instances/{instance_id}/files/watch", auth.HTTPMiddleware(s.aud, http.HandlerFunc(s.WatchFilesHandler)))
	httpMux.Handle("/v1/instances/{instance_id}/resources/-/watch", auth.HTTPMiddleware(s.aud, http.HandlerFunc(s.WatchResourcesHandler)))
	httpMux.Handle("/v1/instances/{instance_id}/ai/complete/stream", auth.HTTPMiddleware(s.aud, http.HandlerFunc(s.CompleteStreamingHandler)))

	// Add Prometheus
	if s.opts.ServePrometheus {
		httpMux.Handle("/metrics", promhttp.Handler())
	}

	// Adds the MCP server handlers.
	// The path without an instance ID is a convenience path intended for Rill Developer (localhost). In this case, the implementation falls back to using the default instance ID.
	mcpHandler := observability.Middleware("runtime", s.logger, auth.HTTPMiddleware(s.aud, s.mcpHandler()))
	observability.MuxHandle(httpMux, "/mcp", mcpHandler)                                    // Routes to the default instance ID (for Rill Developer on localhost)
	observability.MuxHandle(httpMux, "/v1/instances/{instance_id}/mcp", mcpHandler)         // The MCP handler will extract the instance ID from the request path.
	observability.MuxHandle(httpMux, "/mcp/sse", mcpHandler)                                // Backwards compatibility
	observability.MuxHandle(httpMux, "/mcp/message", mcpHandler)                            // Backwards compatibility
	observability.MuxHandle(httpMux, "/v1/instances/{instance_id}/mcp/sse", mcpHandler)     // Backwards compatibility
	observability.MuxHandle(httpMux, "/v1/instances/{instance_id}/mcp/message", mcpHandler) // Backwards compatibility

	// Build CORS options for runtime server

	// If the AllowedOrigins contains a "*" we want to return the requester's origin instead of "*" in the "Access-Control-Allow-Origin" header.
	// This is useful in development. In production, we set AllowedOrigins to non-wildcard values, so this does not have security implications.
	// Details: https://github.com/rs/cors#allow--with-credentials-security-protection
	var allowedOriginFunc func(string) bool
	allowedOrigins := s.opts.AllowedOrigins
	for _, origin := range s.opts.AllowedOrigins {
		if origin == "*" {
			allowedOriginFunc = func(origin string) bool { return true }
			allowedOrigins = nil
			break
		}
	}

	corsOpts := cors.Options{
		AllowedOrigins:  allowedOrigins,
		AllowOriginFunc: allowedOriginFunc,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		// Set max age to 1 hour (default if not set is 5 seconds)
		MaxAge: 60 * 60,
	}

	// Wrap mux with CORS middleware
	handler := cors.New(corsOpts).Handler(httpMux)

	return handler, nil
}

// HTTPErrorHandler wraps gateway.DefaultHTTPErrorHandler to map gRPC unknown errors (i.e. errors without an explicit
// code) to HTTP status code 400 instead of 500.
func HTTPErrorHandler(ctx context.Context, mux *gateway.ServeMux, marshaler gateway.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	s := status.Convert(err)
	if s.Code() == codes.Unknown {
		err = &gateway.HTTPStatusError{HTTPStatus: http.StatusBadRequest, Err: err}
	}
	gateway.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}

func timeoutSelector(fullMethodName string) time.Duration {
	if strings.HasPrefix(fullMethodName, "/rill.runtime.v1.RuntimeService") && (strings.Contains(fullMethodName, "/Trigger") || strings.HasSuffix(fullMethodName, "Reconcile")) {
		return time.Minute * 59 // Not 60 to avoid forced timeout on ingress
	}

	if strings.HasPrefix(fullMethodName, "/rill.runtime.v1.QueryService") ||
		strings.HasPrefix(fullMethodName, "/rill.runtime.v1.ConnectorService") {
		return time.Minute * 5
	}

	if fullMethodName == runtimev1.RuntimeService_WatchFiles_FullMethodName {
		return time.Minute * 30
	}

	if fullMethodName == runtimev1.RuntimeService_WatchResources_FullMethodName {
		return time.Minute * 30
	}

	if fullMethodName == runtimev1.RuntimeService_WatchLogs_FullMethodName {
		return time.Minute * 30
	}

	if fullMethodName == runtimev1.RuntimeService_Complete_FullMethodName || fullMethodName == runtimev1.RuntimeService_CompleteStreaming_FullMethodName {
		return time.Minute * 5
	}

	if fullMethodName == runtimev1.RuntimeService_Health_FullMethodName || fullMethodName == runtimev1.RuntimeService_InstanceHealth_FullMethodName {
		return time.Minute * 3 // Match the default interactive query timeout
	}

	return time.Second * 30
}

// errorMappingUnaryServerInterceptor is an interceptor that applies mapGRPCError.
func errorMappingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		return resp, mapGRPCError(err)
	}
}

// errorMappingUnaryServerInterceptor is an interceptor that applies mapGRPCError.
func errorMappingStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		return mapGRPCError(err)
	}
}

// mapGRPCError rewrites errors returned from gRPC handlers before they are returned to the client.
func mapGRPCError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, err.Error())
	}
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, err.Error())
	}
	if errors.Is(err, queries.ErrForbidden) {
		return ErrForbidden
	}
	if errors.Is(err, runtime.ErrForbidden) {
		return ErrForbidden
	}
	if errors.Is(err, metricsview.ErrForbidden) {
		return ErrForbidden
	}
	return err
}

func (s *Server) checkRateLimit(ctx context.Context) (context.Context, error) {
	// Any request type might be limited separately as it is part of Metadata
	// Any request type might be excluded from this limit check and limited later,
	// e.g. in the corresponding request handler by calling s.limiter.Limit(ctx, "limitKey", redis_rate.PerMinute(100))
	if auth.GetClaims(ctx, "").UserID == "" {
		method, ok := grpc.Method(ctx)
		if !ok {
			return ctx, fmt.Errorf("server context does not have a method")
		}
		limitKey := ratelimit.AnonLimitKey(method, observability.GrpcPeer(ctx))
		if err := s.limiter.Limit(ctx, limitKey, ratelimit.Public); err != nil {
			if errors.As(err, &ratelimit.QuotaExceededError{}) {
				return ctx, status.Error(codes.ResourceExhausted, err.Error())
			}
			return ctx, err
		}
	}

	return ctx, nil
}

func (s *Server) addInstanceRequestAttributes(ctx context.Context, instanceID string) {
	attrs := s.runtime.GetInstanceAttributes(ctx, instanceID)
	observability.AddRequestAttributes(ctx, attrs...)
}

func (s *Server) IssueDevJWT(ctx context.Context, req *runtimev1.IssueDevJWTRequest) (*runtimev1.IssueDevJWTResponse, error) {
	attr := map[string]any{
		"name":   req.Name,
		"email":  req.Email,
		"groups": req.Groups,
		"admin":  req.Admin,
	}

	for k, v := range req.Attributes.AsMap() {
		attr[k] = v
	}

	// If possible, add "domain" inferred from "email"
	email, ok := attr["email"].(string)
	if ok && attr["domain"] == nil {
		attr["domain"] = email[strings.LastIndex(email, "@")+1:]
	}

	jwt, err := auth.NewDevToken(attr, runtime.AllPermissions)
	if err != nil {
		return nil, err
	}

	return &runtimev1.IssueDevJWTResponse{
		Jwt: jwt,
	}, nil
}
