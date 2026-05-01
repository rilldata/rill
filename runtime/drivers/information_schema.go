package drivers

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type InformationSchema interface {
	// ListDatabaseSchemas returns all schemas across databases
	ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*DatabaseSchemaInfo, string, error)
	// ListTables returns tables within the specified database and schema.
	// The `like` parameter is optional and filters results to tables matching the given pattern.
	// Results are paginated using `pageSize` and `pageToken`.
	ListTables(ctx context.Context, database, databaseSchema, like string, pageSize uint32, pageToken string) ([]*TableInfo, string, error)
	// All returns metadata about all tables and views across all schemas.
	// The like argument can optionally be passed to filter the tables by name.
	All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*TableInfo, string, error)
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

// AllFromInformationSchema is a helper that drivers can use to implement All() by iterating ListDatabaseSchemas + ListTables.
func AllFromInformationSchema(ctx context.Context, like string, pageSize uint32, pageToken string, i InformationSchema) ([]*TableInfo, string, error) {
	schemas, token, err := i.ListDatabaseSchemas(ctx, pageSize, pageToken)
	if err != nil {
		return nil, "", err
	}
	tables := make([]*TableInfo, 0)
	for _, schema := range schemas {
		ts, tok, err := i.ListTables(ctx, schema.Database, schema.DatabaseSchema, like, 1000, "")
		if err != nil {
			return nil, "", err
		}
		if tok != "" {
			return nil, "", fmt.Errorf("schema has more than 1000 tables, cannot list all")
		}
		tables = append(tables, ts...)
	}
	return tables, token, nil
}
