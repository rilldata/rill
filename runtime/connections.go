package runtime

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
)

func (r *Runtime) Registry() drivers.RegistryStore {
	registry, _ := r.metastore.RegistryStore() // Verified as registry in New
	return registry
}

func (r *Runtime) Repo(ctx context.Context, instanceID string) (drivers.RepoStore, error) {
	inst, found := r.FindInstance(ctx, instanceID)
	if !found {
		return nil, ErrInstanceNotFound
	}

	conn, err := r.connCache.get(ctx, instanceID, inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, err
	}

	repo, _ := conn.RepoStore() // Verified as repo when instance is created

	return repo, nil
}

func (r *Runtime) OLAP(ctx context.Context, instanceID string) (drivers.OLAPStore, error) {
	inst, found := r.FindInstance(ctx, instanceID)
	if !found {
		return nil, ErrInstanceNotFound
	}

	conn, err := r.connCache.get(ctx, instanceID, inst.OLAPDriver, inst.OLAPDSN)
	if err != nil {
		return nil, err
	}

	olap, _ := conn.OLAPStore() // Verified as OLAP when instance is created

	return olap, nil
}

func (r *Runtime) Catalog(ctx context.Context, instanceID string) (*catalog.Service, error) {
	c, err := r.catalogCache.get(ctx, r, instanceID)
	if err != nil {
		return nil, err
	}

	return c, err
}
