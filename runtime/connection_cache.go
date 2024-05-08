package runtime

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/conncache"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

var (
	connCacheOpens          = observability.Must(meter.Int64Counter("connnection_cache.opens"))
	connCacheCloses         = observability.Must(meter.Int64Counter("connnection_cache.closes"))
	connCacheSizeTotal      = observability.Must(meter.Int64UpDownCounter("connnection_cache.size_total"))
	connCacheSizeLRU        = observability.Must(meter.Int64UpDownCounter("connnection_cache.size_lru"))
	connCacheOpenLatencyMS  = observability.Must(meter.Int64Histogram("connnection_cache.open_latency", metric.WithUnit("ms")))
	connCacheCloseLatencyMS = observability.Must(meter.Int64Histogram("connnection_cache.close_latency", metric.WithUnit("ms")))
)

type cachedConnectionConfig struct {
	instanceID string // Empty if connection is shared
	driver     string
	config     map[string]any
}

// newConnectionCache returns a concurrency-safe cache for open connections.
// Connections should preferably be opened only via the connection cache.
// It's implementation handles issues such as concurrent open/close/eviction of a connection.
// It also monitors for hanging connections.
func (r *Runtime) newConnectionCache() conncache.Cache {
	return conncache.New(conncache.Options{
		MaxIdleConnections:   r.opts.ConnectionCacheSize,
		OpenTimeout:          10 * time.Minute,
		CloseTimeout:         10 * time.Minute,
		CheckHangingInterval: time.Minute,
		OpenFunc: func(ctx context.Context, cfg any) (conncache.Connection, error) {
			x := cfg.(cachedConnectionConfig)
			return r.openAndMigrate(ctx, x)
		},
		KeyFunc: func(cfg any) string {
			x := cfg.(cachedConnectionConfig)
			return generateKey(x)
		},
		HangingFunc: func(cfg any, open bool) {
			x := cfg.(cachedConnectionConfig)
			r.logger.Error("connection cache: connection has been working for too long", zap.String("instance_id", x.instanceID), zap.String("driver", x.driver), zap.Bool("open", open))
		},
		Metrics: conncache.Metrics{
			Opens:          connCacheOpens,
			Closes:         connCacheCloses,
			SizeTotal:      connCacheSizeTotal,
			SizeLRU:        connCacheSizeLRU,
			OpenLatencyMS:  connCacheOpenLatencyMS,
			CloseLatencyMS: connCacheCloseLatencyMS,
		},
	})
}

// getConnection returns a cached connection for the given driver configuration.
// If instanceID is empty, the connection is considered shared (see drivers.Open for details).
func (r *Runtime) getConnection(ctx context.Context, instanceID, driver string, config map[string]any) (drivers.Handle, func(), error) {
	cfg := cachedConnectionConfig{
		instanceID: instanceID,
		driver:     driver,
		config:     config,
	}

	handle, release, err := r.connCache.Acquire(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	return handle.(drivers.Handle), release, nil
}

// evictInstanceConnections evicts all connections for the given instance.
func (r *Runtime) evictInstanceConnections(instanceID string) {
	r.connCache.EvictWhere(func(cfg any) bool {
		x := cfg.(cachedConnectionConfig)
		return x.instanceID == instanceID
	})
}

// openAndMigrate opens a connection and migrates it.
func (r *Runtime) openAndMigrate(ctx context.Context, cfg cachedConnectionConfig) (drivers.Handle, error) {
	logger := r.logger
	activityClient := r.activity
	if cfg.instanceID != "" { // Not shared across multiple instances
		inst, err := r.Instance(ctx, cfg.instanceID)
		if err != nil {
			return nil, err
		}

		logger, err = r.InstanceLogger(ctx, cfg.instanceID)
		if err != nil {
			return nil, err
		}

		activityDims := instanceAnnotationsToAttribs(inst)
		if activityClient != nil {
			activityClient = activityClient.With(activityDims...)
		}
	}

	handle, err := drivers.Open(cfg.driver, cfg.instanceID, cfg.config, activityClient, logger)
	if err == nil && ctx.Err() != nil {
		err = fmt.Errorf("timed out while opening driver %q", cfg.driver)
	}
	if err != nil {
		return nil, err
	}

	err = handle.Migrate(ctx)
	if err != nil {
		handle.Close()
		if errors.Is(err, ctx.Err()) {
			err = fmt.Errorf("timed out while migrating driver %q: %w", cfg.driver, err)
		}
		return nil, err
	}
	return handle, nil
}

func generateKey(cfg cachedConnectionConfig) string {
	sb := strings.Builder{}
	sb.WriteString(cfg.instanceID) // Empty if cfg.shared
	sb.WriteString(cfg.driver)
	keys := maps.Keys(cfg.config)
	slices.Sort(keys)
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteString(":")
		sb.WriteString(fmt.Sprint(cfg.config[key]))
		sb.WriteString(" ")
	}
	return sb.String()
}
