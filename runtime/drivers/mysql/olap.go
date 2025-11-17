package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	if stmt.DryRun {
		return nil
	}
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
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (res *drivers.Result, resErr error) {
	if c.logQueries {
		c.logger.Info("MySQL query", zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)), zap.Any("args", stmt.Args), observability.ZapCtx(ctx))
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
	defer func() {
		if resErr != nil {
			_ = rows.Close()
		}
	}()

	schema, err := rowsToSchema(rows)
	if err != nil {
		return nil, err
	}

	cts, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	mySQLRows := &mysqlRows{
		Rows:     rows,
		scanDest: prepareScanDest(schema),
		colTypes: cts,
	}
	res = &drivers.Result{Rows: mySQLRows, Schema: schema}
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
		rtSchema.Fields = append(rtSchema.Fields, &runtimev1.StructType_Field{
			Name: name,
			Type: databaseTypeToPB(typ, true),
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

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: databaseTypeToPB(ct.DatabaseTypeName(), nullable),
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}

func databaseTypeToPB(dbt string, nullable bool) *runtimev1.Type {
	t := &runtimev1.Type{Nullable: nullable}
	switch strings.ToUpper(dbt) {
	case "DECIMAL":
		t.Code = runtimev1.Type_CODE_STRING
	case "NUMERIC":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "BIT":
		t.Code = runtimev1.Type_CODE_STRING
	case "TINYINT":
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
		t.Code = runtimev1.Type_CODE_INT16
	case "JSON":
		t.Code = runtimev1.Type_CODE_JSON
	case "GEOMETRY":
		t.Code = runtimev1.Type_CODE_STRING
	case "NULL":
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	default:
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	}
	return t
}

// mysqlRows wraps sqlx.Rows to provide MapScan method.
// This is required because if the correct type is not provided to Scan mysql driver just returns byte arrays.
// sqlx driver scans into any instead of actual types.
type mysqlRows struct {
	*sqlx.Rows
	scanDest []any
	colTypes []*sql.ColumnType
}

func (r *mysqlRows) MapScan(dest map[string]any) error {
	err := r.Rows.Scan(r.scanDest...)
	if err != nil {
		return err
	}
	for i, ct := range r.colTypes {
		fieldName := ct.Name()
		valPtr := r.scanDest[i]
		if valPtr == nil {
			dest[fieldName] = nil
			continue
		}
		switch valPtr := valPtr.(type) {
		case *sql.NullBool:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Bool
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullByte:
			if valPtr.Valid {
				if strings.ToUpper(r.colTypes[i].DatabaseTypeName()) == "TINYINT" {
					dest[fieldName] = int8(valPtr.Byte)
				} else {
					dest[fieldName] = valPtr.Byte
				}
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullInt16:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Int16
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullInt32:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Int32
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullInt64:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Int64
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullFloat64:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Float64
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullString:
			if valPtr.Valid {
				dest[fieldName] = valPtr.String
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullTime:
			if valPtr.Valid {
				dest[fieldName] = valPtr.Time
			} else {
				dest[fieldName] = nil
			}
		default:
			dest[fieldName] = *(valPtr.(*any))
		}
	}
	return nil
}

func prepareScanDest(schema *runtimev1.StructType) []any {
	scanList := make([]any, len(schema.Fields))
	for i, field := range schema.Fields {
		var dest any
		switch field.Type.Code {
		case runtimev1.Type_CODE_BOOL:
			dest = &sql.NullBool{}
		case runtimev1.Type_CODE_INT8:
			dest = &sql.NullByte{}
		case runtimev1.Type_CODE_INT16:
			dest = &sql.NullInt16{}
		case runtimev1.Type_CODE_INT32:
			dest = &sql.NullInt32{}
		case runtimev1.Type_CODE_INT64:
			dest = &sql.NullInt64{}
		case runtimev1.Type_CODE_FLOAT64:
			dest = &sql.NullFloat64{}
		case runtimev1.Type_CODE_STRING:
			dest = &sql.NullString{}
		case runtimev1.Type_CODE_DATE:
			dest = &sql.NullString{}
		case runtimev1.Type_CODE_TIME:
			dest = &sql.NullString{}
		case runtimev1.Type_CODE_TIMESTAMP:
			// the driver does not parse time.Time unless parseTime query param is set.
			// even when param is set it can scan into string for TIMESTAMP type.
			dest = &sql.NullString{}
		case runtimev1.Type_CODE_JSON:
			dest = &sql.NullString{}
		default:
			dest = new(any)
		}
		scanList[i] = dest
	}
	return scanList
}
