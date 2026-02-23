package oracle

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	// Query schemas that actually contain tables or views accessible to the current user,
	// rather than listing all users and trying to exclude system accounts.
	q := `SELECT owner FROM (
		SELECT DISTINCT owner FROM all_tables
		UNION
		SELECT DISTINCT owner FROM all_views
	) t`
	var args []any
	nextParam := 1
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += fmt.Sprintf(" WHERE owner > :%d", nextParam)
		args = append(args, startAfter)
	}
	// Oracle does not support bind variables in FETCH FIRST; inline the limit value directly.
	q += fmt.Sprintf(` ORDER BY owner FETCH FIRST %d ROWS ONLY`, limit+1)

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

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
	SELECT table_name, 0 AS is_view FROM all_tables WHERE owner = :1
	UNION ALL
	SELECT view_name, 1 AS is_view FROM all_views WHERE owner = :2
	`
	args := []any{databaseSchema, databaseSchema}

	// Wrap in a subquery for ordering and pagination
	outerQ := fmt.Sprintf(`SELECT table_name, is_view FROM (%s) t`, q)
	nextParam := 3
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		outerQ += fmt.Sprintf(" WHERE table_name > :%d", nextParam)
		args = append(args, startAfter)
	}
	// Oracle does not support bind variables in FETCH FIRST; inline the limit value directly.
	outerQ += fmt.Sprintf(` ORDER BY table_name FETCH FIRST %d ROWS ONLY`, limit+1)

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

	rows, err := db.QueryxContext(ctx, outerQ, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var res []*drivers.TableInfo
	for rows.Next() {
		var name string
		var isView int
		if err := rows.Scan(&name, &isView); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.TableInfo{
			Name: name,
			View: isView == 1,
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
		CASE WHEN EXISTS (SELECT 1 FROM all_views WHERE owner = :1 AND view_name = :2) THEN 1 ELSE 0 END AS is_view,
		c.column_name,
		c.data_type
	FROM all_tab_columns c
	WHERE c.owner = :3 AND c.table_name = :4
	ORDER BY c.column_id
	`

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryxContext(ctx, q, databaseSchema, table, databaseSchema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := &drivers.TableMetadata{
		Schema: make(map[string]string),
	}
	for rows.Next() {
		var isView int
		var colName, dataType string
		if err := rows.Scan(&isView, &colName, &dataType); err != nil {
			return nil, err
		}
		res.View = isView == 1
		res.Schema[colName] = dataType
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
