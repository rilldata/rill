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

func (c *Connection) ListTables(ctx context.Context, database, databaseSchema string) ([]*drivers.TableInfo, error) {
	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := athena.NewFromConfig(awsConfig, func(o *athena.Options) {
		o.TracerProvider = smithyoteltracing.Adapt(otel.GetTracerProvider())
	})

	q := fmt.Sprintf(`
	SELECT
		table_name,
		table_type
	FROM %s.information_schema.tables 
	WHERE table_schema = %s
	`, sqlSafeName(database), escapeStringValue(databaseSchema))

	queryID, err := c.executeQuery(ctx, client, q, c.config.Workgroup, c.config.OutputLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to execute table listing query: %w", err)
	}

	results, err := client.GetQueryResults(ctx, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}
	tables := make([]*drivers.TableInfo, 0, len(results.ResultSet.Rows)-1)
	for _, row := range results.ResultSet.Rows[1:] {
		if len(row.Data) < 2 || row.Data[0].VarCharValue == nil || row.Data[1].VarCharValue == nil {
			continue
		}
		tables = append(tables, &drivers.TableInfo{
			Name: *row.Data[0].VarCharValue,
			View: strings.EqualFold(*row.Data[1].VarCharValue, "VIEW"),
		})
	}

	return tables, nil
}

func (c *Connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := athena.NewFromConfig(awsConfig, func(o *athena.Options) {
		o.TracerProvider = smithyoteltracing.Adapt(otel.GetTracerProvider())
	})

	query := fmt.Sprintf(`
	SELECT
		column_name,
		data_type
	FROM %s.information_schema.columns 
	WHERE table_schema = %s AND table_name = %s
	ORDER BY ordinal_position
	`, sqlSafeName(database), escapeStringValue(databaseSchema), escapeStringValue(table))

	queryID, err := c.executeQuery(ctx, client, query, c.config.Workgroup, c.config.OutputLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to execute columns query: %w", err)
	}

	results, err := client.GetQueryResults(ctx, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	schemaMap := make(map[string]string, len(results.ResultSet.Rows)-1)
	for _, row := range results.ResultSet.Rows[1:] {
		if len(row.Data) < 2 || row.Data[0].VarCharValue == nil || row.Data[1].VarCharValue == nil {
			continue
		}
		schemaMap[*row.Data[0].VarCharValue] = *row.Data[1].VarCharValue
	}

	return &drivers.TableMetadata{
		Schema: schemaMap,
	}, nil
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
			catalog_name,
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

	res := make([]*drivers.DatabaseSchemaInfo, 0, len(results.ResultSet.Rows)-1)
	for _, row := range results.ResultSet.Rows[1:] {
		if len(row.Data) < 2 || row.Data[0].VarCharValue == nil || row.Data[1].VarCharValue == nil {
			continue
		}
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       *row.Data[0].VarCharValue,
			DatabaseSchema: *row.Data[1].VarCharValue,
		})
	}

	return res, nil
}

func sqlSafeName(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf("%q", escaped)
}

func escapeStringValue(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}
