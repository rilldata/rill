package postgres

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

func (c *connection) ListTables(ctx context.Context, database, databaseSchema, like string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := `
	SELECT
		table_schema,
		table_name,
		table_type = 'VIEW' AS is_view,
		current_schema() = table_schema AS is_default_database_schema
	FROM information_schema.tables
	WHERE `

	var args []any
	n := 1
	q += fmt.Sprintf("table_schema = $%d", n)
	args = append(args, databaseSchema)
	n++

	if like != "" {
		q += fmt.Sprintf(" AND LOWER(table_name) LIKE LOWER($%d)", n)
		args = append(args, like)
		n++
	}
	if pageToken != "" {
		var afterSchema, afterName string
		if err := pagination.UnmarshalPageToken(pageToken, &afterSchema, &afterName); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += fmt.Sprintf(" AND (table_schema > $%d OR (table_schema = $%d AND table_name > $%d))", n, n, n+1)
		args = append(args, afterSchema, afterName)
		n += 2
	}
	q += fmt.Sprintf(" ORDER BY table_schema, table_name LIMIT $%d", n)
	args = append(args, limit+1)

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
	var schemaName, name string
	var isView, isDefaultDatabaseSchema bool
	for rows.Next() {
		if err := rows.Scan(&schemaName, &name, &isView, &isDefaultDatabaseSchema); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.TableInfo{
			Database:                database,
			DatabaseSchema:          schemaName,
			Name:                    name,
			View:                    isView,
			IsDefaultDatabase:       true,
			IsDefaultDatabaseSchema: isDefaultDatabaseSchema,
		})
	}
	next := ""
	if len(res) > limit {
		res = res[:limit]
		last := res[len(res)-1]
		next = pagination.MarshalPageToken(last.DatabaseSchema, last.Name)
	}
	return res, next, rows.Err()
}

// All implements drivers.InformationSchema.
func (c *connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// Lookup implements drivers.InformationSchema.
func (c *connection) Lookup(ctx context.Context, database, databaseSchema, table string) (*drivers.TableInfo, error) {
	q := `
	SELECT
		CASE WHEN lower(t.table_type) = 'view' THEN true ELSE false END AS view,
		current_schema() = c.table_schema AS is_default_database_schema,
		c.column_name,
		c.data_type
	FROM information_schema.tables t JOIN information_schema.columns c ON t.table_name = c.table_name AND t.table_schema = c.table_schema
	WHERE c.table_schema = coalesce($1, current_schema()) AND c.table_name = $2
	ORDER BY ordinal_position
	`
	pdb, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	var schemaArg any
	if databaseSchema != "" {
		schemaArg = databaseSchema
	}
	rows, err := pdb.QueryContext(ctx, q, schemaArg, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var isView, isDefaultDatabaseSchema bool
	var fields []*runtimev1.StructType_Field
	for rows.Next() {
		var colName, typ string
		if err := rows.Scan(&isView, &isDefaultDatabaseSchema, &colName, &typ); err != nil {
			return nil, err
		}
		fields = append(fields, &runtimev1.StructType_Field{
			Name: colName,
			Type: databaseTypeToPB(typ),
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
// Note: table.Database is not used; in Postgres, the database is determined by the connection.
func (c *connection) LoadDDL(ctx context.Context, table *drivers.TableInfo) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	schema := table.DatabaseSchema
	if schema == "" {
		if err := db.QueryRowContext(ctx, "SELECT current_schema()").Scan(&schema); err != nil {
			return err
		}
	}

	if table.View {
		// For views: use pg_get_viewdef
		var ddl string
		q := `
			SELECT 'CREATE VIEW ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname) || ' AS ' || pg_get_viewdef(c.oid, true)
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = $1 AND c.relname = $2 AND c.relkind IN ('v', 'm')
		`
		err = db.QueryRowContext(ctx, q, schema, table.Name).Scan(&ddl)
		if err != nil {
			return err
		}
		table.DDL = ddl
		return nil
	}

	// Postgres does not have a built-in way to get the DDL for a table, so we reconstruct a basic CREATE TABLE statement from the available metadata (won't include indexes, constraints, etc.).
	q := `
		SELECT
			'CREATE TABLE ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname) || ' (' ||
			string_agg(
				quote_ident(a.attname) || ' ' || format_type(a.atttypid, a.atttypmod) ||
				CASE WHEN a.attnotnull THEN ' NOT NULL' ELSE '' END,
				', ' ORDER BY a.attnum
			) || ')'
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_attribute a ON a.attrelid = c.oid
		WHERE n.nspname = $1 AND c.relname = $2 AND a.attnum > 0 AND NOT a.attisdropped
		GROUP BY n.nspname, c.relname
	`
	var ddl string
	err = db.QueryRowContext(ctx, q, schema, table.Name).Scan(&ddl)
	if err != nil {
		return err
	}
	table.DDL = ddl
	return nil
}
