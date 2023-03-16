package runtime

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
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
	repo, repoStore, err := r.checkRepoConnection(inst)
	if err != nil {
		return err
	}
	defer repo.Close()

	// Check that it's a driver that supports embedded catalogs
	if inst.EmbedCatalog {
		_, ok := olap.CatalogStore()
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

	c := rillv1beta.New(repoStore, inst.ID)
	proj, err := c.ProjectConfig(ctx)
	if err != nil {
		return err
	}
	inst.ProjectVariables = proj.Variables
	// this is a hack to set allow_host_credentials
	// ideally the runtime should propagate this flag to connectors.Env
	if inst.Variables == nil {
		inst.Variables = make(map[string]string)
	}
	inst.Variables["allow_host_credentials"] = strconv.FormatBool(r.opts.AllowHostCredentials)

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

	svc, err := r.NewCatalogService(ctx, inst.ID)
	if err != nil { // return error if db handlers can't be opened
		return err
	}

	// delete instance related data if catalog is not embedded
	if !inst.EmbedCatalog {
		err := svc.Catalog.DeleteEntries(ctx, instanceID)
		if err != nil {
			return err
		}
	}

	if dropDB {
		// ignoring the dropDB error since if db is already dropped it may not be possible to retry
		err = svc.Olap.DropDB()
		r.logger.Error("could not drop database", zap.Error(err), zap.String("instance_id", instanceID))
	}

	r.evictCaches(ctx, inst)
	// delete instance
	return r.Registry().DeleteInstance(ctx, instanceID)
}

// EditInstance edits exisiting instance.
// Confirming to put api specs, it is expected to send entire existing instance data.
// The API compares and only evicts caches if drivers or dsn is changed.
// This is done to ensure that db handlers are not unnecessarily closed
func (r *Runtime) EditInstance(ctx context.Context, inst *drivers.Instance) error {
	olderInstance, err := r.Registry().FindInstance(ctx, inst.ID)
	if err != nil {
		return err
	}

	// 1. changes in olap driver or olap dsn
	olapChanged := olderInstance.OLAPDriver != inst.OLAPDriver || olderInstance.OLAPDSN != inst.OLAPDSN
	if olapChanged {
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
		olapConn, err := r.connCache.get(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
		if err != nil {
			return err
		}
		_, ok := olapConn.CatalogStore()
		if !ok {
			return errors.New("driver does not support embedded catalogs")
		}
	}

	// 3. changes in repo driver or repo dsn
	repoChanged := inst.RepoDriver != olderInstance.RepoDriver || inst.RepoDSN != olderInstance.RepoDSN
	if repoChanged {
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
	if olapChanged || repoChanged {
		r.evictCaches(ctx, olderInstance)
	}

	// update variables
	if inst.Variables == nil {
		inst.Variables = make(map[string]string)
	}
	inst.Variables["allow_host_credentials"] = strconv.FormatBool(r.opts.AllowHostCredentials)
	// update the entire instance for now to avoid building queries in some complicated way
	return r.Registry().EditInstance(ctx, inst)
}

// TODO :: this is a rudimentary solution and ideally should be done in some producer/consumer pattern
func (r *Runtime) evictCaches(ctx context.Context, inst *drivers.Instance) {
	// evict and close exisiting connection
	r.connCache.evict(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
	r.connCache.evict(ctx, inst.ID, inst.RepoDriver, inst.RepoDSN)

	// evict catalog cache
	r.migrationMetaCache.evict(ctx, inst.ID)
	// query cache can't be evicted since key is a combination of instance ID and other parameters
}

func (r *Runtime) checkRepoConnection(inst *drivers.Instance) (drivers.Connection, drivers.RepoStore, error) {
	repo, err := drivers.Open(inst.RepoDriver, inst.RepoDSN, r.logger)
	if err != nil {
		return nil, nil, err
	}
	repoStore, ok := repo.RepoStore()
	if !ok {
		return nil, nil, fmt.Errorf("not a valid repo driver: '%s'", inst.RepoDriver)
	}

	return repo, repoStore, nil
}

func (r *Runtime) checkOlapConnection(inst *drivers.Instance) (drivers.Connection, drivers.OLAPStore, error) {
	olap, err := drivers.Open(inst.OLAPDriver, inst.OLAPDSN, r.logger)
	if err != nil {
		return nil, nil, err
	}
	olapStore, ok := olap.OLAPStore()
	if !ok {
		return nil, nil, fmt.Errorf("not a valid OLAP driver: '%s'", inst.OLAPDriver)
	}
	return olap, olapStore, nil
}
