package oracle

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
	return drivers.DialectOracle
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
	return false
}

// Query implements drivers.OLAPStore.
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (res *drivers.Result, resErr error) {
	if c.logQueries {
		fields := []zap.Field{
			zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)),
			zap.Any("args", stmt.Args),
			observability.ZapCtx(ctx),
		}
		if len(stmt.QueryAttributes) > 0 {
			fields = append(fields, zap.Any("query_attributes", stmt.QueryAttributes))
		}
		c.logger.Info("Oracle query", fields...)
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	if stmt.DryRun {
		_, err = db.ExecContext(ctx, fmt.Sprintf("EXPLAIN PLAN FOR %s", stmt.Query), stmt.Args...)
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

	oRows := &oracleRows{
		Rows:     rows,
		scanDest: prepareScanDest(schema),
		colTypes: cts,
	}
	res = &drivers.Result{Rows: oRows, Schema: schema}
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

// LoadDDL implements drivers.OLAPInformationSchema.
func (c *connection) LoadDDL(ctx context.Context, table *drivers.OlapTable) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	objectType := "TABLE"
	if table.View {
		objectType = "VIEW"
	}

	var ddl string
	err = db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT DBMS_METADATA.GET_DDL('%s', :1, :2) FROM DUAL", objectType),
		table.Name, table.DatabaseSchema,
	).Scan(&ddl)
	if err != nil {
		// DDL retrieval is optional; don't fail
		return nil
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
	for colName, typ := range meta.Schema {
		rtSchema.Fields = append(rtSchema.Fields, &runtimev1.StructType_Field{
			Name: colName,
			Type: databaseTypeToPB(typ, true),
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
	case "NUMBER":
		// Oracle NUMBER is versatile; map to float64 for general use
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "FLOAT", "BINARY_FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "BINARY_DOUBLE":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "INTEGER", "INT", "SMALLINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "VARCHAR2", "NVARCHAR2", "CHAR", "NCHAR", "LONG", "ROWID", "UROWID":
		t.Code = runtimev1.Type_CODE_STRING
	case "CLOB", "NCLOB":
		t.Code = runtimev1.Type_CODE_STRING
	case "BLOB", "RAW", "LONG RAW":
		t.Code = runtimev1.Type_CODE_BYTES
	case "DATE":
		// Oracle DATE includes time components
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIMESTAMP", "TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITH LOCAL TIME ZONE":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "INTERVAL YEAR TO MONTH", "INTERVAL DAY TO SECOND":
		t.Code = runtimev1.Type_CODE_STRING
	case "XMLTYPE":
		t.Code = runtimev1.Type_CODE_STRING
	case "JSON":
		t.Code = runtimev1.Type_CODE_JSON
	case "BOOLEAN", "BOOL":
		t.Code = runtimev1.Type_CODE_BOOL
	default:
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	}
	return t
}

// oracleRows wraps sqlx.Rows to provide proper type scanning.
// The go-ora driver may return different Go types than expected by sqlx's MapScan,
// so we provide typed scan destinations.
type oracleRows struct {
	*sqlx.Rows
	scanDest []any
	colTypes []*sql.ColumnType
}

func (r *oracleRows) MapScan(dest map[string]any) error {
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
		switch v := valPtr.(type) {
		case *sql.NullBool:
			if v.Valid {
				dest[fieldName] = v.Bool
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullInt64:
			if v.Valid {
				dest[fieldName] = v.Int64
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullFloat64:
			if v.Valid {
				dest[fieldName] = v.Float64
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullString:
			if v.Valid {
				dest[fieldName] = v.String
			} else {
				dest[fieldName] = nil
			}
		case *sql.NullTime:
			if v.Valid {
				dest[fieldName] = v.Time
			} else {
				dest[fieldName] = nil
			}
		default:
			dest[fieldName] = *(v.(*any))
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
		case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64:
			dest = &sql.NullInt64{}
		case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
			dest = &sql.NullFloat64{}
		case runtimev1.Type_CODE_STRING:
			dest = &sql.NullString{}
		case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIME:
			dest = &sql.NullTime{}
		case runtimev1.Type_CODE_JSON:
			dest = &sql.NullString{}
		default:
			dest = new(any)
		}
		scanList[i] = dest
	}
	return scanList
}
