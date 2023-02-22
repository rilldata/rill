package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	_ "github.com/rilldata/rill/admin/database/postgres"
)

type Config struct {
	Env              string `default:"development"`
	DatabaseDriver   string `default:"postgres" split_words:"true"`
	DatabaseURL      string `split_words:"true"`
	HTTPPort         int    `default:"8080" split_words:"true"`
	GRPCPort         int    `default:"9090" split_words:"true"`
	SessionsSecret   string `split_words:"true"`
	AuthDomain       string `split_words:"true"`
	AuthClientID     string `split_words:"true"`
	AuthClientSecret string `split_words:"true"`
	AuthCallbackURL  string `split_words:"true"`
}

func main() {
	// Load .env (note: fails silently if .env has errors)
	_ = godotenv.Load("../.env")

	// Init config
	var conf Config
	err := envconfig.Process("rill_admin", &conf)
	if err != nil {
		fmt.Printf("Failed to load config: %s", err.Error())
		os.Exit(1)
	}

	// Init logger
	var logger *zap.Logger
	if conf.Env == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Printf("Error creating logger: %s", err.Error())
		os.Exit(1)
	}

	// Init db
	db, err := database.Open(conf.DatabaseDriver, conf.DatabaseURL)
	if err != nil {
		logger.Fatal("error connecting to database", zap.Error(err))
	}

	// Auto-run migrations (TODO: don't do this in production)
	err = db.Migrate(context.Background())
	if err != nil {
		logger.Fatal("error migrating database", zap.Error(err))
	}

	srvConf := server.Config{
		HTTPPort:         conf.HTTPPort,
		GRPCPort:         conf.GRPCPort,
		AuthDomain:       conf.AuthDomain,
		AuthClientID:     conf.AuthClientID,
		AuthClientSecret: conf.AuthClientSecret,
		AuthCallbackURL:  conf.AuthCallbackURL,
		SessionsSecret:   conf.SessionsSecret,
	}

	// Init server
	s, err := server.New(logger, db, srvConf)
	if err != nil {
		logger.Fatal("error creating server", zap.Error(err))
	}

	// Run server
	ctx := graceful.WithCancelOnTerminate(context.Background())
	group, cctx := errgroup.WithContext(ctx)
	group.Go(func() error { return s.ServeGRPC(cctx) })
	group.Go(func() error { return s.ServeHTTP(cctx) })
	err = group.Wait()
	if err != nil {
		logger.Fatal("server crashed", zap.Error(err))
	}

	logger.Info("server shutdown gracefully")
}
