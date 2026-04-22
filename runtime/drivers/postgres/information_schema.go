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

func (c *connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := `
	SELECT
		table_name,
		table_type = 'VIEW' AS is_view,
		current_schema() = $1 AS is_default_database_schema
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
	var isView, isDefaultDatabaseSchema bool
	for rows.Next() {
		if err := rows.Scan(&name, &isView, &isDefaultDatabaseSchema); err != nil {
			return nil, "", err
		}
		res = append(res, &drivers.TableInfo{
			Name:                    name,
			View:                    isView,
			IsDefaultDatabase:       true,
			IsDefaultDatabaseSchema: isDefaultDatabaseSchema,
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
		CASE WHEN lower(t.table_type) = 'view' THEN true ELSE false END AS view,
		c.column_name, 
		c.data_type
	FROM information_schema.tables t JOIN information_schema.columns c ON t.table_name = c.table_name AND t.table_schema = c.table_schema
	WHERE c.table_schema = coalesce($1, current_schema()) AND c.table_name = $2
	ORDER BY ordinal_position
	`
	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	var args []any
	if databaseSchema != "" {
		args = append(args, databaseSchema, table)
	} else {
		args = append(args, nil, table)
	}
	rows, err := db.QueryContext(ctx, q, args...)
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

// All implements drivers.InformationSchema.
func (c *connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.InformationSchema.
func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	return nil
}

// LoadDDL implements drivers.InformationSchema.
// Note: table.Database is not used; in Postgres, the database is determined by the connection.
func (c *connection) LoadDDL(ctx context.Context, table *drivers.OlapTable) error {
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

// Lookup implements drivers.InformationSchema.
func (c *connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	meta, err := c.GetTable(ctx, db, schema, name)
	if err != nil {
		return nil, err
	}

	rtSchema := &runtimev1.StructType{}
	for name, typ := range meta.Schema {
		t := databaseTypeToPB(typ)
		rtSchema.Fields = append(rtSchema.Fields, &runtimev1.StructType_Field{
			Name: name,
			Type: t,
		})
	}
	return &drivers.OlapTable{
		Database:          db,
		DatabaseSchema:    schema,
		Name:              name,
		View:              meta.View,
		Schema:            rtSchema,
		UnsupportedCols:   nil,
		PhysicalSizeBytes: 0,
	}, nil
}
