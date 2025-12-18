package postgres

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
		current_database() AS database_name,
		nspname 
	FROM pg_namespace 
	WHERE has_schema_privilege(nspname, 'USAGE') AND ((nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast') AND nspname NOT LIKE 'pg_temp_%' AND nspname NOT LIKE 'pg_toast_temp_%') OR nspname = current_schema())
	`
	var args []any
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += `	AND nspname > $1
		ORDER BY nspname
		LIMIT $2
		`
		args = append(args, startAfter, limit+1)
	} else {
		q += `
		ORDER BY nspname
		LIMIT $1
		`
		args = append(args, limit+1)
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var res []*drivers.DatabaseSchemaInfo
	var database, schema string
	for rows.Next() {
		if err := rows.Scan(&database, &schema); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       database,
			DatabaseSchema: schema,
		})
	}
	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].DatabaseSchema)
	}
	return res, next, rows.Err()
}

func (c *connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := `
	SELECT
		table_name,
		table_type = 'VIEW' AS is_view
	FROM information_schema.tables 
	WHERE table_schema = $1
	`
	var args []any
	args = append(args, databaseSchema)
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += `	AND table_name > $2
		ORDER BY table_name
		LIMIT $3 
		`
		args = append(args, startAfter, limit+1)
	} else {
		q += `
		ORDER BY table_name
		LIMIT $2 
		`
		args = append(args, limit+1)
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var res []*drivers.TableInfo
	var name string
	var isView bool
	for rows.Next() {
		if err := rows.Scan(&name, &isView); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.TableInfo{
			Name: name,
			View: isView,
		})
	}
	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].Name)
	}
	return res, next, rows.Err()
}

func (c *connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	q := `
	SELECT 
		CASE WHEN t.table_type = 'view' THEN true ELSE false END AS view,
		c.column_name, 
		c.data_type
	FROM information_schema.tables t JOIN information_schema.columns c ON t.table_name = c.table_name AND t.table_schema = c.table_schema
	WHERE c.table_schema = $1 AND c.table_name = $2
	ORDER BY ordinal_position
	`
	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, q, databaseSchema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make(map[string]string)
	var name, typ string
	var view bool
	for rows.Next() {
		if err := rows.Scan(&view, &name, &typ); err != nil {
			return nil, err
		}
		columns[name] = typ
	}
	return &drivers.TableMetadata{
		View:   view,
		Schema: columns,
	}, rows.Err()
}
