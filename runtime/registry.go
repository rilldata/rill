package runtime

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/util"
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

	// Check repo connection
	repo, repoStore, err := r.checkRepoConnection(inst)
	if err != nil {
		return err
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
	var dropErrors []error
	if dropDB {
		// drop all ingested data
		// should we insert complete data in one schema so it will be easy to drop the schema ?
		// NOTE : Database file does not free space when tables are dropped issue #1099 in duckdb
		entries := svc.Catalog.FindEntries(ctx, inst.ID, drivers.ObjectTypeUnspecified)
		for _, entry := range entries {
			if entry.Type == drivers.ObjectTypeTable {
				table := entry.GetTable()
				if table.Managed {
					dropErrors = append(dropErrors, svc.Olap.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP TABLE %s", table.Name)}))
				}
			}
		}
		// drop rill schema which deletes embedded catalog as well
		dropErrors = append(dropErrors, svc.Olap.Exec(ctx, &drivers.Statement{Query: "DROP schema rill CASCADE"}))
	}
	// delete instance related data if catalog is not embedded
	if !inst.EmbedCatalog {
		dropErrors = append(dropErrors, svc.Catalog.DeleteInstanceEntries(ctx, instanceID))
	}

	// evict caches
	r.evictCaches(ctx, inst)

	// delete instance
	err4 := r.Registry().DeleteInstance(ctx, instanceID)
	return util.ReturnFirstErr(util.ReturnFirstErr(err1, err4), util.ReturnFirstErr(dropErrors...))
}

// EditInstance edits exisiting instance.
// Confirming to put api specs, it is expected to send entire existing instance data.
// The API compares and only evicts caches if relevant drivers or dsn is changed.
// This is done to ensure that db handlers are not unnecessarily closed
func (r *Runtime) EditInstance(ctx context.Context, inst *drivers.Instance) error {
	olderInstance, err := r.Registry().FindInstance(ctx, inst.ID)
	if err != nil {
		return err
	}

	// 1. changes in olap driver or olap dsn
	var olapConn drivers.Connection
	olapChanged := olderInstance.OLAPDriver != inst.OLAPDriver || olderInstance.OLAPDSN != inst.OLAPDSN
	if olapChanged {
		// Check OLAP connection
		olap, _, err := r.checkOlapConnection(inst)
		if err != nil {
			return err
		}

		// Prepare connections for use
		err = olap.Migrate(ctx)
		if err != nil {
			return fmt.Errorf("failed to prepare instance: %w", err)
		}
		olapConn = olap
	}

	// 2. embedCatalog disabled previously but enabled now
	if inst.EmbedCatalog {
		if olapConn == nil {
			// getting exisiting connection
			olapConn, err = r.connCache.get(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
			if err != nil {
				return err
			}
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

	// update env variables
	if inst.Env == nil {
		inst.Env = make(map[string]string)
	}
	inst.Env["allow_host_credentials"] = strconv.FormatBool(r.opts.AllowHostCredentials)
	// update the entire instance for now to avoid building queries in some complicated way
	return r.Registry().EditInstance(ctx, inst)
}

// TODO :: this is a rudimentary solution and ideally should be done in some producer/consumer pattern
func (r *Runtime) evictCaches(ctx context.Context, inst *drivers.Instance) {
	// evict and close exisiting connection
	r.connCache.evict(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
	r.connCache.evict(ctx, inst.ID, inst.RepoDriver, inst.RepoDSN)

	// evict catalog cache
	r.catalogCache.evict(ctx, inst.ID)
	// query cache can't be evicted since key is a combination of instance ID and other parameters
}

func (r *Runtime) checkRepoConnection(inst *drivers.Instance) (drivers.Connection, drivers.RepoStore, error) {
	// Check repo connection
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
	// Check repo connection
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
