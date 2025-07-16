package drivers

import (
	"context"
)

type InformationSchema interface {
	// ListDatabaseSchemas returns all schemas across databases
	ListDatabaseSchemas(ctx context.Context) ([]*DatabaseSchemaInfo, error)
	// ListTables returns all tables in a schema.
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
