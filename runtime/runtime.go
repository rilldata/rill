package runtime

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
)

type Options struct {
	ConnectionCacheSize     int
	MetastoreConnector      string
	QueryCacheSizeBytes     int64
	SecurityEngineCacheSize int
	AllowHostAccess         bool
	SafeSourceRefresh       bool
	// SystemConnectors are drivers whose handles are shared with all instances
	SystemConnectors []*runtimev1.Connector
}
type Runtime struct {
	opts               *Options
	metastore          drivers.Handle
	logger             *zap.Logger
	connCache          *connectionCache
	migrationMetaCache *migrationMetaCache
	queryCache         *queryCache
	securityEngine     *securityEngine
}

func New(opts *Options, logger *zap.Logger, client activity.Client) (*Runtime, error) {
	rt := &Runtime{
		opts:               opts,
		logger:             logger,
		migrationMetaCache: newMigrationMetaCache(math.MaxInt),
		queryCache:         newQueryCache(opts.QueryCacheSizeBytes),
		securityEngine:     newSecurityEngine(opts.SecurityEngineCacheSize, logger),
	}
	rt.connCache = newConnectionCache(opts.ConnectionCacheSize, logger, rt, client)
	store, _, err := rt.AcquireSystemHandle(context.Background(), opts.MetastoreConnector)
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

func (r *Runtime) Controller(instanceID string) (*Controller, error) {
	panic("not implemented")
}

func (r *Runtime) ResolveMetricsViewSecurity(attributes map[string]any, instanceID string, mv *runtimev1.MetricsView, lastUpdatedOn time.Time) (*ResolvedMetricsViewSecurity, error) {
	return r.securityEngine.resolveMetricsViewSecurity(attributes, instanceID, mv, lastUpdatedOn)
}

func (r *Runtime) ResolveMetricsViewSecurityV2(attributes map[string]any, instanceID string, mv *runtimev1.MetricsViewSpec, lastUpdatedOn time.Time) (*ResolvedMetricsViewSecurity, error) {
	// TODO: Implement actual checks when deprecating ResolveMetricsViewSecurity
	return &ResolvedMetricsViewSecurity{
		Access: true,
	}, nil
}
