package admin

import (
	"context"
	"sync"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/email"
	"github.com/rilldata/rill/admin/provisioner"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
)

type Options struct {
	DatabaseDriver  string
	DatabaseDSN     string
	ProvisionerSpec string
}

type updateTSCache struct {
	cache map[string]bool
	lock  sync.Mutex
}

type Service struct {
	DB             database.DB
	Provisioner    *provisioner.StaticProvisioner
	Email          *email.Client
	opts           *Options
	logger         *zap.Logger
	github         Github
	issuer         *auth.Issuer
	closeCtx       context.Context
	closeCtxCancel context.CancelFunc
	deplTSCache    *updateTSCache
}

func New(ctx context.Context, opts *Options, logger *zap.Logger, issuer *auth.Issuer, emailClient *email.Client, github Github) (*Service, error) {
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

	// Create provisioner
	prov, err := provisioner.NewStatic(opts.ProvisionerSpec, db)
	if err != nil {
		return nil, err
	}

	// Create context that we cancel in Close() (for background reconciles)
	ctx, cancel := context.WithCancel(context.Background())

	adm := &Service{
		DB:             db,
		Provisioner:    prov,
		Email:          emailClient,
		opts:           opts,
		logger:         logger,
		github:         github,
		issuer:         issuer,
		closeCtx:       ctx,
		closeCtxCancel: cancel,
		deplTSCache:    &updateTSCache{cache: make(map[string]bool)},
	}

	adm.LastUsedFlusher(ctx)

	return adm, nil
}

func (s *Service) Close() error {
	s.closeCtxCancel()
	// TODO: Also wait for background items to finish (up to a timeout)

	return s.DB.Close()
}
