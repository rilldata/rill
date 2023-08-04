package runtime

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

const (
	// may be use __system__ prefix ?
	_metastoreDriverName = "metastore"
	_repoDriverName      = "repo"
	_olapDriverName      = "olap"
)

type Options struct {
	ConnectionCacheSize int
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

func (o *Options) OLAPDef(dsn string) (*Connector, bool, error) {
	c, shared, err := o.ConnectorDefByName(_olapDriverName)
	if err != nil {
		return nil, false, fmt.Errorf("dev error, olap connector doesn't exist")
	}
	// TODO :: remove this hack and pass repodsn and olapdsn as variables in form connector.repo.xxxx
	dup := &Connector{Name: c.Name, Type: c.Type, Configs: maps.Clone(c.Configs)}
	if dup.Configs == nil {
		dup.Configs = make(map[string]string)
	}
	dup.Configs["dsn"] = dsn
	return dup, shared, nil
}

func (o *Options) RepoDef(dsn string) (*Connector, bool, error) {
	c, shared, err := o.ConnectorDefByName(_olapDriverName)
	if err != nil {
		return nil, false, fmt.Errorf("dev error, repo connector doesn't exist")
	}

	// TODO :: remove this hack and pass repodsn and olapdsn as variables in form connector.repo.xxxx
	dup := &Connector{Name: c.Name, Type: c.Type, Configs: maps.Clone(c.Configs)}
	if dup.Configs == nil {
		dup.Configs = make(map[string]string)
	}
	dup.Configs["dsn"] = dsn
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
