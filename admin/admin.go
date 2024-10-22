package admin

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/rilldata/rill/admin/ai"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/billing/payment"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs"
	"github.com/rilldata/rill/admin/provisioner"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
)

type Options struct {
	DatabaseDriver            string
	DatabaseDSN               string
	DatabaseEncryptionKeyring string
	ExternalURL               string
	FrontendURL               string
	ProvisionerSetJSON        string
	DefaultProvisioner        string
	VersionNumber             string
	VersionCommit             string
	MetricsProjectOrg         string
	MetricsProjectName        string
	AutoscalerCron            string
	ScaleDownConstraint       int
}

type Service struct {
	DB                  database.DB
	Jobs                jobs.Client
	URLs                *URLs
	ProvisionerSet      map[string]provisioner.Provisioner
	Email               *email.Client
	Github              Github
	AI                  ai.Client
	Assets              *storage.BucketHandle
	Used                *usedFlusher
	Logger              *zap.Logger
	opts                *Options
	issuer              *auth.Issuer
	VersionNumber       string
	VersionCommit       string
	metricsProjectID    string
	AutoscalerCron      string
	ScaleDownConstraint int
	Biller              billing.Biller
	PaymentProvider     payment.Provider
}

func New(ctx context.Context, opts *Options, logger *zap.Logger, issuer *auth.Issuer, emailClient *email.Client, github Github, aiClient ai.Client, assets *storage.BucketHandle, biller billing.Biller, p payment.Provider) (*Service, error) {
	// Init db
	db, err := database.Open(opts.DatabaseDriver, opts.DatabaseDSN, opts.DatabaseEncryptionKeyring)
	if err != nil {
		logger.Fatal("error connecting to database", zap.Error(err))
	}

	// Init URLs
	urls, err := NewURLs(opts.ExternalURL, opts.FrontendURL)
	if err != nil {
		logger.Fatal("error parsing URLs", zap.Error(err))
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
		DB:                  db,
		URLs:                urls,
		ProvisionerSet:      provSet,
		Email:               emailClient,
		Github:              github,
		AI:                  aiClient,
		Assets:              assets,
		Used:                newUsedFlusher(logger, db),
		Logger:              logger,
		opts:                opts,
		issuer:              issuer,
		VersionNumber:       opts.VersionNumber,
		VersionCommit:       opts.VersionCommit,
		metricsProjectID:    metricsProjectID,
		AutoscalerCron:      opts.AutoscalerCron,
		ScaleDownConstraint: opts.ScaleDownConstraint,
		Biller:              biller,
		PaymentProvider:     p,
	}, nil
}

func (s *Service) Close() error {
	s.Used.Close()
	return s.DB.Close()
}
