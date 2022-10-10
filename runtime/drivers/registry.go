package drivers

import (
	"context"
	"time"
)

// RegistryStore is implemented by drivers capable of storing and looking up instances and repos
type RegistryStore interface {
	FindInstances(ctx context.Context) []*Instance
	FindInstance(ctx context.Context, id string) (*Instance, bool)
	CreateInstance(ctx context.Context, instance *Instance) error
	DeleteInstance(ctx context.Context, id string) error
	FindRepos(ctx context.Context) []*Repo
	FindRepo(ctx context.Context, id string) (*Repo, bool)
	CreateRepo(ctx context.Context, repo *Repo) error
	DeleteRepo(ctx context.Context, id string) error
}

// Instance represents one deployment of an OLAP store
type Instance struct {
	// Identifier
	ID string
	// Driver is the driver of the OLAP store to connect to ("druid" or "duckdb" currently)
	Driver string
	// DSN is the connection string for the OLAP store
	DSN string
	// ObjectPrefix will be prepended to tables and views created in the OLAP store through Rill SQL.
	// You can view it as a simpler alternative to using different database schemas.
	ObjectPrefix string `db:"object_prefix"`
	// Exposed indicates that the underlying OLAP infra may be manipulated directly by users.
	// If true, the runtime will continuously poll the infra's information schema to discover tables not created through the runtime.
	Exposed bool
	// EmbedCatalog tells the runtime whether to store the instance's catalog data (such as sources and metrics views)
	// directly in the OLAP datastore instead of in the runtime's metadata store.
	EmbedCatalog bool `db:"embed_catalog"`
	// CreatedOn is when the instance was created
	CreatedOn time.Time `db:"created_on"`
	// UpdatedOn is when the instance was last updated in the registry
	UpdatedOn time.Time `db:"updated_on"`
}

// Repo represents a file artifact store (either a folder on disk or virtualized in a database)
type Repo struct {
	ID        string
	Driver    string
	DSN       string
	CreatedOn time.Time `db:"created_on"`
	UpdatedOn time.Time `db:"updated_on"`
}
