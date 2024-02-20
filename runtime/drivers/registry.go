package drivers

import (
	"context"
	"time"

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
	// StageChanges indicates whether to stage source and model changes before applying them.
	StageChanges bool `db:"stage_changes"`
	// ModelDefaultMaterialize indicates whether to materialize models by default.
	ModelDefaultMaterialize bool `db:"model_default_materialize"`
	// ModelMaterializeDelaySeconds adds a delay before materializing models.
	ModelMaterializeDelaySeconds uint32 `db:"model_materialize_delay_seconds"`
	// IgnoreInitialInvalidProjectError indicates whether to ignore an invalid project error when the instance is initially created.
	IgnoreInitialInvalidProjectError bool `db:"-"`
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
