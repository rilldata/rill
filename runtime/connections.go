package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
)

func (r *Runtime) newMetaStore(ctx context.Context, instanceID string) (drivers.Handle, func(), error) {
	c, shared, err := r.opts.ConnectorDefByName(_metastoreDriverName)
	if err != nil {
		panic(err)
	}

	return r.connCache.get(ctx, instanceID, c.Type, c.Defaults, shared)
}

func (r *Runtime) Registry() drivers.RegistryStore {
	registry, ok := r.metastore.AsRegistry()
	if !ok {
		// Verified as registry in New, so this should never happen
		panic("metastore is not a registry")
	}
	return registry
}

func (r *Runtime) AcquireHandle(ctx context.Context, instanceID, connector string) (drivers.Handle, func(), error) {
	repo, release, err := r.Repo(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}
	defer release()

	// TODO :: parse it and keep a cache in instance to avoid reparsing
	yaml, err := rillv1.ParseRillYAML(ctx, repo, instanceID)
	if err != nil {
		return nil, nil, err
	}

	instance, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}
	// defined in rill.yaml
	for _, c := range yaml.Connectors {
		if c.Name == connector {
			return r.connCache.get(ctx, instanceID, c.Type, variables(connector, c.Defaults, instance.ResolveVariables()), false)
		}
	}
	if c, shared, err := r.opts.ConnectorDefByName(connector); err == nil { // connector found
		// defined in runtime options
		return r.connCache.get(ctx, instanceID, c.Type, variables(connector, c.Defaults, instance.ResolveVariables()), shared)
	}
	// neither defined in rill.yaml nor in runtime options, directly used in source
	return r.connCache.get(ctx, instanceID, connector, variables(connector, nil, instance.ResolveVariables()), false)
}

func (r *Runtime) Repo(ctx context.Context, instanceID string) (drivers.RepoStore, func(), error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	c, shared, err := r.opts.RepoDef(inst.RepoDSN)
	if err != nil {
		return nil, nil, err
	}
	conn, release, err := r.connCache.get(ctx, instanceID, inst.RepoDriver, variables(_repoDriverName, c.Defaults, inst.ResolveVariables()), shared)
	if err != nil {
		return nil, nil, err
	}

	repo, ok := conn.AsRepoStore()
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

	c, shared, err := r.opts.OLAPDef(inst.OLAPDSN)
	if err != nil {
		return nil, nil, err
	}
	conn, release, err := r.connCache.get(ctx, instanceID, inst.OLAPDriver, variables(_olapDriverName, c.Defaults, inst.ResolveVariables()), shared)
	if err != nil {
		return nil, nil, err
	}

	olap, ok := conn.AsOLAP()
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
		c, shared, err := r.opts.OLAPDef(inst.OLAPDSN)
		if err != nil {
			return nil, nil, err
		}
		conn, release, err := r.connCache.get(ctx, instanceID, inst.OLAPDriver, variables(_olapDriverName, c.Defaults, inst.ResolveVariables()), shared)
		if err != nil {
			return nil, nil, err
		}

		store, ok := conn.AsCatalogStore()
		if !ok {
			release()
			// Verified as CatalogStore when instance is created, so this should never happen
			return nil, nil, fmt.Errorf("instance cannot embed catalog")
		}

		return store, release, nil
	}

	store, ok := r.metastore.AsCatalogStore()
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

func variables(name string, def, variables map[string]string) map[string]string {
	vars := make(map[string]string, 0)
	for key, value := range def {
		vars[strings.ToLower(key)] = value
	}

	// connector variables are of format connector.name.var
	// there could also be other variables like allow_host_access, region etc which are global for all connectors
	prefix := fmt.Sprintf("connector.%s.", name)
	for key, value := range variables {
		if !strings.HasPrefix(key, "connector.") { // global variable
			vars[strings.ToLower(key)] = value
		} else if after, found := strings.CutPrefix(key, prefix); found { // connector specific variable
			vars[strings.ToLower(after)] = value
		}
	}
	return vars
}
