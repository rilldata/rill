package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	"github.com/rilldata/rill/admin/database"
	_ "github.com/rilldata/rill/admin/database/postgres"
	"github.com/rilldata/rill/admin/server"
	"github.com/rilldata/rill/runtime/pkg/graceful"
)

type Config struct {
	Env              string `default:"development"`
	DatabaseDriver   string `default:"postgres" split_words:"true"`
	DatabaseURL      string `split_words:"true"`
	Port             int    `default:"8080" split_words:"true"`
	SessionsSecret   string `split_words:"true"`
	AuthDomain       string `split_words:"true"`
	AuthClientID     string `split_words:"true"`
	AuthClientSecret string `split_words:"true"`
	AuthCallbackURL  string `split_words:"true"`
}

func main() {
	// Load .env (note: fails silently if .env has errors)
	_ = godotenv.Load()

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
		Port:             conf.Port,
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
	logger.Info("serving http", zap.Int("port", conf.Port))

	ctx := graceful.WithCancelOnTerminate(context.Background())
	err = s.Serve(ctx, conf.Port)
	if err != nil {
		logger.Error("server crashed", zap.Error(err))
	}

	logger.Info("server shutdown gracefully")
}
