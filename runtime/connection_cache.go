package runtime

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/conncache"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel/attribute"
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
	instanceID    string // Empty if connection is shared
	name          string
	driver        string
	config        map[string]any
	provision     bool
	provisionArgs map[string]any
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
		ErrTTL:               10 * time.Second,
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
			r.Logger.Error("connection cache: connection has been working for too long", zap.String("instance_id", x.instanceID), zap.String("driver", x.driver), zap.Bool("open", open))
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
func (r *Runtime) getConnection(ctx context.Context, cfg cachedConnectionConfig) (drivers.Handle, func(), error) {
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
	logger := r.Logger
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
		if cfg.provision {
			activityDims = append(activityDims, attribute.Bool("managed", true))
		}
		if activityClient != nil {
			activityClient = activityClient.With(activityDims...)
		}

		if cfg.provision {
			if cfg.name == inst.AdminConnector {
				return nil, fmt.Errorf("cannot provision the admin connector (catch-22)")
			}

			// Give the driver a hint that it's a managed connector.
			cfg.config = maps.Clone(cfg.config)
			cfg.config["managed"] = true

			// As a special carve-out, we never try to provision DuckDB through the admin service.
			// (Since the driver just starts an embedded DuckDB when `managed: true`.)
			skipAdminProvisioning := cfg.driver == "duckdb"

			// Provisioning has been requested, but the instance does not have an admin connector.
			if inst.AdminConnector == "" || skipAdminProvisioning {
				// As a fallback, we pass the provision arguments to the driver, giving it a chance to provision itself if it supports it.
				cfg.config["provision"] = true
				cfg.config["provision_args"] = cfg.provisionArgs
			} else {
				// Provision the connector using the admin connector.
				admin, release, err := r.Admin(ctx, cfg.instanceID)
				if err != nil {
					return nil, fmt.Errorf("failed to get admin client: %w", err)
				}
				defer release()

				newConfig, err := admin.ProvisionConnector(ctx, cfg.name, cfg.driver, cfg.provisionArgs)
				if err != nil {
					return nil, fmt.Errorf("failed to provision %q: %w", cfg.name, err)
				}

				// Merge the new provisioned config with the existing one.
				for key, value := range newConfig {
					cfg.config[key] = value
				}
			}
		}
	}

	// Create storage client with a path prefix scoped to the instance and connector.
	// For shared connections, we use "shared" as the path prefix.
	var storage *storage.Client
	if cfg.instanceID != "" {
		storage = r.storage.WithPrefix(cfg.instanceID, cfg.name)
	} else {
		storage = r.storage.WithPrefix("shared", cfg.name)
	}

	r.Logger.Debug("opening connection", zap.String("instance_id", cfg.instanceID), zap.String("driver", cfg.driver), zap.String("name", cfg.name), zap.Bool("provision", cfg.provision))
	handle, err := drivers.Open(cfg.driver, cfg.instanceID, cfg.config, storage, activityClient, logger)
	if err == nil && ctx.Err() != nil {
		err = fmt.Errorf("timed out while opening driver %q", cfg.driver)
	}
	r.activity.Record(ctx, activity.EventTypeLog, "connection_open",
		attribute.String("instance_id", cfg.instanceID),
		attribute.String("driver", cfg.driver),
		attribute.String("name", cfg.name),
		attribute.Bool("provision", cfg.provision),
		attribute.Bool("success", err == nil),
	)
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
	sb.WriteString(":")
	sb.WriteString(cfg.name)
	sb.WriteString(":")
	sb.WriteString(cfg.driver)
	sb.WriteString(":")
	keys := maps.Keys(cfg.config)
	slices.Sort(keys)
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteString(":")
		sb.WriteString(fmt.Sprint(cfg.config[key]))
		sb.WriteString(" ")
	}
	if cfg.provision {
		sb.WriteString(":provision=true:")
		keys := maps.Keys(cfg.provisionArgs)
		slices.Sort(keys)
		for _, key := range keys {
			sb.WriteString(key)
			sb.WriteString(":")
			sb.WriteString(fmt.Sprint(cfg.provisionArgs[key]))
			sb.WriteString(" ")
		}
	}
	return sb.String()
}
