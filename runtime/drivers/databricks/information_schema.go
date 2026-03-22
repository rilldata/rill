package databricks

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

// ListDatabaseSchemas lists catalogs and schemas in Databricks Unity Catalog.
// Databricks hierarchy: catalog > schema > table.
// Catalogs map to Database, schemas map to DatabaseSchema.
func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	db, err := c.getDB(ctx)
	if err != nil {
		return nil, "", err
	}

	// If a catalog is configured, only list schemas in that catalog.
	// Otherwise, list all accessible catalogs and their schemas.
	var catalogs []string
	if c.config.Catalog != "" {
		catalogs = []string{c.config.Catalog}
	} else {
		rows, err := db.QueryxContext(ctx, "SHOW CATALOGS")
		if err != nil {
			return nil, "", fmt.Errorf("failed to list catalogs: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var catalog string
			if err := rows.Scan(&catalog); err != nil {
				return nil, "", err
			}
			catalogs = append(catalogs, catalog)
		}
		if err := rows.Err(); err != nil {
			return nil, "", err
		}
	}

	var res []*drivers.DatabaseSchemaInfo
	for _, catalog := range catalogs {
		rows, err := db.QueryxContext(ctx, fmt.Sprintf("SHOW SCHEMAS IN %s", sqlSafeName(catalog)))
		if err != nil {
			return nil, "", fmt.Errorf("failed to list schemas in catalog %q: %w", catalog, err)
		}

		for rows.Next() {
			var schemaName string
			if err := rows.Scan(&schemaName); err != nil {
				rows.Close()
				return nil, "", err
			}
			// Skip the information_schema unless it's the configured schema.
			if strings.EqualFold(schemaName, "information_schema") && !strings.EqualFold(c.config.Schema, "information_schema") {
				continue
			}
			res = append(res, &drivers.DatabaseSchemaInfo{
				Database:       catalog,
				DatabaseSchema: schemaName,
			})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, "", err
		}
	}

	// Paginate results.
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	start := 0
	if pageToken != "" {
		start, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}
	end := start + limit
	if end > len(res) {
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

// ListTables lists tables in a given catalog and schema.
func (c *connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := fmt.Sprintf(`
		SELECT
			table_name,
			CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS is_view
		FROM %s.information_schema.tables
		WHERE table_schema = ?`, sqlSafeName(database))

	var args []any
	args = append(args, databaseSchema)
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += fmt.Sprintf(` AND table_name > ? ORDER BY table_name LIMIT %d`, limit+1)
		args = append(args, startAfter)
	} else {
		q += fmt.Sprintf(` ORDER BY table_name LIMIT %d`, limit+1)
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
	for rows.Next() {
		var name string
		var view bool
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

// GetTable returns column metadata for a specific table.
func (c *connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	fqn := fmt.Sprintf("%s.%s.%s", sqlSafeName(database), sqlSafeName(databaseSchema), sqlSafeName(table))

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryxContext(ctx, fmt.Sprintf("DESCRIBE TABLE %s", fqn))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	t := &drivers.TableMetadata{
		Schema: make(map[string]string),
	}
	for rows.Next() {
		var colName, colType string
		var comment sql.NullString
		if err := rows.Scan(&colName, &colType, &comment); err != nil {
			return nil, err
		}
		// DESCRIBE TABLE may include partition/metadata rows with empty col names.
		if colName == "" || strings.HasPrefix(colName, "#") {
			break
		}
		t.Schema[colName] = colType
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return t, nil
}

// sqlSafeName quotes an identifier to prevent injection.
func sqlSafeName(name string) string {
	escaped := strings.ReplaceAll(name, "`", "``")
	return fmt.Sprintf("`%s`", escaped)
}
