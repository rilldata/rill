package runtime

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var ErrAdminNotConfigured = fmt.Errorf("an admin service is not configured for this instance")

var ErrAINotConfigured = fmt.Errorf("an AI service is not configured for this instance")

func (r *Runtime) AcquireSystemHandle(ctx context.Context, connector string) (drivers.Handle, func(), error) {
	for _, c := range r.opts.SystemConnectors {
		if c.Name == connector {
			cfg := make(map[string]any, len(c.Config)+1)
			for k, v := range c.Config {
				cfg[strings.ToLower(k)] = v
			}
			cfg["allow_host_access"] = r.opts.AllowHostAccess
			return r.getConnection(ctx, "", c.Type, cfg, true)
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
	if ctx.Err() != nil {
		// Many code paths around connection acquisition leverage caches that won't actually touch the ctx.
		// So we take this moment to make sure the ctx gets checked for cancellation at least every once in a while.
		return nil, nil, ctx.Err()
	}
	return r.getConnection(ctx, instanceID, driver, cfg, false)
}

func (r *Runtime) Repo(ctx context.Context, instanceID string) (drivers.RepoStore, func(), error) {
	inst, err := r.Instance(ctx, instanceID)
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
		return nil, release, fmt.Errorf("connector %q is not a valid project file store", inst.RepoConnector)
	}

	return repo, release, nil
}

func (r *Runtime) Admin(ctx context.Context, instanceID string) (drivers.AdminService, func(), error) {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	// The admin connector is optional
	if inst.AdminConnector == "" {
		return nil, nil, ErrAdminNotConfigured
	}

	conn, release, err := r.AcquireHandle(ctx, instanceID, inst.AdminConnector)
	if err != nil {
		return nil, nil, err
	}

	admin, ok := conn.AsAdmin(instanceID)
	if !ok {
		release()
		return nil, nil, fmt.Errorf("connector %q is not a valid admin service", inst.AdminConnector)
	}

	return admin, release, nil
}

func (r *Runtime) AI(ctx context.Context, instanceID string) (drivers.AIService, func(), error) {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	// The AI connector is optional
	if inst.AIConnector == "" {
		return nil, nil, ErrAINotConfigured
	}

	conn, release, err := r.AcquireHandle(ctx, instanceID, inst.AIConnector)
	if err != nil {
		return nil, nil, err
	}

	ai, ok := conn.AsAI(instanceID)
	if !ok {
		release()
		return nil, nil, fmt.Errorf("connector %q is not a valid AI service", inst.AIConnector)
	}

	return ai, release, nil
}

func (r *Runtime) OLAP(ctx context.Context, instanceID string) (drivers.OLAPStore, func(), error) {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	conn, release, err := r.AcquireHandle(ctx, instanceID, inst.ResolveOLAPConnector())
	if err != nil {
		return nil, nil, err
	}

	olap, ok := conn.AsOLAP(instanceID)
	if !ok {
		release()
		return nil, nil, fmt.Errorf("connector %q is not a valid OLAP data store", inst.ResolveOLAPConnector())
	}

	return olap, release, nil
}

func (r *Runtime) Catalog(ctx context.Context, instanceID string) (drivers.CatalogStore, func(), error) {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	if inst.EmbedCatalog {
		conn, release, err := r.AcquireHandle(ctx, instanceID, inst.ResolveOLAPConnector())
		if err != nil {
			return nil, nil, err
		}

		store, ok := conn.AsCatalogStore(instanceID)
		if !ok {
			release()
			return nil, nil, fmt.Errorf("can't embed catalog because it is not supported by the connector %q", inst.ResolveOLAPConnector())
		}

		return store, release, nil
	}

	if inst.CatalogConnector == "" {
		store, ok := r.metastore.AsCatalogStore(instanceID)
		if !ok {
			return nil, nil, fmt.Errorf("metastore cannot serve as catalog")
		}
		return store, func() {}, nil
	}

	conn, release, err := r.AcquireHandle(ctx, instanceID, inst.CatalogConnector)
	if err != nil {
		return nil, nil, err
	}

	store, ok := conn.AsCatalogStore(instanceID)
	if !ok {
		release()
		return nil, nil, fmt.Errorf("connector %q is not a valid catalog store", inst.CatalogConnector)
	}

	return store, release, nil
}

func (r *Runtime) connectorConfig(ctx context.Context, instanceID, name string) (string, map[string]any, error) {
	inst, err := r.Instance(ctx, instanceID)
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
	case "s3", "athena":
		setIfNil(cfg, "aws_access_key_id", vars["aws_access_key_id"])
		setIfNil(cfg, "aws_secret_access_key", vars["aws_secret_access_key"])
		setIfNil(cfg, "aws_session_token", vars["aws_session_token"])
	case "azure":
		setIfNil(cfg, "azure_storage_account", vars["azure_storage_account"])
		setIfNil(cfg, "azure_storage_key", vars["azure_storage_key"])
		setIfNil(cfg, "azure_storage_sas_token", vars["azure_storage_sas_token"])
		setIfNil(cfg, "azure_storage_connection_string", vars["azure_storage_connection_string"])
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
