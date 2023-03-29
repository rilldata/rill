package admin

import (
	"context"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
)

type Options struct {
	DatabaseDriver      string
	DatabaseDSN         string
	GithubAppID         int64
	GithubAppPrivateKey string
	ProvisionerSpec     string
}

type Service struct {
	DB             database.DB
	opts           *Options
	logger         *zap.Logger
	Github         *github.Client
	provisioner    provisioner.Provisioner
	issuer         *auth.Issuer
	closeCtx       context.Context
	closeCtxCancel context.CancelFunc
}

func New(opts *Options, logger *zap.Logger, issuer *auth.Issuer) (*Service, error) {
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

	// Create provisioner
	prov, err := provisioner.NewStatic(opts.ProvisionerSpec, logger, db, issuer)
	if err != nil {
		return nil, err
	}

	// Create context that we cancel in Close() (for background reconciles)
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		DB:             db,
		opts:           opts,
		logger:         logger,
		Github:         gh,
		provisioner:    prov,
		issuer:         issuer,
		closeCtx:       ctx,
		closeCtxCancel: cancel,
	}, nil
}

func (s *Service) Close() error {
	err := s.provisioner.Close()
	if err != nil {
		return err
	}

	s.closeCtxCancel()
	// TODO: Also wait for background items to finish (up to a timeout)

	return s.DB.Close()
}
