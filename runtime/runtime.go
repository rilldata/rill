package runtime

import (
	"context"
	"fmt"
	"math"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type Options struct {
	ConnectionCacheSize   int
	MetastoreDriver       string
	MetastoreDSN          string
	QueryCacheSizeInBytes int64
	AllowHostAccess       bool
	SafeSourceRefresh     bool
}

type Runtime struct {
	opts               *Options
	metastore          drivers.Connection
	logger             *zap.Logger
	connCache          *connectionCache
	migrationMetaCache *migrationMetaCache
	queryCache         *queryCache
}

func New(opts *Options, logger *zap.Logger) (*Runtime, error) {
	// Open metadata db connection
	metastore, err := drivers.Open(opts.MetastoreDriver, opts.MetastoreDSN, logger)
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
		opts:               opts,
		metastore:          metastore,
		logger:             logger,
		connCache:          newConnectionCache(opts.ConnectionCacheSize, logger),
		migrationMetaCache: newMigrationMetaCache(math.MaxInt),
		queryCache:         newQueryCache(opts.QueryCacheSizeInBytes),
	}, nil
}

func (r *Runtime) Close() error {
	err1 := r.metastore.Close()
	err2 := r.connCache.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
