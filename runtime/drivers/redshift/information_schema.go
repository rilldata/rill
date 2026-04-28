package redshift

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

func (c *Connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	condFilter := ""
	if pageToken != "" {
		var tokDB, tokSchema string
		if err := pagination.UnmarshalPageToken(pageToken, &tokDB, &tokSchema); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		condFilter = fmt.Sprintf(" AND (database_name > %s OR (database_name = %s AND schema_name > %s))", escapeStringValue(tokDB), escapeStringValue(tokDB), escapeStringValue(tokSchema))
	}
	q := fmt.Sprintf(`
	SELECT 
		database_name, 
		schema_name 
	FROM svv_all_tables 
	WHERE (schema_name NOT IN ('information_schema', 'pg_catalog') OR schema_name = current_schema()) %s
	GROUP BY database_name, schema_name 
	ORDER BY database_name, schema_name
	LIMIT %d 
	`, condFilter, limit+1)

	client, err := c.getClient(ctx)
	if err != nil {
		return nil, "", err
	}

	out, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list schemas: %w", err)
	}
	queryExecutionID := *out.Id

	result, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{
		Id: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get query results: %w", err)
	}

	var res []*drivers.DatabaseSchemaInfo
	for _, record := range result.Records {
		dbField, ok := record[0].(*types.FieldMemberStringValue)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type for database_name field")
		}
		schemaField, ok := record[1].(*types.FieldMemberStringValue)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type for schema_name field")
		}
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       dbField.Value,
			DatabaseSchema: schemaField.Value,
		})
	}
	next := ""
	if len(res) > limit {
		res = res[:limit]
		last := res[len(res)-1]
		next = pagination.MarshalPageToken(last.Database, last.DatabaseSchema)
	}
	return res, next, nil
}

func (c *Connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	condFilter := ""
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		condFilter = fmt.Sprintf("AND table_name > %s", escapeStringValue(startAfter))
	}
	q := fmt.Sprintf(`
	SELECT
		table_name,
		CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS view,
		CASE WHEN database_name = current_database() THEN true ELSE false END AS is_default_database,
		CASE WHEN schema_name = current_schema() THEN true ELSE false END AS is_default_database_schema
	FROM svv_all_tables
	WHERE database_name = %s AND schema_name = %s %s
	ORDER BY table_name
	LIMIT %d
	`, escapeStringValue(database), escapeStringValue(databaseSchema), condFilter, limit+1)

	client, err := c.getClient(ctx)
	if err != nil {
		return nil, "", err
	}

	out, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list schemas: %w", err)
	}
	queryExecutionID := *out.Id

	result, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{
		Id: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get query results: %w", err)
	}

	var res []*drivers.TableInfo
	for _, record := range result.Records {
		nameField, ok := record[0].(*types.FieldMemberStringValue)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type for table name field")
		}
		viewField, ok := record[1].(*types.FieldMemberBooleanValue)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type for view field")
		}
		isDefaultDatabaseField, ok := record[2].(*types.FieldMemberBooleanValue)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type for is_default_database field")
		}
		isDefaultDatabaseSchemaField, ok := record[3].(*types.FieldMemberBooleanValue)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type for is_default_database_schema field")
		}
		res = append(res, &drivers.TableInfo{
			Name:                    nameField.Value,
			View:                    viewField.Value,
			IsDefaultDatabase:       isDefaultDatabaseField.Value,
			IsDefaultDatabaseSchema: isDefaultDatabaseSchemaField.Value,
		})
	}
	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].Name)
	}
	return res, next, nil
}

func (c *Connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	q := fmt.Sprintf(`
SELECT 
	column_name, 
	data_type,
	CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS view
FROM svv_all_columns c
JOIN svv_all_tables t
	ON t.database_name = c.database_name AND t.schema_name = c.schema_name AND t.table_name = c.table_name
WHERE c.database_name = %s AND c.schema_name = %s AND c.table_name = %s
ORDER BY ordinal_position;
`, escapeStringValue(database), escapeStringValue(databaseSchema), escapeStringValue(table))

	client, err := c.getClient(ctx)
	if err != nil {
		return nil, err
	}

	out, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}
	queryExecutionID := *out.Id

	result, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{
		Id: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	var column, dataType string
	var isView bool
	schemaMap := make(map[string]string, len(result.Records))
	for _, record := range result.Records {
		colField, ok := record[0].(*types.FieldMemberStringValue)
		if !ok {
			return nil, fmt.Errorf("unexpected type for column_name field")
		}
		typeField, ok := record[1].(*types.FieldMemberStringValue)
		if !ok {
			return nil, fmt.Errorf("unexpected type for data_type field")
		}
		viewField, ok := record[2].(*types.FieldMemberBooleanValue)
		if !ok {
			return nil, fmt.Errorf("unexpected type for view field")
		}
		column = colField.Value
		dataType = typeField.Value
		isView = viewField.Value
		schemaMap[column] = dataType
	}

	return &drivers.TableMetadata{
		Schema: schemaMap,
		View:   isView,
	}, nil
}

// All implements drivers.InformationSchema.
func (c *Connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.InformationSchema.
func (c *Connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	return nil
}

// LoadDDL implements drivers.InformationSchema.
func (c *Connection) LoadDDL(ctx context.Context, table *drivers.OlapTable) error {
	return nil // Not implemented
}

// Lookup implements drivers.InformationSchema.
func (c *Connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	meta, err := c.GetTable(ctx, db, schema, name)
	if err != nil {
		return nil, err
	}
	runtimeSchema := &runtimev1.StructType{
		Fields: make([]*runtimev1.StructType_Field, 0, len(meta.Schema)),
	}
	for name, typ := range meta.Schema {
		runtimeSchema.Fields = append(runtimeSchema.Fields, &runtimev1.StructType_Field{
			Name: name,
			Type: redshiftTypeToRuntimeType(typ),
		})
	}
	return &drivers.OlapTable{
		Database:          db,
		DatabaseSchema:    schema,
		Name:              name,
		View:              meta.View,
		Schema:            runtimeSchema,
		UnsupportedCols:   nil,
		PhysicalSizeBytes: 0,
	}, nil
}

func escapeStringValue(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}
