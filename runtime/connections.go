package runtime

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
)

var ErrAdminNotConfigured = fmt.Errorf("an admin service is not configured for this instance")

var ErrAINotConfigured = fmt.Errorf("an AI service is not configured for this instance")

func (r *Runtime) AcquireSystemHandle(ctx context.Context, connector string) (drivers.Handle, func(), error) {
	for _, c := range r.opts.SystemConnectors {
		if c.Name != connector {
			continue
		}
		raw := make(map[string]any)
		if c.Config != nil {
			raw = c.Config.AsMap()
		}
		cfg := make(map[string]any, len(raw)+1)
		for k, v := range raw {
			cfg[strings.ToLower(k)] = v
		}
		cfg["allow_host_access"] = r.opts.AllowHostAccess

		return r.getConnection(ctx, cachedConnectionConfig{
			instanceID: "",
			name:       connector,
			driver:     c.Type,
			config:     cfg,
		})
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
	return r.getConnection(ctx, cachedConnectionConfig{
		instanceID:    instanceID,
		name:          connector,
		driver:        cfg.Driver,
		config:        cfg.Resolve(),
		provision:     cfg.Provision,
		provisionArgs: cfg.ProvisionArgs,
	})
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

	aiConnector := inst.ResolveAIConnector()
	// The AI connector is optional
	if aiConnector == "" {
		return nil, nil, ErrAINotConfigured
	}

	conn, release, err := r.AcquireHandle(ctx, instanceID, aiConnector)
	if err != nil {
		return nil, nil, err
	}

	ai, ok := conn.AsAI(instanceID)
	if !ok {
		release()
		return nil, nil, fmt.Errorf("connector %q is not a valid AI service", aiConnector)
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
			if c.Config != nil {
				res.Preset = c.Config.AsMap()
			}
			if c.Provision {
				res.Provision = c.Provision
				res.ProvisionArgs = c.ProvisionArgs.AsMap()
			}
			break
		}
	}

	// Search for connector definitions from YAML files
	for _, c := range inst.ProjectConnectors {
		if c.Name != name {
			continue
		}

		res.Driver = c.Type
		res.Project, err = resolveConnectorProperties(inst.Environment, inst.ResolveVariables(false), c)
		if err != nil {
			return nil, err
		}
		if c.Provision {
			res.Provision = c.Provision
			res.ProvisionArgs = c.ProvisionArgs.AsMap()
		}

		break
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
	vars := inst.ResolveVariables(true)
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
			rootPath, err := repo.Root(ctx)
			if err != nil {
				release()
				return nil, fmt.Errorf("failed to get root path: %w", err)
			}
			res.setPreset("dsn", rootPath, true)
			release()
		}
	}

	// Apply built-in system-wide config
	res.setPreset("allow_host_access", strconv.FormatBool(r.opts.AllowHostAccess), true)

	// Done
	return res, nil
}

// resolveConnectorProperties resolves templating in the provided connector's properties.
// It always returns a clone of the properties, even if no templating is found, so the output is safe for further mutations.
func resolveConnectorProperties(environment string, vars map[string]string, c *runtimev1.Connector) (map[string]any, error) {
	if c.Config == nil {
		return make(map[string]any), nil
	}
	res := c.Config.AsMap()

	td := parser.TemplateData{
		Environment: environment,
		Variables:   vars,
	}

	for _, k := range c.TemplatedProperties {
		v, ok := res[k]
		if !ok {
			continue
		}
		v, err := parser.ResolveTemplateRecursively(v, td, true)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve template: %w", err)
		}
		res[k] = v
	}

	return res, nil
}

// ConnectorConfig holds and resolves connector configuration.
// We support three levels of configuration:
// 1. Preset: provided when creating the instance (or set by the system, such as allow_host_access). Cannot be overridden.
// 2. Project: defined in the rill.yaml file. Can be overridden by the env.
// 3. Env: defined in the instance's variables (in the format "connector.name.var").
type ConnectorConfig struct {
	Driver  string
	Preset  map[string]any
	Project map[string]any
	Env     map[string]string
	// Provision will cause it to request the admin service to provision the connector.
	Provision bool
	// ProvisionArgs provide provisioning args for when ProvisionName is set.
	ProvisionArgs map[string]any
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

// setPreset sets a preset value.
// If the provided value is empty, it will not be added unless force is true.
func (c *ConnectorConfig) setPreset(k, v string, force bool) {
	if v == "" && !force {
		return
	}
	if c.Preset == nil {
		c.Preset = make(map[string]any)
	}
	c.Preset[k] = v
}
