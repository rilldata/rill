package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/pkg/securetoken"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/server")

var ErrForbidden = status.Error(codes.Unauthenticated, "action not allowed")

type Options struct {
	HTTPPort         int
	GRPCPort         int
	AllowedOrigins   []string
	ServePrometheus  bool
	SessionKeyPairs  [][]byte
	AuthEnable       bool
	AuthIssuerURL    string
	AuthAudienceURL  string
	DownloadRowLimit *int64
	TLSCertPath      string
	TLSKeyPath       string
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
	activity activity.Client
}

var (
	_ runtimev1.RuntimeServiceServer   = (*Server)(nil)
	_ runtimev1.QueryServiceServer     = (*Server)(nil)
	_ runtimev1.ConnectorServiceServer = (*Server)(nil)
)

// NewServer creates a new runtime server.
// The provided ctx is used for the lifetime of the server for background refresh of the JWKS that is used to validate auth tokens.
func NewServer(ctx context.Context, opts *Options, rt *runtime.Runtime, logger *zap.Logger, limiter ratelimit.Limiter, activityClient activity.Client) (*Server, error) {
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

	err := s.activity.Close()

	return err
}

// Ping implements RuntimeService
func (s *Server) Ping(ctx context.Context, req *runtimev1.PingRequest) (*runtimev1.PingResponse, error) {
	resp := &runtimev1.PingResponse{
		Version: "", // TODO: Return version
		Time:    timestamppb.New(time.Now()),
	}
	return resp, nil
}

// ServeGRPC Starts the gRPC server.
func (s *Server) ServeGRPC(ctx context.Context) error {
	server := grpc.NewServer(
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

	runtimev1.RegisterRuntimeServiceServer(server, s)
	runtimev1.RegisterQueryServiceServer(server, s)
	runtimev1.RegisterConnectorServiceServer(server, s)
	s.logger.Sugar().Infof("serving runtime gRPC on port:%v", s.opts.GRPCPort)
	return graceful.ServeGRPC(ctx, server, s.opts.GRPCPort)
}

// Starts the HTTP server.
func (s *Server) ServeHTTP(ctx context.Context, registerAdditionalHandlers func(mux *http.ServeMux)) error {
	handler, err := s.HTTPHandler(ctx, registerAdditionalHandlers)
	if err != nil {
		return err
	}

	server := &http.Server{Handler: handler}
	s.logger.Sugar().Infof("serving HTTP on port:%v", s.opts.HTTPPort)
	options := graceful.ServeOptions{
		Port:     s.opts.HTTPPort,
		CertPath: s.opts.TLSCertPath,
		KeyPath:  s.opts.TLSKeyPath,
	}
	return graceful.ServeHTTP(ctx, server, options)
}

// HTTPHandler HTTP handler serving REST gateway.
func (s *Server) HTTPHandler(ctx context.Context, registerAdditionalHandlers func(mux *http.ServeMux)) (http.Handler, error) {
	// Create REST gateway
	gwMux := gateway.NewServeMux(gateway.WithErrorHandler(HTTPErrorHandler))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcAddress := fmt.Sprintf("localhost:%d", s.opts.GRPCPort)
	err := runtimev1.RegisterRuntimeServiceHandlerFromEndpoint(ctx, gwMux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}
	err = runtimev1.RegisterQueryServiceHandlerFromEndpoint(ctx, gwMux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}
	err = runtimev1.RegisterConnectorServiceHandlerFromEndpoint(ctx, gwMux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}

	// One-off REST-only path for multipart file upload
	// NOTE: It's local only and we should deprecate it in favor of a cloud-friendly alternative.
	err = gwMux.HandlePath("POST", "/v1/instances/{instance_id}/files/upload/-/{path=**}", auth.GatewayMiddleware(s.aud, s.UploadMultipartFile))
	if err != nil {
		panic(err)
	}

	// Call callback to register additional paths
	// NOTE: This is so ugly, but not worth refactoring it properly right now.
	httpMux := http.NewServeMux()
	if registerAdditionalHandlers != nil {
		registerAdditionalHandlers(httpMux)
	}

	// Add gRPC-gateway on httpMux
	httpMux.Handle("/v1/", gwMux)

	// Add HTTP handler for query export downloads
	observability.MuxHandle(httpMux, "/v1/download", observability.Middleware("runtime", s.logger, auth.HTTPMiddleware(s.aud, http.HandlerFunc(s.downloadHandler))))

	// Add Prometheus
	if s.opts.ServePrometheus {
		httpMux.Handle("/metrics", promhttp.Handler())
	}

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

	if strings.HasPrefix(fullMethodName, "/rill.runtime.v1.QueryService") {
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
	return err
}

func (s *Server) checkRateLimit(ctx context.Context) (context.Context, error) {
	// Any request type might be limited separately as it is part of Metadata
	// Any request type might be excluded from this limit check and limited later,
	// e.g. in the corresponding request handler by calling s.limiter.Limit(ctx, "limitKey", redis_rate.PerMinute(100))
	if auth.GetClaims(ctx).Subject() == "" {
		method, ok := grpc.Method(ctx)
		if !ok {
			return ctx, fmt.Errorf("server context does not have a method")
		}
		limitKey := ratelimit.AnonLimitKey(method, observability.GrpcPeer(ctx))
		if err := s.limiter.Limit(ctx, limitKey, ratelimit.Public); err != nil {
			if errors.As(err, &ratelimit.QuotaExceededError{}) {
				return ctx, status.Errorf(codes.ResourceExhausted, err.Error())
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
		"domain": req.Email[strings.LastIndex(req.Email, "@")+1:],
		"groups": req.Groups,
		"admin":  req.Admin,
	}

	jwt, err := auth.NewDevToken(attr)
	if err != nil {
		return nil, err
	}
	return &runtimev1.IssueDevJWTResponse{
		Jwt: jwt,
	}, nil
}
