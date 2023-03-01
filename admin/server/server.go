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
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/eventhandler"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// hi
type Server struct {
	adminv1.UnsafeAdminServiceServer
	logger  *zap.Logger
	db      database.DB
	conf    Config
	auth    *Authenticator
	handler eventhandler.Handler
}

var _ adminv1.AdminServiceServer = (*Server)(nil)

type Config struct {
	HTTPPort         int
	GRPCPort         int
	AuthDomain       string
	AuthClientID     string
	AuthClientSecret string
	AuthCallbackURL  string
	SessionSecret    string
	GithubSecretKey  string // used to validate github events
}

func New(logger *zap.Logger, db database.DB, conf Config) (*Server, error) {
	auth, err := newAuthenticator(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	handler, err := eventhandler.NewGithubHandler(db)
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:  logger,
		db:      db,
		conf:    conf,
		auth:    auth,
		handler: handler,
	}, nil
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
		),
		grpc.ChainUnaryInterceptor(
			tracing.UnaryServerInterceptor(opentracing.InterceptorTracer()),
			metrics.UnaryServerInterceptor(metrics.NewServerMetrics()),
			logging.UnaryServerInterceptor(grpczaplog.InterceptorLogger(s.logger), logging.WithCodes(ErrorToCode), logging.WithLevels(GRPCCodeToLevel)),
			recovery.UnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
		),
	)

	adminv1.RegisterAdminServiceServer(server, s)
	s.logger.Sugar().Infof("serving admin gRPC on port:%v", s.conf.GRPCPort)
	return graceful.ServeGRPC(ctx, server, s.conf.GRPCPort)
}

// Starts the HTTP server.
func (s *Server) ServeHTTP(ctx context.Context) error {
	handler, err := s.HTTPHandler(ctx)
	if err != nil {
		return err
	}

	server := &http.Server{Handler: handler}
	s.logger.Sugar().Infof("serving admin HTTP on port:%v", s.conf.HTTPPort)
	return graceful.ServeHTTP(ctx, server, s.conf.HTTPPort)
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
func (s *Server) HTTPHandler(ctx context.Context) (http.Handler, error) {
	// Create REST gateway
	mux := gateway.NewServeMux(gateway.WithErrorHandler(HTTPErrorHandler))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcAddress := fmt.Sprintf(":%d", s.conf.GRPCPort)
	err := adminv1.RegisterAdminServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}

	// One-off REST-only path for multipart file upload
	err = mux.HandlePath("GET", "/auth/login", s.authLogin)
	if err != nil {
		panic(err)
	}

	// this MAY be a common API for all events originating from multiple sources like github,gitlab etc
	err = mux.HandlePath("POST", "/event_handler/{origin}", s.handleEvent)
	if err != nil {
		panic(err)
	}

	err = mux.HandlePath("GET", "/auth/callback", s.callback)
	if err != nil {
		panic(err)
	}

	err = mux.HandlePath("GET", "/auth/user", s.user)
	if err != nil {
		panic(err)
	}

	err = mux.HandlePath("GET", "/auth/logout", s.logout)
	if err != nil {
		panic(err)
	}

	err = mux.HandlePath("GET", "/auth/logout/callback", s.logoutCallback)
	if err != nil {
		panic(err)
	}

	// Register CORS
	handler := cors(mux)

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

// Ping implements AdminService
func (s *Server) Ping(ctx context.Context, req *adminv1.PingRequest) (*adminv1.PingResponse, error) {
	resp := &adminv1.PingResponse{
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
