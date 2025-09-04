package athena

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/aws/smithy-go"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/sync/errgroup"
)

func (c *Connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, "", err
	}

	catalogs, err := c.listCatalogs(ctx, client)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list catalogs: %w", err)
	}
	// if no catalogs query current catalog by passing empty string
	if len(catalogs) == 0 {
		items, err := c.listSchemasForCatalog(ctx, client, "")
		if err != nil {
			return nil, "", err
		}
		return paginateSchemas(items, pageSize, pageToken)
	}
	var (
		mu  sync.Mutex
		res []*drivers.DatabaseSchemaInfo
	)
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(5)
	for _, catalog := range catalogs {
		catalogName := catalog
		g.Go(func() error {
			schemas, err := c.listSchemasForCatalog(ctx, client, catalogName)
			if err != nil {
				return fmt.Errorf("failed to list schemas for catalog %q: %w", catalog, err)
			}
			mu.Lock()
			res = append(res, schemas...)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, "", err
	}
	return paginateSchemas(res, pageSize, pageToken)
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
		table_type
	FROM %s.information_schema.tables
	WHERE table_schema = %s 
	ORDER BY table_name
	OFFSET %d
	LIMIT %d 
	`, sqlSafeName(database), escapeStringValue(databaseSchema), offset, pageSize+1)

	client, err := c.getClient(ctx)
	if err != nil {
		return nil, "", err
	}

	queryID, err := c.executeQuery(ctx, client, q, c.config.Workgroup, c.config.OutputLocation)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute table listing query: %w", err)
	}

	input := &athena.GetQueryResultsInput{QueryExecutionId: queryID}
	results, err := client.GetQueryResults(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get query results: %w", err)
	}
	// first row is header of skipping it
	res := make([]*drivers.TableInfo, 0, len(results.ResultSet.Rows)-1)
	for _, row := range results.ResultSet.Rows[1:] {
		if len(row.Data) < 2 || row.Data[0].VarCharValue == nil || row.Data[1].VarCharValue == nil {
			continue
		}
		res = append(res, &drivers.TableInfo{
			Name: *row.Data[0].VarCharValue,
			View: strings.EqualFold(*row.Data[1].VarCharValue, "VIEW"),
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
	q := fmt.Sprintf(`
	SELECT
		column_name,
		data_type
	FROM %s.information_schema.columns
	WHERE table_schema = %s AND table_name = %s
	ORDER BY ordinal_position
	`, sqlSafeName(database), escapeStringValue(databaseSchema), escapeStringValue(table))

	client, err := c.getClient(ctx)
	if err != nil {
		return nil, err
	}

	queryID, err := c.executeQuery(ctx, client, q, c.config.Workgroup, c.config.OutputLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to execute columns query: %w", err)
	}

	results, err := client.GetQueryResults(ctx, &athena.GetQueryResultsInput{
		QueryExecutionId: queryID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}
	// first row is header of skipping it
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

func (c *Connection) listCatalogs(ctx context.Context, client *athena.Client) ([]string, error) {
	// NOTE: In Athena, catalogs are similar to databases in most traditional DBs.
	var catalogs []string
	paginator := athena.NewListDataCatalogsPaginator(client, &athena.ListDataCatalogsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			var ae smithy.APIError
			if errors.As(err, &ae) {
				switch ae.ErrorCode() {
				case "AccessDeniedException", "NotAuthorized":
					// Return nil
					return nil, nil
				}
			}
			return nil, err
		}
		for _, summary := range page.DataCatalogsSummary {
			if summary.Status == types.DataCatalogStatusCreateComplete && summary.Type == types.DataCatalogTypeGlue {
				catalogs = append(catalogs, *summary.CatalogName)
			}
		}
	}

	return catalogs, nil
}

func (c *Connection) listSchemasForCatalog(ctx context.Context, client *athena.Client, catalog string) ([]*drivers.DatabaseSchemaInfo, error) {
	// Use catalog if specified
	var q string
	if catalog != "" {
		q = fmt.Sprintf(`
		SELECT
			catalog_name,
			schema_name
		FROM %s.information_schema.schemata
		WHERE schema_name NOT IN ('information_schema', 'performance_schema', 'sys') OR schema_name = current_schema
		`, sqlSafeName(catalog))
	} else {
		q = `
		SELECT 
			catalog_name, 
			schema_name 
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('information_schema', 'performance_schema', 'sys') OR schema_name = current_schema
		`
	}

	// Execute the query
	queryID, err := c.executeQuery(ctx, client, q, c.config.Workgroup, c.config.OutputLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to execute schema listing query: %w", err)
	}

	// Fetch results
	results, err := client.GetQueryResults(ctx, &athena.GetQueryResultsInput{
		QueryExecutionId: queryID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	// first row is header of skipping it
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

func paginateSchemas(res []*drivers.DatabaseSchemaInfo, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	// sort by database and schema befor paginating
	sort.Slice(res, func(i, j int) bool {
		if res[i].Database == res[j].Database {
			return res[i].DatabaseSchema < res[j].DatabaseSchema
		}
		return res[i].Database < res[j].Database
	})
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
	end := offset + int(pageSize)
	if end > len(res) {
		end = len(res)
	}
	next := ""
	if end < len(res) {
		next = fmt.Sprintf("%d", end)
	}
	return res[offset:end], next, nil
}

func sqlSafeName(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf("%q", escaped)
}

func escapeStringValue(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}
