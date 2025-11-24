package starrocks

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
	return drivers.DialectStarRocks
}

// MayBeScaledToZero implements drivers.OLAPStore.
func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return false // StarRocks instances are typically always running
}

// WithConnection implements drivers.OLAPStore.
func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	// StarRocks supports connection affinity for temp tables
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	conn, err := db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create wrapped context with connection
	wrappedCtx := context.WithValue(ctx, connCtxKey{}, conn)
	ensuredCtx := context.WithValue(context.Background(), connCtxKey{}, conn)

	return fn(wrappedCtx, ensuredCtx)
}

// connCtxKey is used to store connection in context for WithConnection.
type connCtxKey struct{}

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

// Query implements drivers.OLAPStore.
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (res *drivers.Result, resErr error) {
	if c.logQueries {
		c.logger.Info("StarRocks query",
			zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)),
			zap.Any("args", stmt.Args),
			observability.ZapCtx(ctx))
	}

	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// Handle dry run with EXPLAIN
	if stmt.DryRun {
		_, err = db.ExecContext(ctx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
		return nil, err
	}

	// Check if we have a connection from WithConnection
	var rows *sqlx.Rows
	if conn, ok := ctx.Value(connCtxKey{}).(*sqlx.Conn); ok && conn != nil {
		rows, err = conn.QueryxContext(ctx, stmt.Query, stmt.Args...)
	} else {
		rows, err = db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	}
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

	starrocksRows := &starrocksRows{
		Rows:     rows,
		scanDest: prepareScanDest(schema),
		colTypes: cts,
	}
	res = &drivers.Result{Rows: starrocksRows, Schema: schema}
	return res, nil
}

