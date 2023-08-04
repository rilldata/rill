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
	// Check OLAP connection
	olap, _, err := r.checkOlapConnection(inst)
	if err != nil {
		return err
	}
	defer olap.Close()

	// Check repo connection
	repo, _, err := r.checkRepoConnection(inst)
	if err != nil {
		return err
	}
	defer repo.Close()

	// Check that it's a driver that supports embedded catalogs
	if inst.EmbedCatalog {
		_, ok := olap.AsCatalogStore()
		if !ok {
			return errors.New("driver does not support embedded catalogs")
		}
	}

	// Prepare connections for use
	err = olap.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare instance: %w", err)
	}
	err = repo.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare instance: %w", err)
	}

	// this is a hack to set variables and pass to connectors
	// ideally the runtime should propagate this flag to connectors.Env
	if inst.Variables == nil {
		inst.Variables = make(map[string]string)
	}
	inst.Variables["allow_host_access"] = strconv.FormatBool(r.opts.AllowHostAccess)

	// Create instance
	err = r.Registry().CreateInstance(ctx, inst)
	if err != nil {
		return err
	}

	return nil
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
			err = catalog.DeleteEntries(ctx, instanceID)
			release()
		}
		if err != nil {
			r.logger.Error("delete instance: error deleting catalog", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		}
	}

	// Drop the underlying data store
	if dropDB {
		c, shared, _ := r.OLAPDef(inst)
		vars := variables(inst.OLAPDriver, c.Configs, inst.ResolveVariables())
		conn, release, err := r.connCache.get(ctx, instanceID, c.Type, vars, shared)
		if err == nil {
			release()
			err = conn.Close()
			if err != nil {
				r.logger.Error("delete instance: error closing connection", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
			}
		} else {
			r.logger.Error("delete instance: error getting connection", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		}

		err = drivers.Drop(c.Type, convert(vars), r.logger)
		if err != nil {
			r.logger.Error("could not drop database", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		}
	}

	// Evict cached data and connections for the instance
	r.evictCaches(ctx, inst)

	return r.Registry().DeleteInstance(ctx, instanceID)
}

// EditInstance edits exisiting instance.
// The API compares and only evicts caches if drivers or dsn is changed.
// This is done to ensure that db handlers are not unnecessarily closed
func (r *Runtime) EditInstance(ctx context.Context, inst *drivers.Instance) error {
	olderInstance, err := r.Registry().FindInstance(ctx, inst.ID)
	if err != nil {
		return err
	}

	evict := false
	// 1. changes in olap driver or olap dsn
	if r.olapChanged(olderInstance, inst) {
		evict = true
		// Check OLAP connection
		olap, _, err := r.checkOlapConnection(inst)
		if err != nil {
			return err
		}
		defer olap.Close()

		// Prepare connections for use
		err = olap.Migrate(ctx)
		if err != nil {
			return fmt.Errorf("failed to prepare instance: %w", err)
		}
	}

	// 2. Check that it's a driver that supports embedded catalogs
	if inst.EmbedCatalog {
		olapConn, _, err := r.checkOlapConnection(inst)
		if err != nil {
			return err
		}
		defer olapConn.Close()
		_, ok := olapConn.AsCatalogStore()
		if !ok {
			return errors.New("driver does not support embedded catalogs")
		}
	}

	// 3. changes in repo driver or repo dsn
	if r.repoChanged(olderInstance, inst) {
		evict = true
		// Check repo connection
		repo, _, err := r.checkRepoConnection(inst)
		if err != nil {
			return err
		}
		defer repo.Close()

		// Prepare connections for use
		err = repo.Migrate(ctx)
		if err != nil {
			return fmt.Errorf("failed to prepare instance: %w", err)
		}
	}

	// evict caches if connections need to be updated
	if evict {
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

// TODO :: this is a rudimentary solution and ideally should be done in some producer/consumer pattern
func (r *Runtime) evictCaches(ctx context.Context, inst *drivers.Instance) {
	// evict and close exisiting connection
	c, _, _ := r.OLAPDef(inst)
	r.connCache.evict(ctx, inst.ID, c.Type, variables(inst.OLAPDriver, c.Configs, inst.ResolveVariables()))
	c, _, _ = r.RepoDef(inst)
	r.connCache.evict(ctx, inst.ID, c.Type, variables(inst.RepoDriver, c.Configs, inst.ResolveVariables()))

	// evict catalog cache
	r.migrationMetaCache.evict(ctx, inst.ID)
	// query cache can't be evicted since key is a combination of instance ID and other parameters
}

func (r *Runtime) checkRepoConnection(inst *drivers.Instance) (drivers.Handle, drivers.RepoStore, error) {
	c, _, err := r.RepoDef(inst)
	if err != nil {
		return nil, nil, err
	}
	repo, err := drivers.Open(c.Type, convert(variables(c.Name, c.Configs, inst.ResolveVariables())), r.logger)
	if err != nil {
		return nil, nil, err
	}
	repoStore, ok := repo.AsRepoStore()
	if !ok {
		return nil, nil, fmt.Errorf("not a valid repo driver: '%s'", inst.RepoDriver)
	}

	return repo, repoStore, nil
}

func (r *Runtime) checkOlapConnection(inst *drivers.Instance) (drivers.Handle, drivers.OLAPStore, error) {
	c, _, err := r.OLAPDef(inst)
	if err != nil {
		return nil, nil, err
	}
	olap, err := drivers.Open(c.Type, convert(variables(c.Name, c.Configs, inst.ResolveVariables())), r.logger)
	if err != nil {
		return nil, nil, err
	}
	olapStore, ok := olap.AsOLAP()
	if !ok {
		return nil, nil, fmt.Errorf("not a valid OLAP driver: '%s'", inst.OLAPDriver)
	}
	return olap, olapStore, nil
}

func (r *Runtime) OLAPDef(inst *drivers.Instance) (*runtimev1.ConnectorDef, bool, error) {
	for _, c := range inst.Connectors {
		if c.Name == inst.OLAPDriver {
			return c, false, nil
		}
	}
	if c, _, err := r.opts.ConnectorDefByName(inst.OLAPDriver); err == nil {
		return &runtimev1.ConnectorDef{Name: c.Name, Type: c.Type, Configs: c.Configs}, true, nil
	}
	return nil, false, fmt.Errorf("dev error, olap connector doesn't exist")
}

func (r *Runtime) RepoDef(inst *drivers.Instance) (*runtimev1.ConnectorDef, bool, error) {
	for _, c := range inst.Connectors {
		if c.Name == inst.RepoDriver {
			return c, false, nil
		}
	}
	if c, _, err := r.opts.ConnectorDefByName(inst.RepoDriver); err == nil {
		return &runtimev1.ConnectorDef{Name: c.Name, Type: c.Type, Configs: c.Configs}, true, nil
	}
	return nil, false, fmt.Errorf("dev error, repo connector doesn't exist")
}

func (r *Runtime) repoChanged(a, b *drivers.Instance) bool {
	o1, s1, _ := r.RepoDef(a)
	o2, s2, _ := r.RepoDef(b)
	return a.RepoDriver != b.RepoDriver || s1 != s2 || !equal(o1, o2)
}

func (r *Runtime) olapChanged(a, b *drivers.Instance) bool {
	o1, s1, _ := r.OLAPDef(a)
	o2, s2, _ := r.OLAPDef(b)
	return a.OLAPDriver != b.OLAPDriver || s1 != s2 || !equal(o1, o2)
}

func equal(a, b *runtimev1.ConnectorDef) bool {
	if (a != nil) != (b != nil) {
		return false
	}
	return a.Name == b.Name && a.Type == b.Type && maps.Equal(a.Configs, b.Configs)
}
