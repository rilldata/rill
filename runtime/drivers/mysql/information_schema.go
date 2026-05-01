package mysql

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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

func (c *connection) ListTables(ctx context.Context, _, databaseSchema, like string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := `
	SELECT
		table_schema,
		table_name,
		CASE WHEN table_type = 'VIEW' THEN true ELSE false END AS view,
		DATABASE() = table_schema AS is_default_database_schema
	FROM information_schema.tables
	WHERE `

	var args []any
	q += "table_schema = ?"
	args = append(args, databaseSchema)

	if like != "" {
		q += " AND LOWER(table_name) LIKE LOWER(?)"
		args = append(args, like)
	}
	if pageToken != "" {
		var afterSchema, afterName string
		if err := pagination.UnmarshalPageToken(pageToken, &afterSchema, &afterName); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += " AND (table_schema > ? OR (table_schema = ? AND table_name > ?))"
		args = append(args, afterSchema, afterSchema, afterName)
	}
	q += `
	ORDER BY table_schema, table_name
	LIMIT ?
	`
	args = append(args, limit+1)

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
		var schemaName, name string
		var typ, isDefaultDatabaseSchema bool
		if err := rows.Scan(&schemaName, &name, &typ, &isDefaultDatabaseSchema); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.TableInfo{
			Database:                "",
			DatabaseSchema:          schemaName,
			Name:                    name,
			View:                    typ,
			IsDefaultDatabase:       true,
			IsDefaultDatabaseSchema: isDefaultDatabaseSchema,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	next := ""
	if len(res) > limit {
		res = res[:limit]
		last := res[len(res)-1]
		next = pagination.MarshalPageToken(last.DatabaseSchema, last.Name)
	}
	return res, next, nil
}

// Lookup implements drivers.InformationSchema.
func (c *connection) Lookup(ctx context.Context, database, databaseSchema, table string) (*drivers.TableInfo, error) {
	q := `
	SELECT
		CASE WHEN t.table_type = 'VIEW' THEN true ELSE false END AS view,
		DATABASE() = c.table_schema AS is_default_database_schema,
		c.column_name,
		c.data_type
	FROM information_schema.tables t
	JOIN information_schema.columns c
	ON t.table_schema = c.table_schema AND t.table_name = c.table_name
	WHERE c.table_schema = ? AND c.table_name = ?
	ORDER BY ordinal_position
	`

	mdb, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := mdb.QueryxContext(ctx, q, databaseSchema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var isView, isDefaultDatabaseSchema bool
	var fields []*runtimev1.StructType_Field
	for rows.Next() {
		var colName, dataType string
		if err := rows.Scan(&isView, &isDefaultDatabaseSchema, &colName, &dataType); err != nil {
			return nil, err
		}
		fields = append(fields, &runtimev1.StructType_Field{
			Name: colName,
			Type: databaseTypeToPB(dataType, true),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, drivers.ErrNotFound
	}
	return &drivers.TableInfo{
		Database:                database,
		DatabaseSchema:          databaseSchema,
		Name:                    table,
		View:                    isView,
		IsDefaultDatabase:       true,
		IsDefaultDatabaseSchema: isDefaultDatabaseSchema,
		Schema:                  &runtimev1.StructType{Fields: fields},
		UnsupportedCols:         nil,
		PhysicalSizeBytes:       0,
	}, nil
}

// LoadPhysicalSize implements drivers.InformationSchema.
func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.TableInfo) error {
	return nil
}

// LoadDDL implements drivers.InformationSchema.
func (c *connection) LoadDDL(ctx context.Context, table *drivers.TableInfo) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	// SHOW CREATE TABLE works for both tables and views in MySQL.
	// For tables it returns columns: [Table, Create Table].
	// For views it returns columns: [View, Create View, character_set_client, collation_connection].
	// We extract the DDL by column name to avoid depending on column order or count.
	rows, err := db.QueryxContext(ctx, fmt.Sprintf("SHOW CREATE TABLE %s", c.Dialect().EscapeTable(table.Database, table.DatabaseSchema, table.Name)))
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		res := make(map[string]any)
		if err := rows.MapScan(res); err != nil {
			return err
		}
		for _, key := range []string{"Create Table", "Create View"} {
			if v, ok := res[key]; ok && v != nil {
				if b, ok := v.([]byte); ok {
					table.DDL = string(b)
				}
				break
			}
		}
	}
	return rows.Err()
}
