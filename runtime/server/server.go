package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	grpczaplog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zap/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/server/metrics"
	"github.com/rs/cors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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
	logger  *otelzap.Logger
	aud     *auth.Audience
}

var (
	_ runtimev1.RuntimeServiceServer = (*Server)(nil)
	_ runtimev1.QueryServiceServer   = (*Server)(nil)
)

func NewServer(opts *Options, rt *runtime.Runtime, logger *otelzap.Logger) (*Server, error) {
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
	si, ui, err := metrics.InitCustomMetricsInterceptors(ctx)
	if err != nil {
		return err
	}

	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			si,
			otelgrpc.StreamServerInterceptor(),
			logging.StreamServerInterceptor(grpczaplog.InterceptorLogger(s.logger.Logger), logging.WithCodes(ErrorToCode), logging.WithLevels(GRPCCodeToLevel)),
			recovery.StreamServerInterceptor(),
			grpc_validator.StreamServerInterceptor(),
			auth.StreamServerInterceptor(s.aud),
		),
		grpc.ChainUnaryInterceptor(
			ui,
			otelgrpc.UnaryServerInterceptor(),
			logging.UnaryServerInterceptor(grpczaplog.InterceptorLogger(s.logger.Logger), logging.WithCodes(ErrorToCode), logging.WithLevels(GRPCCodeToLevel)),
			recovery.UnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
			auth.UnaryServerInterceptor(s.aud),
		),
	)
	runtimev1.RegisterRuntimeServiceServer(server, s)
	runtimev1.RegisterQueryServiceServer(server, s)
	s.logger.Ctx(ctx).Sugar().Infof("serving runtime gRPC on port:%v", s.opts.GRPCPort)
	return graceful.ServeGRPC(ctx, server, s.opts.GRPCPort)
}

// Starts the HTTP server.
func (s *Server) ServeHTTP(ctx context.Context, registerAdditionalHandlers func(mux *http.ServeMux)) error {
	handler, err := s.HTTPHandler(ctx, registerAdditionalHandlers)
	if err != nil {
		return err
	}

	server := &http.Server{Handler: handler}
	s.logger.Ctx(ctx).Sugar().Infof("serving HTTP on port:%v", s.opts.HTTPPort)
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

// Ping implements RuntimeService
func (s *Server) Ping(ctx context.Context, req *runtimev1.PingRequest) (*runtimev1.PingResponse, error) {
	resp := &runtimev1.PingResponse{
		Version: "", // TODO: Return version
		Time:    timestamppb.New(time.Now()),
	}
	return resp, nil
}

// Initializes providers and exporters.
// Global providers accumulate metrics/traces.
// The providers export data from Runtime using the exporters.
func InitOpenTelemetry(endpoint string) (*metric.MeterProvider, *sdktrace.TracerProvider, error) {
	if endpoint == "" {
		return nil, nil, nil
	}

	err := otelruntime.Start(otelruntime.WithMinimumReadMemStatsInterval(time.Minute))
	if err != nil {
		return nil, nil, err
	}

	exporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(exporter, metric.WithInterval(8*time.Second)),
		),
	)

	global.SetMeterProvider(meterProvider)

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()))
	traceExp, err := otlptrace.New(context.Background(), traceClient)
	if err != nil {
		return nil, nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	return meterProvider, tracerProvider, nil
}

func OtelHandler(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "otel-instrumented")
}
