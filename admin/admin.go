package admin

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin/ai"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
)

type Options struct {
	DatabaseDriver     string
	DatabaseDSN        string
	ProvisionerSetJSON string
	DefaultProvisioner string
	ExternalURL        string
	VersionNumber      string
	MetricsProjectOrg  string
	MetricsProjectName string
}

type Service struct {
	DB               database.DB
	ProvisionerSet   map[string]provisioner.Provisioner
	Email            *email.Client
	Github           Github
	AI               ai.Client
	Used             *usedFlusher
	Logger           *zap.Logger
	opts             *Options
	issuer           *auth.Issuer
	VersionNumber    string
	metricsProjectID string
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

	// Create provisioner set
	provSet, err := provisioner.NewSet(opts.ProvisionerSetJSON, db, logger)
	if err != nil {
		return nil, err
	}

	// Verify that the specified default provisioner is in the provisioner set
	_, ok := provSet[opts.DefaultProvisioner]
	if !ok {
		return nil, fmt.Errorf("default provisioner %q is not in the provisioner set", opts.DefaultProvisioner)
	}

	// Look for the optional metrics project
	var metricsProjectID string
	if opts.MetricsProjectOrg != "" && opts.MetricsProjectName != "" {
		proj, err := db.FindProjectByName(ctx, opts.MetricsProjectOrg, opts.MetricsProjectName)
		if err != nil {
			return nil, fmt.Errorf("error looking up metrics project: %w", err)
		}
		metricsProjectID = proj.ID
	}

	return &Service{
		DB:               db,
		ProvisionerSet:   provSet,
		Email:            emailClient,
		Github:           github,
		AI:               aiClient,
		Used:             newUsedFlusher(logger, db),
		Logger:           logger,
		opts:             opts,
		issuer:           issuer,
		VersionNumber:    opts.VersionNumber,
		metricsProjectID: metricsProjectID,
	}, nil
}

func (s *Service) Close() error {
	s.Used.Close()
	return s.DB.Close()
}
