package postgres

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	q := `
	SELECT 
		datname 
	FROM pg_database 
	WHERE has_database_privilege(datname, 'CONNECT') AND datname NOT IN ('template0', 'template1')
	ORDER BY datname
	`
	db, err := c.getDB("")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databaseSchemas []*drivers.DatabaseSchemaInfo
	for rows.Next() {
		var database string
		if err := rows.Scan(&database); err != nil {
			return nil, err
		}

		schemas, err := c.fetchSchemasForDatabase(ctx, database)
		if err != nil {
			return nil, err
		}
		databaseSchemas = append(databaseSchemas, schemas...)
	}
	return databaseSchemas, rows.Err()
}

func (c *connection) ListTables(ctx context.Context, database, databaseSchema string) ([]*drivers.TableInfo, error) {
	q := `
	SELECT
		table_name,
		table_type = 'VIEW' AS is_view
	FROM information_schema.tables 
	WHERE table_schema = $1
	ORDER BY table_name
	`
	db, err := c.getDB(database)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q, databaseSchema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*drivers.TableInfo
	for rows.Next() {
		var name string
		var isView bool
		if err := rows.Scan(&name, &isView); err != nil {
			return nil, err
		}
		result = append(result, &drivers.TableInfo{
			Name: name,
			View: isView,
		})
	}

	return result, rows.Err()
}

func (c *connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	q := `
	SELECT 
		column_name, 
		data_type
	FROM information_schema.columns
	WHERE table_schema = $1 AND table_name = $2
	ORDER BY ordinal_position
	`
	db, err := c.getDB(database)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q, databaseSchema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make(map[string]string)
	for rows.Next() {
		var name, typ string
		if err := rows.Scan(&name, &typ); err != nil {
			return nil, err
		}
		columns[name] = typ
	}
	return &drivers.TableMetadata{
		Schema: columns,
	}, rows.Err()
}

// fetchSchemasForDatabase fetches schemas from the given database and returns them
func (c *connection) fetchSchemasForDatabase(ctx context.Context, database string) ([]*drivers.DatabaseSchemaInfo, error) {
	q := `
	SELECT 
		nspname 
	FROM pg_namespace 
	WHERE has_schema_privilege(nspname, 'USAGE') AND nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast', 'pg_temp_1', 'pg_toast_temp_1')
	ORDER BY nspname
	`

	db, err := c.getDB(database)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []*drivers.DatabaseSchemaInfo
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, err
		}
		schemas = append(schemas, &drivers.DatabaseSchemaInfo{
			Database:       database,
			DatabaseSchema: schema,
		})
	}
	return schemas, rows.Err()
}
