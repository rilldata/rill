package runtime

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	// Load infra drivers and connectors for runtime
	_ "github.com/rilldata/rill/runtime/connectors/gcs"
	_ "github.com/rilldata/rill/runtime/connectors/https"
	_ "github.com/rilldata/rill/runtime/connectors/s3"
	_ "github.com/rilldata/rill/runtime/drivers/druid"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/github"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
)

// Config describes runtime server config derived from environment variables.
// Env var keys must be prefixed with RILL_RUNTIME_ and are converted from snake_case to CamelCase.
// For example RILL_RUNTIME_HTTP_PORT is mapped to Config.HTTPPort.
type Config struct {
	HTTPPort             int                    `default:"8080" split_words:"true"`
	GRPCPort             int                    `default:"9090" split_words:"true"`
	LogLevel             zapcore.Level          `default:"info" split_words:"true"`
	MetricsExporter      observability.Exporter `default:"prometheus" split_words:"true"`
	TracesExporter       observability.Exporter `default:"" split_words:"true"`
	MetastoreDriver      string                 `default:"sqlite"`
	MetastoreURL         string                 `default:"file:rill?mode=memory&cache=shared" split_words:"true"`
	AllowedOrigins       []string               `default:"*" split_words:"true"`
	AuthEnable           bool                   `default:"false" split_words:"true"`
	AuthIssuerURL        string                 `default:"" split_words:"true"`
	AuthAudienceURL      string                 `default:"" split_words:"true"`
	SafeSourceRefresh    bool                   `default:"false" split_words:"true"`
	ConnectionCacheSize  int                    `default:"100" split_words:"true"`
	QueryCacheSize       int                    `default:"10000" split_words:"true"`
	AllowHostCredentials bool                   `default:"false" split_words:"true"`
}

// StartCmd starts a stand-alone runtime server. It only allows configuration using environment variables.
func StartCmd(cliCfg *config.Config) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start stand-alone runtime server",
		Run: func(cmd *cobra.Command, args []string) {
			// Load .env (note: fails silently if .env has errors)
			_ = godotenv.Load()

			// Init config
			var conf Config
			err := envconfig.Process("rill_runtime", &conf)
			if err != nil {
				fmt.Printf("failed to load config: %s", err.Error())
				os.Exit(1)
			}

			// Init logger
			cfg := zap.NewProductionConfig()
			cfg.Level.SetLevel(conf.LogLevel)
			logger, err := cfg.Build()
			if err != nil {
				fmt.Printf("error: failed to create logger: %s", err.Error())
				os.Exit(1)
			}

			// Init telemetry
			shutdown, err := observability.Start(&observability.Options{
				MetricsExporter: conf.MetricsExporter,
				TracesExporter:  conf.TracesExporter,
				ServiceName:     "runtime-server",
				ServiceVersion:  cliCfg.Version.String(),
			})
			if err != nil {
				logger.Fatal("error starting telemetry", zap.Error(err))
			}
			defer func() {
				err := shutdown(context.Background())
				if err != nil {
					logger.Error("telemetry shutdown failed", zap.Error(err))
				}
			}()

			// Init runtime
			opts := &runtime.Options{
				ConnectionCacheSize:  conf.ConnectionCacheSize,
				MetastoreDriver:      conf.MetastoreDriver,
				MetastoreDSN:         conf.MetastoreURL,
				QueryCacheSize:       conf.QueryCacheSize,
				AllowHostCredentials: conf.AllowHostCredentials,
				SafeSourceRefresh:    conf.SafeSourceRefresh,
			}
			rt, err := runtime.New(opts, logger)
			if err != nil {
				logger.Fatal("error: could not create runtime", zap.Error(err))
			}
			defer rt.Close()

			// Init server
			srvOpts := &server.Options{
				HTTPPort:        conf.HTTPPort,
				GRPCPort:        conf.GRPCPort,
				AllowedOrigins:  conf.AllowedOrigins,
				AuthEnable:      conf.AuthEnable,
				AuthIssuerURL:   conf.AuthIssuerURL,
				AuthAudienceURL: conf.AuthAudienceURL,
			}
			s, err := server.NewServer(srvOpts, rt, logger)
			if err != nil {
				logger.Fatal("error: could not create server", zap.Error(err))
			}

			// Run server
			ctx := graceful.WithCancelOnTerminate(context.Background())
			group, cctx := errgroup.WithContext(ctx)
			group.Go(func() error { return s.ServeGRPC(cctx) })
			group.Go(func() error { return s.ServeHTTP(cctx, nil) })
			err = group.Wait()
			if err != nil {
				logger.Fatal("server crashed", zap.Error(err))
			}

			logger.Info("server shutdown gracefully")
		},
	}
	return startCmd
}
