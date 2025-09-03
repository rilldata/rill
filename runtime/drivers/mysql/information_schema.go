package mysql

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
		schema_name
	FROM information_schema.schemata
	WHERE schema_name not in ('information_schema', 'performance_schema', 'sys') OR schema_name = DATABASE()
	ORDER BY schema_name
	LIMIT ? OFFSET ?
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

	var res []*drivers.DatabaseSchemaInfo
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       "",
			DatabaseSchema: schema,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	next := ""
	if len(res) > int(pageSize) {
		res = res[:pageSize]
		next = fmt.Sprintf("%d", offset+int(pageSize))
	}
	return res, next, nil
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
		CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS view
	FROM information_schema.tables
	WHERE table_schema = ?
	ORDER BY table_name
	LIMIT ? OFFSET ?
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
	for rows.Next() {
		var name string
		var typ bool
		if err := rows.Scan(&name, &typ); err != nil {
			return nil, "", err
		}
		result = append(result, &drivers.TableInfo{
			Name: name,
			View: typ,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	next := ""
	if len(result) > int(pageSize) {
		result = result[:pageSize]
		next = fmt.Sprintf("%d", offset+int(pageSize))
	}
	return result, next, nil
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
