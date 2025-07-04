package athena

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/rilldata/rill/runtime/drivers"
)

func (c *Connection) ListDatabaseSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	var res []*drivers.DatabaseSchemaInfo

	catalogs, err := c.listCatalogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list catalogs: %w", err)
	}

	for _, catalog := range catalogs {
		schemas, err := c.listSchemasForCatalog(ctx, catalog)
		if err != nil {
			return nil, fmt.Errorf("failed to list schemas for catalog %s: %w", catalog, err)
		}

		for _, schema := range schemas {
			res = append(res, &drivers.DatabaseSchemaInfo{
				Database:       catalog,
				DatabaseSchema: schema,
			})
		}
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
			return nil, err
		}
		for _, summary := range page.DataCatalogsSummary {
			catalogs = append(catalogs, *summary.CatalogName)
		}
	}

	return catalogs, nil
}

func (c *Connection) listSchemasForCatalog(ctx context.Context, catalog string) ([]string, error) {
	// NOTE: In Athena, databases are similar to schemas in most traditional DBs.
	var databases []string

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}
	client := athena.NewFromConfig(awsConfig)
	input := &athena.ListDatabasesInput{
		CatalogName: &catalog,
	}
	paginator := athena.NewListDatabasesPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, summary := range page.DatabaseList {
			databases = append(databases, *summary.Name)
		}
	}

	return databases, nil
}
