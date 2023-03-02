package admin

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

type Options struct {
	DatabaseDriver string
	DatabaseDSN    string
}

type Service struct {
	DB     database.DB
	opts   *Options
	logger *zap.Logger
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

	return &Service{
		DB:     db,
		opts:   opts,
		logger: logger,
	}, nil
}

func (s *Service) Close() error {
	return s.DB.Close()
}
