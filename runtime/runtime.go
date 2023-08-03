package runtime

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type Options struct {
	ConnectionCacheSize int
	QueryCacheSizeBytes int64
	AllowHostAccess     bool
	SafeSourceRefresh   bool
	// GlobalDrivers are drivers whose handles are shared with all instances
	GlobalDrivers []*rillv1.ConnectorDef
	// PrivateDrivers are drivers whose handles are private to an instance
	PrivateDrivers []*rillv1.ConnectorDef
}

// ConnectorDefByName return the connector definition and whether it should be shared or not
func (o *Options) ConnectorDefByName(name string) (*rillv1.ConnectorDef, bool, error) {
	for _, c := range o.GlobalDrivers {
		if c.Name == name {
			return c, true, nil
		}
	}
	for _, c := range o.PrivateDrivers {
		if c.Name == name {
			return c, false, nil
		}
	}
	return nil, false, fmt.Errorf("connector %s doesn't exist", name)
}

func (o *Options) OLAPDef(dsn string) (*rillv1.ConnectorDef, bool, error) {
	c, shared, err := o.ConnectorDefByName("olap")
	if err != nil {
		return nil, false, fmt.Errorf("dev error, olap connector doesn't exist")
	}
	// TODO :: remove this hack and pass repodsn and olapdsn as variables in form connector.repo.xxxx
	dup := &rillv1.ConnectorDef{Name: c.Name, Type: c.Type, Defaults: maps.Clone(c.Defaults)}
	if dup.Defaults == nil {
		dup.Defaults = make(map[string]string)
	}
	dup.Defaults["dsn"] = dsn
	return dup, shared, nil
}

func (o *Options) RepoDef(dsn string) (*rillv1.ConnectorDef, bool, error) {
	c, shared, err := o.ConnectorDefByName("repo")
	if err != nil {
		return nil, false, fmt.Errorf("dev error, repo connector doesn't exist")
	}

	// TODO :: remove this hack and pass repodsn and olapdsn as variables in form connector.repo.xxxx
	dup := &rillv1.ConnectorDef{Name: c.Name, Type: c.Type, Defaults: maps.Clone(c.Defaults)}
	if dup.Defaults == nil {
		dup.Defaults = make(map[string]string)
	}
	dup.Defaults["dsn"] = dsn
	return dup, shared, nil
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
