package runtime

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
)

func (r *Runtime) AcquireSystemHandle(ctx context.Context, connector string) (drivers.Handle, func(), error) {
	for _, c := range r.opts.SystemConnectors {
		if c.Name == connector {
			cfg := make(map[string]any, len(c.Config)+1)
			for k, v := range c.Config {
				cfg[strings.ToLower(k)] = v
			}
			cfg["allow_host_access"] = r.opts.AllowHostAccess
			return r.connCache.get(ctx, "", c.Type, cfg, true)
		}
	}
	return nil, nil, fmt.Errorf("connector %s doesn't exist", connector)
}

// AcquireHandle returns instance specific handle
func (r *Runtime) AcquireHandle(ctx context.Context, instanceID, connector string) (drivers.Handle, func(), error) {
	driver, cfg, err := r.connectorConfig(ctx, instanceID, connector)
	if err != nil {
		return nil, nil, err
	}
	return r.connCache.get(ctx, instanceID, driver, cfg, false)
}

// EvictHandle flushes the db handle for the specific connector from the cache
func (r *Runtime) EvictHandle(ctx context.Context, instanceID, connector string, drop bool) error {
	driver, cfg, err := r.connectorConfig(ctx, instanceID, connector)
	if err != nil {
		return err
	}
	r.connCache.evict(ctx, instanceID, driver, cfg)
	if drop {
		return drivers.Drop(driver, cfg, r.logger)
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

	conn, release, err := r.AcquireHandle(ctx, instanceID, inst.RepoConnector)
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

	conn, release, err := r.AcquireHandle(ctx, instanceID, inst.OLAPConnector)
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
		conn, release, err := r.AcquireHandle(ctx, instanceID, inst.OLAPConnector)
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

func (r *Runtime) connectorConfig(ctx context.Context, instanceID, name string) (string, map[string]any, error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return "", nil, err
	}

	// Search for connector definition in instance
	var connector *runtimev1.Connector
	for _, c := range inst.Connectors {
		if c.Name == name {
			connector = c
			break
		}
	}

	// Search for connector definition in rill.yaml
	if connector == nil {
		for _, c := range inst.ProjectConnectors {
			if c.Name == name {
				connector = c
				break
			}
		}
	}

	// Search for implicit connectors (where the name matches a driver name)
	if connector == nil {
		_, ok := drivers.Drivers[name]
		if ok {
			connector = &runtimev1.Connector{
				Type: name,
				Name: name,
			}
		}
	}

	// Return if search for connector was unsuccessful
	if connector == nil {
		return "", nil, fmt.Errorf("unknown connector %q", name)
	}

	// Build connector config
	cfg := make(map[string]any)

	// Apply config from definition
	for key, value := range connector.Config {
		cfg[strings.ToLower(key)] = value
	}

	// Instance variables matching the format "connector.name.var" are applied to the connector config
	vars := inst.ResolveVariables()
	prefix := fmt.Sprintf("connector.%s.", name)
	for k, v := range vars {
		if after, found := strings.CutPrefix(k, prefix); found {
			cfg[strings.ToLower(after)] = v
		}
	}

	// For backwards compatibility, certain root-level variables apply to certain implicit connectors.
	// NOTE: This switches on connector.Name, not connector.Type, because this only applies to implicit connectors.
	switch connector.Name {
	case "s3":
		setIfNil(cfg, "aws_access_key_id", vars["aws_access_key_id"])
		setIfNil(cfg, "aws_secret_access_key", vars["aws_secret_access_key"])
		setIfNil(cfg, "aws_session_token", vars["aws_session_token"])
	case "gcs":
		setIfNil(cfg, "google_application_credentials", vars["google_application_credentials"])
	case "bigquery":
		setIfNil(cfg, "google_application_credentials", vars["google_application_credentials"])
	case "motherduck":
		setIfNil(cfg, "token", vars["token"])
		setIfNil(cfg, "dsn", "")
	}

	// Apply built-in connector config
	cfg["allow_host_access"] = r.opts.AllowHostAccess

	// The "local_file" connector needs to know the repo root.
	// TODO: This is an ugly hack. But how can we get rid of it?
	if connector.Name == "local_file" {
		if inst.RepoConnector != "local_file" { // The RepoConnector shouldn't be named "local_file", but let's still try to avoid infinite recursion
			repo, release, err := r.Repo(ctx, instanceID)
			if err != nil {
				return "", nil, err
			}
			cfg["dsn"] = repo.Root()
			release()
		}
	}

	return connector.Type, cfg, nil
}

func setIfNil(m map[string]any, key string, value any) {
	if _, ok := m[key]; !ok {
		m[key] = value
	}
}
