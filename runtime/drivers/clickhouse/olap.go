package clickhouse

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

const (
	numRetries = 3
	retryWait  = 300 * time.Millisecond
)

var colDataTypePattern = regexp.MustCompile(`Nullable\(([^)]+)\)`)

var _ drivers.OLAPStore = &connection{}

// AddTableColumn implements drivers.OLAPStore.
func (c *connection) AddTableColumn(ctx context.Context, tableName, columnName, typ string) error {
	return fmt.Errorf("clickhouse: data transformation not yet supported")
}

// AlterTableColumn implements drivers.OLAPStore.
func (c *connection) AlterTableColumn(ctx context.Context, tableName, columnName, newType string) error {
	return fmt.Errorf("clickhouse: data transformation not yet supported")
}

// CreateTableAsSelect implements drivers.OLAPStore.
func (c *connection) CreateTableAsSelect(ctx context.Context, name string, view bool, sql string) error {
	if view {
		return c.Exec(ctx, &drivers.Statement{
			Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", safeSQLName(name), sql),
		})
	}
	return c.Exec(ctx, &drivers.Statement{
		Query: fmt.Sprintf("CREATE OR REPLACE TABLE %s ENGINE = MergeTree ORDER BY tuple() AS %s", safeSQLName(name), sql),
	})
}

// DropTable implements drivers.OLAPStore.
func (c *connection) DropTable(ctx context.Context, name string, view bool) error {
	var typ string
	if view {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}
	return c.Exec(ctx, &drivers.Statement{
		Query: fmt.Sprintf("DROP %s %s", typ, safeSQLName(name)),
	})
}

// InsertTableAsSelect implements drivers.OLAPStore.
func (c *connection) InsertTableAsSelect(ctx context.Context, name string, byName bool, sql string) error {
	return fmt.Errorf("clickhouse: data transformation not yet supported")
}

// RenameTable implements drivers.OLAPStore.
func (c *connection) RenameTable(ctx context.Context, name, newName string, view bool) error {
	if !view {
		return c.Exec(ctx, &drivers.Statement{
			Query: fmt.Sprintf("RENAME TABLE %s TO %s", safeSQLName(name), safeSQLName(newName)),
		})
	}
	res, err := c.Execute(ctx, &drivers.Statement{
		Query: fmt.Sprintf("SHOW CREATE VIEW %s", safeSQLName(name)),
	})
	if err != nil {
		return err
	}

	var sql string
	if res.Next() {
		if err := res.Scan(&sql); err != nil {
			res.Close()
			return err
		}
	}
	res.Close()

	sql = strings.Replace(sql, name, safeSQLName(newName), 1)
	return c.Exec(ctx, &drivers.Statement{
		Query: sql,
	})
}

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectClickHouse
}

func (c *connection) WithConnection(ctx context.Context, priority int, longRunning, tx bool, fn drivers.WithConnectionFunc) error {
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	return fn(ctx, ctx, conn)
}

func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	var cancelFunc context.CancelFunc
	if stmt.ExecutionTimeout != 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, stmt.ExecutionTimeout)
	}
	_, err := c.db.ExecContext(ctx, stmt.Query, stmt.Args...)
	if cancelFunc != nil {
		cancelFunc()
	}
	return err
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

	var cancelFunc context.CancelFunc
	if stmt.ExecutionTimeout != 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, stmt.ExecutionTimeout)
	}

	var rows *sqlx.Rows
	var err error

	re := retrier.New(retrier.ExponentialBackoff(numRetries, retryWait), retryErrClassifier{})
	err = re.RunCtx(ctx, func(ctx2 context.Context) error {
		rows, err = c.db.QueryxContext(ctx2, stmt.Query, stmt.Args...)
		return err
	})

	if err != nil {
		if cancelFunc != nil {
			cancelFunc()
		}
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		rows.Close()
		if cancelFunc != nil {
			cancelFunc()
		}

		return nil, err
	}

	r := &drivers.Result{Rows: rows, Schema: schema}
	r.SetCleanupFunc(func() error {
		if cancelFunc != nil {
			cancelFunc()
		}

		return nil
	})

	return r, nil
}

func (c *connection) DropDB() error {
	return fmt.Errorf("dropping database not supported")
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
		WHERE T.TABLE_SCHEMA = 'default'
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
			C.IS_NULLABLE = '1' AS IS_NULLABLE
		FROM INFORMATION_SCHEMA.TABLES T 
		JOIN INFORMATION_SCHEMA.COLUMNS C ON T.TABLE_SCHEMA = C.TABLE_SCHEMA AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.TABLE_SCHEMA = 'default' AND T.TABLE_NAME = ?
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
	// for nullable the datatype is Nullable(X)
	match := colDataTypePattern.FindStringSubmatch(dbt)
	if len(match) >= 2 {
		dbt = match[1]
	}

	dbt = strings.ToUpper(dbt)
	t := &runtimev1.Type{Nullable: nullable}
	// int256 and uint256
	switch dbt {
	case "BOOL":
		t.Code = runtimev1.Type_CODE_BOOL
	case "INT8":
		t.Code = runtimev1.Type_CODE_INT8
	case "INT16":
		t.Code = runtimev1.Type_CODE_INT16
	case "INT32":
		t.Code = runtimev1.Type_CODE_INT32
	case "INT64":
		t.Code = runtimev1.Type_CODE_INT64
	case "INT128":
		t.Code = runtimev1.Type_CODE_INT128
	case "UINT8":
		t.Code = runtimev1.Type_CODE_UINT8
	case "UINT16":
		t.Code = runtimev1.Type_CODE_UINT16
	case "UINT32":
		t.Code = runtimev1.Type_CODE_UINT32
	case "UINT64":
		t.Code = runtimev1.Type_CODE_UINT64
	case "UINT128":
		t.Code = runtimev1.Type_CODE_UINT128
	case "FLOAT32":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "FLOAT64":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "DECIMAL": // todo :: check decimal
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "FIXEDSTRING":
		t.Code = runtimev1.Type_CODE_STRING
	case "STRING":
		t.Code = runtimev1.Type_CODE_STRING
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "DATE32":
		t.Code = runtimev1.Type_CODE_DATE
	case "DATETIME":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "DATETIME64":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "JSON":
		t.Code = runtimev1.Type_CODE_JSON
	case "UUID":
		t.Code = runtimev1.Type_CODE_UUID
	case "OTHER":
		t.Code = runtimev1.Type_CODE_JSON
	}

	return t, nil
}

// retryErrClassifier classifies 429 errors as retryable and all other errors as non retryable
type retryErrClassifier struct{}

func (retryErrClassifier) Classify(err error) retrier.Action {
	if err == nil {
		return retrier.Succeed
	}

	if strings.Contains(err.Error(), "QueryCapacityExceededException") {
		return retrier.Retry
	}

	return retrier.Fail
}
