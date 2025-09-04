package redshift

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	"github.com/rilldata/rill/runtime/drivers"
)

func (c *Connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	if pageSize == 0 {
		pageSize = drivers.DefaultPageSize
	}
	offset := 0
	if pageToken != "" {
		var err error
		offset, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}
	q := fmt.Sprintf(`
	SELECT 
		database_name, 
		schema_name 
	FROM svv_all_tables 
	WHERE schema_name NOT IN ('information_schema', 'pg_catalog') OR schema_name = current_schema()
	GROUP BY database_name, schema_name 
	ORDER BY database_name, schema_name
	LIMIT %d 
	OFFSET %d
	`, pageSize+1, offset)

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := redshiftdata.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list schemas: %w", err)
	}

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
	if len(res) > int(pageSize) {
		res = res[:pageSize]
		next = strconv.Itoa(offset + int(pageSize))
	}
	return res, next, nil
}

func (c *Connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	if pageSize == 0 {
		pageSize = drivers.DefaultPageSize
	}
	offset := 0
	if pageToken != "" {
		var err error
		offset, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}
	q := fmt.Sprintf(`
	SELECT
		table_name,
		CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS view
	FROM svv_all_tables
	WHERE database_name = %s AND schema_name = %s 
	ORDER BY table_name
	LIMIT %d 
	OFFSET %d
	`, escapeStringValue(database), escapeStringValue(databaseSchema), pageSize+1, offset)

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := redshiftdata.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list schemas: %w", err)
	}

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
		res = append(res, &drivers.TableInfo{
			Name: nameField.Value,
			View: viewField.Value,
		})
	}
	next := ""
	if len(res) > int(pageSize) {
		res = res[:pageSize]
		next = strconv.Itoa(offset + int(pageSize))
	}
	return res, next, nil
}

func (c *Connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	// Query to get column name and data type
	q := fmt.Sprintf(`
	SELECT 
		column_name, 
		data_type
	FROM svv_all_columns
	WHERE database_name = %s AND schema_name = %s AND table_name = %s
	ORDER BY ordinal_position;
	`, escapeStringValue(database), escapeStringValue(databaseSchema), escapeStringValue(table))

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := redshiftdata.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}

	result, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{
		Id: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	var column, dataType string
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
		column = colField.Value
		dataType = typeField.Value
		schemaMap[column] = dataType
	}

	return &drivers.TableMetadata{
		Schema: schemaMap,
	}, nil
}

func escapeStringValue(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}
