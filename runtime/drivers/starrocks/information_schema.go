package starrocks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

// informationSchema implements drivers.OLAPInformationSchema for StarRocks.
// Uses fully qualified names (catalog.information_schema.tables) instead of SET CATALOG/USE.
type informationSchema struct {
	c *connection
}

var _ drivers.OLAPInformationSchema = (*informationSchema)(nil)

// All returns metadata about all tables and views.
// For StarRocks, we query from the configured catalog's information_schema.
func (i *informationSchema) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	db, err := i.c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

	catalog := i.c.configProp.Catalog

	// Build query using fully qualified information_schema path
	// Pattern: catalog.information_schema.tables
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
		WHERE table_schema NOT IN ('information_schema', '_statistics_', 'mysql', 'sys')
	`, safeSQLName(catalog))

	args := []any{}

	if like != "" {
		q += " AND table_name LIKE ?"
		args = append(args, like)
	}

	if pageToken != "" {
		q += " AND table_name > ?"
		args = append(args, pageToken)
	}

	q += " ORDER BY table_schema, table_name"

	if pageSize > 0 {
		q += fmt.Sprintf(" LIMIT %d", pageSize+1)
	}

	rows, err := db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var tables []*drivers.OlapTable
	for rows.Next() {
		var schema, name string
		var isView bool
		if err := rows.Scan(&schema, &name, &isView); err != nil {
			return nil, "", err
		}

		tables = append(tables, &drivers.OlapTable{
			Database:       catalog, // StarRocks catalog -> Rill database
			DatabaseSchema: schema,  // StarRocks database -> Rill databaseSchema
			Name:           name,
			View:           isView,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	// Handle pagination
	var nextToken string
	if pageSize > 0 && uint32(len(tables)) > pageSize {
		tables = tables[:pageSize]
		nextToken = tables[len(tables)-1].Name
	}

	return tables, nextToken, nil
}

// Lookup returns metadata about a specific table or view.
// database parameter = catalog, schema parameter = database in StarRocks terms.
func (i *informationSchema) Lookup(ctx context.Context, database, schema, name string) (*drivers.OlapTable, error) {
	db, err := i.c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// StarRocks mapping: database parameter = catalog
	// If database is empty, use connector's configured catalog
	catalog := database
	if catalog == "" {
		catalog = i.c.configProp.Catalog
	}

	// StarRocks mapping: schema parameter = database
	// If schema is empty, use connector's configured database
	dbSchema := schema
	if dbSchema == "" {
		dbSchema = i.c.configProp.Database
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
	err = db.QueryRowxContext(ctx, tableQuery, dbSchema, name).Scan(&tableSchema, &tableName, &isView)
	if err != nil {
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

	rows, err := db.QueryxContext(ctx, columnsQuery, dbSchema, name)
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

		runtimeType, err := i.c.databaseTypeToRuntimeType(dataType)
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
	return &drivers.OlapTable{
		Database:                catalog,
		DatabaseSchema:          tableSchema,
		IsDefaultDatabase:       false,
		IsDefaultDatabaseSchema: false,
		Name:                    tableName,
		View:                    isView,
		Schema:                  &runtimev1.StructType{Fields: fields},
		UnsupportedCols:         unsupportedCols,
	}, nil
}

// LoadPhysicalSize populates the PhysicalSizeBytes field of table metadata.
// For external catalogs, this may not be available.
func (i *informationSchema) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	// StarRocks doesn't easily expose physical size for external tables
	// For internal tables, we could query be_tablets but it's complex
	// Return without error, leaving PhysicalSizeBytes as 0
	return nil
}

// InformationSchema interface implementation for drivers.InformationSchema

var _ drivers.InformationSchema = (*informationSchemaImpl)(nil)

// informationSchemaImpl implements drivers.InformationSchema for StarRocks
type informationSchemaImpl struct {
	c *connection
}

// ListDatabaseSchemas returns a list of database schemas in StarRocks.
// StarRocks structure: Catalog -> Database -> Table
// We map: Database = catalog, DatabaseSchema = database
func (i *informationSchemaImpl) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	db, err := i.c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

	catalog := i.c.configProp.Catalog

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

	if pageSize > 0 {
		q += fmt.Sprintf(" LIMIT %d", pageSize+1)
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
	if pageSize > 0 && uint32(len(schemas)) > pageSize {
		schemas = schemas[:pageSize]
		nextToken = schemas[len(schemas)-1].DatabaseSchema
	}

	return schemas, nextToken, nil
}

// ListTables returns a list of tables in a specific database schema.
// database parameter = catalog, databaseSchema parameter = database
func (i *informationSchemaImpl) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	db, err := i.c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

	// StarRocks mapping: database parameter = catalog
	catalog := database

	// StarRocks mapping: databaseSchema parameter = database
	dbSchema := databaseSchema

	// Query information_schema.tables using fully qualified path
	q := fmt.Sprintf(`
		SELECT
			table_name,
			CASE
				WHEN table_type = 'VIEW' THEN true
				WHEN table_type = 'MATERIALIZED VIEW' THEN true
				ELSE false
			END AS is_view
		FROM %s.information_schema.tables
		WHERE table_schema = ?
	`, safeSQLName(catalog))

	args := []any{dbSchema}

	if pageToken != "" {
		q += " AND table_name > ?"
		args = append(args, pageToken)
	}

	q += " ORDER BY table_name"

	if pageSize > 0 {
		q += fmt.Sprintf(" LIMIT %d", pageSize+1)
	}

	rows, err := db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var tables []*drivers.TableInfo
	for rows.Next() {
		var tableName string
		var isView bool
		if err := rows.Scan(&tableName, &isView); err != nil {
			return nil, "", err
		}

		tables = append(tables, &drivers.TableInfo{
			Name: tableName,
			View: isView,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	// Handle pagination
	var nextToken string
	if pageSize > 0 && uint32(len(tables)) > pageSize {
		tables = tables[:pageSize]
		nextToken = tables[len(tables)-1].Name
	}

	return tables, nextToken, nil
}

// GetTable returns metadata about a specific table.
func (i *informationSchemaImpl) GetTable(ctx context.Context, database, databaseSchema, tableName string) (*drivers.TableMetadata, error) {
	db, err := i.c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// StarRocks mapping: database parameter = catalog
	catalog := database

	// StarRocks mapping: databaseSchema parameter = database
	dbSchema := databaseSchema

	// Query table metadata and columns using JOIN
	query := fmt.Sprintf(`
		SELECT
			CASE
				WHEN t.table_type = 'VIEW' THEN true
				WHEN t.table_type = 'MATERIALIZED VIEW' THEN true
				ELSE false
			END AS is_view,
			c.column_name,
			c.data_type
		FROM %s.information_schema.tables t
		JOIN %s.information_schema.columns c
			ON t.table_schema = c.table_schema
			AND t.table_name = c.table_name
		WHERE t.table_schema = ? AND LOWER(t.table_name) = LOWER(?)
		ORDER BY c.ordinal_position
	`, safeSQLName(catalog), safeSQLName(catalog))

	rows, err := db.QueryxContext(ctx, query, dbSchema, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}
	defer rows.Close()

	schema := make(map[string]string)
	var isView bool
	hasRows := false

	for rows.Next() {
		var colName, dataType string
		if err := rows.Scan(&isView, &colName, &dataType); err != nil {
			return nil, err
		}
		schema[colName] = dataType
		hasRows = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(schema) > 0 {
		return nil, fmt.Errorf("table not found")
	}

	return &drivers.TableMetadata{
		View:   isView,
		Schema: schema,
	}, nil
}
