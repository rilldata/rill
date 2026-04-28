package drivers

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type InformationSchema interface {
	// ListDatabaseSchemas returns all schemas across databases
	ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*DatabaseSchemaInfo, string, error)
	// ListTables returns all tables in a schema.
	ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*TableInfo, string, error)
	// All returns metadata about all tables and views.
	// The like argument can optionally be passed to filter the tables by name.
	All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*TableInfo, string, error)
	// Lookup returns metadata about a specific tables and views.
	Lookup(ctx context.Context, db, schema, name string) (*TableInfo, error)
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
	// Schema is the table schema. It is only set when only single table is looked up. It is not set when listing all tables.
	Schema            *runtimev1.StructType
	UnsupportedCols   map[string]string
	PhysicalSizeBytes int64
	DDL               string
}

type TableMetadata struct {
	View   bool // TODO: populate for other drivers
	Schema map[string]string
}

// AllFromInformationSchema is a helper function that drivers implementing InformationSchema can use to implement Olap.All()
func AllFromInformationSchema(ctx context.Context, like string, pageSize uint32, pageToken string, i InformationSchema) ([]*TableInfo, string, error) {
	if like != "" {
		return nil, "", fmt.Errorf("like filter not supported")
	}
	schemas, token, err := i.ListDatabaseSchemas(ctx, pageSize, pageToken)
	if err != nil {
		return nil, "", err
	}
	tables := make([]*TableInfo, 0)
	for _, schema := range schemas {
		ts, token, err := i.ListTables(ctx, schema.Database, schema.DatabaseSchema, 1000, "")
		if err != nil {
			return nil, "", err
		}
		if token != "" {
			// we don't support pagination across multiple schemas
			return nil, "", fmt.Errorf("schema has more than 1000 tables can not list all")
		}
		for _, t := range ts {
			table := &TableInfo{
				Database:                schema.Database,
				DatabaseSchema:          schema.DatabaseSchema,
				IsDefaultDatabase:       t.IsDefaultDatabase,
				IsDefaultDatabaseSchema: t.IsDefaultDatabaseSchema,
				Name:                    t.Name,
				View:                    t.View,
				Schema:                  nil,
				UnsupportedCols:         nil,
				PhysicalSizeBytes:       0,
			}
			tables = append(tables, table)
		}
	}
	return tables, token, nil
}
