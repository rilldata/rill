package drivers

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// RegistryStore is implemented by drivers capable of storing and looking up instances and repos.
type RegistryStore interface {
	FindInstances(ctx context.Context) ([]*Instance, error)
	FindInstance(ctx context.Context, id string) (*Instance, error)
	CreateInstance(ctx context.Context, instance *Instance) error
	DeleteInstance(ctx context.Context, id string) error
	EditInstance(ctx context.Context, instance *Instance) error
}

// Instance represents a single data project, meaning one OLAP connection, one repo connection,
// and one catalog connection.
type Instance struct {
	// Identifier
	ID string
	// Environment is the environment that the instance represents
	Environment string
	// Driver name to connect to for OLAP
	OLAPConnector string
	// ProjectOLAPConnector is an override of OLAPConnector that may be set in rill.yaml.
	// NOTE: Hopefully we can merge this with OLAPConnector if/when we remove the ability to set OLAPConnector using flags.
	ProjectOLAPConnector string
	// Driver name for reading/editing code artifacts
	RepoConnector string
	// Driver name for the admin service managing the deployment (optional)
	AdminConnector string
	// Driver name for the AI service (optional)
	AIConnector string
	// Driver name for catalog
	CatalogConnector string
	// CreatedOn is when the instance was created
	CreatedOn time.Time `db:"created_on"`
	// UpdatedOn is when the instance was last updated in the registry
	UpdatedOn time.Time `db:"updated_on"`
	// Instance specific connectors
	Connectors []*runtimev1.Connector `db:"connectors"`
	// ProjectVariables contains default connectors from rill.yaml
	ProjectConnectors []*runtimev1.Connector `db:"project_connectors"`
	// Variables contains user-provided variables
	Variables map[string]string `db:"variables"`
	// ProjectVariables contains default variables from rill.yaml
	// (NOTE: This can always be reproduced from rill.yaml, so it's really just a handy cache of the values.)
	ProjectVariables map[string]string `db:"project_variables"`
	// Annotations to enrich activity events (like usage tracking)
	Annotations map[string]string
	// EmbedCatalog tells the runtime to store the instance's catalog in its OLAP store instead
	// of in the runtime's metadata store. Currently only supported for the duckdb driver.
	EmbedCatalog bool `db:"embed_catalog"`
	// WatchRepo indicates whether to watch the repo for file changes and reconcile them automatically.
	WatchRepo bool `db:"watch_repo"`
	// IgnoreInitialInvalidProjectError indicates whether to ignore an invalid project error when the instance is initially created.
	IgnoreInitialInvalidProjectError bool `db:"-"`
}

// InstanceConfig contains dynamic configuration for an instance.
// It is configured by parsing instance variables prefixed with "rill.".
// For example, a variable "rill.stage_changes=true" would set the StageChanges field to true.
// InstanceConfig should only be used for config that the user is allowed to change dynamically at runtime.
type InstanceConfig struct {
	// StageChanges indicates whether to keep previously ingested tables for sources/models, and only override them if ingestion of a new table is successful.
	StageChanges bool `mapstructure:"rill.stage_changes"`
	// ModelDefaultMaterialize indicates whether to materialize models by default.
	ModelDefaultMaterialize bool `mapstructure:"rill.model_default_materialize"`
	// ModelMaterializeDelaySeconds adds a delay before materializing models.
	ModelMaterializeDelaySeconds uint32 `mapstructure:"rill.model_materialize_delay_seconds"`
	// DownloadRowLimit is the row limit for interactive data exports. If set to 0, there is no limit.
	DownloadRowLimit int64 `mapstructure:"rill.download_row_limit"`
	// PivotCellLimit is the maximum number of cells allowed in a single pivot query.
	// Note that it does not limit the UI's pivot table because it paginates the requests.
	PivotCellLimit int64 `mapstructure:"rill.pivot_cell_limit"`
}

// ResolveOLAPConnector resolves the OLAP connector to default to for the instance.
func (i *Instance) ResolveOLAPConnector() string {
	if i.ProjectOLAPConnector != "" {
		return i.ProjectOLAPConnector
	}
	return i.OLAPConnector
}

// ResolveVariables returns the final resolved variables
func (i *Instance) ResolveVariables() map[string]string {
	r := make(map[string]string, len(i.ProjectVariables))

	// set ProjectVariables first i.e. Project defaults
	for k, v := range i.ProjectVariables {
		r[k] = v
	}

	// override with instance Variables
	for k, v := range i.Variables {
		r[k] = v
	}
	return r
}

// Config resolves the current dynamic config properties for the instance.
// See InstanceConfig for details.
func (i *Instance) Config() (InstanceConfig, error) {
	// Default config
	res := InstanceConfig{
		StageChanges:                 true,
		ModelDefaultMaterialize:      false,
		ModelMaterializeDelaySeconds: 0,
		DownloadRowLimit:             10_000,
		PivotCellLimit:               1_000_000,
	}

	// Backwards compatibility: Use "__materialize_default" as alias for "rill.model_default_materialize".
	if i.Variables != nil {
		if v, ok := i.Variables["__materialize_default"]; ok {
			if _, ok := i.Variables["rill.model_default_materialize"]; !ok {
				i.Variables["rill.model_default_materialize"] = v
			}
		}
	}

	// Decode variables into config.
	// Minor optimization: Instead of calling ResolveVariables (and allocating a new map), just call Decode in on the underlying variable maps in the same order as in ResolveVariables.
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &res,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return InstanceConfig{}, fmt.Errorf("failed to parse instance config: %w", err)
	}
	err = dec.Decode(i.ProjectVariables)
	if err != nil {
		return InstanceConfig{}, fmt.Errorf("failed to parse instance config: %w", err)
	}
	err = dec.Decode(i.Variables)
	if err != nil {
		return InstanceConfig{}, fmt.Errorf("failed to parse instance config: %w", err)
	}

	return res, nil
}
