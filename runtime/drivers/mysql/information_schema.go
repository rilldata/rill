package mysql

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := `
	SELECT
		schema_name
	FROM information_schema.schemata
	WHERE (schema_name NOT IN ('information_schema', 'performance_schema', 'sys') OR schema_name = DATABASE())
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
	q += `
	ORDER BY schema_name 
	LIMIT ?
	`
	args = append(args, limit+1)

	db, err := c.getDB()
	if err != nil {
		return nil, "", err
	}
	defer db.Close()

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
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       "",
			DatabaseSchema: schema,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].DatabaseSchema)
	}
	return res, next, nil
}

func (c *connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := `
	SELECT
		table_name,
		CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS view
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
	q += `
	ORDER BY table_name 
	LIMIT ?
	`
	args = append(args, limit+1)

	db, err := c.getDB()
	if err != nil {
		return nil, "", err
	}
	defer db.Close()

	rows, err := db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var res []*drivers.TableInfo
	for rows.Next() {
		var name string
		var typ bool
		if err := rows.Scan(&name, &typ); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.TableInfo{
			Name: name,
			View: typ,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].Name)
	}
	return res, next, nil
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
