package runtime

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type Options struct {
	ConnectionCacheSize int
	MetastoreDriver     string
	QueryCacheSizeBytes int64
	AllowHostAccess     bool
	SafeSourceRefresh   bool
	// GlobalDrivers are drivers whose handles are shared with all instances
	GlobalDrivers []*Connector
	// PrivateDrivers are drivers whose handles are private to an instance
	PrivateDrivers []*Connector
}

type Connector struct {
	Type    string
	Name    string
	Configs map[string]string
}

// ConnectorDefByName return the connector definition and whether it should be shared or not
func (o *Options) ConnectorDefByName(name string) (*Connector, bool, error) {
	for _, c := range o.GlobalDrivers {
		if c.Name == name {
			return c, true, nil
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
	rt := &Runtime{
		opts:               opts,
		logger:             logger,
		connCache:          newConnectionCache(opts.ConnectionCacheSize, logger),
		migrationMetaCache: newMigrationMetaCache(math.MaxInt),
		queryCache:         newQueryCache(opts.QueryCacheSizeBytes),
	}
	store, _, err := rt.newMetaStore(context.Background(), "")
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
