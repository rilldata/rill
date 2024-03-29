package clickhouse

import (
	"context"
	"errors"
	"slices"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var ignoreSchemas = []string{"INFORMATION_SCHEMA", "information_schema", "system"}

type informationSchema struct {
	c *connection
}

func (c *connection) InformationSchema() drivers.InformationSchema {
	return informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context) ([]*drivers.Table, error) {
	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	var databaseSchemas []string
	var defaultSchema string

	row := conn.QueryRowxContext(ctx, "SELECT currentDatabase()")
	if err := row.Scan(&defaultSchema); err != nil {
		return nil, err
	}

	rows, err := conn.QueryxContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	var schema string
	for rows.Next() {
		if err := rows.Scan(&schema); err != nil {
			rows.Close()
			return nil, err
		}
		if slices.Contains(ignoreSchemas, schema) {
			continue
		}
		databaseSchemas = append(databaseSchemas, schema)
	}
	rows.Close()

	var res []*drivers.Table
	for _, databaseSchema := range databaseSchemas {
		q := `
		SELECT
			T.table_schema AS SCHEMA,
			T.table_name AS NAME,
			T.table_type AS TABLE_TYPE, 
			C.column_name AS COLUMNS,
			C.data_type AS COLUMN_TYPE,
			C.ordinal_position as ORDINAL_POSITION
		FROM information_schema.tables T 
		JOIN information_schema.columns C ON T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema = ?
		ORDER BY SCHEMA, NAME, TABLE_TYPE, ORDINAL_POSITION
	`

		rows, err := conn.QueryxContext(ctx, q, databaseSchema)
		if err != nil {
			return nil, err
		}

		tables, err := i.scanTables(rows, databaseSchema == defaultSchema)
		if err != nil {
			rows.Close()
			return nil, err
		}

		rows.Close()
		res = append(res, tables...)
	}
	return res, nil
}

func (i informationSchema) Lookup(ctx context.Context, db, schema, name string) (*drivers.Table, error) {
	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	var q string
	var args []any
	// table_catalog and table_schema both means the name of the database in which the table is located in clickhouse
	// we map it to our internal schema field
	var isDefaultSchema bool
	if schema == "" {
		q = `
		SELECT
			T.table_schema AS SCHEMA,
			T.table_name AS NAME,
			T.table_type AS TABLE_TYPE, 
			C.column_name AS COLUMNS,
			C.data_type AS COLUMN_TYPE,
			C.ordinal_position as ORDINAL_POSITION
		FROM information_schema.tables T 
		JOIN information_schema.columns C ON T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema = currentDatabase() AND T.table_name = ?
		ORDER BY SCHEMA, NAME, TABLE_TYPE, ORDINAL_POSITION
	`
		args = append(args, name)
		isDefaultSchema = true
	} else {
		q = `
		SELECT
			T.table_schema AS SCHEMA,
			T.table_name AS NAME,
			T.table_type AS TABLE_TYPE, 
			C.column_name AS COLUMNS,
			C.data_type AS COLUMN_TYPE,
			C.ordinal_position as ORDINAL_POSITION
		FROM information_schema.tables T 
		JOIN information_schema.columns C ON T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema = ? AND T.table_name = ?
		ORDER BY SCHEMA, NAME, TABLE_TYPE, ORDINAL_POSITION
	`
		args = append(args, schema, name)

		// get current database
		row := conn.QueryRowContext(ctx, "SELECT currentDatabase()")
		var currentSchema string
		if err := row.Scan(&currentSchema); err != nil {
			return nil, err
		}
		isDefaultSchema = schema == currentSchema
	}

	rows, err := conn.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := i.scanTables(rows, isDefaultSchema)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (i informationSchema) scanTables(rows *sqlx.Rows, isDefaultDatabaseSchema bool) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for rows.Next() {
		var databaseSchema string
		var name string
		var tableType string
		var columnName string
		var columnType string
		var oridinalPosition int

		err := rows.Scan(&databaseSchema, &name, &tableType, &columnName, &columnType, &oridinalPosition)
		if err != nil {
			return nil, err
		}

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.Table
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.DatabaseSchema == databaseSchema && t.Name == name) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.Table{
				DatabaseSchema:          databaseSchema,
				Name:                    name,
				View:                    tableType == "VIEW",
				Schema:                  &runtimev1.StructType{},
				IsDefaultDatabaseSchema: isDefaultDatabaseSchema,
			}
			res = append(res, t)
		}

		// parse column type
		colType, err := databaseTypeToPB(columnType, false)
		if err != nil {
			if !errors.Is(err, errUnsupportedType) {
				return nil, err
			}
			if t.UnsupportedCols == nil {
				t.UnsupportedCols = make(map[string]string)
			}
			t.UnsupportedCols[columnName] = columnType
			continue
		}

		// append column
		t.Schema.Fields = append(t.Schema.Fields, &runtimev1.StructType_Field{
			Name: columnName,
			Type: colType,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
