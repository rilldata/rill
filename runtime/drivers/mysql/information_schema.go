package mysql

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	q := `
	SELECT
		schema_name
	FROM information_schema.schemata
	WHERE schema_name not in ('information_schema', 'performance_schema', 'sys') OR schema_name = DATABASE()
	ORDER BY schema_name
	`

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*drivers.DatabaseSchemaInfo
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, err
		}
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       "",
			DatabaseSchema: schema,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *connection) ListTables(ctx context.Context, database, databaseSchema string) ([]*drivers.TableInfo, error) {
	q := `
	SELECT
		table_name,
		CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS view
	FROM information_schema.tables
	WHERE table_schema = ?
	ORDER BY table_name
	`

	db, err := c.getDB()
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
		var typ bool
		if err := rows.Scan(&name, &typ); err != nil {
			return nil, err
		}
		result = append(result, &drivers.TableInfo{
			Name: name,
			View: typ,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	q := `
	SELECT
		column_name,
		data_type
	FROM information_schema.columns
	WHERE table_schema = ? AND table_name = ?
	ORDER BY ordinal_position
	`

	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q, databaseSchema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schemaMap := make(map[string]string)
	for rows.Next() {
		var colName, dataType string
		if err := rows.Scan(&colName, &dataType); err != nil {
			return nil, err
		}
		schemaMap[colName] = dataType
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &drivers.TableMetadata{
		Schema: schemaMap,
	}, nil
}
