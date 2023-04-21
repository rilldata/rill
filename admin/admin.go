package admin

import (
	"context"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/email"
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
	github         *github.Client
	provisioner    provisioner.Provisioner
	issuer         *auth.Issuer
	closeCtx       context.Context
	closeCtxCancel context.CancelFunc
	email          *email.Client
}

func New(ctx context.Context, opts *Options, logger *zap.Logger, issuer *auth.Issuer, emailClient *email.Client) (*Service, error) {
	// Init db
	db, err := database.Open(opts.DatabaseDriver, opts.DatabaseDSN)
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
		github:         gh,
		provisioner:    prov,
		issuer:         issuer,
		closeCtx:       ctx,
		closeCtxCancel: cancel,
		email:          emailClient,
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