// QuerySchema implements drivers.OLAPStore.
func (c *connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	// Use EXPLAIN to get schema without executing
	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// Execute with LIMIT 0 to get schema (only add if not already present)
	finalQuery := query
	if !strings.Contains(strings.ToUpper(query), "LIMIT") {
		finalQuery = query + " LIMIT 0"
	}
	rows, err := db.QueryxContext(ctx, finalQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToSchema(rows)
}

// InformationSchema implements drivers.OLAPStore.
func (c *connection) InformationSchema() drivers.OLAPInformationSchema {
	return c
}

// All implements drivers.OLAPInformationSchema.
func (c *connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.OLAPInformationSchema.
func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	// StarRocks doesn't easily expose physical size per table in information_schema
	// This could be extended to query system tables if needed
	return nil
}

// Lookup implements drivers.OLAPInformationSchema.
func (c *connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	// Use default database if schema is empty
	// In StarRocks, schema is equivalent to database
	if schema == "" {
		schema = c.configProp.Database
	}

	// StarRocks doesn't use the catalog concept for table references
	// Always pass empty string for db to GetTable
	meta, err := c.GetTable(ctx, "", schema, name)
	if err != nil {
		return nil, err
	}

	rtSchema := &runtimev1.StructType{}
	for colName, colType := range meta.Schema {
		rtSchema.Fields = append(rtSchema.Fields, &runtimev1.StructType_Field{
			Name: colName,
			Type: databaseTypeToPB(colType, true),
		})
	}

	return &drivers.OlapTable{
		Database:          "", // StarRocks doesn't use catalog in table references
		DatabaseSchema:    schema,
		Name:              name,
		View:              meta.View,
		Schema:            rtSchema,
		UnsupportedCols:   nil,
		PhysicalSizeBytes: 0,
	}, nil
}

// rowsToSchema extracts schema from query result.
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

// databaseTypeToPB converts StarRocks types to Rill's generic schema type.
// StarRocks supports the following data types:
//
// Numeric types:
//   - BOOLEAN: 1-byte boolean
//   - TINYINT: 1-byte signed integer (-128 to 127)
//   - SMALLINT: 2-byte signed integer
//   - INT/INTEGER: 4-byte signed integer
//   - BIGINT: 8-byte signed integer
//   - LARGEINT: 16-byte signed integer (128-bit)
//   - FLOAT: 4-byte floating point
//   - DOUBLE: 8-byte floating point
//   - DECIMAL/DECIMALV2/DECIMAL32/DECIMAL64/DECIMAL128/DECIMAL256: Fixed-point decimal
//
// String types:
//   - CHAR: Fixed-length string (1-255 bytes)
//   - VARCHAR: Variable-length string
//   - STRING/TEXT: Alias for VARCHAR(1048576)
//   - BINARY/VARBINARY: Binary strings
//
// Date/Time types (Note: StarRocks does NOT support TIME type):
//   - DATE: Calendar date ('0000-01-01' to '9999-12-31'), format: YYYY-MM-DD
//   - DATETIME: Date and time ('0000-01-01 00:00:00' to '9999-12-31 23:59:59')
//
// Semi-structured types:
//   - JSON: JSON data (since v2.2.0)
//   - ARRAY: Array of elements (since v3.1)
//   - MAP: Key-value pairs (since v3.1)
//   - STRUCT: Named fields collection (since v3.1)
//
// Special types:
//   - HLL: HyperLogLog for approximate distinct count
//   - BITMAP: Bitmap for set operations
//   - PERCENTILE: Percentile aggregation type
//
// Reference: https://docs.starrocks.io/docs/sql-reference/data-types/
func databaseTypeToPB(dbt string, nullable bool) *runtimev1.Type {
	t := &runtimev1.Type{Nullable: nullable}
	upperDbt := strings.ToUpper(dbt)

	// Handle parameterized types (e.g., DECIMAL(10,2), VARCHAR(255))
	baseDbt := upperDbt
	if idx := strings.Index(upperDbt, "("); idx != -1 {
		baseDbt = upperDbt[:idx]
	}

	switch baseDbt {
	// Boolean
	case "BOOLEAN", "BOOL":
		t.Code = runtimev1.Type_CODE_BOOL

	// Integer types
	case "TINYINT":
		t.Code = runtimev1.Type_CODE_INT8
	case "SMALLINT":
		t.Code = runtimev1.Type_CODE_INT16
	case "INT", "INTEGER":
		t.Code = runtimev1.Type_CODE_INT32
	case "BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "LARGEINT":
		t.Code = runtimev1.Type_CODE_INT128

	// Floating point types
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "DOUBLE":
		t.Code = runtimev1.Type_CODE_FLOAT64

	// Decimal types
	case "DECIMAL", "DECIMALV2", "DECIMAL32", "DECIMAL64", "DECIMAL128", "DECIMAL256":
		t.Code = runtimev1.Type_CODE_DECIMAL

	// String types
	case "CHAR", "VARCHAR", "STRING", "TEXT":
		t.Code = runtimev1.Type_CODE_STRING
	case "BINARY", "VARBINARY":
		t.Code = runtimev1.Type_CODE_BYTES

	// Date/Time types
	// StarRocks only supports DATE and DATETIME, not TIME
	// Reference: https://docs.starrocks.io/docs/sql-reference/data-types/
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "DATETIME":
		t.Code = runtimev1.Type_CODE_TIMESTAMP

	// Semi-structured types (StarRocks specific)
	case "JSON", "JSONB":
		t.Code = runtimev1.Type_CODE_JSON
	case "ARRAY":
		t.Code = runtimev1.Type_CODE_ARRAY
	case "MAP":
		t.Code = runtimev1.Type_CODE_MAP
	case "STRUCT":
		t.Code = runtimev1.Type_CODE_STRUCT

	// Special types
	case "HLL":
		t.Code = runtimev1.Type_CODE_STRING // HyperLogLog for approximate distinct count
	case "BITMAP":
		t.Code = runtimev1.Type_CODE_STRING // Bitmap for set operations
	case "PERCENTILE":
		t.Code = runtimev1.Type_CODE_STRING // Percentile aggregation type

	// NULL type
	case "NULL":
		t.Code = runtimev1.Type_CODE_UNSPECIFIED

	default:
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	}

	return t
}

// starrocksRows wraps sqlx.Rows to provide MapScan method with proper type handling.
// This is required because MySQL driver returns byte arrays if correct types aren't provided.
type starrocksRows struct {
	*sqlx.Rows
	scanDest []any
	colTypes []*sql.ColumnType
}

func (r *starrocksRows) MapScan(dest map[string]any) error {
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
				dbType := strings.ToUpper(r.colTypes[i].DatabaseTypeName())
				if dbType == "TINYINT" {
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

// prepareScanDest creates scan destinations based on schema.
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
		case runtimev1.Type_CODE_INT64, runtimev1.Type_CODE_INT128:
			dest = &sql.NullInt64{}
		case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
			dest = &sql.NullFloat64{}
		case runtimev1.Type_CODE_DECIMAL:
			dest = &sql.NullString{} // Decimals are returned as strings for precision
		case runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_BYTES:
			dest = &sql.NullString{}
		case runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIMESTAMP:
			// StarRocks DATE and DATETIME types
			dest = &sql.NullTime{}
		case runtimev1.Type_CODE_JSON:
			dest = &sql.NullString{}
		case runtimev1.Type_CODE_ARRAY, runtimev1.Type_CODE_MAP, runtimev1.Type_CODE_STRUCT:
			dest = &sql.NullString{} // Complex types returned as JSON strings
		default:
			dest = new(any)
		}
		scanList[i] = dest
	}
	return scanList
}
