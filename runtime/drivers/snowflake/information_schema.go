package snowflake

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context) ([]*drivers.DatabaseSchemaInfo, error) {
	db, err := c.getDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Get all databases (TERSE for faster results)
	dbRows, err := db.QueryxContext(ctx, "SHOW TERSE DATABASES")
	if err != nil {
		return nil, err
	}
	defer dbRows.Close()

	var results []*drivers.DatabaseSchemaInfo

	for dbRows.Next() {
		cols := make([]interface{}, 5)
		for i := range cols {
			var v interface{}
			cols[i] = &v
		}
		if err := dbRows.Scan(cols...); err != nil {
			return nil, err
		}
		dbName := fmt.Sprintf("%v", *(cols[1].(*interface{}))) // column 1 = database name

		schemaQuery := fmt.Sprintf("SHOW TERSE SCHEMAS IN DATABASE %s", sqlSafeName(dbName))
		schemaRows, err := db.QueryxContext(ctx, schemaQuery)
		if err != nil {
			return nil, err
		}

		for schemaRows.Next() {
			schemaCols := make([]interface{}, 5)
			for i := range schemaCols {
				var v interface{}
				schemaCols[i] = &v
			}
			if err := schemaRows.Scan(schemaCols...); err != nil {
				schemaRows.Close()
				return nil, err
			}
			schemaName := fmt.Sprintf("%v", *(schemaCols[1].(*interface{}))) // column 1 = schema name

			// Skip INFORMATION_SCHEMA and other system schemas
			if strings.EqualFold(schemaName, "INFORMATION_SCHEMA") {
				continue
			}

			results = append(results, &drivers.DatabaseSchemaInfo{
				Database:       dbName,
				DatabaseSchema: schemaName,
			})
		}
		schemaRows.Close()
	}

	if err := dbRows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (c *connection) ListTables(ctx context.Context, database, databaseSchema string) ([]*drivers.TableInfo, error) {
	q := fmt.Sprintf(`
		SELECT
			table_name,
			CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS view
		FROM %s.INFORMATION_SCHEMA.TABLES
		WHERE table_schema = ?
		ORDER BY table_name
	`, sqlSafeName(database))

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

	var res []*drivers.TableInfo
	for rows.Next() {
		var name string
		var view bool
		if err := rows.Scan(&name, &view); err != nil {
			return nil, err
		}
		res = append(res, &drivers.TableInfo{
			Name: name,
			View: view,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	q := fmt.Sprintf(`
		SELECT
			column_name,
			data_type
		FROM %s.INFORMATION_SCHEMA.COLUMNS
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position
	`, sqlSafeName(database))

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
		var colName, colType string
		if err := rows.Scan(&colName, &colType); err != nil {
			return nil, err
		}
		schemaMap[colName] = colType
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &drivers.TableMetadata{
		Schema: schemaMap,
	}, nil
}

func sqlSafeName(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf("%q", escaped)
}
