package admin

import (
	"context"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

type Options struct {
	DatabaseDriver      string
	DatabaseDSN         string
	GithubAppID         int64
	GithubAppPrivateKey string
}

type Service struct {
	DB     database.DB
	opts   *Options
	logger *zap.Logger
	github *github.Client
}

func New(opts *Options, logger *zap.Logger) (*Service, error) {
	// Init db
	db, err := database.Open(opts.DatabaseDriver, opts.DatabaseDSN)
	if err != nil {
		logger.Fatal("error connecting to database", zap.Error(err))
	}

	// Auto-run migrations
	err = db.Migrate(context.Background())
	if err != nil {
		logger.Fatal("error migrating database", zap.Error(err))
	}

	// Create Github client
	itr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, opts.GithubAppID, []byte(opts.GithubAppPrivateKey))
	if err != nil {
		return nil, err
	}
	gh := github.NewClient(&http.Client{Transport: itr})

	return &Service{
		DB:     db,
		opts:   opts,
		logger: logger,
		github: gh,
	}, nil
}

func (s *Service) Close() error {
	return s.DB.Close()
}
