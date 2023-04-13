package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	metrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/providers/opentracing/v2"
	grpczaplog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zap/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-version"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const cliVersionConstraint = ">= 0.20.0"

type Options struct {
	HTTPPort               int
	GRPCPort               int
	ExternalURL            string
	FrontendURL            string
	SessionKeyPairs        [][]byte
	AllowedOrigins         []string
	AuthDomain             string
	AuthClientID           string
	AuthClientSecret       string
	GithubAppName          string
	GithubAppWebhookSecret string
	GithubClientID         string
	GithubClientSecret     string
}

type Server struct {
	adminv1.UnsafeAdminServiceServer
	logger        *zap.Logger
	admin         *admin.Service
	opts          *Options
	cookies       *sessions.CookieStore
	authenticator *auth.Authenticator
	issuer        *runtimeauth.Issuer
	urls          *externalURLs
}

var _ adminv1.AdminServiceServer = (*Server)(nil)

func New(opts *Options, logger *zap.Logger, adm *admin.Service, issuer *runtimeauth.Issuer) (*Server, error) {
	externalURL, err := url.Parse(opts.ExternalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse external URL: %w", err)
	}

	if len(opts.SessionKeyPairs) == 0 {
		return nil, fmt.Errorf("provided SessionKeyPairs is empty")
	}

	cookies := sessions.NewCookieStore(opts.SessionKeyPairs...)
	cookies.Options.MaxAge = 60 * 60 * 24 * 365 * 10 // 10 years
	cookies.Options.Secure = externalURL.Scheme == "https"
	cookies.Options.HttpOnly = true

	authenticator, err := auth.NewAuthenticator(logger, adm, cookies, &auth.AuthenticatorOptions{
		AuthDomain:       opts.AuthDomain,
		AuthClientID:     opts.AuthClientID,
		AuthClientSecret: opts.AuthClientSecret,
		ExternalURL:      opts.ExternalURL,
		FrontendURL:      opts.FrontendURL,
	})
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:        logger,
		admin:         adm,
		opts:          opts,
		cookies:       cookies,
		authenticator: authenticator,
		issuer:        issuer,
		urls:          newURLRegistry(opts),
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
			s.authenticator.StreamServerInterceptor(),
			grpc_auth.StreamServerInterceptor(CheckUserAgent),
		),
		grpc.ChainUnaryInterceptor(
			tracing.UnaryServerInterceptor(opentracing.InterceptorTracer()),
			metrics.UnaryServerInterceptor(metrics.NewServerMetrics()),
			logging.UnaryServerInterceptor(grpczaplog.InterceptorLogger(s.logger), logging.WithCodes(ErrorToCode), logging.WithLevels(GRPCCodeToLevel)),
			recovery.UnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
			s.authenticator.UnaryServerInterceptor(),
			grpc_auth.UnaryServerInterceptor(CheckUserAgent),
		),
	)

	adminv1.RegisterAdminServiceServer(server, s)
	s.logger.Sugar().Infof("serving admin gRPC on port:%v", s.opts.GRPCPort)
	return graceful.ServeGRPC(ctx, server, s.opts.GRPCPort)
}

// Starts the HTTP server.
func (s *Server) ServeHTTP(ctx context.Context) error {
	handler, err := s.HTTPHandler(ctx)
	if err != nil {
		return err
	}

	server := &http.Server{Handler: handler}
	s.logger.Sugar().Infof("serving admin HTTP on port:%v", s.opts.HTTPPort)
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
func (s *Server) HTTPHandler(ctx context.Context) (http.Handler, error) {
	// Create REST gateway
	mux := gateway.NewServeMux(
		gateway.WithErrorHandler(HTTPErrorHandler),
		gateway.WithMetadata(s.authenticator.Annotator),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcAddress := fmt.Sprintf(":%d", s.opts.GRPCPort)
	err := adminv1.RegisterAdminServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}

	// Add auth endpoints (not gRPC handlers, just regular endpoints on /auth/*)
	err = s.authenticator.RegisterEndpoints(mux)
	if err != nil {
		return nil, err
	}

	// Add Github-related endpoints (not gRPC handlers, just regular endpoints on /github/*)
	err = s.registerGithubEndpoints(mux)
	if err != nil {
		return nil, err
	}

	// Server public JWKS for runtime JWT verification
	err = mux.HandlePath("GET", "/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		s.issuer.WellKnownHandleFunc(w, r)
	})
	if err != nil {
		return nil, err
	}

	// Build CORS options for admin server

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
		AllowedHeaders: []string{"*"},
		// We use cookies for browser sessions, so this is required to allow ui.rilldata.com to make authenticated requests to admin.rilldata.com
		AllowCredentials: true,
		// Set max age to 1 hour (default if not set is 5 seconds)
		MaxAge: 60 * 60,
	}

	// Wrap mux with CORS middleware
	handler := cors.New(corsOpts).Handler(mux)

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

func CheckUserAgent(ctx context.Context) (context.Context, error) {
	userAgent := strings.Split(metautils.ExtractIncoming(ctx).Get("user-agent"), " ")
	var ver string
	for _, s := range userAgent {
		if strings.HasPrefix(s, "rill-cli/") {
			ver = strings.TrimPrefix(s, "rill-cli/")
		}
	}

	// Check if build from source
	if ver == "unknown" || ver == "" {
		return ctx, nil
	}

	v1, err := version.NewVersion(ver)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("could not parse rill-cli version: %s", err.Error()))
	}

	constraints, err := version.NewConstraint(cliVersionConstraint)
	if err != nil {
		panic(err)
	}

	if !constraints.Check(v1) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("Rill %s is no longer supported, please upgrade to the latest version", v1))
	}

	return ctx, nil
}

type externalURLs struct {
	githubConnect         string
	githubConnectRetry    string
	githubConnectRequest  string
	githubConnectSuccess  string
	githubAppInstallation string
	githubAuth            string
	githubAuthCallback    string
	githubAuthSuccess     string
	githubAuthRetry       string
	authLogin             string
}

func newURLRegistry(opts *Options) *externalURLs {
	return &externalURLs{
		githubConnect:         mustJoinURL(opts.ExternalURL, "/github/connect"),
		githubConnectRetry:    mustJoinURL(opts.FrontendURL, "/-/github/connect/retry"),
		githubConnectRequest:  mustJoinURL(opts.FrontendURL, "/-/github/connect/request"),
		githubConnectSuccess:  mustJoinURL(opts.FrontendURL, "/-/github/connect/success"),
		githubAppInstallation: fmt.Sprintf("https://github.com/apps/%s/installations/new", opts.GithubAppName),
		githubAuth:            mustJoinURL(opts.ExternalURL, "/github/auth/login"),
		githubAuthCallback:    mustJoinURL(opts.ExternalURL, "/github/auth/callback"),
		githubAuthSuccess:     mustJoinURL(opts.FrontendURL, "/-/github/auth/success"),
		githubAuthRetry:       mustJoinURL(opts.FrontendURL, "/-/github/auth/retry"),
		authLogin:             mustJoinURL(opts.ExternalURL, "/auth/login"),
	}
}

func urlWithQuery(urlString string, query map[string]string) (string, error) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	qry := parsedURL.Query()
	for key, value := range query {
		qry.Set(key, value)
	}
	parsedURL.RawQuery = qry.Encode()
	return parsedURL.String(), nil
}

func mustJoinURL(base string, elem ...string) string {
	joinedURL, err := url.JoinPath(base, elem...)
	if err != nil {
		panic(err)
	}
	return joinedURL
}
