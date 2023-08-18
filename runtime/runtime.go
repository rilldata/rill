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
	ConnectionCacheSize   int
	MetastoreConnector    string
	QueryCacheSizeBytes   int64
	PolicyEngineCacheSize int
	AllowHostAccess       bool
	SafeSourceRefresh     bool
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
	policyEngine       *policyEngine
}

func New(opts *Options, logger *zap.Logger, client activity.Client) (*Runtime, error) {
	rt := &Runtime{
		opts:               opts,
		logger:             logger,
		connCache:          newConnectionCache(opts.ConnectionCacheSize, logger, client),
		migrationMetaCache: newMigrationMetaCache(math.MaxInt),
		queryCache:         newQueryCache(opts.QueryCacheSizeBytes),
		policyEngine:       newPolicyEngine(opts.PolicyEngineCacheSize, logger),
	}
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

func (r *Runtime) ResolveMetricsViewPolicy(attributes map[string]any, instanceID string, mv *runtimev1.MetricsView, lastUpdatedOn time.Time) (*ResolvedMetricsViewPolicy, error) {
	return r.policyEngine.resolveMetricsViewPolicy(attributes, instanceID, mv, lastUpdatedOn)
}
