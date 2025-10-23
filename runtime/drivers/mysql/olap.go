package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

var _ drivers.OLAPStore = (*connection)(nil)

// Dialect implements drivers.OLAPStore.
func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectMySQL
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
	return false // MySQL instances are typically always running
}

// Query implements drivers.OLAPStore.
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	if c.logQueries {
		c.logger.Info("MySQL query", zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)), zap.Any("args", stmt.Args), observability.ZapCtx(ctx))
	}

	db, err := c.acquireDB(ctx)
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

	res := &drivers.Result{Rows: rows, Schema: schema}
	return res, nil
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

// Lookup implements drivers.OLAPInformationSchema.
func (c *connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	meta, err := c.GetTable(ctx, db, schema, name)
	if err != nil {
		return nil, err
	}

	rtSchema := &runtimev1.StructType{}
	for name, typ := range meta.Schema {
		t, err := databaseTypeToPB(typ, true)
		if err != nil {
			return nil, err
		}
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

func databaseTypeToPB(dbt string, nullable bool) (*runtimev1.Type, error) {
	t := &runtimev1.Type{Nullable: nullable}
	switch dbt {
	case "DECIMAL":
		t.Code = runtimev1.Type_CODE_STRING
	case "NUMERIC":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "BIT", "TINYINT":
		t.Code = runtimev1.Type_CODE_INT8
	case "SMALLINT":
		t.Code = runtimev1.Type_CODE_INT16
	case "MEDIUMINT":
		t.Code = runtimev1.Type_CODE_INT32
	case "INT", "INTEGER":
		t.Code = runtimev1.Type_CODE_INT32
	case "BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "UNSIGNED BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "BOOLEAN", "BOOL":
		t.Code = runtimev1.Type_CODE_BOOL
	case "FLOAT", "DOUBLE", "REAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "CHAR", "VARCHAR", "BINARY", "VARBINARY", "TINYBLOB", "BLOB", "MEDIUMBLOB", "LONGBLOB", "TINYTEXT", "TEXT", "MEDIUMTEXT", "LONGTEXT":
		t.Code = runtimev1.Type_CODE_STRING
	case "ENUM", "SET":
		t.Code = runtimev1.Type_CODE_STRING
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "TIME":
		t.Code = runtimev1.Type_CODE_TIME
	case "DATETIME", "TIMESTAMP":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "YEAR":
		t.Code = runtimev1.Type_CODE_INT64
	case "JSON":
		t.Code = runtimev1.Type_CODE_JSON
	case "GEOMETRY":
		t.Code = runtimev1.Type_CODE_STRING
	case "NULL":
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	default:
		return nil, fmt.Errorf("unhandled MySQL type: %s", dbt)
	}
	return t, nil
}

var (
	scanTypeFloat32   = reflect.TypeOf(float32(0))
	scanTypeFloat64   = reflect.TypeOf(float64(0))
	scanTypeInt8      = reflect.TypeOf(int8(0))
	scanTypeInt16     = reflect.TypeOf(int16(0))
	scanTypeInt32     = reflect.TypeOf(int32(0))
	scanTypeInt64     = reflect.TypeOf(int64(0))
	scanTypeNullFloat = reflect.TypeOf(sql.NullFloat64{})
	scanTypeNullInt   = reflect.TypeOf(sql.NullInt64{})
	scanTypeUint8     = reflect.TypeOf(uint8(0))
	scanTypeUint16    = reflect.TypeOf(uint16(0))
	scanTypeUint32    = reflect.TypeOf(uint32(0))
	scanTypeUint64    = reflect.TypeOf(uint64(0))
)

func scanTypeForDatabaseType(dbt string, nullable bool) (reflect.Type, error) {
	t := &runtimev1.Type{Nullable: nullable}
	switch dbt {
	case "DECIMAL":
		
	case "NUMERIC":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "BIT", "TINYINT":
		t.Code = runtimev1.Type_CODE_INT8
	case "SMALLINT":
		t.Code = runtimev1.Type_CODE_INT16
	case "MEDIUMINT":
		t.Code = runtimev1.Type_CODE_INT32
	case "INT", "INTEGER":
		t.Code = runtimev1.Type_CODE_INT32
	case "BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "UNSIGNED BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "BOOLEAN", "BOOL":
		t.Code = runtimev1.Type_CODE_BOOL
	case "FLOAT", "DOUBLE", "REAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "CHAR", "VARCHAR", "BINARY", "VARBINARY", "TINYBLOB", "BLOB", "MEDIUMBLOB", "LONGBLOB", "TINYTEXT", "TEXT", "MEDIUMTEXT", "LONGTEXT":
		t.Code = runtimev1.Type_CODE_STRING
	case "ENUM", "SET":
		t.Code = runtimev1.Type_CODE_STRING
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "TIME":
		t.Code = runtimev1.Type_CODE_TIME
	case "DATETIME", "TIMESTAMP":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "YEAR":
		t.Code = runtimev1.Type_CODE_INT64
	case "JSON":
		t.Code = runtimev1.Type_CODE_JSON
	case "GEOMETRY":
		t.Code = runtimev1.Type_CODE_STRING
	case "NULL":
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	default:
		return nil, fmt.Errorf("unhandled MySQL type: %s", dbt)
	}
	return t, nil
}
