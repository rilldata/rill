package drivers

import (
	"context"
	"fmt"
)

type InformationSchema interface {
	// ListDatabaseSchemas returns all schemas across databases
	ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*DatabaseSchemaInfo, string, error)
	// ListTables returns all tables in a schema.
	ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*TableInfo, string, error)
	// GetTable returns metadata about a specific table.
	GetTable(ctx context.Context, database, databaseSchema, table string) (*TableMetadata, error)
}

const (
	// DefaultPageSize is the default page size used when pageSize is not defined
	DefaultPageSize = 100
)

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
	View   bool // TODO: populate for other drivers
	Schema map[string]string
}

// AllFromInformationSchema is a helper function that drivers implementing InformationSchema can use to implement Olap.All()
// This is a short term solution. Longer term we should merge OLAPInformationSchema and InformationSchema interfaces.
func AllFromInformationSchema(ctx context.Context, like string, pageSize uint32, pageToken string, i InformationSchema) ([]*OlapTable, string, error) {
	if like != "" {
		return nil, "", fmt.Errorf("like filter not supported")
	}
	schemas, token, err := i.ListDatabaseSchemas(ctx, pageSize, pageToken)
	if err != nil {
		return nil, "", err
	}
	tables := make([]*OlapTable, 0)
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
			table := &OlapTable{
				Database:          schema.Database,
				DatabaseSchema:    schema.DatabaseSchema,
				Name:              t.Name,
				View:              t.View,
				Schema:            nil,
				UnsupportedCols:   nil,
				PhysicalSizeBytes: 0,
			}
			tables = append(tables, table)
		}
	}
	return tables, token, nil
}
