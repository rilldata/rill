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

type Service struct {
	DB             database.DB
	Provisioner    *provisioner.StaticProvisioner
	Email          *email.Client
	Github         Github
	opts           *Options
	logger         *zap.Logger
	issuer         *auth.Issuer
	closeCtx       context.Context
	closeCtxCancel context.CancelFunc
	reconcileWg    sync.WaitGroup
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

	return &Service{
		DB:             db,
		Provisioner:    prov,
		Email:          emailClient,
		Github:         github,
		opts:           opts,
		logger:         logger,
		issuer:         issuer,
		closeCtx:       ctx,
		closeCtxCancel: cancel,
	}, nil
}

func (s *Service) Close() error {
	s.closeCtxCancel()
	s.reconcileWg.Wait()

	return s.DB.Close()
}

// UnsafeWaitForReconciles waits for all background reconciles to finish.
// It is unsafe because while it is running, no new reconciles should be started.
// It's a temporary solution until the runtime is able to reconcile asynchronously.
// Unlike s.Close(), it does not cancel currently running reconciles, it just waits for them to finish.
func (s *Service) UnsafeWaitForReconciles() {
	s.reconcileWg.Wait()
}
