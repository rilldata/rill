package runtime

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/singleflight"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter                        = otel.Meter("github.com/rilldata/rill/runtime")
	queryCacheHitsCounter        = observability.Must(meter.Int64ObservableCounter("query_cache.hits"))
	queryCacheMissesCounter      = observability.Must(meter.Int64ObservableCounter("query_cache.misses"))
	queryCacheItemCountGauge     = observability.Must(meter.Int64ObservableGauge("query_cache.items"))
	queryCacheSizeBytesGauge     = observability.Must(meter.Int64ObservableGauge("query_cache.size", metric.WithUnit("bytes")))
	queryCacheEntrySizeHistogram = observability.Must(meter.Int64Histogram("query_cache.entry_size", metric.WithUnit("bytes")))
)

type QueryResult struct {
	Value any
	Bytes int64
}

type ExportOptions struct {
	Format       runtimev1.ExportFormat
	Priority     int
	PreWriteHook func(filename string) error
}

type Query interface {
	// Key should return a cache key that uniquely identifies the query
	Key() string
	// Deps should return the resource names that the query targets.
	// It's used to invalidate cached queries when the underlying data changes.
	// If a dependency doesn't exist, it is ignored. (So if the underlying resource kind is unknown, it can return all possible dependency names.)
	Deps() []*runtimev1.ResourceName
	// MarshalResult should return the query result and estimated cost in bytes for caching
	MarshalResult() *QueryResult
	// UnmarshalResult should populate a query with a cached result
	UnmarshalResult(v any) error
	// Resolve should execute the query against the instance's infra.
	// Error can be nil along with a nil result in general, i.e. when a model contains no rows aggregation results can be nil.
	Resolve(ctx context.Context, rt *Runtime, instanceID string, priority int) error
	// Export resolves the query and serializes the result to the writer.
	Export(ctx context.Context, rt *Runtime, instanceID string, w io.Writer, opts *ExportOptions) error
}

func (r *Runtime) Query(ctx context.Context, instanceID string, query Query, priority int) error {
	qk := query.Key()
	// If key is empty, skip caching
	if qk == "" {
		return query.Resolve(ctx, r, instanceID, priority)
	}

	// Skip caching for specific named drivers.
	// TODO: Make this configurable with a default provided by the driver.
	olap, release, err := r.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	// TODO :: check enabling caching at clickhouse level
	if olap.Dialect() == drivers.DialectDruid || olap.Dialect() == drivers.DialectClickHouse {
		release()
		return query.Resolve(ctx, r, instanceID, priority)
	}
	release()

	// Get dependency cache keys
	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		return err
	}
	deps := query.Deps()
	depKeys := make([]string, 0, len(deps))
	for _, dep := range deps {
		res, err := ctrl.Get(ctx, dep, false)
		if err != nil {
			// Deps are approximate, not exact (see docstring for Deps()), so they may not all exist
			continue
		}
		// Using StateUpdatedOn instead of StateVersion because the state version is reset when the resource is deleted and recreated.
		key := fmt.Sprintf("%s:%s:%d:%d", res.Meta.Name.Kind, res.Meta.Name.Name, res.Meta.StateUpdatedOn.Seconds, res.Meta.StateUpdatedOn.Nanos/int32(time.Millisecond))
		depKeys = append(depKeys, key)
	}

	// If there were no known dependencies, skip caching
	if len(depKeys) == 0 {
		return query.Resolve(ctx, r, instanceID, priority)
	}

	// Build cache key
	depKey := strings.Join(depKeys, ";")
	key := queryCacheKey{
		instanceID:    instanceID,
		queryKey:      qk,
		dependencyKey: depKey,
	}.String()

	// Try to get from cache
	if val, ok := r.queryCache.cache.Get(key); ok {
		observability.AddRequestAttributes(ctx, attribute.Bool("query.cache_hit", true))
		return query.UnmarshalResult(val)
	}
	observability.AddRequestAttributes(ctx, attribute.Bool("query.cache_hit", false))

	// Load with singleflight
	owner := false
	val, err := r.queryCache.singleflight.Do(ctx, key, func(ctx context.Context) (any, error) {
		// Try cache again
		if val, ok := r.queryCache.cache.Get(key); ok {
			return val, nil
		}

		// Load
		err := query.Resolve(ctx, r, instanceID, priority)
		if err != nil {
			return nil, err
		}

		owner = true
		res := query.MarshalResult()
		r.queryCache.cache.Set(key, res.Value, res.Bytes)
		queryCacheEntrySizeHistogram.Record(ctx, res.Bytes, metric.WithAttributes(attribute.String("query", queryName(query))))
		return res.Value, nil
	})
	if err != nil {
		return err
	}

	if !owner {
		return query.UnmarshalResult(val)
	}
	return nil
}

type queryCacheKey struct {
	instanceID    string
	queryKey      string
	dependencyKey string
}

func (k queryCacheKey) String() string {
	return fmt.Sprintf("inst:%s deps:%s qry:%s", k.instanceID, k.dependencyKey, k.queryKey)
}

type queryCache struct {
	cache        *ristretto.Cache
	singleflight *singleflight.Group[string, any]
	metrics      metric.Registration
}

func newQueryCache(sizeInBytes int64) *queryCache {
	if sizeInBytes <= 100 {
		panic(fmt.Sprintf("invalid cache size should be greater than 100: %v", sizeInBytes))
	}
	cache, err := ristretto.NewCache(&ristretto.Config{
		// Use 5% of cache memory for storing counters. Each counter takes roughly 3 bytes.
		// Recommended value is 10x the number of items in cache when full.
		// Tune this again based on metrics.
		NumCounters: int64(float64(sizeInBytes) * 0.05 / 3),
		MaxCost:     int64(float64(sizeInBytes) * 0.95),
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		panic(err)
	}

	metrics := observability.Must(meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		observer.ObserveInt64(queryCacheHitsCounter, int64(cache.Metrics.Hits()))
		observer.ObserveInt64(queryCacheMissesCounter, int64(cache.Metrics.Misses()))
		observer.ObserveInt64(queryCacheItemCountGauge, int64(cache.Metrics.KeysAdded()-cache.Metrics.KeysEvicted()))
		observer.ObserveInt64(queryCacheSizeBytesGauge, int64(cache.Metrics.CostAdded()-cache.Metrics.CostEvicted()))
		return nil
	}, queryCacheHitsCounter, queryCacheMissesCounter, queryCacheItemCountGauge, queryCacheSizeBytesGauge))

	return &queryCache{
		cache:        cache,
		singleflight: &singleflight.Group[string, any]{},
		metrics:      metrics,
	}
}

func (c *queryCache) close() error {
	c.cache.Close()
	return c.metrics.Unregister()
}

func queryName(q Query) string {
	nameWithPkg := fmt.Sprintf("%T", q)
	_, after, _ := strings.Cut(nameWithPkg, ".")
	return after
}
