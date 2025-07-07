package athena

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/tracing/smithyoteltracing"
	"github.com/rilldata/rill/runtime/drivers"
	"go.opentelemetry.io/otel"
)

func (c *Connection) ListDatabaseSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	var res []*drivers.DatabaseSchemaInfo

	catalogs, err := c.listCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list catalogs: %w", err)
	}
	// if no catalogs query current catalog by passing empty string
	if len(catalogs) == 0 {
		return c.listSchemasForCatalog(ctx, "")
	}

	for _, catalog := range catalogs {
		schemas, err := c.listSchemasForCatalog(ctx, catalog)
		if err != nil {
			return nil, fmt.Errorf("failed to list schemas for catalog %s: %w", catalog, err)
		}

		res = append(res, schemas...)
	}

	return res, nil
}

func (c *Connection) ListTables(ctx context.Context, database, schema string) ([]*drivers.TableInfo, error) {
	return nil, nil
}

func (c *Connection) GetTable(ctx context.Context, database, schema, table string) (*drivers.TableMetadata, error) {
	return nil, nil
}

func (c *Connection) listCatalogs(ctx context.Context) ([]string, error) {
	// NOTE: In Athena, catalogs are similar to databases in most traditional DBs.
	var catalogs []string

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}
	client := athena.NewFromConfig(awsConfig)
	input := &athena.ListDataCatalogsInput{}
	paginator := athena.NewListDataCatalogsPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			var ae smithy.APIError
			if errors.As(err, &ae) {
				switch ae.ErrorCode() {
				case "AccessDeniedException", "NotAuthorized":
					// Return whatever empty catalogs
					return catalogs, nil
				}
			}
			return nil, err
		}
		for _, summary := range page.DataCatalogsSummary {
			catalogs = append(catalogs, *summary.CatalogName)
		}
	}

	return catalogs, nil
}

func (c *Connection) listSchemasForCatalog(ctx context.Context, catalog string) ([]*drivers.DatabaseSchemaInfo, error) {
	// Get AWS config with configured region
	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	// Create Athena client
	client := athena.NewFromConfig(awsConfig, func(o *athena.Options) {
		o.TracerProvider = smithyoteltracing.Adapt(otel.GetTracerProvider())
	})

	// Use catalog if specified
	var q string
	if catalog != "" {
		q = fmt.Sprintf(`
		SELECT
			schema_name 
		FROM %s.information_schema.schemata
		`, sqlSafeName(catalog))
	} else {
		q = `
		SELECT 
			catalog_name, 
			schema_name 
		FROM information_schema.schemata
		`
	}

	// Execute the query
	queryID, err := c.executeQuery(ctx, client, q, c.config.Workgroup, c.config.OutputLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to execute schema listing query: %w", err)
	}

	// Fetch results
	results, err := client.GetQueryResults(ctx, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	var res []*drivers.DatabaseSchemaInfo
	// Parse rows; skip header
	for _, row := range results.ResultSet.Rows[1:] {
		data := row.Data

		if catalog != "" && len(data) >= 1 && data[0].VarCharValue != nil {
			schema := *data[0].VarCharValue
			res = append(res, &drivers.DatabaseSchemaInfo{
				Database:       catalog,
				DatabaseSchema: schema,
			})
		} else if len(data) >= 2 && data[0].VarCharValue != nil && data[1].VarCharValue != nil {
			catalogName := *data[0].VarCharValue
			schemaName := *data[1].VarCharValue
			res = append(res, &drivers.DatabaseSchemaInfo{
				Database:       catalogName,
				DatabaseSchema: schemaName,
			})
		}
	}

	return res, nil
}

func sqlSafeName(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf("%q", escaped)
}
