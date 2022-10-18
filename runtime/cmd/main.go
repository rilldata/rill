package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/druid"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/server"
	_ "github.com/rilldata/rill/runtime/sql"
)

type Config struct {
	Env             string        `default:"development"`
	HTTPPort        int           `default:"8080" split_words:"true"`
	GRPCPort        int           `default:"9090" split_words:"true"`
	LogLevel        zapcore.Level `default:"info" split_words:"true"`
	DatabaseDriver  string        `default:"sqlite"`
	DatabaseURL     string        `default:":memory:" split_words:"true"`
	LocalMode       bool          `default:"false" split_words:"true"`
	OLAPDatabaseUrl string        `default:"stage.db" split_words:"true"`
}

func main() {
	// Load .env (note: fails silently if .env has errors)
	godotenv.Load()

	// Init config
	var conf Config
	err := envconfig.Process("rill_runtime", &conf)
	if err != nil {
		fmt.Printf("failed to load config: %s", err.Error())
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
		fmt.Printf("error: failed to create logger: %s", err.Error())
		os.Exit(1)
	}

	// Open metadata db connection
	metastore, err := drivers.Open(conf.DatabaseDriver, conf.DatabaseURL)
	if err != nil {
		logger.Fatal("error: could not connect to metadata db", zap.Error(err))
	}
	err = metastore.Migrate(context.Background())
	if err != nil {
		logger.Fatal("error: metadata db migration", zap.Error(err))
	}

	// Init server
	opts := &server.ServerOptions{
		HTTPPort:            conf.HTTPPort,
		GRPCPort:            conf.GRPCPort,
		ConnectionCacheSize: 100,
	}
	s, err := server.NewServer(opts, metastore, logger)
	if err != nil {
		logger.Fatal("error: could not create server", zap.Error(err))
	}

	if conf.LocalMode {
		err = server.CreateLocalInstance(s, "duckdb", conf.OLAPDatabaseUrl)
		if err != nil {
			logger.Error("failed to create default instance", zap.Error(err))
		}
	}

	// Run server
	ctx := graceful.WithCancelOnTerminate(context.Background())
	err = s.Serve(ctx)
	if err != nil {
		logger.Error("server crashed", zap.Error(err))
	}

	logger.Info("server shutdown gracefully")
}
