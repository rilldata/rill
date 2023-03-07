package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	metrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/providers/opentracing/v2"
	grpczaplog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zap/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrForbidden = status.Error(codes.Unauthenticated, "action not allowed")

type Options struct {
	HTTPPort        int
	GRPCPort        int
	AllowedOrigins  []string
	AuthEnable      bool
	AuthIssuerURL   string
	AuthAudienceURL string
}

type Server struct {
	runtimev1.UnsafeRuntimeServiceServer
	runtimev1.UnsafeQueryServiceServer
	runtime *runtime.Runtime
	opts    *Options
	logger  *zap.Logger
	aud     *auth.Audience
}

var (
	_ runtimev1.RuntimeServiceServer = (*Server)(nil)
	_ runtimev1.QueryServiceServer   = (*Server)(nil)
)

func NewServer(opts *Options, rt *runtime.Runtime, logger *zap.Logger) (*Server, error) {
	srv := &Server{
		opts:    opts,
		runtime: rt,
		logger:  logger,
	}

	if opts.AuthEnable {
		aud, err := auth.OpenAudience(logger, opts.AuthIssuerURL, opts.AuthAudienceURL)
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

// ServeGRPC Starts the gRPC server.
func (s *Server) ServeGRPC(ctx context.Context) error {
	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			tracing.StreamServerInterceptor(opentracing.InterceptorTracer()),
			metrics.StreamServerInterceptor(metrics.NewServerMetrics()),
			logging.StreamServerInterceptor(grpczaplog.InterceptorLogger(s.logger), logging.WithCodes(ErrorToCode), logging.WithLevels(GRPCCodeToLevel)),
			recovery.StreamServerInterceptor(),
			grpc_validator.StreamServerInterceptor(),
			auth.StreamServerInterceptor(s.aud),
		),
		grpc.ChainUnaryInterceptor(
			tracing.UnaryServerInterceptor(opentracing.InterceptorTracer()),
			metrics.UnaryServerInterceptor(metrics.NewServerMetrics()),
			logging.UnaryServerInterceptor(grpczaplog.InterceptorLogger(s.logger), logging.WithCodes(ErrorToCode), logging.WithLevels(GRPCCodeToLevel)),
			recovery.UnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
			auth.UnaryServerInterceptor(s.aud),
		),
	)
	runtimev1.RegisterRuntimeServiceServer(server, s)
	runtimev1.RegisterQueryServiceServer(server, s)
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
	return graceful.ServeHTTP(ctx, server, s.opts.HTTPPort)
}

// ErrorToCode maps an error to a gRPC code for logging. It wraps the default behavior and adds handling of context errors.
func ErrorToCode(err error) codes.Code {
	if errors.Is(err, context.DeadlineExceeded) {
		return codes.DeadlineExceeded
	}
	if errors.Is(err, context.Canceled) {
		return codes.Canceled
	}
	return logging.DefaultErrorToCode(err)
}

// GRPCCodeToLevel overrides the log level of various gRPC codes.
// We're currently not doing very granular error handling, so we get quite a lot of codes.Unknown errors, which we do not want to emit as error logs.
func GRPCCodeToLevel(code codes.Code) logging.Level {
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

// HTTPHandler HTTP handler serving REST gateway.
func (s *Server) HTTPHandler(ctx context.Context, registerAdditionalHandlers func(mux *http.ServeMux)) (http.Handler, error) {
	// Create REST gateway
	mux := gateway.NewServeMux(gateway.WithErrorHandler(HTTPErrorHandler))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcAddress := fmt.Sprintf(":%d", s.opts.GRPCPort)
	err := runtimev1.RegisterRuntimeServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}

	err = runtimev1.RegisterQueryServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}

	// One-off REST-only path for multipart file upload
	err = mux.HandlePath("POST", "/v1/instances/{instance_id}/files/upload/-/{path=**}", auth.HTTPMiddleware(s.aud, s.UploadMultipartFile))
	if err != nil {
		panic(err)
	}

	// One-off REST-only path for file export
	err = mux.HandlePath("GET", "/v1/instances/{instance_id}/table/{table_name}/export/{format}", auth.HTTPMiddleware(s.aud, s.ExportTable))
	if err != nil {
		panic(err)
	}

	// Call callback to register additional paths
	// NOTE: This is so ugly, but not worth refactoring it properly right now.
	httpMux := http.NewServeMux()
	if registerAdditionalHandlers != nil {
		registerAdditionalHandlers(httpMux)
	}

	// Add gRPC-gateway mux on /v1
	httpMux.Handle("/v1/", mux)

	// Build CORS options for runtime server

	// If the AllowedOrigins contains a "*" we want to return the requester's origin instead of "*" in the "Access-Control-Allow-Origin" header.
	// This circumvents a special case where the browser will not follow redirects for wildcard allow origins. This impacts CORS API calls that get redirected.
	// Resolving "*" to the requester's origin is normally discouraged, but is useful in development on localhost (where different ports are considered different origins).
	// We should always set AllowedOrigins to non-wildcard values in production, so this should not have security implications.
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
		MaxAge:           60 * 60, // 1 hour (default if not set is 5 seconds)
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

// Ping implements RuntimeService
func (s *Server) Ping(ctx context.Context, req *runtimev1.PingRequest) (*runtimev1.PingResponse, error) {
	resp := &runtimev1.PingResponse{
		Version: "", // TODO: Return version
		Time:    timestamppb.New(time.Now()),
	}
	return resp, nil
}
