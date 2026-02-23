package databricks

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

var _ drivers.OLAPStore = (*connection)(nil)

// Dialect implements drivers.OLAPStore.
func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectDatabricks
}

// Exec implements drivers.OLAPStore.
func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Query(ctx, stmt)
	if err != nil {
		return err
	}
	if res != nil {
		return res.Close()
	}
	return nil
}

// InformationSchema implements drivers.OLAPStore.
func (c *connection) InformationSchema() drivers.OLAPInformationSchema {
	return c
}

// MayBeScaledToZero implements drivers.OLAPStore.
func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return true
}

// Query implements drivers.OLAPStore.
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	if c.config.LogQueries {
		c.logger.Info("Databricks query",
			zap.String("sql", stmt.Query),
			zap.Any("args", stmt.Args),
			observability.ZapCtx(ctx),
		)
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	if stmt.DryRun {
		_, err = db.ExecContext(ctx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
		return nil, err
	}

	rows, err := db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		_ = rows.Close()
		return nil, err
	}

	return &drivers.Result{Rows: rows, Schema: schema}, nil
}

// QuerySchema implements drivers.OLAPStore.
func (c *connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	return nil, drivers.ErrNotImplemented
}

// WithConnection implements drivers.OLAPStore.
func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	return drivers.ErrNotImplemented
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

	fqn := drivers.DialectDatabricks.EscapeTable(table.Database, table.DatabaseSchema, table.Name)
	var ddl string
	err = db.QueryRowContext(ctx, fmt.Sprintf("SHOW CREATE TABLE %s", fqn)).Scan(&ddl)
	if err != nil {
		return err
	}
	table.DDL = ddl
	return nil
}

// Lookup implements drivers.OLAPInformationSchema.
func (c *connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	meta, err := c.GetTable(ctx, db, schema, name)
	if err != nil {
		return nil, err
	}

	rtSchema := &runtimev1.StructType{}
	for colName, colType := range meta.Schema {
		t, err := databaseTypeToPB(colType, true)
		if err != nil {
			return nil, err
		}
		rtSchema.Fields = append(rtSchema.Fields, &runtimev1.StructType_Field{
			Name: colName,
			Type: t,
		})
	}

	return &drivers.OlapTable{
		Database:       db,
		DatabaseSchema: schema,
		Name:           name,
		View:           meta.View,
		Schema:         rtSchema,
	}, nil
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

// databaseTypeToPB maps Databricks/Spark SQL types to proto types.
func databaseTypeToPB(dbt string, nullable bool) (*runtimev1.Type, error) {
	t := &runtimev1.Type{Nullable: nullable}
	switch dbt {
	case "BOOLEAN":
		t.Code = runtimev1.Type_CODE_BOOL
	case "TINYINT", "BYTE":
		t.Code = runtimev1.Type_CODE_INT8
	case "SMALLINT", "SHORT":
		t.Code = runtimev1.Type_CODE_INT16
	case "INT", "INTEGER":
		t.Code = runtimev1.Type_CODE_INT32
	case "BIGINT", "LONG":
		t.Code = runtimev1.Type_CODE_INT64
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "DOUBLE":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "DECIMAL", "DEC", "NUMERIC":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "STRING", "VARCHAR", "CHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "BINARY":
		t.Code = runtimev1.Type_CODE_BYTES
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "TIMESTAMP", "TIMESTAMP_NTZ":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "INTERVAL":
		t.Code = runtimev1.Type_CODE_INTERVAL
	case "ARRAY", "MAP", "STRUCT", "VARIANT":
		t.Code = runtimev1.Type_CODE_JSON
	case "NULL", "VOID":
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	default:
		t.Code = runtimev1.Type_CODE_STRING
	}
	return t, nil
}
