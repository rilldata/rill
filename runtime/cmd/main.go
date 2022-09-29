package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rilldata/rill/runtime"
	_ "github.com/rilldata/rill/runtime/infra/druid"
	_ "github.com/rilldata/rill/runtime/infra/duckdb"
	"github.com/rilldata/rill/runtime/metadata"
	_ "github.com/rilldata/rill/runtime/metadata/postgres"
	_ "github.com/rilldata/rill/runtime/metadata/sqlite"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	_ "github.com/rilldata/rill/runtime/sql"
)

type Config struct {
	Env            string        `default:"development"`
	LogLevel       zapcore.Level `default:"info" split_words:"true"`
	DatabaseDriver string        `default:"sqlite"`
	DatabaseURL    string        `default:":memory:" split_words:"true"`
	GRPCPort       int           `default:"9090" split_words:"true"`
	HTTPPort       int           `default:"8080" split_words:"true"`
}

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Failed to load .env: %s", err.Error())
		os.Exit(1)
	}

	// Init config
	var conf Config
	err = envconfig.Process("rill_runtime", &conf)
	if err != nil {
		fmt.Printf("Failed to load config: %s", err.Error())
		os.Exit(1)
	}

	// Init logger
	var logger *zap.Logger
	if conf.Env == "production" {
		logger, err = zap.NewProduction(zap.IncreaseLevel(conf.LogLevel))
	} else {
		logger, err = zap.NewDevelopment(zap.IncreaseLevel(conf.LogLevel))
	}
	if err != nil {
		fmt.Printf("Error creating logger: %s", err.Error())
		os.Exit(1)
	}

	// Init db
	db, err := metadata.Open(conf.DatabaseDriver, conf.DatabaseURL)
	if err != nil {
		logger.Fatal("error connecting to database", zap.Error(err))
	}

	// Auto-run migrations
	err = db.Migrate(context.Background())
	if err != nil {
		logger.Fatal("error migrating database", zap.Error(err))
	}

	// Init runtime
	rt := runtime.New(db, logger)

	// Init server
	opts := &runtime.ServerOptions{
		GRPCPort: conf.GRPCPort,
		HTTPPort: conf.HTTPPort,
	}
	server := runtime.NewServer(opts, rt, logger)

	// Run server
	ctx := graceful.WithCancelOnTerminate(context.Background())
	err = server.Serve(ctx)
	if err != nil {
		logger.Error("server crashed", zap.Error(err))
	}

	logger.Info("server shutdown gracefully")
}
