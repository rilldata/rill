package runtime

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
)

func (r *Runtime) Registry() drivers.RegistryStore {
	registry, ok := r.metastore.RegistryStore()
	if !ok {
		// Verified as registry in New, so this should never happen
		panic("metastore is not a registry")
	}
	return registry
}

func (r *Runtime) Repo(ctx context.Context, instanceID string) (drivers.RepoStore, error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	conn, err := r.connCache.get(ctx, instanceID, inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, err
	}

	repo, ok := conn.RepoStore()
	if !ok {
		// Verified as repo when instance is created, so this should never happen
		return nil, fmt.Errorf("connection for instance '%s' is not a repo", instanceID)
	}

	return repo, nil
}

func (r *Runtime) OLAP(ctx context.Context, instanceID string) (drivers.OLAPStore, error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	conn, err := r.connCache.get(ctx, instanceID, inst.OLAPDriver, inst.OLAPDSN)
	if err != nil {
		return nil, err
	}

	olap, ok := conn.OLAPStore()
	if !ok {
		// Verified as OLAP when instance is created, so this should never happen
		return nil, fmt.Errorf("connection for instance '%s' is not an olap", instanceID)
	}

	return olap, nil
}

func (r *Runtime) Catalog(ctx context.Context, instanceID string) (*catalog.Service, error) {
	c, err := r.catalogCache.get(ctx, r, instanceID)
	if err != nil {
		return nil, err
	}

	return c, err
}

func (r *Runtime) Close() error {
	c := r.connCache
	for _, key := range c.cache.Keys() {
		val, _ := c.cache.Get(key)
		err := val.(drivers.Connection).Close()
		if err != nil {
			return err
		}
	}
	return nil
}
