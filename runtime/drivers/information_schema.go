package drivers

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type InformationSchema interface {
	// ListSchemas returns metadata about all schemas.
	ListSchemas(ctx context.Context) ([]*DatabaseSchemaInfo, error)
	// ListTables returns metadata about all tables in a schema.
	ListTables(ctx context.Context, database, databaseSchema string) ([]*TableInfo, error)
	// GetTable returns metadata about a specific table.
	GetTable(ctx context.Context, database, databaseSchema, table string) (*TableMetadata, error)
}

// SchemaInfo represents a schema in an information schema.
type DatabaseSchemaInfo struct {
	Database       string
	DatabaseSchema string
}

// TableInfo represents a table in an information schema.
type TableInfo struct {
	Name string
	View bool
}

type TableMetadata struct {
	Schema map[string]string
}

// OlapInformationSchema contains information about existing tables in an OLAP driver.
// Table lookups should be case insensitive.
type OlapInformationSchema interface {
	// All returns metadata about all tables and views.
	// The like argument can optionally be passed to filter the tables by name.
	All(ctx context.Context, like string) ([]*OlapTable, error)
	// Lookup returns metadata about a specific tables and views.
	Lookup(ctx context.Context, db, schema, name string) (*OlapTable, error)
	// LoadPhysicalSize populates the PhysicalSizeBytes field of table metadata.
	// It should be called after All or Lookup and not on manually created tables.
	LoadPhysicalSize(ctx context.Context, tables []*OlapTable) error
}

// OlapTable represents a table in an information schema.
type OlapTable struct {
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
