package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/admin/server/cookies"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_minCliVersion         = version.Must(version.NewVersion("0.20.0"))
	_minCliVersionByMethod = map[string]*version.Version{
		"/rill.admin.v1.AdminService/UpdateProject":      version.Must(version.NewVersion("0.28.0")),
		"/rill.admin.v1.AdminService/UpdateOrganization": version.Must(version.NewVersion("0.28.0")),
	}
)

type Options struct {
	HTTPPort               int
	GRPCPort               int
	ExternalURL            string
	FrontendURL            string
	AllowedOrigins         []string
	SessionKeyPairs        [][]byte
	ServePrometheus        bool
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
	adminv1.UnsafeAIServiceServer
	adminv1.UnsafeTelemetryServiceServer
	logger        *zap.Logger
	admin         *admin.Service
	opts          *Options
	cookies       *cookies.Store
	authenticator *auth.Authenticator
	issuer        *runtimeauth.Issuer
	urls          *externalURLs
	limiter       ratelimit.Limiter
	activity      *activity.Client
}

var _ adminv1.AdminServiceServer = (*Server)(nil)

var _ adminv1.AIServiceServer = (*Server)(nil)

var _ adminv1.TelemetryServiceServer = (*Server)(nil)

func New(logger *zap.Logger, adm *admin.Service, issuer *runtimeauth.Issuer, limiter ratelimit.Limiter, activityClient *activity.Client, opts *Options) (*Server, error) {
	externalURL, err := url.Parse(opts.ExternalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse external URL: %w", err)
	}

	if len(opts.SessionKeyPairs) == 0 {
		return nil, fmt.Errorf("provided SessionKeyPairs is empty")
	}

	cookieStore := cookies.New(logger, opts.SessionKeyPairs...)

	// Auth tokens are validated against the DB on each request, so we can set a long MaxAge.
	cookieStore.MaxAge(60 * 60 * 24 * 365 * 10) // 10 years

	// Set Secure if the admin service is served over HTTPS (will resolve to true in production and false in local dev environments).
	cookieStore.Options.Secure = externalURL.Scheme == "https"

	// Only the admin server reads its cookies, so we can set HttpOnly (i.e. UI should not access cookie contents).
	cookieStore.Options.HttpOnly = true

	// Only the admin server reads its cookies, so we can set Domain to be the admin server's sub-domain (e.g. admin.rilldata.com).
	// That is automatically accomplished when Domain is not set.
	cookieStore.Options.Domain = ""

	// We need to protect against CSRF and clickjacking attacks, but still support requests from the UI to the admin service.
	// This is accomplished by setting SameSite=Lax (note that "site" just means the same root domain, not sub-domain).
	// For example, cookies will be passed on requests from ui.rilldata.com to admin.rilldata.com (or localhost:3000 to localhost:8080),
	// but not for requests from a different site AND NOT from an iframe of ui.rilldata.com on a different site.
	//
	// Note: We use Lax instead of Strict because we need cookies to be passed on redirects to the admin service from external providers, namely Auth0 and Github.
	//
	// Note on embedding: When embedding our UI, requests are only made to the runtime using the ephemeral JWT generated for the iframe. So we do not need cookies to be passed.
	// In the future, if iframes need to communicate with the admin service, we should introduce a scheme involving ephemeral tokens and not rely on cookies.
	cookieStore.Options.SameSite = http.SameSiteLaxMode

	authenticator, err := auth.NewAuthenticator(logger, adm, cookieStore, &auth.AuthenticatorOptions{
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
		cookies:       cookieStore,
		authenticator: authenticator,
		issuer:        issuer,
		urls:          newURLRegistry(opts),
		limiter:       limiter,
		activity:      activityClient,
	}, nil
}

// ServeGRPC Starts the gRPC server.
func (s *Server) ServeGRPC(ctx context.Context) error {
	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			middleware.TimeoutStreamServerInterceptor(timeoutSelector),
			observability.LoggingStreamServerInterceptor(s.logger),
			errorMappingStreamServerInterceptor(),
			grpc_auth.StreamServerInterceptor(checkUserAgent),
			grpc_validator.StreamServerInterceptor(),
			s.authenticator.StreamServerInterceptor(),
			grpc_auth.StreamServerInterceptor(s.checkRateLimit),
		),
		grpc.ChainUnaryInterceptor(
			middleware.TimeoutUnaryServerInterceptor(timeoutSelector),
			observability.LoggingUnaryServerInterceptor(s.logger),
			errorMappingUnaryServerInterceptor(),
			grpc_auth.UnaryServerInterceptor(checkUserAgent),
			grpc_validator.UnaryServerInterceptor(),
			s.authenticator.UnaryServerInterceptor(),
			grpc_auth.UnaryServerInterceptor(s.checkRateLimit),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	adminv1.RegisterAdminServiceServer(server, s)
	adminv1.RegisterAIServiceServer(server, s)
	adminv1.RegisterTelemetryServiceServer(server, s)
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

	return graceful.ServeHTTP(ctx, server, graceful.ServeOptions{
		Port: s.opts.HTTPPort,
	})
}

// HTTPHandler HTTP handler serving REST gateway.
func (s *Server) HTTPHandler(ctx context.Context) (http.Handler, error) {
	// Create REST gateway
	gwMux := gateway.NewServeMux(
		gateway.WithErrorHandler(httpErrorHandler),
		gateway.WithMetadata(s.authenticator.Annotator),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcAddress := fmt.Sprintf(":%d", s.opts.GRPCPort)
	err := adminv1.RegisterAdminServiceHandlerFromEndpoint(ctx, gwMux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}
	err = adminv1.RegisterAIServiceHandlerFromEndpoint(ctx, gwMux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}
	err = adminv1.RegisterTelemetryServiceHandlerFromEndpoint(ctx, gwMux, grpcAddress, opts)
	if err != nil {
		return nil, err
	}

	// Create regular http mux and mount gwMux on it
	mux := http.NewServeMux()
	mux.Handle("/v1/", gwMux)

	// Add runtime proxy
	observability.MuxHandle(mux, "/v1/orgs/{org}/projects/{project}/runtime/{path...}",
		observability.Middleware(
			"runtime-proxy",
			s.logger,
			s.authenticator.HTTPMiddlewareLenient(httputil.Handler(s.runtimeProxyForOrgAndProject)),
		),
	)

	// Temporary endpoint for testing headers.
	// TODO: Remove this.
	mux.HandleFunc("/v1/dump-headers", func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			fmt.Fprintf(w, "%s: %v\n", k, v)
		}
	})

	// Add Prometheus
	if s.opts.ServePrometheus {
		mux.Handle("/metrics", promhttp.Handler())
	}

	// Server public JWKS for runtime JWT verification
	mux.Handle("/.well-known/jwks.json", s.issuer.WellKnownHandler())

	// Add auth endpoints (not gRPC handlers, just regular endpoints on /auth/*)
	s.authenticator.RegisterEndpoints(mux, s.limiter)

	// Add Github-related endpoints (not gRPC handlers, just regular endpoints on /github/*)
	s.registerGithubEndpoints(mux)

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

// Ping implements AdminService
func (s *Server) Ping(ctx context.Context, req *adminv1.PingRequest) (*adminv1.PingResponse, error) {
	resp := &adminv1.PingResponse{
		Version: "", // TODO: Return version
		Time:    timestamppb.New(time.Now()),
	}
	return resp, nil
}

func (s *Server) checkRateLimit(ctx context.Context) (context.Context, error) {
	method, ok := grpc.Method(ctx)
	if !ok {
		return ctx, fmt.Errorf("server context does not have a method")
	}

	var limitKey string
	if auth.GetClaims(ctx).OwnerType() == auth.OwnerTypeAnon {
		limitKey = ratelimit.AnonLimitKey(method, observability.GrpcPeer(ctx))
	} else {
		limitKey = ratelimit.AuthLimitKey(method, auth.GetClaims(ctx).OwnerID())
	}

	limit := ratelimit.Default
	if strings.HasPrefix(method, "/rill.admin.v1.AIService") {
		limit = ratelimit.Sensitive
	}

	if err := s.limiter.Limit(ctx, limitKey, limit); err != nil {
		if errors.As(err, &ratelimit.QuotaExceededError{}) {
			return ctx, status.Errorf(codes.ResourceExhausted, err.Error())
		}
		return ctx, err
	}

	return ctx, nil
}

func (s *Server) jwtAttributesForUser(ctx context.Context, userID, orgID string, projectPermissions *adminv1.ProjectPermissions) (map[string]any, error) {
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	groups, err := s.admin.DB.FindUsergroupsForUser(ctx, user.ID, orgID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Using []any instead of []string since attr must be compatible with structpb.NewStruct
	groupNames := make([]any, len(groups))
	for i, group := range groups {
		groupNames[i] = group.Name
	}

	attr := map[string]any{
		"name":   user.DisplayName,
		"email":  user.Email,
		"domain": user.Email[strings.LastIndex(user.Email, "@")+1:],
		"groups": groupNames,
		"admin":  projectPermissions.ManageProject,
	}

	return attr, nil
}

// httpErrorHandler wraps gateway.DefaultHTTPErrorHandler to map gRPC unknown errors (i.e. errors without an explicit
// code) to HTTP status code 400 instead of 500.
func httpErrorHandler(ctx context.Context, mux *gateway.ServeMux, marshaler gateway.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	s := status.Convert(err)
	if s.Code() == codes.Unknown {
		err = &gateway.HTTPStatusError{HTTPStatus: http.StatusBadRequest, Err: err}
	}
	gateway.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}

func timeoutSelector(fullMethodName string) time.Duration {
	if strings.HasPrefix(fullMethodName, "/rill.admin.v1.AIService") {
		return time.Minute * 2
	}
	return time.Minute
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
	return err
}

// checkUserAgent is an interceptor that checks rejects from requests from old versions of the Rill CLI.
func checkUserAgent(ctx context.Context) (context.Context, error) {
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

	v, err := version.NewVersion(ver)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("could not parse rill-cli version: %s", err.Error()))
	}

	method, _ := grpc.Method(ctx)
	minVersion, ok := _minCliVersionByMethod[method]
	if !ok {
		minVersion = _minCliVersion
	}

	if v.LessThan(minVersion) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("Rill %s is no longer supported for given operation, please upgrade to the latest version", v))
	}

	return ctx, nil
}

