package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	metrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/providers/opentracing/v2"
	grpczaplog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zap/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/graceful"
)

type ServerOptions struct {
	HTTPPort            int
	GRPCPort            int
	ConnectionCacheSize int
}

type Server struct {
	api.UnsafeRuntimeServiceServer
	opts         *ServerOptions
	metastore    drivers.Connection
	logger       *zap.Logger
	connCache    *connectionCache
	serviceCache *servicesCache
}

var _ api.RuntimeServiceServer = (*Server)(nil)

func NewServer(opts *ServerOptions, metastore drivers.Connection, logger *zap.Logger) (*Server, error) {
	_, ok := metastore.RegistryStore()
	if !ok {
		return nil, fmt.Errorf("server metastore must be a valid registry")
	}

	return &Server{
		opts:         opts,
		metastore:    metastore,
		logger:       logger,
		connCache:    newConnectionCache(opts.ConnectionCacheSize),
		serviceCache: newServicesCache(),
	}, nil
}

// Starts the gRPC server
func (s *Server) ServeGRPC(ctx context.Context) error {
	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			tracing.StreamServerInterceptor(opentracing.InterceptorTracer()),
			metrics.StreamServerInterceptor(metrics.NewServerMetrics()),
			logging.StreamServerInterceptor(grpczaplog.InterceptorLogger(s.logger)),
			recovery.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			tracing.UnaryServerInterceptor(opentracing.InterceptorTracer()),
			metrics.UnaryServerInterceptor(metrics.NewServerMetrics()),
			logging.UnaryServerInterceptor(grpczaplog.InterceptorLogger(s.logger)),
			recovery.UnaryServerInterceptor(),
		),
	)
	api.RegisterRuntimeServiceServer(server, s)
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

// HTTP handler serving REST gateway
func (s *Server) HTTPHandler(ctx context.Context) (http.Handler, error) {
	// Create REST gateway
	mux := gateway.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcAddress := fmt.Sprintf(":%d", s.opts.GRPCPort)
	err := api.RegisterRuntimeServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}

	// One-off REST-only path for multipart file upload
	mux.HandlePath(
		"POST",
		"/v1/repos/{repo_id}/objects/file/-/{path=**}",
		s.UploadMultipartFile,
	)

	// Register CORS
	handler := cors(mux)

	return handler, nil
}

// Metrics APIs
func (s *Server) EstimateRollupInterval(ctx context.Context, req *api.EstimateRollupIntervalRequest) (*api.EstimateRollupIntervalResponse, error) {
	return &api.EstimateRollupIntervalResponse{}, nil
}

// Ping implements RuntimeService
func (s *Server) Ping(ctx context.Context, req *api.PingRequest) (*api.PingResponse, error) {
	resp := &api.PingResponse{
		Version: runtime.Version,
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
