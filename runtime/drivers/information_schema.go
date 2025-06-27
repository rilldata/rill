package drivers

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// InformationSchema contains information about existing databases, schemas and tables
type InformationSchema interface {
	// ListDatabases returns a list of all databases available in the OLAP driver.
	ListDatabases(ctx context.Context) ([]string, error)
	// ListSchemas returns a list of all schemas available in the specified database.
	// If database is empty, it returns schemas from the current/default database.
	ListSchemas(ctx context.Context, database string) ([]string, error)
	// All returns metadata about all tables and views.
	// The like argument can optionally be passed to filter the tables by name.
	All(ctx context.Context, like string) ([]*Table, error)
	// Lookup returns metadata about a specific tables and views.
	Lookup(ctx context.Context, db, schema, name string) (*Table, error)
	// LoadPhysicalSize populates the PhysicalSizeBytes field of table metadata.
	// It should be called after All or Lookup and not on manually created tables.
	LoadPhysicalSize(ctx context.Context, tables []*Table) error
}

// Table represents a table in an information schema.
type Table struct {
	Database                string
	DatabaseSchema          string
	IsDefaultDatabase       bool
	IsDefaultDatabaseSchema bool
	Name                    string
	View                    bool
	Schema                  *runtimev1.StructType
	UnsupportedCols         map[string]string
	PhysicalSizeBytes       int64
}
