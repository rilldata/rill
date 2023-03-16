package drivers

import (
	"context"
	"time"
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
	// Driver to connect to for OLAP (options: duckdb, druid)
	OLAPDriver string
	// DSN for connection to OLAP
	OLAPDSN string
	// Driver for reading/editing code artifacts (options: file, metastore)
	RepoDriver string
	// DSN for connecting to repo
	RepoDSN string
	// EmbedCatalog tells the runtime to store the instance's catalog in its OLAP store instead
	// of in the runtime's metadata store. Currently only supported for the duckdb driver.
	EmbedCatalog bool `db:"embed_catalog"`
	// CreatedOn is when the instance was created
	CreatedOn time.Time `db:"created_on"`
	// UpdatedOn is when the instance was last updated in the registry
	UpdatedOn time.Time `db:"updated_on"`
	// Variables contains user-provided variables
	Variables map[string]string `db:"variables"`
	// ProjectVariables contains default variables from rill.yaml
	// (NOTE: This can always be reproduced from rill.yaml, so it's really just a handy cache of the values.)
	ProjectVariables map[string]string `db:"project_variables"`
	// IngestionLimitBytes is total data allowed to ingest across all sources
	// 0 means there is no limit
	IngestionLimitBytes int64 `db:"ingestion_limit_in_bytes"`
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
