package drivers

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type InformationSchema interface {
	// ListDatabaseSchemas returns all schemas across databases
	ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*DatabaseSchemaInfo, string, error)
	// ListTables returns tables based on the provided scope.
	// If both `database` and `databaseSchema` are empty, it lists tables across all databases.
	// Otherwise, it lists tables within the specified database and schema.
	// The `like` parameter is optional and filters results to tables matching the given pattern.
	// Results are paginated using `pageSize` and `pageToken`.
	ListTables(ctx context.Context, database, databaseSchema, like string, pageSize uint32, pageToken string) ([]*TableInfo, string, error)
	// Lookup returns metadata about a specific tables and views.
	Lookup(ctx context.Context, database, databaseSchema, table string) (*TableInfo, error)
	// LoadPhysicalSize populates the PhysicalSizeBytes field of table metadata.
	// It should be called after All or Lookup and not on manually created tables.
	LoadPhysicalSize(ctx context.Context, tables []*TableInfo) error
	// LoadDDL populates the DDL field of a single table's metadata.
	// Drivers that don't support DDL retrieval should return nil (leaving DDL empty).
	LoadDDL(ctx context.Context, table *TableInfo) error
}

const (
	// DefaultPageSize is the default page size used when pageSize is not defined
	DefaultPageSize           = 100
	DefaultPageSizeForObjects = 1000
)

// SchemaInfo represents a schema in an information schema.
type DatabaseSchemaInfo struct {
	Database       string
	DatabaseSchema string
}

// TableInfo represents a table in an information schema.
type TableInfo struct {
	Database                string
	DatabaseSchema          string
	IsDefaultDatabase       bool
	IsDefaultDatabaseSchema bool
	Name                    string
	View                    bool

	// Schema contains the table schema.
	// It is only populated after calling Lookup and is nil when listing tables.
	Schema *runtimev1.StructType

	// UnsupportedCols lists columns that could not be mapped to a supported runtimev1.Type.
	// The key is the column name and the value is the original raw type.
	// It is only populated after calling Lookup and is empty when listing tables.
	UnsupportedCols map[string]string

	// PhysicalSizeBytes is the on-disk size of the table in bytes.
	// It is only populated after calling LoadPhysicalSize.
	PhysicalSizeBytes int64

	// DDL contains the CREATE TABLE/VIEW statement for the table/view.
	// It is only populated after calling LoadDDL.
	DDL string
}
