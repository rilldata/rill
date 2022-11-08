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
	"golang.org/x/sync/errgroup"
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
	opts      *ServerOptions
	metastore drivers.Connection
	logger    *zap.Logger
	cache     *connectionCache
	os        drivers.OLAPStore // todo should be a pool, see current usage
}

var _ api.RuntimeServiceServer = (*Server)(nil)

func NewServer(opts *ServerOptions, metastore drivers.Connection, logger *zap.Logger) (*Server, error) {
	_, ok := metastore.RegistryStore()
	if !ok {
		return nil, fmt.Errorf("server metastore must be a valid registry")
	}

	return &Server{
		opts:      opts,
		metastore: metastore,
		logger:    logger,
		cache:     newConnectionCache(opts.ConnectionCacheSize),
	}, nil
}

func (s *Server) ServeNoWait(ctx context.Context) (*errgroup.Group, error) {
	group, cctx := errgroup.WithContext(ctx)

	// Start the gRPC server
	group.Go(func() error {
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
		s.logger.Info("serving gRPC", zap.Int("port", s.opts.GRPCPort))
		return graceful.ServeGRPC(cctx, server, s.opts.GRPCPort)
	})

	// Start the HTTP gateway targetting the gRPC server
	group.Go(func() error {
		mux := gateway.NewServeMux()
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		grpcAddress := fmt.Sprintf(":%d", s.opts.GRPCPort)
		err := api.RegisterRuntimeServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
		if err != nil {
			return err
		}
		mux.HandlePath(
			"POST",
			"/v1/repos/{repo_id}/objects/file/-/{path=**}",
			s.PutRepoObjectFromHTTPRequest,
		)
		handler := cors(mux)
		server := &http.Server{Handler: handler}
		s.logger.Info("serving HTTP", zap.Int("port", s.opts.HTTPPort))
		return graceful.ServeHTTP(cctx, server, s.opts.HTTPPort)
	})
	return group, nil
}

// Serve starts a gRPC server and a gRPC REST gateway server
func (s *Server) Serve(ctx context.Context) error {
	group, _ := s.ServeNoWait(ctx)
	return group.Wait()
}

func (s *Server) Cardinality(ctx context.Context, req *api.CardinalityRequest) (*api.CardinalityResponse, error) {
	bb := s.os == nil
	s.logger.Info("nil " + fmt.Sprintf("%v", bb))

	rows, err := s.os.Execute(ctx, &drivers.Statement{
		Query: "select count(*) from " + req.TableName,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int64
	for rows.Next() {
		// note that city can be NULL, so we use the NullString type
		err := rows.Scan(&count)
		if err != nil {
			return nil, err
		}
	}
	return &api.CardinalityResponse{
		Cardinality: count,
	}, nil
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