type externalURLs struct {
	external              string
	frontend              string
	githubConnectUI       string
	githubConnect         string
	githubConnectRetry    string
	githubConnectRequest  string
	githubConnectSuccess  string
	githubAppInstallation string
	githubAuth            string
	githubAuthCallback    string
	githubAuthRetry       string
	authLogin             string
}

func newURLRegistry(opts *Options) *externalURLs {
	return &externalURLs{
		external:              opts.ExternalURL,
		frontend:              opts.FrontendURL,
		githubConnectUI:       urlutil.MustJoinURL(opts.FrontendURL, "/-/github/connect"),
		githubConnect:         urlutil.MustJoinURL(opts.ExternalURL, "/github/connect"),
		githubConnectRetry:    urlutil.MustJoinURL(opts.FrontendURL, "/-/github/connect/retry-install"),
		githubConnectRequest:  urlutil.MustJoinURL(opts.FrontendURL, "/-/github/connect/request"),
		githubConnectSuccess:  urlutil.MustJoinURL(opts.FrontendURL, "/-/github/connect/success"),
		githubAppInstallation: fmt.Sprintf("https://github.com/apps/%s/installations/new", opts.GithubAppName),
		githubAuth:            urlutil.MustJoinURL(opts.ExternalURL, "/github/auth/login"),
		githubAuthCallback:    urlutil.MustJoinURL(opts.ExternalURL, "/github/auth/callback"),
		githubAuthRetry:       urlutil.MustJoinURL(opts.FrontendURL, "/-/github/connect/retry-auth"),
		authLogin:             urlutil.MustJoinURL(opts.ExternalURL, "/auth/login"),
	}
}

func (u *externalURLs) reportOpen(org, project, report string, executionTime time.Time) string {
	reportURL := urlutil.MustJoinURL(u.frontend, org, project, "-", "reports", report, "open")
	reportURL += fmt.Sprintf("?execution_time=%s", executionTime.UTC().Format(time.RFC3339))
	return reportURL
}

func (u *externalURLs) reportExport(org, project, report string) string {
	return urlutil.MustJoinURL(u.frontend, org, project, "-", "reports", report, "export")
}

func (u *externalURLs) reportEdit(org, project, report string) string {
	return urlutil.MustJoinURL(u.frontend, org, project, "-", "reports", report)
}

func (u *externalURLs) alertOpen(org, project, alert string) string {
	return urlutil.MustJoinURL(u.frontend, org, project, "-", "alerts", alert)
}

func (u *externalURLs) alertEdit(org, project, alert string) string {
	return urlutil.MustJoinURL(u.frontend, org, project, "-", "alerts", alert)
}
