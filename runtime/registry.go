package runtime

import (
	"context"
	"errors"
	"strconv"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func (r *Runtime) FindInstances(ctx context.Context) ([]*drivers.Instance, error) {
	return r.Registry().FindInstances(ctx)
}

func (r *Runtime) FindInstance(ctx context.Context, instanceID string) (*drivers.Instance, error) {
	return r.Registry().FindInstance(ctx, instanceID)
}

func (r *Runtime) CreateInstance(ctx context.Context, inst *drivers.Instance) error {
	// this is a hack to set variables and pass to connectors
	// remove this once sources start calling runtime.AcquireHandle in all cases
	if inst.Variables == nil {
		inst.Variables = make(map[string]string)
	}
	inst.Variables["allow_host_access"] = strconv.FormatBool(r.opts.AllowHostAccess)

	// Create instance
	return r.Registry().CreateInstance(ctx, inst)
}

func (r *Runtime) DeleteInstance(ctx context.Context, instanceID string, dropDB bool) error {
	inst, err := r.Registry().FindInstance(ctx, instanceID)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return nil
		}
		return err
	}

	// For idempotency, it's ok for some steps to fail

	// Delete instance related data if catalog is not embedded
	if !inst.EmbedCatalog {
		catalog, release, err := r.Catalog(ctx, instanceID)
		if err == nil {
			err = catalog.DeleteEntries(ctx)
			release()
		}
		if err != nil {
			r.logger.Error("delete instance: error deleting catalog", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		}
	}

	// Evict cached data and connections for the instance
	r.evictCaches(ctx, inst)

	// Drop the underlying data store
	if dropDB {
		err = r.EvictHandle(ctx, instanceID, inst.OLAPConnector, true)
		if err != nil {
			r.logger.Error("could not drop database", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		}
	}

	return r.Registry().DeleteInstance(ctx, instanceID)
}

// EditInstance edits exisiting instance. Calling it will evict all connection handles associated with the instance.
func (r *Runtime) EditInstance(ctx context.Context, inst *drivers.Instance) error {
	// evict caches
	oldInstance, err := r.Registry().FindInstance(ctx, inst.ID)
	if err != nil {
		return err
	}
	r.evictCaches(ctx, oldInstance)

	// update variables
	if inst.Variables == nil {
		inst.Variables = make(map[string]string)
	}
	inst.Variables["allow_host_access"] = strconv.FormatBool(r.opts.AllowHostAccess)

	// update the entire instance for now to avoid building queries in some complicated way
	return r.Registry().EditInstance(ctx, inst)
}

func (r *Runtime) evictCaches(ctx context.Context, inst *drivers.Instance) {
	// evict and close instance connections
	_ = r.EvictHandle(ctx, inst.ID, inst.OLAPConnector, false)
	_ = r.EvictHandle(ctx, inst.ID, inst.RepoConnector, false)
	// evict catalog cache
	r.migrationMetaCache.evict(ctx, inst.ID)
	// query cache can't be evicted since key is a combination of instance ID and other parameters
}

// GetInstanceAttributes fetches an instance and converts its annotations to attributes
// nil is returned if an error occurred or instance was not found
func (r *Runtime) GetInstanceAttributes(ctx context.Context, instanceID string) []attribute.KeyValue {
	instance, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil
	}

	return instanceAnnotationsToAttribs(instance)
}

func instanceAnnotationsToAttribs(instance *drivers.Instance) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(instance.Annotations)+1)
	attrs = append(attrs, attribute.String("instance_id", instance.ID))
	for k, v := range instance.Annotations {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}
