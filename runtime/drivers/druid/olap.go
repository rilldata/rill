package druid

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/connectors/sources"
	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (*sqlx.Rows, error) {
	if stmt.DryRun {
		// TODO: Find way to validate with args
		prepared, err := c.db.PrepareContext(ctx, stmt.Query)
		if err != nil {
			return nil, err
		}
		prepared.Close()
		return nil, nil
	}

	rows, err := c.db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

type informationSchema struct {
	c *connection
}

func (c *connection) InformationSchema() drivers.InformationSchema {
	return informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context) ([]*drivers.Table, error) {
	q := `
		SELECT
			T.TABLE_CATALOG AS DATABASE,
			T.TABLE_SCHEMA AS SCHEMA,
			T.TABLE_NAME AS NAME,
			T.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMNS,
			C.DATA_TYPE AS COLUMN_TYPE,
			C.IS_NULLABLE = 'YES' AS IS_NULLABLE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA NOT IN ('INFORMATION_SCHEMA', 'sys')
		ORDER BY DATABASE, SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`

	rows, err := i.c.db.QueryxContext(ctx, q)
	if err != nil {
		return nil, err
	}

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

func (i informationSchema) Lookup(ctx context.Context, name string) (*drivers.Table, error) {
	q := `
		SELECT
			T.TABLE_CATALOG AS DATABASE,
			T.TABLE_SCHEMA AS SCHEMA,
			T.TABLE_NAME AS NAME,
			T.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMN_NAME,
			C.DATA_TYPE AS COLUMN_TYPE,
			C.IS_NULLABLE = 'YES' AS IS_NULLABLE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA NOT IN ('INFORMATION_SCHEMA', 'sys') AND T.TABLE_NAME = ?
		ORDER BY DATABASE, SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`

	rows, err := i.c.db.QueryxContext(ctx, q, name)
	if err != nil {
		return nil, err
	}

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
		var nullable bool

		err := rows.Scan(&database, &schema, &name, &tableType, &columnName, &columnType, &nullable)
		if err != nil {
			return nil, err
		}

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.Table
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.Database == database && t.Schema == schema && t.Name == name && t.Type == tableType) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.Table{
				Database: database,
				Schema:   schema,
				Name:     name,
				Type:     tableType,
			}
			res = append(res, t)
		}

		// append column
		t.Columns = append(t.Columns, drivers.Column{
			Name:     columnName,
			Type:     columnType,
			Nullable: nullable,
		})
	}

	return res, nil
}

func (c *connection) Ingest(ctx context.Context, source sources.Source) (*sqlx.Rows, error) {
	return nil, drivers.ErrUnsupportedConnector
}
