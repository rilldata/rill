package databricks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)

	q := `
	SELECT
		catalog_name,
		schema_name
	FROM system.information_schema.schemata
	WHERE (catalog_name NOT IN ('samples', 'system') OR catalog_name = ?)
		AND (schema_name != 'information_schema' OR schema_name = ?)
	`
	args := []any{c.config.Catalog, c.config.Schema}
	if pageToken != "" {
		var afterCatalog, afterSchema string
		if err := pagination.UnmarshalPageToken(pageToken, &afterCatalog, &afterSchema); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		q += "	AND (catalog_name > ? OR (catalog_name = ? AND schema_name > ?))"
		args = append(args, afterCatalog, afterCatalog, afterSchema)
	}
	q += `
	ORDER BY catalog_name, schema_name
	LIMIT CAST(? AS INT)
	`
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
		last := res[len(res)-1]
		next = pagination.MarshalPageToken(last.Database, last.DatabaseSchema)
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
		LIMIT CAST(? AS INT)
		`
		args = append(args, startAfter, limit+1)
	} else {
		q += `
		ORDER BY table_name
		LIMIT CAST(? AS INT)
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

func (c *connection) Lookup(ctx context.Context, database, databaseSchema, name string) (*drivers.OlapTable, error) {
	prefix := catalogPrefix(database)

	conn, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// Query the table type and columns separately rather than JOINing
	// information_schema.tables and information_schema.columns. The join forces a
	// shuffle that Lakehouse//RT's Photon rejects (PHOTON_INTERNAL_ERROR), after
	// which RT refuses the retry; two filtered point-lookups avoid the shuffle and
	// are equivalent on DBSQL.
	var tableType string
	err = conn.QueryRowContext(ctx,
		fmt.Sprintf("SELECT table_type FROM %sinformation_schema.tables WHERE table_schema = ? AND table_name = ?", prefix),
		databaseSchema, name,
	).Scan(&tableType)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	rows, err := conn.QueryContext(ctx,
		fmt.Sprintf("SELECT column_name, data_type FROM %sinformation_schema.columns WHERE table_schema = ? AND table_name = ? ORDER BY ordinal_position", prefix),
		databaseSchema, name,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []*runtimev1.StructType_Field
	var colName, colType string
	for rows.Next() {
		if err := rows.Scan(&colName, &colType); err != nil {
			return nil, err
		}
		fields = append(fields, &runtimev1.StructType_Field{
			Name: colName,
			Type: databaseTypeToPB(colType),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &drivers.OlapTable{
		Database:       database,
		DatabaseSchema: databaseSchema,
		Name:           name,
		View:           tableType == "VIEW",
		Schema:         &runtimev1.StructType{Fields: fields},
	}, nil
}

// All implements drivers.OLAPInformationSchema.
func (c *connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.OLAPInformationSchema.
func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	return nil
}

// LoadDDL implements drivers.OLAPInformationSchema.
func (c *connection) LoadDDL(ctx context.Context, table *drivers.OlapTable) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	fqn := DialectDatabricks.EscapeTable(table.Database, table.DatabaseSchema, table.Name)

	objectType := "TABLE"
	if table.View {
		objectType = "VIEW"
	}

	var ddl string
	err = db.QueryRowContext(ctx, fmt.Sprintf("SHOW CREATE %s %s", objectType, fqn)).Scan(&ddl)
	if err != nil {
		return err
	}
	table.DDL = ddl
	return nil
}

// catalogPrefix returns "<catalog>." if catalog is non-empty, or "" otherwise.
func catalogPrefix(catalog string) string {
	if catalog == "" {
		return ""
	}
	return DatabricksEscapeIdentifier(catalog) + "."
}
