package starrocks

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

var _ drivers.InformationSchema = (*connection)(nil)

// ListDatabaseSchemas returns a list of database schemas in StarRocks.
// StarRocks structure: Catalog -> Database -> Table
// We map: Database = catalog, DatabaseSchema = database
func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	// Determine catalog (defaultCatalog if not configured)
	catalog := c.configProp.Catalog
	if catalog == "" {
		catalog = defaultCatalog
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

	// Switch to the configured catalog (each catalog has its own information_schema)
	if err := switchCatalogContext(ctx, db, catalog, ""); err != nil {
		return nil, "", err
	}

	// Query information_schema.schemata to list databases
	// Exclude system databases: information_schema, _statistics_, mysql
	q := `
	SELECT
		schema_name
	FROM information_schema.schemata
	WHERE schema_name NOT IN ('information_schema', '_statistics_', 'mysql', 'sys')
	`
	args := []any{}

	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += "	AND schema_name > ?"
		args = append(args, startAfter)
	}

	q += fmt.Sprintf(`
	ORDER BY schema_name
	LIMIT %d
	`, limit+1)

	rows, err := db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var res []*drivers.DatabaseSchemaInfo
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, "", err
		}
		// StarRocks mapping: Database = catalog, DatabaseSchema = database
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       catalog, // Catalog name (e.g., default_catalog, iceberg_catalog)
			DatabaseSchema: schema,  // Database name (e.g., sales, analytics)
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	// Handle pagination
	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].DatabaseSchema)
	}

	return res, next, nil
}

// ListTables returns a list of tables in a specific database schema.
// Includes both regular tables and materialized views.
// database parameter = catalog, databaseSchema parameter = database
func (c *connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	// database parameter is the catalog
	catalog := database
	if catalog == "" {
		catalog = c.configProp.Catalog
		if catalog == "" {
			catalog = defaultCatalog
		}
	}

	// databaseSchema is the actual database name
	if databaseSchema == "" {
		databaseSchema = c.configProp.Database
	}

	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

	// Switch to the correct catalog (each catalog has its own information_schema)
	if err := switchCatalogContext(ctx, db, catalog, ""); err != nil {
		return nil, "", err
	}

	// Query information_schema.tables
	// StarRocks table_type values: BASE TABLE, VIEW, MATERIALIZED VIEW
	q := `
	SELECT
		table_name,
		CASE
			WHEN table_type = 'VIEW' THEN true
			WHEN table_type = 'MATERIALIZED VIEW' THEN true
			ELSE false
		END AS is_view
	FROM information_schema.tables
	WHERE table_schema = ?
	`
	args := []any{databaseSchema}

	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += "	AND table_name > ?"
		args = append(args, startAfter)
	}

	q += fmt.Sprintf(`
	ORDER BY table_name
	LIMIT %d
	`, limit+1)

	rows, err := db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var res []*drivers.TableInfo
	for rows.Next() {
		var name string
		var isView bool
		if err := rows.Scan(&name, &isView); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.TableInfo{
			Name: name,
			View: isView,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	// Handle pagination
	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].Name)
	}

	return res, next, nil
}

// GetTable returns metadata for a specific table including column information.
// StarRocks mapping: database parameter = catalog, databaseSchema = database, table = table
// Each catalog has its own information_schema.
func (c *connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	// database parameter is the catalog
	catalog := database
	if catalog == "" {
		catalog = c.configProp.Catalog
		if catalog == "" {
			catalog = defaultCatalog
		}
	}

	// databaseSchema is the actual database name
	if databaseSchema == "" {
		databaseSchema = c.configProp.Database
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// Switch to the correct catalog (each catalog has its own information_schema)
	if err := switchCatalogContext(ctx, db, catalog, ""); err != nil {
		return nil, err
	}

	// Query to get table type and column information
	q := `
	SELECT
		CASE
			WHEN t.table_type = 'VIEW' THEN true
			WHEN t.table_type = 'MATERIALIZED VIEW' THEN true
			ELSE false
		END AS is_view,
		c.column_name,
		c.data_type
	FROM information_schema.tables t
	JOIN information_schema.columns c
		ON t.table_schema = c.table_schema AND t.table_name = c.table_name
	WHERE c.table_schema = ? AND c.table_name = ?
	ORDER BY c.ordinal_position
	`

	rows, err := db.QueryxContext(ctx, q, databaseSchema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := &drivers.TableMetadata{
		Schema: make(map[string]string),
	}

	for rows.Next() {
		var colName, dataType string
		if err := rows.Scan(&res.View, &colName, &dataType); err != nil {
			return nil, err
		}
		res.Schema[colName] = dataType
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Check if table was found
	if len(res.Schema) == 0 {
		return nil, drivers.ErrNotFound
	}

	return res, nil
}
