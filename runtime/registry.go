package runtime

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/drivers"
)

func (r *Runtime) FindInstances(ctx context.Context) ([]*drivers.Instance, error) {
	return r.Registry().FindInstances(ctx)
}

func (r *Runtime) FindInstance(ctx context.Context, instanceID string) (*drivers.Instance, error) {
	return r.Registry().FindInstance(ctx, instanceID)
}

func (r *Runtime) CreateInstance(ctx context.Context, inst *drivers.Instance) error {
	// Check OLAP connection
	olap, err := drivers.Open(inst.OLAPDriver, inst.OLAPDSN, r.logger)
	if err != nil {
		return err
	}
	_, ok := olap.OLAPStore()
	if !ok {
		return fmt.Errorf("not a valid OLAP driver: '%s'", inst.OLAPDriver)
	}

	// Check repo connection
	repo, err := drivers.Open(inst.RepoDriver, inst.RepoDSN, r.logger)
	if err != nil {
		return err
	}
	repoStore, ok := repo.RepoStore()
	if !ok {
		return fmt.Errorf("not a valid repo driver: '%s'", inst.RepoDriver)
	}

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
	inst.ProjectEnv = proj.Env
	// this is a hack to set allow_host_credentials
	// ideally the runtime should propagate this flag to connectors.Env
	if inst.Env == nil {
		inst.Env = make(map[string]string)
	}
	inst.Env["allow_host_credentials"] = strconv.FormatBool(r.opts.AllowHostCredentials)

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
		return err
	}
	svc, err1 := r.catalogCache.get(ctx, r, instanceID)
	// drop tables and instance related data
	err2 := svc.Catalog.DeleteInstanceEntries(ctx, instanceID)
	err3 := svc.Olap.Drop(ctx)

	// evict and close exisiting connection
	r.connCache.evict(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
	r.connCache.evict(ctx, inst.ID, inst.RepoDriver, inst.RepoDSN)

	// evict catalog cache
	r.catalogCache.evict(ctx, instanceID)
	// query cache can't be evicted since key is a combination of instance ID and other parameters

	// delete instances
	err4 := r.Registry().DeleteInstance(ctx, instanceID)
	return returnFirstErr(err1, err2, err3, err4)
}

func (r *Runtime) EditInstance(ctx context.Context, inst *drivers.Instance) error {
	olderInstance, err := r.Registry().FindInstance(ctx, inst.ID)
	if err != nil {
		return err
	}

	// 1. changes in olap driver or olap dsn
	olapChanged := olderInstance.OLAPDriver != inst.OLAPDriver || olderInstance.OLAPDSN != inst.OLAPDSN
	var olapConn drivers.Connection
	if olapChanged {
		// evict exisiting connection
		r.connCache.evict(ctx, olderInstance.ID, olderInstance.OLAPDriver, olderInstance.OLAPDSN)
		// Check OLAP connection
		olap, err := drivers.Open(inst.OLAPDriver, inst.OLAPDSN, r.logger)
		if err != nil {
			return err
		}
		_, ok := olap.OLAPStore()
		if !ok {
			return fmt.Errorf("not a valid OLAP driver: '%s'", inst.OLAPDriver)
		}
		// Prepare connections for use
		err = olap.Migrate(ctx)
		if err != nil {
			return fmt.Errorf("failed to prepare instance: %w", err)
		}
		olderInstance.OLAPDSN = inst.OLAPDSN
		olderInstance.OLAPDriver = inst.OLAPDriver
		olapConn = olap
	} else {
		// getting exisiting connection
		olapConn, err = r.connCache.get(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
		if err != nil {
			return err
		}
	}

	// 2. embedCatalog disabled previously but enabled now
	if inst.EmbedCatalog && !olderInstance.EmbedCatalog {
		_, ok := olapConn.CatalogStore()
		if !ok {
			return errors.New("driver does not support embedded catalogs")
		}
	}
	olderInstance.EmbedCatalog = inst.EmbedCatalog

	// 3. changes in repo driver or repo dsn
	repoChanged := inst.RepoDriver != olderInstance.RepoDriver || inst.RepoDSN != olderInstance.RepoDSN
	if repoChanged {
		// evict exisiting connection
		r.connCache.evict(ctx, inst.ID, inst.RepoDriver, inst.RepoDSN)
		// Check repo connection
		repo, err := drivers.Open(inst.RepoDriver, inst.RepoDSN, r.logger)
		if err != nil {
			return err
		}
		_, ok := repo.RepoStore()
		if !ok {
			return fmt.Errorf("not a valid repo driver: '%s'", inst.RepoDriver)
		}
		// Prepare connections for use
		err = repo.Migrate(ctx)
		if err != nil {
			return fmt.Errorf("failed to prepare instance: %w", err)
		}
		olderInstance.RepoDSN = inst.RepoDSN
		olderInstance.RepoDriver = inst.RepoDriver
	}

	// 4. changes in env variables
	if !reflect.DeepEqual(olderInstance.Env, inst.Env) {
		// this is a hack to set allow_host_credentials
		// ideally the runtime should propagate this flag to connectors.Env
		if inst.Env == nil {
			inst.Env = make(map[string]string)
		}
		inst.Env["allow_host_credentials"] = strconv.FormatBool(r.opts.AllowHostCredentials)
		olderInstance.Env = inst.Env
	}
	// update the entire instance for now to avoid building complex queries
	return r.Registry().EditInstance(ctx, olderInstance)
}

func returnFirstErr(errs ...error) error {
	for _, r := range errs {
		if r != nil {
			return r
		}
	}
	return nil
}
