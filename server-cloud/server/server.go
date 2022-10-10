package server

import (
	"context"

	// "encoding/gob"
	"fmt"
	"net/http"

	// oapimiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mikespook/gorbac"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/labstack/echo-contrib/session"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/server-cloud/api"
	"github.com/rilldata/rill/server-cloud/database"
	"github.com/rilldata/rill/server-cloud/ent"
)

type Server struct {
	logger     *zap.Logger
	db         database.DB
	conf       Config
	auth       *Authenticator
	client     *ent.Client
	authorizer *authorizer
}
type Config struct {
	Port             int
	AuthDomain       string
	AuthClientID     string
	AuthClientSecret string
	AuthCallbackURL  string
	SessionsSecret   string
}

func New(logger *zap.Logger, db database.DB, conf Config, client *ent.Client) (*Server, error) {
	auth, err := newAuthenticator(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	users, _ := LoadUsers(context.Background(), client)
	dbRoles, _ := LoadRoles(context.Background(), client)

	rbac := gorbac.New()
	permissions := make(gorbac.Permissions)

	roles := dbRoles.(map[string][]string)
	// Build roles and add them to goRBAC instance
	for rid, pids := range roles {
		role := gorbac.NewStdRole(rid)
		for _, pid := range pids {
			_, ok := permissions[pid]
			if !ok {
				permissions[pid] = gorbac.NewStdPermission(pid)
			}
			role.Assign(permissions[pid])
		}
		rbac.Add(role)
	}

	authorizer := &authorizer{users: users, rbac: rbac, permissions: permissions}

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		logger.Fatal("failed creating schema resources: %v", zap.Error(err))
	}

	return &Server{
		logger:     logger,
		db:         db,
		conf:       conf,
		auth:       auth,
		client:     client,
		authorizer: authorizer,
	}, nil
}

func (s *Server) Serve(ctx context.Context, port int) error {
	e := echo.New()

	// Request logging middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogLatency:      true,
		LogProtocol:     true,
		LogRemoteIP:     true,
		LogMethod:       true,
		LogURI:          true,
		LogRoutePath:    true,
		LogUserAgent:    true,
		LogStatus:       true,
		LogError:        true,
		LogResponseSize: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			s.logger.Info("request",
				zap.String("ip", v.RemoteIP),
				zap.String("protocol", v.Protocol),
				zap.Int("status", v.Status),
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.String("route", v.RoutePath),
				zap.Error(v.Error),
				zap.String("elapsed", v.Latency.String()),
				zap.String("user_agent", v.UserAgent),
				zap.Int64("response_size", v.ResponseSize),
			)
			return nil
		},
	}))

	// Recover middleware that uses zap
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					if r == http.ErrAbortHandler {
						panic(r)
					}
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					s.logger.Error("request panic", zap.Error(err), zap.Stack("stacktrace"))

					c.Error(err)
				}
			}()
			return next(c)
		}
	})

	// CORS (TODO: configure approriately)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{}))

	// Prometheus middleware
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)

	store := sessions.NewCookieStore([]byte(s.conf.SessionsSecret))
	e.Use(session.Middleware(store))

	e.GET("/auth/login", s.authLogin)
	e.GET("/auth/callback", s.callback)
	e.GET("/auth/logout", s.logout)
	e.GET("/auth/logout/callback", s.logoutCallback)
	e.GET("/auth/user", s.user, IsAuthenticated)

	// Register OpenAPI handlers
	// spec, err := api.GetSwagger()
	// if err != nil {
	// 	return err
	// }
	// e.Use(oapimiddleware.OapiRequestValidator(spec))

	// Adding basic Authentication check for every other routes in openAPI
	// api.RegisterHandlers(e.Group("", IsAuthenticated), s)
	api.RegisterHandlers(e.Group("", s.IsAuthorized), s)
	// api.RegisterHandlers(e, s)

	// Start serer
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: h2c.NewHandler(e, &http2.Server{}),
	}
	return graceful.ServeHTTP(ctx, srv, port)
}
