package clickhouse

import (
	"context"
	"strings"

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
	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	var databases []string
	dbs, ok := i.c.config["databases"].(string)
	if ok {
		databases = strings.Split(dbs, ",")
	} else {
		row := conn.QueryRowxContext(ctx, "SELECT currentDatabase()")
		var db string
		if err := row.Scan(&db); err != nil {
			return nil, err
		}
		databases = append(databases, db)
	}

	var res []*drivers.Table
	for _, database := range databases {
		q := `
		SELECT
			T.TABLE_CATALOG AS DATABASE,
			T.TABLE_NAME AS NAME,
			T.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMNS,
			C.DATA_TYPE AS COLUMN_TYPE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA = ?
		ORDER BY DATABASE, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`

		rows, err := conn.QueryxContext(ctx, q, database)
		if err != nil {
			return nil, err
		}

		tables, err := i.scanTables(rows)
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
	var q string
	var args []any
	// table_catalog and table_schema both means the name of the database in which the table is located in clickhouse.
	// we use either db or schema arg to set table_schema
	if db == "" && schema == "" {
		q = `
		SELECT
			T.TABLE_CATALOG AS DATABASE,
			T.TABLE_NAME AS NAME,
			T.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMN_NAME,
			C.DATA_TYPE AS COLUMN_TYPE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA = currentDatabase() AND T.TABLE_NAME = ?
		ORDER BY DATABASE, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`
		args = append(args, name)
	} else {
		q = `
		SELECT
			T.TABLE_CATALOG AS DATABASE,
			T.TABLE_NAME AS NAME,
			T.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMN_NAME,
			C.DATA_TYPE AS COLUMN_TYPE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA = ? AND T.TABLE_NAME = ?
		ORDER BY DATABASE, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`
		if db == "" {
			args = append(args, schema, name)
		} else {
			args = append(args, db, name)
		}
	}

	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	rows, err := conn.QueryxContext(ctx, q, args...)
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
		var name string
		var tableType string
		var columnName string
		var columnType string

		err := rows.Scan(&database, &name, &tableType, &columnName, &columnType)
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
				Database: database,
				Name:     name,
				View:     tableType == "VIEW",
				Schema:   &runtimev1.StructType{},
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
