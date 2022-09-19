package server

import (
	"context"
	"fmt"
	"net/http"

	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/server-cloud/api"
	"github.com/rilldata/rill/server-cloud/database"
)

type Server struct {
	logger *zap.Logger
	db     database.DB
}

func New(logger *zap.Logger, db database.DB) *Server {
	return &Server{
		logger: logger,
		db:     db,
	}
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

	// Add other routes here...

	// Register OpenAPI handlers
	spec, err := api.GetSwagger()
	if err != nil {
		return err
	}
	e.Use(oapimiddleware.OapiRequestValidator(spec))
	api.RegisterHandlers(e, s)

	// Start serer
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: h2c.NewHandler(e, &http2.Server{}),
	}
	return graceful.ServeHTTP(ctx, srv, port)
}
