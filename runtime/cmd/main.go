package main

import (
	"context"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metadata"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	_ "github.com/rilldata/rill/runtime/sql"

	_ "github.com/rilldata/rill/runtime/infra/duckdb"
	_ "github.com/rilldata/rill/runtime/metadata/sqlite"
)

type Config struct {
	Env            string `default:"development"`
	LogLevel       string `default:"info" split_words:"true"`
	DatabaseDriver string `default:"sqlite"`
	DatabaseURL    string `default:":memory:" split_words:"true"`
	GRPCPort       int    `default:"9090" split_words:"true"`
	HTTPPort       int    `default:"8080" split_words:"true"`
}

func main() {
	// Init config
	var conf Config
	err := envconfig.Process("rill_runtime", &conf)
	if err != nil {
		fmt.Printf("Failed to load config: %s", err.Error())
		os.Exit(1)
	}

	// Init logger
	level, err := zerolog.ParseLevel(conf.LogLevel)
	if err != nil {
		fmt.Printf("Error parsing log level: %s", err.Error())
		os.Exit(1)
	}
	var logger zerolog.Logger
	if conf.Env == "production" {
		logger = zerolog.New(os.Stderr).Level(level)
	} else {
		logger = zerolog.New(zerolog.NewConsoleWriter()).Level(level)
	}

	// Init db
	driver, ok := metadata.Drivers[conf.DatabaseDriver]
	if !ok {
		logger.Fatal().Msgf("Unknown db driver '%s'", conf.DatabaseDriver)
	}
	db, err := driver.Open(conf.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to DB")
	}

	// Migrate db
	// TODO: Move to separate command and only auto-migrate in development
	mig, err := migrate.NewWithSourceInstance("iofs", driver.Migrations(), fmt.Sprintf("%s://%s", conf.DatabaseDriver, conf.DatabaseURL))
	if err != nil {
		logger.Fatal().Err(err).Msg("Migrator failed to connect to DB")
	}
	err = mig.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Fatal().Err(err).Msg("Failed to migrate DB")
	}

	// Init runtime and server
	rt := runtime.New(db, logger)
	opts := &runtime.ServerOptions{
		GRPCPort: conf.GRPCPort,
		HTTPPort: conf.HTTPPort,
	}
	server := runtime.NewServer(opts, rt, logger)

	// Run server
	ctx := graceful.WithCancelOnTerminate(context.Background())
	err = server.Serve(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Server failed")
	}

	logger.Info().Msg("Server shutdown gracefully")
}
