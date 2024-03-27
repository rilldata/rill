package clickhouse

import (
	"context"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

type informationSchema struct {
	c *connection
}

func (c *connection) InformationSchema() drivers.InformationSchema {
	return informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context) ([]*drivers.Table, error) {
	q := `
		SELECT
			T.table_catalog AS DATABASE,
			T.table_schema AS SCHEMA,
			T.table_name AS NAME,
			T.table_type AS TABLE_TYPE, 
			C.column_name AS COLUMNS,
			C.data_type AS COLUMN_TYPE,
			C.ordinal_position as ORDINAL_POSITION
		FROM information_schema.tables T 
		JOIN information_schema.columns C ON T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema = currentDatabase()
		ORDER BY DATABASE, SCHEMA, NAME, TABLE_TYPE, ORDINAL_POSITION
	`

	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	rows, err := conn.QueryxContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

func (i informationSchema) Lookup(ctx context.Context, name string) (*drivers.Table, error) {
	q := `
		SELECT
			T.table_catalog AS DATABASE,
			T.table_schema AS SCHEMA,
			T.table_name AS NAME,
			T.table_type AS TABLE_TYPE, 
			C.column_name AS COLUMNS,
			C.data_type AS COLUMN_TYPE,
			C.ordinal_position as ORDINAL_POSITION
		FROM information_schema.tables T 
		JOIN information_schema.columns C ON T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema = currentDatabase() AND T.table_name = ?
		ORDER BY DATABASE, SCHEMA, NAME, TABLE_TYPE, ORDINAL_POSITION
	`

	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	rows, err := conn.QueryxContext(ctx, q, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (i informationSchema) scanTables(rows *sqlx.Rows) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for rows.Next() {
		var database string
		var schema string
		var name string
		var tableType string
		var columnName string
		var columnType string
		var oridinalPosition int

		err := rows.Scan(&database, &schema, &name, &tableType, &columnName, &columnType, &oridinalPosition)
		if err != nil {
			return nil, err
		}

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.Table
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.Database == database && t.DatabaseSchema == schema && t.Name == name) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.Table{
				Database:       database,
				DatabaseSchema: schema,
				Name:           name,
				View:           tableType == "VIEW",
				Schema:         &runtimev1.StructType{},
			}
			res = append(res, t)
		}

		// parse column type
		colType, err := databaseTypeToPB(columnType, false)
		if err != nil {
			return nil, err
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
