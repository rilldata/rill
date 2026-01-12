package starrocks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

var errUnsupportedType = errors.New("encountered unsupported StarRocks type")

var _ drivers.OLAPStore = (*connection)(nil)

// Dialect implements drivers.OLAPStore.
func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectStarRocks
}

// MayBeScaledToZero implements drivers.OLAPStore.
func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return false
}

// WithConnection implements drivers.OLAPStore.
// StarRocks is a read-only OLAP connector and does not support connection affinity operations.
// This is only used by model executors, which are not supported for StarRocks.
func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	return fmt.Errorf("starrocks: WithConnection not supported")
}

// Exec implements drivers.OLAPStore.
func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	db := c.db
	var err error

	if c.configProp.LogQueries {
		c.logger.Info("StarRocks exec",
			zap.String("query", stmt.Query),
			zap.Any("args", stmt.Args))
	}

	// Handle DryRun: validate query without execution
	if stmt.DryRun {
		_, err = db.ExecContext(ctx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
		return err
	}

	_, err = db.ExecContext(ctx, stmt.Query, stmt.Args...)
	return err
}

// Query implements drivers.OLAPStore.
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	db := c.db

	if c.configProp.LogQueries {
		c.logger.Info("StarRocks query",
			zap.String("query", stmt.Query),
			zap.Any("args", stmt.Args))
	}

	// Handle DryRun: validate query without execution
	if stmt.DryRun {
		rows, err := db.QueryxContext(ctx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
		if err != nil {
			return nil, err
		}
		rows.Close()
		// Return nil result for dry run (query is valid)
		return nil, nil
	}

	rows, err := db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		return nil, err
	}

	schema, err := c.rowsToSchema(rows)
	if err != nil {
		rows.Close()
		return nil, err
	}

	cts, err := rows.ColumnTypes()
	if err != nil {
		rows.Close()
		return nil, err
	}

	starrocksRows := &starrocksRows{
		Rows:     rows,
		scanDest: prepareScanDest(schema),
		colTypes: cts,
	}

	return &drivers.Result{
		Rows:   starrocksRows,
		Schema: schema,
	}, nil
}

// QuerySchema implements drivers.OLAPStore.
func (c *connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	db := c.db

	// Use LIMIT 0 to get schema without data
	schemaQuery := fmt.Sprintf("SELECT * FROM (%s) AS _schema_query LIMIT 0", query)

	rows, err := db.QueryxContext(ctx, schemaQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return c.rowsToSchema(rows)
}

// InformationSchema implements drivers.OLAPStore.
func (c *connection) InformationSchema() drivers.OLAPInformationSchema {
	return &informationSchema{c: c}
}

// rowsToSchema converts SQL rows to StructType schema.
func (c *connection) rowsToSchema(rows *sqlx.Rows) (*runtimev1.StructType, error) {
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(colTypes))
	for i, ct := range colTypes {
		runtimeType, err := c.databaseTypeToRuntimeType(ct.DatabaseTypeName())
		if err != nil {
			if errors.Is(err, errUnsupportedType) {
				// Skip unsupported types or handle gracefully
				return nil, fmt.Errorf("unsupported type %q for column %q: %w", ct.DatabaseTypeName(), ct.Name(), err)
			}
			return nil, err
		}
		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: runtimeType,
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}

// databaseTypeToRuntimeType converts StarRocks/MySQL types to runtime types.
// Returns an error for unsupported types instead of falling back to string.
func (c *connection) databaseTypeToRuntimeType(dbType string) (*runtimev1.Type, error) {
	dbType = strings.ToUpper(dbType)

	// Handle parameterized types
	if idx := strings.Index(dbType, "("); idx != -1 {
		dbType = dbType[:idx]
	}

	switch dbType {
	case "BOOLEAN", "BOOL":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}, nil
	case "TINYINT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT8}, nil
	case "SMALLINT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}, nil
	case "INT", "INTEGER":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}, nil
	case "BIGINT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}, nil
	case "LARGEINT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT128}, nil
	case "FLOAT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT32}, nil
	case "DOUBLE":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}, nil
	case "DECIMAL":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
	case "CHAR", "VARCHAR", "STRING", "TEXT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
	case "DATE":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}, nil
	case "DATETIME", "TIMESTAMP":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}, nil
	case "JSON", "JSONB":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}, nil
	case "ARRAY":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY}, nil
	case "MAP":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_MAP}, nil
	case "STRUCT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT}, nil
	case "BINARY", "VARBINARY", "BLOB":
		// Note: StarRocks doesn't have BLOB type, but MySQL driver may report VARBINARY as BLOB
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BYTES}, nil
	default:
		return nil, errUnsupportedType
	}
}

// starrocksRows wraps sqlx.Rows to provide MapScan method.
// This is required because if the correct type is not provided to Scan
// mysql driver just returns byte arrays.
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
		// Safety guard: prepareScanDest always allocates, but check anyway
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
			// Handle ARRAY, MAP, STRUCT, BYTES and other complex types
			// These are scanned into *any in prepareScanDest
			if ptr, ok := valPtr.(*any); ok {
				dest[fieldName] = *ptr
			} else {
				// Fallback: store the pointer's underlying value directly
				dest[fieldName] = valPtr
			}
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
			dest = &sql.NullInt16{}
		case runtimev1.Type_CODE_INT16:
			dest = &sql.NullInt16{}
		case runtimev1.Type_CODE_INT32:
			dest = &sql.NullInt32{}
		case runtimev1.Type_CODE_INT64, runtimev1.Type_CODE_INT128:
			dest = &sql.NullInt64{}
		case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
			dest = &sql.NullFloat64{}
		case runtimev1.Type_CODE_STRING:
			dest = &sql.NullString{}
		case runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIME:
			dest = &sql.NullString{}
		case runtimev1.Type_CODE_TIMESTAMP:
			// MySQL driver returns DATETIME as []byte unless parseTime=true in DSN
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
