package devtool

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/cli/cmd/admin"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func MigrateCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Used to validate and run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load .env (note: fails silently if .env has errors)
			_ = godotenv.Load()
			ctx := cmd.Context()

			// Init config
			var conf admin.Config
			err := envconfig.Process("rill_admin", &conf)
			if err != nil {
				fmt.Printf("failed to load config: %s\n", err.Error())
				os.Exit(1)
			}

			// Init logger
			cfg := zap.NewProductionConfig()
			cfg.Level.SetLevel(conf.LogLevel)
			cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			logger, err := cfg.Build()
			if err != nil {
				fmt.Printf("error: failed to create logger: %s\n", err.Error())
				os.Exit(1)
			}

			// Init db
			db, err := database.Open(conf.DatabaseDriver, conf.DatabaseURL, "")
			if err != nil {
				logger.Fatal("error connecting to database", zap.Error(err))
			}

			// Auto-run migrations
			v1, err := db.FindMigrationVersion(ctx)
			if err != nil {
				logger.Fatal("error getting migration version", zap.Error(err))
			}
			err = db.Migrate(ctx)
			if err != nil {
				logger.Fatal("error migrating database", zap.Error(err))
			}
			v2, err := db.FindMigrationVersion(ctx)
			if err != nil {
				logger.Fatal("error getting migration version", zap.Error(err))
			}
			if v1 == v2 {
				logger.Info("database is up to date", zap.Int("version", v2))
			} else {
				logger.Info("database migrated", zap.Int("from_version", v1), zap.Int("to_version", v2))
			}

			return nil
		},
	}
	return cmd
}
