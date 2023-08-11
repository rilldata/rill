package runtime

import (
	"context"
	"errors"
	"fmt"
	"math"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type Options struct {
	ConnectionCacheSize int
	MetastoreDriver     string
	QueryCacheSizeBytes int64
	AllowHostAccess     bool
	SafeSourceRefresh   bool
	// SystemConnectors are drivers whose handles are shared with all instances
	SystemConnectors []*runtimev1.ConnectorDef
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
	rt := &Runtime{
		opts:               opts,
		logger:             logger,
		connCache:          newConnectionCache(opts.ConnectionCacheSize, logger),
		migrationMetaCache: newMigrationMetaCache(math.MaxInt),
		queryCache:         newQueryCache(opts.QueryCacheSizeBytes),
	}
	store, _, err := rt.AcquireGlobalHandle(context.Background(), "metastore")
	if err != nil {
		return nil, err
	}

	// Check the metastore is a registry
	_, ok := store.AsRegistry()
	if !ok {
		return nil, fmt.Errorf("server metastore must be a valid registry")
	}
	rt.metastore = store
	return rt, nil
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
