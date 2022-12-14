package runtime

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type Options struct {
	ConnectionCacheSize int
	MetastoreDriver     string
	MetastoreDSN        string
	QueryCacheSize      int
}

type Runtime struct {
	opts         *Options
	metastore    drivers.Connection
	logger       *zap.Logger
	connCache    *connectionCache
	catalogCache *catalogCache
	queryCache   *queryCache
}

func New(opts *Options, logger *zap.Logger) (*Runtime, error) {
	// Open metadata db connection
	metastore, err := drivers.Open(opts.MetastoreDriver, opts.MetastoreDSN)
	if err != nil {
		return nil, fmt.Errorf("could not connect to metadata db: %w", err)
	}
	err = metastore.Migrate(context.Background())
	if err != nil {
		return nil, fmt.Errorf("metadata db migration: %w", err)
	}

	// Check the metastore is a registry
	_, ok := metastore.RegistryStore()
	if !ok {
		return nil, fmt.Errorf("server metastore must be a valid registry")
	}

	return &Runtime{
		opts:         opts,
		metastore:    metastore,
		logger:       logger,
		connCache:    newConnectionCache(opts.ConnectionCacheSize),
		catalogCache: newCatalogCache(),
		queryCache:   newQueryCache(opts.QueryCacheSize),
	}, nil
}
