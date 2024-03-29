package clickhouse

import (
	"context"
	"errors"
	"slices"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var ignoreDatabases = []string{"INFORMATION_SCHEMA", "information_schema", "system"}

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

	var databases []string
	var defaultDatabase string

	row := conn.QueryRowxContext(ctx, "SELECT currentDatabase()")
	if err := row.Scan(&defaultDatabase); err != nil {
		return nil, err
	}

	rows, err := conn.QueryxContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	var db string
	for rows.Next() {
		if err := rows.Scan(&db); err != nil {
			rows.Close()
			return nil, err
		}
		if slices.Contains(ignoreDatabases, db) {
			continue
		}
		databases = append(databases, db)
	}
	rows.Close()

	var res []*drivers.Table
	for _, database := range databases {
		q := `
		SELECT
			T.table_catalog AS DATABASE,
			T.table_name AS NAME,
			T.table_type AS TABLE_TYPE, 
			C.column_name AS COLUMNS,
			C.data_type AS COLUMN_TYPE,
			C.ordinal_position as ORDINAL_POSITION
		FROM information_schema.tables T 
		JOIN information_schema.columns C ON T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema = ?
		ORDER BY DATABASE, NAME, TABLE_TYPE, ORDINAL_POSITION
	`

		rows, err := conn.QueryxContext(ctx, q, database)
		if err != nil {
			return nil, err
		}

		tables, err := i.scanTables(rows, database == defaultDatabase)
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
	// table_catalog and table_schema both means the name of the database in which the table is located in clickhouse.
	// we use either db or schema arg to set table_schema
	var isDefaultDatabase bool
	if db == "" {
		q = `
		SELECT
			T.table_catalog AS DATABASE,
			T.table_name AS NAME,
			T.table_type AS TABLE_TYPE, 
			C.column_name AS COLUMNS,
			C.data_type AS COLUMN_TYPE,
			C.ordinal_position as ORDINAL_POSITION
		FROM information_schema.tables T 
		JOIN information_schema.columns C ON T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema = currentDatabase() AND T.table_name = ?
		ORDER BY DATABASE, NAME, TABLE_TYPE, ORDINAL_POSITION
	`
		args = append(args, name)
		isDefaultDatabase = true
	} else {
		q = `
		SELECT
			T.table_catalog AS DATABASE,
			T.table_name AS NAME,
			T.table_type AS TABLE_TYPE, 
			C.column_name AS COLUMNS,
			C.data_type AS COLUMN_TYPE,
			C.ordinal_position as ORDINAL_POSITION
		FROM information_schema.tables T 
		JOIN information_schema.columns C ON T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema = ? AND T.table_name = ?
		ORDER BY DATABASE, NAME, TABLE_TYPE, ORDINAL_POSITION
	`
		args = append(args, db, name)

		// get current database
		row := conn.QueryRowContext(ctx, "SELECT currentDatabase()")
		var currentDatabase string
		if err := row.Scan(&currentDatabase); err != nil {
			return nil, err
		}
		isDefaultDatabase = db == currentDatabase
	}

	rows, err := conn.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := i.scanTables(rows, isDefaultDatabase)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (i informationSchema) scanTables(rows *sqlx.Rows, isDefaultDatabase bool) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for rows.Next() {
		var database string
		var name string
		var tableType string
		var columnName string
		var columnType string
		var oridinalPosition int

		err := rows.Scan(&database, &name, &tableType, &columnName, &columnType, &oridinalPosition)
		if err != nil {
			return nil, err
		}

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.Table
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.Database == database && t.Name == name) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.Table{
				Database:          database,
				Name:              name,
				View:              tableType == "VIEW",
				Schema:            &runtimev1.StructType{},
				IsDefaultDatabase: isDefaultDatabase,
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
