package snowflake

import (
	"context"
	"database/sql"
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

	curDBName, curSchemaName, err := getCurrentDatabaseAndSchema(ctx, db.DB)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryxContext(ctx, "SHOW TERSE SCHEMAS IN ACCOUNT")
	if err != nil {
		return nil, fmt.Errorf("failed to execute SHOW TERSE SCHEMAS IN ACCOUNT: %w", err)
	}
	defer rows.Close()

	var results []*drivers.DatabaseSchemaInfo
	var schemaName, dbName string
	var createdOn, kind, sn any
	for rows.Next() {
		if err := rows.Scan(&createdOn, &schemaName, &kind, &dbName, &sn); err != nil {
			return nil, fmt.Errorf("failed to scan schema row: %w", err)
		}

		// Skip the SNOWFLAKE database and INFORMATION_SCHEMA schema unless they are the current database or schema in use.
		if (strings.EqualFold(dbName, "SNOWFLAKE") && !strings.EqualFold(curDBName, "SNOWFLAKE")) || (strings.EqualFold(schemaName, "INFORMATION_SCHEMA") && !strings.EqualFold(curSchemaName, "INFORMATION_SCHEMA")) {
			continue
		}

		results = append(results, &drivers.DatabaseSchemaInfo{
			Database:       dbName,
			DatabaseSchema: schemaName,
		})
	}
	return results, rows.Err()
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
	var name string
	var view bool
	for rows.Next() {
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
	var colName, colType string
	for rows.Next() {
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
