package postgres

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
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
	q := `
	SELECT
		current_database() AS database_name,
		nspname 
	FROM pg_namespace 
	WHERE has_schema_privilege(nspname, 'USAGE') AND ((nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast') AND nspname NOT LIKE 'pg_temp_%' AND nspname NOT LIKE 'pg_toast_temp_%') OR nspname = current_schema())
	ORDER BY nspname
	LIMIT $1 OFFSET $2
	`
	db, err := c.getDB()
	if err != nil {
		return nil, "", err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q, int(pageSize)+1, offset)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var schemas []*drivers.DatabaseSchemaInfo
	var database, schema string
	for rows.Next() {
		if err := rows.Scan(&database, &schema); err != nil {
			return nil, "", err
		}
		schemas = append(schemas, &drivers.DatabaseSchemaInfo{
			Database:       database,
			DatabaseSchema: schema,
		})
	}
	next := ""
	if len(schemas) > int(pageSize) {
		schemas = schemas[:pageSize]
		next = fmt.Sprintf("%d", offset+int(pageSize))
	}
	return schemas, next, rows.Err()
}

func (c *connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
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
	q := `
	SELECT
		table_name,
		table_type = 'VIEW' AS is_view
	FROM information_schema.tables 
	WHERE table_schema = $1
	ORDER BY table_name
	LIMIT $2 OFFSET $3
	`
	db, err := c.getDB()
	if err != nil {
		return nil, "", err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q, databaseSchema, int(pageSize)+1, offset)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var result []*drivers.TableInfo
	var name string
	var isView bool
	for rows.Next() {
		if err := rows.Scan(&name, &isView); err != nil {
			return nil, "", err
		}
		result = append(result, &drivers.TableInfo{
			Name: name,
			View: isView,
		})
	}
	next := ""
	if len(result) > int(pageSize) {
		result = result[:pageSize]
		next = fmt.Sprintf("%d", offset+int(pageSize))
	}
	return result, next, rows.Err()
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

	columns := make(map[string]string)
	var name, typ string
	for rows.Next() {
		if err := rows.Scan(&name, &typ); err != nil {
			return nil, err
		}
		columns[name] = typ
	}
	return &drivers.TableMetadata{
		Schema: columns,
	}, rows.Err()
}
