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
	"github.com/rilldata/rill/runtime/pkg/pagination"
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
		table_type
	FROM %s.information_schema.tables
	WHERE table_schema = %s %s 
	ORDER BY table_name
	LIMIT %d 
	`, sqlSafeName(database), escapeStringValue(databaseSchema), condFilter, limit+1)

	client, err := c.getClient(ctx)
	if err != nil {
		return nil, "", err
	}

	queryID, err := c.executeQuery(ctx, client, q, c.config.Workgroup, c.config.OutputLocation, nil)
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
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].Name)
	}
	return res, next, nil
}

func (c *Connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	q := fmt.Sprintf(`
SELECT
	CASE t.table_type WHEN 'VIEW' THEN true ELSE false END AS view,
	column_name,
	data_type
FROM %s.information_schema.columns c
JOIN %s.information_schema.tables t
	ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE c.table_schema = ? AND c.table_name = ?
ORDER BY c.ordinal_position
`, sqlSafeName(database), sqlSafeName(database))

	rows, err := c.Query(ctx, &drivers.Statement{
		Query: q,
		Args:  []any{databaseSchema, table},
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := &drivers.TableMetadata{
		Schema: make(map[string]string),
	}
	var col, typ string
	for rows.Next() {
		err = rows.Scan(&res.View, &col, &typ)
		if err != nil {
			return nil, err
		}
		res.Schema[col] = typ
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
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
	queryID, err := c.executeQuery(ctx, client, q, c.config.Workgroup, c.config.OutputLocation, nil)
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
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	start := 0
	if pageToken != "" {
		var err error
		start, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}
	end := start + limit
	if end >= len(res) {
		return res[start:], "", nil
	}
	return res[start:end], fmt.Sprintf("%d", end), nil
}

func sqlSafeName(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf("%q", escaped)
}

func escapeStringValue(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}
