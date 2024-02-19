package admin

import (
	"context"

	"github.com/rilldata/rill/admin/ai"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
)

type Options struct {
	DatabaseDriver  string
	DatabaseDSN     string
	ProvisionerSpec string
	ExternalURL     string
}

type Service struct {
	DB          database.DB
	Provisioner *provisioner.StaticProvisioner
	Email       *email.Client
	Github      Github
	AI          ai.Client
	Used        *usedFlusher
	Logger      *zap.Logger
	opts        *Options
	issuer      *auth.Issuer
}

func New(ctx context.Context, opts *Options, logger *zap.Logger, issuer *auth.Issuer, emailClient *email.Client, github Github, aiClient ai.Client) (*Service, error) {
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

	return &Service{
		DB:          db,
		Provisioner: prov,
		Email:       emailClient,
		Github:      github,
		AI:          aiClient,
		Used:        newUsedFlusher(logger, db),
		Logger:      logger,
		opts:        opts,
		issuer:      issuer,
	}, nil
}

func (s *Service) Close() error {
	s.Used.Close()
	return s.DB.Close()
}
