package starrocks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

// StarRocks Uses fully qualified names (catalog.database_schema.tables) instead of SET CATALOG/USE.
// ListDatabaseSchemas returns a list of database schemas in StarRocks.
// StarRocks structure: Catalog -> Database -> Table
// We map: Database = catalog, DatabaseSchema = database
func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	db := c.db

	catalog := c.configProp.Catalog

	// Query information_schema.schemata using fully qualified path
	q := fmt.Sprintf(`
		SELECT schema_name
		FROM %s.information_schema.schemata
		WHERE schema_name NOT IN ('information_schema', '_statistics_', 'mysql', 'sys')
	`, safeSQLName(catalog))

	args := []any{}

	if pageToken != "" {
		q += " AND schema_name > ?"
		args = append(args, pageToken)
	}

	q += " ORDER BY schema_name"

	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit+1)
	}

	rows, err := db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var schemas []*drivers.DatabaseSchemaInfo
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, "", err
		}

		// StarRocks mapping: Database = catalog, DatabaseSchema = database
		schemas = append(schemas, &drivers.DatabaseSchemaInfo{
			Database:       catalog,    // Catalog name (e.g., default_catalog, iceberg_catalog)
			DatabaseSchema: schemaName, // Database name (e.g., sales, analytics)
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	// Handle pagination
	var nextToken string
	if limit > 0 && len(schemas) > limit {
		schemas = schemas[:limit]
		nextToken = schemas[len(schemas)-1].DatabaseSchema
	}

	return schemas, nextToken, nil
}

// ListTables returns a list of tables in a specific database schema.
// database parameter = catalog, databaseSchema parameter = database.
// When databaseSchema is empty, tables from all schemas are returned.
func (c *connection) ListTables(ctx context.Context, database, databaseSchema, like string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	db := c.db
	catalog := database
	if catalog == "" {
		catalog = c.configProp.Catalog
	}

	q := fmt.Sprintf(`
		SELECT
			table_schema,
			table_name,
			CASE
				WHEN table_type = 'VIEW' THEN true
				WHEN table_type = 'MATERIALIZED VIEW' THEN true
				ELSE false
			END AS is_view
		FROM %s.information_schema.tables
		WHERE `, safeSQLName(catalog))

	var args []any
	q += " table_schema = ?"
	args = append(args, databaseSchema)

	if like != "" {
		q += " AND LOWER(table_name) LIKE LOWER(?)"
		args = append(args, like)
	}

	if pageToken != "" {
		var afterSchema, afterName string
		if err := pagination.UnmarshalPageToken(pageToken, &afterSchema, &afterName); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += " AND (table_schema > ? OR (table_schema = ? AND table_name > ?))"
		args = append(args, afterSchema, afterSchema, afterName)
	}

	q += " ORDER BY table_schema, table_name"

	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit+1)
	}

	rows, err := db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var tables []*drivers.TableInfo
	for rows.Next() {
		var schemaName, tableName string
		var isView bool
		if err := rows.Scan(&schemaName, &tableName, &isView); err != nil {
			return nil, "", err
		}
		tables = append(tables, &drivers.TableInfo{
			Name:                    tableName,
			View:                    isView,
			Database:                catalog,
			DatabaseSchema:          schemaName,
			IsDefaultDatabase:       false,
			IsDefaultDatabaseSchema: false,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextToken string
	if limit > 0 && len(tables) > limit {
		tables = tables[:limit]
		last := tables[len(tables)-1]
		nextToken = pagination.MarshalPageToken(last.DatabaseSchema, last.Name)
	}

	return tables, nextToken, nil
}

// Lookup returns metadata about a specific table or view.
// database parameter = catalog, schema parameter = database in StarRocks terms.
func (c *connection) Lookup(ctx context.Context, database, databaseSchema, table string) (*drivers.TableInfo, error) {
	db := c.db
	// StarRocks mapping: database parameter = catalog
	// If database is empty, use connector's configured catalog
	catalog := database
	if catalog == "" {
		catalog = c.configProp.Catalog
	}
	// StarRocks mapping: databaseSchema parameter = database
	// If databaseSchema is empty, use connector's configured database
	dbSchema := databaseSchema
	if dbSchema == "" {
		dbSchema = c.configProp.Database
	}

	// Query table metadata using fully qualified information_schema path
	tableQuery := fmt.Sprintf(`
		SELECT
			table_schema,
			table_name,
			CASE
				WHEN table_type = 'VIEW' THEN true
				WHEN table_type = 'MATERIALIZED VIEW' THEN true
				ELSE false
			END AS is_view
		FROM %s.information_schema.tables
		WHERE table_schema = ? AND LOWER(table_name) = LOWER(?)
	`, safeSQLName(catalog))

	var tableSchema, tableName string
	var isView bool
	err := db.QueryRowxContext(ctx, tableQuery, dbSchema, table).Scan(&tableSchema, &tableName, &isView)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, drivers.ErrNotFound
		}
		return nil, fmt.Errorf("table not found: %w", err)
	}

	// Query column information using fully qualified information_schema path
	columnsQuery := fmt.Sprintf(`
		SELECT
			column_name,
			data_type
		FROM %s.information_schema.columns
		WHERE table_schema = ? AND LOWER(table_name) = LOWER(?)
		ORDER BY ordinal_position
	`, safeSQLName(catalog))

	rows, err := db.QueryxContext(ctx, columnsQuery, dbSchema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	var fields []*runtimev1.StructType_Field
	unsupportedCols := make(map[string]string)

	for rows.Next() {
		var colName, dataType string
		if err := rows.Scan(&colName, &dataType); err != nil {
			return nil, err
		}

		runtimeType, err := c.databaseTypeToRuntimeType(dataType)
		if err != nil {
			if errors.Is(err, errUnsupportedType) {
				unsupportedCols[colName] = dataType
				continue // Skip unsupported types
			}
			return nil, err
		}

		fields = append(fields, &runtimev1.StructType_Field{
			Name: colName,
			Type: runtimeType,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// StarRocks: Always save database (catalog) and databaseSchema (database) in metrics view YAML
	// because StarRocks queries require fully qualified table names (catalog.database.table)
	return &drivers.TableInfo{
		Database:                catalog,
		DatabaseSchema:          tableSchema,
		IsDefaultDatabase:       false, // Because StarRocks uses fully qualified names (catalog.database_schema.tables)
		IsDefaultDatabaseSchema: false, // Because StarRocks uses fully qualified names (catalog.database_schema.tables)
		Name:                    tableName,
		View:                    isView,
		Schema:                  &runtimev1.StructType{Fields: fields},
		UnsupportedCols:         unsupportedCols,
	}, nil
}

// LoadPhysicalSize populates the PhysicalSizeBytes field of table metadata.
// For external catalogs, this may not be available.
func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.TableInfo) error {
	// StarRocks doesn't easily expose physical size for external tables
	// For internal tables, we could query be_tablets but it's complex
	// Return without error, leaving PhysicalSizeBytes as 0
	return nil
}

// LoadDDL implements drivers.InformationSchema.
func (c *connection) LoadDDL(ctx context.Context, table *drivers.TableInfo) error {
	db := c.db

	catalog := table.Database
	if catalog == "" {
		catalog = c.configProp.Catalog
	}
	schema := table.DatabaseSchema
	if schema == "" {
		schema = c.configProp.Database
	}

	// SHOW CREATE TABLE works for both tables and views in StarRocks.
	// For tables it returns columns: [Table, Create Table].
	// For views it returns columns: [View, Create View, character_set_client, collation_connection].
	// We extract the DDL by column name to avoid depending on column order or count.
	rows, err := db.QueryxContext(ctx, fmt.Sprintf("SHOW CREATE TABLE %s", c.Dialect().EscapeTable(catalog, schema, table.Name)))
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		res := make(map[string]any)
		if err := rows.MapScan(res); err != nil {
			return err
		}
		for _, key := range []string{"Create Table", "Create View"} {
			if v, ok := res[key]; ok && v != nil {
				if b, ok := v.([]byte); ok {
					table.DDL = string(b)
				}
				break
			}
		}
	}
	return rows.Err()
}
