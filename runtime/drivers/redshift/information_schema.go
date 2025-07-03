package redshift

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	"github.com/rilldata/rill/runtime/drivers"
)

func (c *Connection) ListDatabaseSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	q := `
	SELECT database_name,schema_name 
	FROM svv_all_tables 
	WHERE schema_name NOT IN ('information_schema','pg_catalog') 
	GROUP BY database_name,schema_name`

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := redshiftdata.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}

	result, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{
		Id: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	var res []*drivers.DatabaseSchemaInfo
	for _, record := range result.Records {
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       record[0].(*types.FieldMemberStringValue).Value,
			DatabaseSchema: record[1].(*types.FieldMemberStringValue).Value,
		})
	}
	return res, nil
}

func (c *Connection) ListTables(ctx context.Context, database, databaseSchema string) ([]*drivers.TableInfo, error) {
	q := fmt.Sprintf(`
	SELECT
	table_name,
	CASE
		WHEN table_type = 'VIEW' THEN true
		ELSE false
	END AS view
	FROM svv_all_tables
	WHERE database_name = '%s'
	AND schema_name = '%s' 
	ORDER BY table_name;
	`, database, databaseSchema)

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := redshiftdata.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}

	result, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{
		Id: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	var res []*drivers.TableInfo
	for _, record := range result.Records {
		res = append(res, &drivers.TableInfo{
			Name: record[0].(*types.FieldMemberStringValue).Value,
			View: record[1].(*types.FieldMemberBooleanValue).Value,
		})
	}
	return res, nil
}

func (c *Connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	// Query to get column name and data type
	q := fmt.Sprintf(`
	SELECT column_name, data_type
	FROM svv_all_columns
	WHERE database_name = '%s'
	AND schema_name = '%s'
	AND table_name = '%s'
	ORDER BY ordinal_position;
	`, database, databaseSchema, table)

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

	// Build schema map
	schemaMap := make(map[string]string)
	for _, record := range result.Records {
		column := record[0].(*types.FieldMemberStringValue).Value
		dataType := record[1].(*types.FieldMemberStringValue).Value
		schemaMap[column] = dataType
	}

	return &drivers.TableMetadata{
		Schema: schemaMap,
	}, nil
}
