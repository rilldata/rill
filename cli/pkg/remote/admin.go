package remote

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config describes admin server config derived from environment variables.
// Env var keys must be prefixed with RILL_ADMIN_ and are converted from snake_case to CamelCase.
// For example RILL_ADMIN_HTTP_PORT is mapped to Config.HTTPPort.
type Config struct {
	DatabaseDriver         string        `default:"postgres" split_words:"true"`
	DatabaseURL            string        `split_words:"true"`
	HTTPPort               int           `default:"8080" split_words:"true"`
	GRPCPort               int           `default:"9090" split_words:"true"`
	LogLevel               zapcore.Level `default:"info" split_words:"true"`
	ExternalURL            string        `default:"http://localhost:8080" split_words:"true"`
	FrontendURL            string        `default:"http://localhost:3000" split_words:"true"`
	SessionKeyPairs        []string      `split_words:"true"`
	AllowedOrigins         []string      `default:"*" split_words:"true"`
	AuthDomain             string        `split_words:"true"`
	AuthClientID           string        `split_words:"true"`
	AuthClientSecret       string        `split_words:"true"`
	GithubAppID            int64         `split_words:"true"`
	GithubAppName          string        `split_words:"true"`
	GithubAppPrivateKey    string        `split_words:"true"`
	GithubAppWebhookSecret string        `split_words:"true"`
	ProvisionerSpec        string        `split_words:"true"`
	SigningJWKS            string        `split_words:"true"`
	SigningKeyID           string        `split_words:"true"`
}

func NewAdminService() (*admin.Service, error) {
	// Load .env (note: fails silently if .env has errors)
	_ = godotenv.Load()

	// Init config
	var conf Config
	err := envconfig.Process("rill_admin", &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Init logger
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(conf.LogLevel)
	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("error: failed to create logger: %w", err)
	}

	// Init runtime JWT issuer
	issuer, err := auth.NewIssuer(conf.ExternalURL, conf.SigningKeyID, []byte(conf.SigningJWKS))
	if err != nil {
		return nil, fmt.Errorf("error creating runtime jwt issuer: %w", err)
	}

	// Init admin service
	admOpts := &admin.Options{
		DatabaseDriver:      conf.DatabaseDriver,
		DatabaseDSN:         conf.DatabaseURL,
		GithubAppID:         conf.GithubAppID,
		GithubAppPrivateKey: conf.GithubAppPrivateKey,
		ProvisionerSpec:     conf.ProvisionerSpec,
	}
	adm, err := admin.New(admOpts, logger, issuer)
	if err != nil {
		return nil, fmt.Errorf("error creating service: %w", err)
	}
	return adm, nil
}
