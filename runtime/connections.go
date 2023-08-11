package runtime

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
)

func (r *Runtime) AcquireGlobalHandle(ctx context.Context, connector string) (drivers.Handle, func(), error) {
	for _, c := range r.opts.SystemConnectors {
		if c.Name == connector {
			return r.connCache.get(ctx, "", c.Type, r.connectorConfig(r.opts.MetastoreDriver, c.Config, nil), true)
		}
	}
	return nil, nil, fmt.Errorf("connector %s doesn't exist", connector)
}

// AcquireHandle returns instance specific handle
func (r *Runtime) AcquireHandle(ctx context.Context, instanceID, connector string) (drivers.Handle, func(), error) {
	instance, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	if c, err := r.connectorDef(instance, connector); err == nil {
		return r.connCache.get(ctx, instanceID, c.Type, r.connectorConfig(connector, c.Config, instance.ResolveVariables()), false)
	}

	// neither defined in rill.yaml nor set in instance, directly used in source
	return r.connCache.get(ctx, instanceID, connector, r.connectorConfig(connector, nil, instance.ResolveVariables()), false)
}

// FlushHandle flushes the db handle for the specific connector from the cache
func (r *Runtime) FlushHandle(ctx context.Context, instanceID, connector string, drop bool) error {
	instance, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return err
	}

	var driverType string
	var connectorConfig map[string]any
	if c, err := r.connectorDef(instance, connector); err == nil {
		driverType = c.Type
		connectorConfig = r.connectorConfig(connector, c.Config, instance.ResolveVariables())
	} else {
		driverType = connector
		connectorConfig = r.connectorConfig(connector, nil, instance.ResolveVariables())
	}
	r.connCache.evict(ctx, instanceID, driverType, connectorConfig)
	if drop {
		return drivers.Drop(driverType, connectorConfig, r.logger)
	}
	return nil
}

func (r *Runtime) Registry() drivers.RegistryStore {
	registry, ok := r.metastore.AsRegistry()
	if !ok {
		// Verified as registry in New, so this should never happen
		panic("metastore is not a registry")
	}
	return registry
}

func (r *Runtime) Repo(ctx context.Context, instanceID string) (drivers.RepoStore, func(), error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	conn, release, err := r.AcquireHandle(ctx, instanceID, inst.RepoDriver)
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

	conn, release, err := r.AcquireHandle(ctx, instanceID, inst.OLAPDriver)
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
		c, err := r.connectorDef(inst, inst.OLAPDriver)
		if err != nil {
			return nil, nil, err
		}
		conn, release, err := r.connCache.get(ctx, instanceID, c.Type, r.connectorConfig(inst.OLAPDriver, c.Config, inst.ResolveVariables()), false)
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

func (r *Runtime) connectorDef(inst *drivers.Instance, name string) (*runtimev1.Connector, error) {
	for _, c := range inst.Connectors {
		// set in instance
		if c.Name == name {
			return c, nil
		}
	}

	// defined in rill.yaml
	for _, c := range inst.ProjectConnectors {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, fmt.Errorf("connector %s doesn't exist", name)
}

// TODO :: these can also be generated during reconcile itself ?
func (r *Runtime) connectorConfig(name string, def, variables map[string]string) map[string]any {
	vars := make(map[string]any, 0)
	for key, value := range def {
		vars[strings.ToLower(key)] = value
	}

	// connector variables are of format connector.name.var
	prefix := fmt.Sprintf("connector.%s.", name)
	for key, value := range variables {
		if strings.EqualFold(key, "allow_host_access") { // global variable
			vars[strings.ToLower(key)] = value
		} else if after, found := strings.CutPrefix(key, prefix); found { // connector specific variable
			vars[strings.ToLower(after)] = value
		}
	}
	vars["allow_host_access"] = r.opts.AllowHostAccess
	return vars
}
