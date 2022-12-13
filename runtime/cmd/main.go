package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rilldata/rill/runtime"
	_ "github.com/rilldata/rill/runtime/connectors/gcs"
	_ "github.com/rilldata/rill/runtime/connectors/https"
	_ "github.com/rilldata/rill/runtime/connectors/s3"
	_ "github.com/rilldata/rill/runtime/drivers/druid"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	Env                 string        `default:"development"`
	HTTPPort            int           `default:"8080" split_words:"true"`
	GRPCPort            int           `default:"9090" split_words:"true"`
	LogLevel            zapcore.Level `default:"info" split_words:"true"`
	DatabaseDriver      string        `default:"sqlite"`
	DatabaseURL         string        `default:"file:rill?mode=memory&cache=shared" split_words:"true"`
	ConnectionCacheSize int           `default:"100" split_words:"true"`
	QueryCacheSize      int           `default:"10000" split_words:"true"`
}

func main() {
	// Load .env (note: fails silently if .env has errors)
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("failed to load godotenv: %s", err.Error())
	}

	// Init config
	var conf Config
	err = envconfig.Process("rill_runtime", &conf)
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

	// Init runtime
	opts := &runtime.Options{
		ConnectionCacheSize: conf.ConnectionCacheSize,
		MetastoreDriver:     conf.DatabaseDriver,
		MetastoreDSN:        conf.DatabaseURL,
		QueryCacheSize:      conf.QueryCacheSize,
	}
	rt, err := runtime.New(opts, logger)
	if err != nil {
		logger.Fatal("error: could not create runtime", zap.Error(err))
	}

	// Init server
	srvOpts := &server.Options{
		HTTPPort: conf.HTTPPort,
		GRPCPort: conf.GRPCPort,
	}
	server, err := server.NewServer(srvOpts, rt, logger)
	if err != nil {
		logger.Fatal("error: could not create server", zap.Error(err))
	}

	// Run server
	ctx := graceful.WithCancelOnTerminate(context.Background())
	group, cctx := errgroup.WithContext(ctx)
	group.Go(func() error { return server.ServeGRPC(cctx) })
	group.Go(func() error { return server.ServeHTTP(cctx) })
	err = group.Wait()
	if err != nil {
		logger.Fatal("server crashed", zap.Error(err))
	}

	logger.Info("server shutdown gracefully")
}
