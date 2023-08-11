package runtime

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

func (r *Runtime) FindInstances(ctx context.Context) ([]*drivers.Instance, error) {
	return r.Registry().FindInstances(ctx)
}

func (r *Runtime) FindInstance(ctx context.Context, instanceID string) (*drivers.Instance, error) {
	return r.Registry().FindInstance(ctx, instanceID)
}

func (r *Runtime) CreateInstance(ctx context.Context, inst *drivers.Instance) error {
	_, err := r.connectorDef(inst, inst.OLAPDriver)
	if err != nil {
		return fmt.Errorf("invalid olap driver")
	}

	_, err = r.connectorDef(inst, inst.RepoDriver)
	if err != nil {
		return fmt.Errorf("invalid repo driver")
	}

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
		err = r.FlushHandle(ctx, instanceID, inst.OLAPDriver, true)
		if err != nil {
			r.logger.Error("could not drop database", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		}
	}

	return r.Registry().DeleteInstance(ctx, instanceID)
}

// EditInstance edits exisiting instance.
// The API compares and only evicts caches if drivers or dsn is changed.
// This is done to ensure that db handlers are not unnecessarily closed
func (r *Runtime) EditInstance(ctx context.Context, inst *drivers.Instance) error {
	_, err := r.connectorDef(inst, inst.OLAPDriver)
	if err != nil {
		return fmt.Errorf("invalid olap driver")
	}

	_, err = r.connectorDef(inst, inst.RepoDriver)
	if err != nil {
		return fmt.Errorf("invalid repo driver")
	}

	olderInstance, err := r.Registry().FindInstance(ctx, inst.ID)
	if err != nil {
		return err
	}

	// evict caches if connections need to be updated
	if r.olapChanged(ctx, olderInstance, inst) || r.repoChanged(ctx, olderInstance, inst) {
		r.evictCaches(ctx, olderInstance)
	}

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
	_ = r.FlushHandle(ctx, inst.ID, inst.OLAPDriver, false)
	_ = r.FlushHandle(ctx, inst.ID, inst.RepoDriver, false)
	// evict catalog cache
	r.migrationMetaCache.evict(ctx, inst.ID)
	// query cache can't be evicted since key is a combination of instance ID and other parameters
}

func (r *Runtime) repoChanged(ctx context.Context, a, b *drivers.Instance) bool {
	o1, _ := r.connectorDef(a, a.RepoDriver)
	o2, _ := r.connectorDef(b, b.RepoDriver)
	return a.RepoDriver != b.RepoDriver || !equal(o1, o2)
}

func (r *Runtime) olapChanged(ctx context.Context, a, b *drivers.Instance) bool {
	o1, _ := r.connectorDef(a, a.OLAPDriver)
	o2, _ := r.connectorDef(b, b.OLAPDriver)
	return a.OLAPDriver != b.OLAPDriver || !equal(o1, o2)
}

func equal(a, b *runtimev1.Connector) bool {
	if (a != nil) != (b != nil) {
		return false
	}
	return a.Name == b.Name && a.Type == b.Type && maps.Equal(a.Config, b.Config)
}
