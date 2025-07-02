package postgres

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	q := `SELECT datname FROM pg_database 
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

func (c *connection) ListTables(ctx context.Context, database, schema string) ([]*drivers.TableInfo, error) {
	return nil, nil
}

func (c *connection) GetTable(ctx context.Context, database, schema, table string) (*drivers.TableMetadata, error) {
	return nil, nil
}

// fetchSchemasForDatabase fetches schemas from the given database and returns them
func (c *connection) fetchSchemasForDatabase(ctx context.Context, database string) ([]*drivers.DatabaseSchemaInfo, error) {
	db, err := c.getDB(database)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	q := `
	SELECT n.nspname FROM pg_namespace n 
	WHERE has_schema_privilege(n.nspname, 'USAGE') AND n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast', 'pg_temp_1', 'pg_toast_temp_1')
	ORDER BY n.nspname
	`
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
