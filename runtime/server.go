package runtime

import (
	"context"
	"fmt"
	"net/http"

	metrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/providers/opentracing/v2"
	grpczerolog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rilldata/rill/runtime/pkg/graceful"
	proto "github.com/rilldata/rill/runtime/proto"
)

type ServerOptions struct {
	HTTPPort int
	GRPCPort int
}

type Server struct {
	proto.UnimplementedRuntimeServiceServer
	opts    *ServerOptions
	runtime *Runtime
	logger  zerolog.Logger
}

func NewServer(opts *ServerOptions, runtime *Runtime, logger zerolog.Logger) *Server {
	return &Server{
		opts:    opts,
		runtime: runtime,
		logger:  logger,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	group, cctx := errgroup.WithContext(ctx)

	// Start the gRPC server
	group.Go(func() error {
		server := grpc.NewServer(
			grpc.ChainStreamInterceptor(
				tracing.StreamServerInterceptor(opentracing.InterceptorTracer()),
				metrics.StreamServerInterceptor(metrics.NewServerMetrics()),
				logging.StreamServerInterceptor(grpczerolog.InterceptorLogger(s.logger)),
				recovery.StreamServerInterceptor(),
			),
			grpc.ChainUnaryInterceptor(
				tracing.UnaryServerInterceptor(opentracing.InterceptorTracer()),
				metrics.UnaryServerInterceptor(metrics.NewServerMetrics()),
				logging.UnaryServerInterceptor(grpczerolog.InterceptorLogger(s.logger)),
				recovery.UnaryServerInterceptor(),
			),
		)
		proto.RegisterRuntimeServiceServer(server, s)
		s.logger.Info().Int("port", s.opts.GRPCPort).Msg("Serving gRPC")
		return graceful.ServeGRPC(cctx, server, s.opts.GRPCPort)
	})

	// Start the HTTP gateway targetting the gRPC server
	group.Go(func() error {
		mux := gateway.NewServeMux()
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		grpcAddress := fmt.Sprintf(":%d", s.opts.GRPCPort)
		err := proto.RegisterRuntimeServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
		if err != nil {
			return err
		}
		server := &http.Server{Handler: mux}
		s.logger.Info().Int("port", s.opts.HTTPPort).Msg("Serving HTTP")
		return graceful.ServeHTTP(cctx, server, s.opts.HTTPPort)
	})

	return group.Wait()
}
