package runtime

import (
	"context"
	"fmt"
	"maps"
	"path/filepath"
	"strconv"
	"strings"

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
			return r.getConnection(ctx, "", c.Type, cfg)
		}
	}
	return nil, nil, fmt.Errorf("connector %s doesn't exist", connector)
}

// AcquireHandle returns instance specific handle
func (r *Runtime) AcquireHandle(ctx context.Context, instanceID, connector string) (drivers.Handle, func(), error) {
	cfg, err := r.ConnectorConfig(ctx, instanceID, connector)
	if err != nil {
		return nil, nil, err
	}
	if ctx.Err() != nil {
		// Many code paths around connection acquisition leverage caches that won't actually touch the ctx.
		// So we take this moment to make sure the ctx gets checked for cancellation at least every once in a while.
		return nil, nil, ctx.Err()
	}
	return r.getConnection(ctx, instanceID, cfg.Driver, cfg.Resolve())
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

// OLAP returns a handle for an OLAP data store.
// The connector argument is optional. If not provided, the instance's default OLAP connector is used.
func (r *Runtime) OLAP(ctx context.Context, instanceID, connector string) (drivers.OLAPStore, func(), error) {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	if connector == "" {
		connector = inst.ResolveOLAPConnector()
	}

	conn, release, err := r.AcquireHandle(ctx, instanceID, connector)
	if err != nil {
		return nil, nil, err
	}

	olap, ok := conn.AsOLAP(instanceID)
	if !ok {
		release()
		return nil, nil, fmt.Errorf("connector %q is not a valid OLAP data store", connector)
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

func (r *Runtime) ConnectorConfig(ctx context.Context, instanceID, name string) (*ConnectorConfig, error) {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	res := &ConnectorConfig{}

	// Search for connector definition in instance
	for _, c := range inst.Connectors {
		if c.Name == name {
			res.Driver = c.Type
			res.Preset = maps.Clone(c.Config) // Cloning because Preset may be mutated later, but the inst object is shared.
			break
		}
	}

	// Search for connector definition in rill.yaml
	for _, c := range inst.ProjectConnectors {
		if c.Name == name {
			res.Driver = c.Type
			res.Project = maps.Clone(c.Config) // Cloning because Project may be mutated later, but the inst object is shared.
			break
		}
	}

	// Search for implicit connectors (where the name matches a driver name)
	if res.Driver == "" {
		_, ok := drivers.Drivers[name]
		if ok {
			res.Driver = name
		}
	}

	// Return if search for connector driver was unsuccessful
	if res.Driver == "" {
		return nil, fmt.Errorf("unknown connector %q", name)
	}

	// Build res.Env config based on instance variables matching the format "connector.name.var"
	vars := inst.ResolveVariables()
	prefix := fmt.Sprintf("connector.%s.", name)
	for k, v := range vars {
		if after, found := strings.CutPrefix(k, prefix); found {
			if res.Env == nil {
				res.Env = make(map[string]string)
			}
			res.Env[after] = v
		}
	}

	// For backwards compatibility, certain root-level variables apply to certain implicit connectors.
	// NOTE: This switches on connector.Name, not connector.Type, because this only applies to implicit connectors.
	switch name {
	case "s3", "athena", "redshift":
		res.setPreset("aws_access_key_id", vars["aws_access_key_id"], false)
		res.setPreset("aws_secret_access_key", vars["aws_secret_access_key"], false)
		res.setPreset("aws_session_token", vars["aws_session_token"], false)
	case "azure":
		res.setPreset("azure_storage_account", vars["azure_storage_account"], false)
		res.setPreset("azure_storage_key", vars["azure_storage_key"], false)
		res.setPreset("azure_storage_sas_token", vars["azure_storage_sas_token"], false)
		res.setPreset("azure_storage_connection_string", vars["azure_storage_connection_string"], false)
	case "gcs":
		res.setPreset("google_application_credentials", vars["google_application_credentials"], false)
	case "bigquery":
		res.setPreset("google_application_credentials", vars["google_application_credentials"], false)
	case "motherduck":
		res.setPreset("token", vars["token"], false)
		res.setPreset("dsn", "", true)
	case "local_file":
		// The "local_file" connector needs to know the repo root.
		// TODO: This is an ugly hack. But how can we get rid of it?
		if inst.RepoConnector != "local_file" { // The RepoConnector shouldn't be named "local_file", but let's still try to avoid infinite recursion
			repo, release, err := r.Repo(ctx, instanceID)
			if err != nil {
				return nil, err
			}
			res.setPreset("dsn", repo.Root(), true)
			release()
		}
	}

	// Apply built-in system-wide config
	res.setPreset("allow_host_access", strconv.FormatBool(r.opts.AllowHostAccess), true)
	// data_dir stores persistent data
	res.setPreset("data_dir", filepath.Join(r.opts.DataDir, instanceID, name), true)
	// temp_dir stores temporary data. The logic that creates any temporary file here should also delete them.
	// The contents will also be deleted on runtime restarts.
	res.setPreset("temp_dir", filepath.Join(r.opts.DataDir, instanceID, "tmp"), true)

	// Done
	return res, nil
}

// ConnectorConfig holds and resolves connector configuration.
// We support three levels of configuration:
// 1. Preset: provided when creating the instance (or set by the system, such as allow_host_access). Cannot be overridden.
// 2. Project: defined in the rill.yaml file. Can be overridden by the env.
// 3. Env: defined in the instance's variables (in the format "connector.name.var").
type ConnectorConfig struct {
	Driver  string
	Preset  map[string]string
	Project map[string]string
	Env     map[string]string
}

// Resolve returns the final resolved connector configuration.
// It guarantees that all keys in the result are lowercase.
func (c *ConnectorConfig) Resolve() map[string]any {
	n := len(c.Preset) + len(c.Project) + len(c.Env)
	if n == 0 {
		return nil
	}

	cfg := make(map[string]any, n)
	for k, v := range c.Project {
		cfg[strings.ToLower(k)] = v
	}
	for k, v := range c.Env {
		cfg[strings.ToLower(k)] = v
	}
	for k, v := range c.Preset {
		cfg[strings.ToLower(k)] = v
	}
	return cfg
}

// ResolveString is similar to Resolve, but it returns a map of strings.
func (c *ConnectorConfig) ResolveStrings() map[string]string {
	n := len(c.Preset) + len(c.Project) + len(c.Env)
	if n == 0 {
		return nil
	}

	cfg := make(map[string]string, n)
	for k, v := range c.Project {
		cfg[strings.ToLower(k)] = v
	}
	for k, v := range c.Env {
		cfg[strings.ToLower(k)] = v
	}
	for k, v := range c.Preset {
		cfg[strings.ToLower(k)] = v
	}
	return cfg
}

// setPreset sets a preset value.
// If the provided value is empty, it will not be added unless force is true.
func (c *ConnectorConfig) setPreset(k, v string, force bool) {
	if v == "" && !force {
		return
	}
	if c.Preset == nil {
		c.Preset = make(map[string]string)
	}
	c.Preset[k] = v
}
