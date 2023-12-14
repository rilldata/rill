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
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type cachedConnectionConfig struct {
	instanceID string
	driver     string
	shared     bool
	config     map[string]any
}

// newConnectionCache returns a concurrency-safe cache for open connections.
// Connections should preferably be opened only via the connection cache.
// It's implementation handles issues such as concurrent open/close/eviction of a connection.
// It also monitors for hanging connections.
func (r *Runtime) newConnectionCache() conncache.Cache {
	return conncache.New(conncache.Options{
		MaxConnectionsIdle:   r.opts.ConnectionCacheSize,
		OpenTimeout:          2 * time.Minute,
		CloseTimeout:         5 * time.Minute,
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
	})
}

// getConnection returns a cached connection for the given driver configuration.
func (r *Runtime) getConnection(ctx context.Context, instanceID, driver string, config map[string]any, shared bool) (drivers.Handle, func(), error) {
	cfg := cachedConnectionConfig{
		instanceID: instanceID,
		driver:     driver,
		shared:     shared,
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
	if cfg.instanceID != "default" {
		logger = r.logger.With(zap.String("instance_id", cfg.instanceID), zap.String("driver", cfg.driver))
	}

	activityClient := r.activity
	if !cfg.shared {
		inst, err := r.Instance(ctx, cfg.instanceID)
		if err != nil {
			return nil, err
		}

		activityDims := instanceAnnotationsToAttribs(inst)
		if activityClient != nil {
			activityClient = activityClient.With(activityDims...)
		}
	}

	handle, err := drivers.Open(cfg.driver, cfg.config, cfg.shared, activityClient, logger)
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
