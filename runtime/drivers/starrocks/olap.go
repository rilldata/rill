package starrocks

import (
	"context"
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

	return &drivers.Result{
		Rows:   rows,
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
	case "BOOLEAN", "BOOL", "TINYINT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}, nil
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
	case "DECIMAL", "DECIMALV2", "DECIMAL32", "DECIMAL64", "DECIMAL128":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DECIMAL}, nil
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
