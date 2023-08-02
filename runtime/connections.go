package runtime

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
)

func (r *Runtime) Registry() drivers.RegistryStore {
	registry, ok := r.metastore.AsRegistry()
	if !ok {
		// Verified as registry in New, so this should never happen
		panic("metastore is not a registry")
	}
	return registry
}

func (r *Runtime) AcquireHandle(ctx context.Context, instanceID string, connector string) (drivers.Handle, func(), error) {
	repo, release, err := r.Repo(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}
	defer release()

	// TODO :: do we have to parse yaml again and again
	yaml, err := rillv1.ParseRillYAML(ctx, repo, instanceID)
	if err != nil {
		return nil, nil, err
	}

	for _, c := range yaml.Connectors {
		if c.Name == connector {
			// TODO :: check this ?
			dsn := c.Defaults["dsn"]
			return r.connCache.get(ctx, instanceID, c.Type, dsn, false)
		}
	}
	return nil, nil, fmt.Errorf("connector %s doesn't exist", connector)
}

func (r *Runtime) Repo(ctx context.Context, instanceID string) (drivers.RepoStore, func(), error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	conn, release, err := r.connCache.get(ctx, instanceID, inst.RepoDriver, inst.RepoDSN, false)
	if err != nil {
		return nil, nil, err
	}

	repo, ok := conn.AsRepoStore(instanceID)
	if !ok {
		release()
		// Verified as repo when instance is created, so this should never happen
		return nil, release, fmt.Errorf("connection for instance '%s' is not a repo", instanceID)
	}

	return repo, release, nil
}

func (r *Runtime) OLAP(ctx context.Context, instanceID string) (drivers.OLAPStore, func(), error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	conn, release, err := r.connCache.get(ctx, instanceID, inst.OLAPDriver, inst.OLAPDSN, false)
	if err != nil {
		return nil, nil, err
	}

	olap, ok := conn.AsOLAP(instanceID)
	if !ok {
		release()
		// Verified as OLAP when instance is created, so this should never happen
		return nil, nil, fmt.Errorf("connection for instance '%s' is not an olap", instanceID)
	}

	return olap, release, nil
}

func (r *Runtime) Catalog(ctx context.Context, instanceID string) (drivers.CatalogStore, func(), error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	if inst.EmbedCatalog {
		conn, release, err := r.connCache.get(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN, false)
		if err != nil {
			return nil, nil, err
		}

		store, ok := conn.AsCatalogStore(instanceID)
		if !ok {
			release()
			// Verified as CatalogStore when instance is created, so this should never happen
			return nil, nil, fmt.Errorf("instance cannot embed catalog")
		}

		return store, release, nil
	}

	store, ok := r.metastore.AsCatalogStore(instanceID)
	if !ok {
		return nil, nil, fmt.Errorf("metastore cannot serve as catalog")
	}
	return store, func() {}, nil
}

func (r *Runtime) NewCatalogService(ctx context.Context, instanceID string) (*catalog.Service, error) {
	// get all stores
	olapStore, releaseOLAP, err := r.OLAP(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	catalogStore, releaseCatalog, err := r.Catalog(ctx, instanceID)
	if err != nil {
		releaseOLAP()
		return nil, err
	}

	repoStore, releaseRepo, err := r.Repo(ctx, instanceID)
	if err != nil {
		releaseOLAP()
		releaseCatalog()
		return nil, err
	}

	registry := r.Registry()

	migrationMetadata := r.migrationMetaCache.get(instanceID)
	releaseFunc := func() {
		releaseOLAP()
		releaseCatalog()
		releaseRepo()
	}
	return catalog.NewService(catalogStore, repoStore, olapStore, registry, instanceID, r.logger, migrationMetadata, releaseFunc), nil
}
