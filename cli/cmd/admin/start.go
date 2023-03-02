package admin

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/server"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	// Load database drivers for admin
	_ "github.com/rilldata/rill/admin/database/postgres"
)

// Config describes admin server config derived from environment variables.
// Env var keys must be prefixed with RILL_ADMIN_ and are converted from snake_case to CamelCase.
// For example RILL_ADMIN_HTTP_PORT is mapped to Config.HTTPPort.
type Config struct {
	DatabaseDriver          string        `default:"postgres" split_words:"true"`
	DatabaseURL             string        `split_words:"true"`
	HTTPPort                int           `default:"8080" split_words:"true"`
	GRPCPort                int           `default:"9090" split_words:"true"`
	LogLevel                zapcore.Level `default:"info" split_words:"true"`
	SessionSecret           string        `split_words:"true"`
	AuthDomain              string        `split_words:"true"`
	AuthClientID            string        `split_words:"true"`
	AuthClientSecret        string        `split_words:"true"`
	AuthCallbackURL         string        `split_words:"true"`
	GithubAppSecret         string        `split_words:"true"`
	GithubAppID             int64         `split_words:"true"`
	GithubAppPrivateKeyPath string        `split_words:"true"`
	GithubAppName           string        `default:"test-rill-webhooks" split_words:"true"`
	UIHost                  string        `split_words:"true"`
}

// StartCmd starts an admin server. It only allows configuration using environment variables.
func StartCmd(cliCfg *config.Config) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start admin server",
		Run: func(cmd *cobra.Command, args []string) {
			// Load .env (note: fails silently if .env has errors)
			_ = godotenv.Load()

			// Init config
			var conf Config
			err := envconfig.Process("rill_admin", &conf)
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

			// Init admin service
			admOpts := &admin.Options{
				DatabaseDriver: conf.DatabaseDriver,
				DatabaseDSN:    conf.DatabaseURL,
			}
			adm, err := admin.New(admOpts, logger)
			if err != nil {
				logger.Fatal("error creating service", zap.Error(err))
			}
			defer adm.Close()

			// Init admin server
			srvConf := server.Config{
				HTTPPort:                conf.HTTPPort,
				GRPCPort:                conf.GRPCPort,
				AuthDomain:              conf.AuthDomain,
				AuthClientID:            conf.AuthClientID,
				AuthClientSecret:        conf.AuthClientSecret,
				AuthCallbackURL:         conf.AuthCallbackURL,
				SessionSecret:           conf.SessionSecret,
				GithubAPISecretKey:      []byte(conf.GithubAppSecret),
				GithubAppID:             conf.GithubAppID,
				GithubAppPrivateKeyPath: conf.GithubAppPrivateKeyPath,
				GithubAppName:           conf.GithubAppName,
				UIHost:                  conf.UIHost,
			}
			srv, err := server.New(logger, adm, srvConf)
			if err != nil {
				logger.Fatal("error creating server", zap.Error(err))
			}

			// Run server
			ctx := graceful.WithCancelOnTerminate(context.Background())
			group, cctx := errgroup.WithContext(ctx)
			group.Go(func() error { return srv.ServeGRPC(cctx) })
			group.Go(func() error { return srv.ServeHTTP(cctx) })
			err = group.Wait()
			if err != nil {
				logger.Fatal("server crashed", zap.Error(err))
			}

			logger.Info("server shutdown gracefully")
		},
	}
	return startCmd
}
