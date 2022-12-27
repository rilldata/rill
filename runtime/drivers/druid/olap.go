package druid

import (
	"context"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectDruid
}

func (c *connection) Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	return drivers.ErrUnsupportedConnector
}

func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	panic("not implemented")
}

func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Execute(ctx, stmt)
	if err != nil {
		return err
	}
	return res.Close()
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	if stmt.DryRun {
		// TODO: Find way to validate with args
		prepared, err := c.db.PrepareContext(ctx, stmt.Query)
		if err != nil {
			return nil, err
		}
		return nil, prepared.Close()
	}

	rows, err := c.db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		return nil, err
	}

	return &drivers.Result{Rows: rows, Schema: schema}, nil
}

func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	if r == nil {
		return nil, nil
	}

	cts, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(cts))
	for i, ct := range cts {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		t, err := databaseTypeToPB(ct.DatabaseTypeName(), nullable)
		if err != nil {
			return nil, err
		}

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: t,
		}
	}

	err = r.Err()
	if err != nil {
		return nil, err
	}

	return &runtimev1.StructType{Fields: fields}, nil
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
		WHERE T.TABLE_SCHEMA = 'druid'
		ORDER BY DATABASE, SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`

	rows, err := i.c.db.QueryxContext(ctx, q)
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
			T.TABLE_CATALOG AS DATABASE,
			T.TABLE_SCHEMA AS SCHEMA,
			T.TABLE_NAME AS NAME,
			T.TABLE_TYPE AS TABLE_TYPE, 
			C.COLUMN_NAME AS COLUMN_NAME,
			C.DATA_TYPE AS COLUMN_TYPE,
			C.IS_NULLABLE = 'YES' AS IS_NULLABLE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA = 'druid' AND T.TABLE_NAME = ?
		ORDER BY DATABASE, SCHEMA, NAME, TABLE_TYPE, C.ORDINAL_POSITION
	`

	rows, err := i.c.db.QueryxContext(ctx, q, name)
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
		var nullable bool

		err := rows.Scan(&database, &schema, &name, &tableType, &columnName, &columnType, &nullable)
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
				Schema:         &runtimev1.StructType{},
			}
			res = append(res, t)
		}

		// parse column type
		colType, err := databaseTypeToPB(columnType, nullable)
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

func databaseTypeToPB(dbt string, nullable bool) (*runtimev1.Type, error) {
	t := &runtimev1.Type{Nullable: nullable}
	switch dbt {
	case "BOOLEAN":
		t.Code = runtimev1.Type_CODE_BOOL
	case "TINYINT":
		t.Code = runtimev1.Type_CODE_INT8
	case "SMALLINT":
		t.Code = runtimev1.Type_CODE_INT16
	case "INTEGER":
		t.Code = runtimev1.Type_CODE_INT32
	case "BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "DOUBLE":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "REAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "DECIMAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "CHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "VARCHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "TIMESTAMP":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "DATE":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "OTHER":
		t.Code = runtimev1.Type_CODE_JSON
	}

	return t, nil
}
