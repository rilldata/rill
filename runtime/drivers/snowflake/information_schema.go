package snowflake

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	db, err := c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

	curDBName, curSchemaName, err := getCurrentDatabaseAndSchema(ctx, db.DB)
	if err != nil {
		return nil, "", err
	}
	rows, err := db.QueryxContext(ctx, "SHOW TERSE SCHEMAS IN ACCOUNT")
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute SHOW TERSE SCHEMAS IN ACCOUNT: %w", err)
	}
	defer rows.Close()

	var res []*drivers.DatabaseSchemaInfo
	var schemaName, dbName string
	var createdOn, kind, sn any
	for rows.Next() {
		if err := rows.Scan(&createdOn, &schemaName, &kind, &dbName, &sn); err != nil {
			return nil, "", fmt.Errorf("failed to scan schema row: %w", err)
		}

		// Skip the SNOWFLAKE database and INFORMATION_SCHEMA schema unless they are the current database or schema in use.
		if (strings.EqualFold(dbName, "SNOWFLAKE") && !strings.EqualFold(curDBName, "SNOWFLAKE")) || (strings.EqualFold(schemaName, "INFORMATION_SCHEMA") && !strings.EqualFold(curSchemaName, "INFORMATION_SCHEMA")) {
			continue
		}

		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       dbName,
			DatabaseSchema: schemaName,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	start := 0
	if pageToken != "" {
		var err error
		start, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}
	end := start + limit
	if end >= len(res) {
		end = len(res)
	}

	if start >= len(res) {
		return []*drivers.DatabaseSchemaInfo{}, "", nil
	}

	next := ""
	if end < len(res) {
		next = fmt.Sprintf("%d", end)
	}

	return res[start:end], next, nil
}

func (c *connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := fmt.Sprintf(`
		SELECT
			table_name,
			CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS view
		FROM %s.INFORMATION_SCHEMA.TABLES
		WHERE table_schema = ?`, sqlSafeName(database))
	var args []any
	args = append(args, databaseSchema)
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += `	AND table_name > ?
		ORDER BY table_name
		LIMIT ?
		`
		args = append(args, startAfter, limit+1)
	} else {
		q += `
		ORDER BY table_name
		LIMIT ?
		`
		args = append(args, limit+1)
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}
	rows, err := db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var res []*drivers.TableInfo
	var name string
	var view bool
	for rows.Next() {
		if err := rows.Scan(&name, &view); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.TableInfo{
			Name: name,
			View: view,
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
	q := fmt.Sprintf(`
		SELECT
			CASE WHEN t.table_type = 'VIEW' THEN true ELSE false END as is_view,
			c.column_name,
			c.data_type
		FROM %s.INFORMATION_SCHEMA.TABLES t
		JOIN %s.INFORMATION_SCHEMA.COLUMNS c
		ON t.table_schema = c.table_schema AND t.table_name = c.table_name 
		WHERE t.table_schema = ? AND t.table_name = ?
		ORDER BY c.ordinal_position
	`, sqlSafeName(database), sqlSafeName(database))

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryxContext(ctx, q, databaseSchema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	t := &drivers.TableMetadata{
		Schema: make(map[string]string),
	}
	var (
		colName, colType string
		isView           bool
	)
	for rows.Next() {
		if err := rows.Scan(&isView, &colName, &colType); err != nil {
			return nil, err
		}
		t.Schema[colName] = colType
		t.View = isView
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return t, nil
}

func getCurrentDatabaseAndSchema(ctx context.Context, db *sql.DB) (string, string, error) {
	query := "SELECT CURRENT_DATABASE(), CURRENT_SCHEMA()"

	var currentDB, currentSchema sql.NullString
	err := db.QueryRowContext(ctx, query).Scan(&currentDB, &currentSchema)
	if err != nil {
		return "", "", fmt.Errorf("failed to get current database and schema: %w", err)
	}
	var dbName string
	if currentDB.Valid {
		dbName = currentDB.String
	}
	var schemaName string
	if currentSchema.Valid {
		schemaName = currentSchema.String
	}
	return dbName, schemaName, nil
}

func sqlSafeName(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf("%q", escaped)
}
