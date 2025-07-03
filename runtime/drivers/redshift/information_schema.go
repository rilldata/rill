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
	q := `SELECT database_name,schema_name 
	FROM svv_all_tables where schema_name not in ('information_schema','pg_catalog') 
	group by database_name,schema_name`

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

func (c *Connection) ListTables(ctx context.Context, database, schema string) ([]*drivers.TableInfo, error) {
	return nil, nil
}

func (c *Connection) GetTable(ctx context.Context, database, schema, table string) (*drivers.TableMetadata, error) {
	return nil, nil
}
