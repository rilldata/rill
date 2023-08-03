package runtime

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type Options struct {
	ConnectionCacheSize int
	MetastoreDriver     string
	MetastoreDSN        string
	QueryCacheSizeBytes int64
	AllowHostAccess     bool
	SafeSourceRefresh   bool
	GlobalConnectors    []*rillv1.ConnectorDef
	PrivateConnectors   []*rillv1.ConnectorDef
}

func (o *Options) ConnectorByName(name string) (*rillv1.ConnectorDef, bool, error) {
	for _, c := range o.GlobalConnectors {
		if c.Name == name {
			return c, true, nil
		}
	}
	for _, c := range o.PrivateConnectors {
		if c.Name == name {
			return c, false, nil
		}
	}
	return nil, false, fmt.Errorf("connector %s doesn't exist", name)
}

type Runtime struct {
	opts               *Options
	metastore          drivers.Handle
	logger             *zap.Logger
	connCache          *connectionCache
	migrationMetaCache *migrationMetaCache
	queryCache         *queryCache
}

func New(opts *Options, logger *zap.Logger) (*Runtime, error) {
	// Open metadata db connection
	c, _, err := opts.ConnectorByName("metastore")
	if err != nil {
		return nil, err
	}
	metastore, err := drivers.Open(c.Type, convert(c.Defaults), logger)
	if err != nil {
		return nil, fmt.Errorf("could not connect to metadata db: %w", err)
	}
	err = metastore.Migrate(context.Background())
	if err != nil {
		return nil, fmt.Errorf("metadata db migration: %w", err)
	}

	// Check the metastore is a registry
	_, ok := metastore.AsRegistry()
	if !ok {
		return nil, fmt.Errorf("server metastore must be a valid registry")
	}

	return &Runtime{
		opts:               opts,
		metastore:          metastore,
		logger:             logger,
		connCache:          newConnectionCache(opts.ConnectionCacheSize, logger),
		migrationMetaCache: newMigrationMetaCache(math.MaxInt),
		queryCache:         newQueryCache(opts.QueryCacheSizeBytes),
	}, nil
}

func (r *Runtime) AllowHostAccess() bool {
	return r.opts.AllowHostAccess
}

func (r *Runtime) Close() error {
	return errors.Join(
		r.metastore.Close(),
		r.connCache.Close(),
		r.queryCache.close(),
	)
}
