package databricks

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
		catalog_name,
		schema_name
	FROM information_schema.schemata
	WHERE schema_name != 'information_schema'
	`
	var args []any
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += `	AND schema_name > ?
		ORDER BY catalog_name, schema_name
		LIMIT ?
		`
		args = append(args, startAfter, limit+1)
	} else {
		q += `
		ORDER BY catalog_name, schema_name
		LIMIT ?
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
	var catalog, schema string
	for rows.Next() {
		if err := rows.Scan(&catalog, &schema); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.DatabaseSchemaInfo{
			Database:       catalog,
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

	q := fmt.Sprintf(`
	SELECT
		table_name,
		CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS is_view
	FROM %sinformation_schema.tables
	WHERE table_schema = ?
	`, catalogPrefix(database))
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
	prefix := catalogPrefix(database)
	q := fmt.Sprintf(`
	SELECT
		CASE WHEN t.table_type = 'VIEW' THEN true ELSE false END AS is_view,
		c.column_name,
		c.data_type
	FROM %sinformation_schema.tables t
	JOIN %sinformation_schema.columns c
	ON t.table_schema = c.table_schema AND t.table_name = c.table_name
	WHERE t.table_schema = ? AND t.table_name = ?
	ORDER BY c.ordinal_position
	`, prefix, prefix)

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, q, databaseSchema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	t := &drivers.TableMetadata{
		Schema: make(map[string]string),
	}
	var colName, colType string
	var isView bool
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

// catalogPrefix returns "<catalog>." if catalog is non-empty, or "" otherwise.
func catalogPrefix(catalog string) string {
	if catalog == "" {
		return ""
	}
	return drivers.DialectDatabricks.EscapeIdentifier(catalog) + "."
}
