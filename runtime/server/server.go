package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	metrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/providers/opentracing/v2"
	grpczaplog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zap/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Options struct {
	HTTPPort int
	GRPCPort int
}

type Server struct {
	runtimev1.UnsafeRuntimeServiceServer
	runtime *runtime.Runtime
	opts    *Options
	logger  *zap.Logger
}

var _ runtimev1.RuntimeServiceServer = (*Server)(nil)

func NewServer(opts *Options, rt *runtime.Runtime, logger *zap.Logger) (*Server, error) {
	return &Server{
		opts:    opts,
		runtime: rt,
		logger:  logger,
	}, nil
}

// ServeGRPC Starts the gRPC server
func (s *Server) ServeGRPC(ctx context.Context) error {
	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			tracing.StreamServerInterceptor(opentracing.InterceptorTracer()),
			metrics.StreamServerInterceptor(metrics.NewServerMetrics()),
			logging.StreamServerInterceptor(grpczaplog.InterceptorLogger(s.logger), logging.WithCodes(ErrorToCode), logging.WithLevels(GRPCCodeToLevel)),
			recovery.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			tracing.UnaryServerInterceptor(opentracing.InterceptorTracer()),
			metrics.UnaryServerInterceptor(metrics.NewServerMetrics()),
			logging.UnaryServerInterceptor(grpczaplog.InterceptorLogger(s.logger), logging.WithCodes(ErrorToCode), logging.WithLevels(GRPCCodeToLevel)),
			recovery.UnaryServerInterceptor(),
		),
	)
	runtimev1.RegisterRuntimeServiceServer(server, s)
	s.logger.Sugar().Infof("serving runtime gRPC on port:%v", s.opts.GRPCPort)
	return graceful.ServeGRPC(ctx, server, s.opts.GRPCPort)
}

// Starts the HTTP server
func (s *Server) ServeHTTP(ctx context.Context) error {
	handler, err := s.HTTPHandler(ctx)
	if err != nil {
		return err
	}

	server := &http.Server{Handler: handler}
	s.logger.Sugar().Infof("serving HTTP on port:%v", s.opts.HTTPPort)
	return graceful.ServeHTTP(ctx, server, s.opts.HTTPPort)
}

// ErrorToCode returns the Code of the error if it is a Status error
// otherwise use status.FromContextError to determine the Code.
// Log level for error codes is defined in logging.DefaultServerCodeToLevel
func ErrorToCode(err error) codes.Code {
	if se, ok := err.(interface {
		GRPCStatus() *status.Status
	}); ok {
		return se.GRPCStatus().Code()
	}
	contextStatus := status.FromContextError(err)
	return contextStatus.Code()
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

// HTTPHandler HTTP handler serving REST gateway
func (s *Server) HTTPHandler(ctx context.Context) (http.Handler, error) {
	// Create REST gateway
	mux := gateway.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcAddress := fmt.Sprintf(":%d", s.opts.GRPCPort)
	err := runtimev1.RegisterRuntimeServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}

	// One-off REST-only path for multipart file upload
	err = mux.HandlePath("POST", "/v1/instances/{instance_id}/files/upload/-/{path=**}", s.UploadMultipartFile)
	if err != nil {
		return nil, err
	}

	// One-off REST-only path for file export
	err = mux.HandlePath("GET", "/v1/instances/{instance_id}/table/{table_name}/export/{format}", s.ExportTable)
	if err != nil {
		return nil, err
	}

	// Register CORS
	handler := cors(mux)

	return handler, nil
}

// Ping implements RuntimeService
func (s *Server) Ping(ctx context.Context, req *runtimev1.PingRequest) (*runtimev1.PingResponse, error) {
	resp := &runtimev1.PingResponse{
		Version: "", // TODO: Return version
		Time:    timestamppb.New(time.Now()),
	}
	return resp, nil
}

func cors(h http.Handler) http.Handler {
	// TODO: Hack for local - not production-ready
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Headers", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, PATCH, DELETE")
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
