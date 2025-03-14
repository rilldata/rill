package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
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
		"/rill.admin.v1.AdminService/UpdateProject":          version.Must(version.NewVersion("0.28.0")),
		"/rill.admin.v1.AdminService/UpdateOrganization":     version.Must(version.NewVersion("0.28.0")),
		"/rill.admin.v1.AdminService/UpdateProjectVariables": version.Must(version.NewVersion("0.51.0")),
	}
)

type Options struct {
	HTTPPort               int
	GRPCPort               int
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
	// AssetsBucket is the path on gcs where rill managed project artifacts are stored.
	AssetsBucket string
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
	limiter       ratelimit.Limiter
	activity      *activity.Client
}

var _ adminv1.AdminServiceServer = (*Server)(nil)

var _ adminv1.AIServiceServer = (*Server)(nil)

var _ adminv1.TelemetryServiceServer = (*Server)(nil)

func New(logger *zap.Logger, adm *admin.Service, issuer *runtimeauth.Issuer, limiter ratelimit.Limiter, activityClient *activity.Client, opts *Options) (*Server, error) {
	if len(opts.SessionKeyPairs) == 0 {
		return nil, fmt.Errorf("provided SessionKeyPairs is empty")
	}

	cookieStore := cookies.New(logger, opts.SessionKeyPairs...)

	// Auth tokens are validated against the DB on each request, so we can set a long MaxAge.
	cookieStore.MaxAge(60 * 60 * 24 * 365 * 10) // 10 years

	// Set Secure if the admin service is served over HTTPS (will resolve to true in production and false in local dev environments).
	cookieStore.Options.Secure = adm.URLs.IsHTTPS()

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
		gateway.WithOutgoingHeaderMatcher(func(s string) (string, bool) {
			// grpc gateway adds gateway.MetadataHeaderPrefix to all outgoing headers
			// we want to skip that for `x-trace-id` set in response
			if s == observability.TracingHeader {
				return s, true
			}
			// default matcher logic
			return fmt.Sprintf("%s%s", gateway.MetadataHeaderPrefix, s), true
		}),
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

	// Add project assets endpoint.
	mux.Handle("/v1/assets/{asset_id}/download", observability.Middleware("assets", s.logger, s.authenticator.HTTPMiddleware(httputil.Handler(s.assetHandler))))

	// Add biller webhook handler if any
	if s.admin.Biller != nil {
		handlerFunc := s.admin.Biller.WebhookHandlerFunc(ctx, s.admin.Jobs)
		if handlerFunc != nil {
			inner := http.NewServeMux()
			observability.MuxHandle(inner, "/billing/webhook", handlerFunc)
			mux.Handle("/billing/webhook", observability.Middleware("admin", s.logger, inner))
		}
	}

	// Add payment webhook handler if any
	if s.admin.PaymentProvider != nil {
		handlerFunc := s.admin.PaymentProvider.WebhookHandlerFunc(ctx, s.admin.Jobs)
		if handlerFunc != nil {
			inner := http.NewServeMux()
			observability.MuxHandle(inner, "/payment/webhook", handlerFunc)
			mux.Handle("/payment/webhook", observability.Middleware("admin", s.logger, inner))
		}
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

// AwaitServing waits for both the HTTP and gRPC servers to be reachable on their localhost ports.
func (s *Server) AwaitServing(ctx context.Context) error {
	// Since the HTTP server proxies the ping endpoint to the gRPC server,
	// it is sufficient to check that endpoint on the HTTP server.
	client := &http.Client{}
	pingURL := fmt.Sprintf("http://localhost:%d/v1/ping", s.opts.HTTPPort)

	// Check every 100ms for 15s
	ticker := time.NewTicker(100 * time.Millisecond)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, pingURL, http.NoBody)
			if err != nil {
				return err
			}
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
				return nil
			}
		}
	}
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

	// If the connection to the cache server is lost, skip rate limiting.
	if err := s.limiter.Ping(ctx); err != nil {
		s.logger.Warn("skipping rate limiting due to cache connection error", zap.Error(err))
		return ctx, nil
	}

	// Don't rate limit superusers. This is useful for scripting.
	if auth.GetClaims(ctx).Superuser(ctx) {
		return ctx, nil
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
			return ctx, status.Error(codes.ResourceExhausted, err.Error())
		}
		return ctx, err
	}

	return ctx, nil
}

func (s *Server) jwtAttributesForUser(ctx context.Context, userID, orgID string, projectPermissions *adminv1.ProjectPermissions) (map[string]any, error) {
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	groups, err := s.admin.DB.FindUsergroupsForUser(ctx, user.ID, orgID)
	if err != nil {
		return nil, err
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
	switch fullMethodName {
	case
		"/rill.admin.v1.AdminService/CreateProject",
		"/rill.admin.v1.AdminService/UpdateProject",
		"/rill.admin.v1.AdminService/RedeployProject",
		"/rill.admin.v1.AdminService/TriggerRedeploy":
		return time.Minute * 5
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
	if _, ok := status.FromError(err); ok {
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, err.Error())
	}
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, err.Error())
	}
	if errors.Is(err, database.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, database.ErrNotUnique) {
		return status.Error(codes.AlreadyExists, err.Error())
	}
	if errors.Is(err, database.ErrValidation) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
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
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("Rill %s is no longer supported for this operation, run `rill upgrade` to upgrade to the latest version", v))
	}

	return ctx, nil
}
