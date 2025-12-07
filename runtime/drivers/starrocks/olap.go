package starrocks

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

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
func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	// Acquire semaphore for connection affinity
	if err := c.querySem.Acquire(ctx, 1); err != nil {
		return err
	}
	defer c.querySem.Release(1)

	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	conn, err := db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Close()

	// Create ensured context that won't be cancelled
	ensuredCtx := context.WithoutCancel(ctx)

	return fn(ctx, ensuredCtx)
}

// Exec implements drivers.OLAPStore.
func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	if c.configProp.LogQueries {
		c.logger.Info("StarRocks exec",
			zap.String("query", stmt.Query),
			zap.Any("args", stmt.Args))
	}

	// Remove deadline but preserve cancellation (ClickHouse pattern)
	execCtx := contextWithoutDeadline(ctx)

	_, err = db.ExecContext(execCtx, stmt.Query, stmt.Args...)
	return err
}

// Query implements drivers.OLAPStore.
func (c *connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	if c.configProp.LogQueries {
		c.logger.Info("StarRocks query",
			zap.String("query", stmt.Query),
			zap.Any("args", stmt.Args))
	}

	// Remove deadline but preserve cancellation (ClickHouse pattern)
	queryCtx := contextWithoutDeadline(ctx)

	rows, err := db.QueryxContext(queryCtx, stmt.Query, stmt.Args...)
	if err != nil {
		return nil, err
	}

	schema, err := c.rowsToSchema(rows)
	if err != nil {
		rows.Close()
		return nil, err
	}

	return &drivers.Result{
		Rows:   &sqlRows{rows},
		Schema: schema,
	}, nil
}

// QuerySchema implements drivers.OLAPStore.
func (c *connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	db, err := c.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// Use LIMIT 0 to get schema without data
	schemaQuery := fmt.Sprintf("SELECT * FROM (%s) AS _schema_query LIMIT 0", query)

	// Remove deadline but preserve cancellation (ClickHouse pattern)
	queryCtx := contextWithoutDeadline(ctx)

	rows, err := db.QueryxContext(queryCtx, schemaQuery, args...)
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
		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: c.databaseTypeToRuntimeType(ct.DatabaseTypeName()),
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}

// databaseTypeToRuntimeType converts StarRocks/MySQL types to runtime types.
func (c *connection) databaseTypeToRuntimeType(dbType string) *runtimev1.Type {
	dbType = strings.ToUpper(dbType)

	// Handle nullable types
	if strings.HasPrefix(dbType, "NULLABLE(") {
		dbType = strings.TrimPrefix(dbType, "NULLABLE(")
		dbType = strings.TrimSuffix(dbType, ")")
	}

	// Handle parameterized types
	if idx := strings.Index(dbType, "("); idx != -1 {
		dbType = dbType[:idx]
	}

	switch dbType {
	case "BOOLEAN", "BOOL", "TINYINT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}
	case "SMALLINT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}
	case "INT", "INTEGER":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}
	case "BIGINT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}
	case "LARGEINT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT128}
	case "FLOAT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT32}
	case "DOUBLE":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}
	case "DECIMAL", "DECIMALV2", "DECIMAL32", "DECIMAL64", "DECIMAL128":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DECIMAL}
	case "CHAR", "VARCHAR", "STRING", "TEXT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}
	case "DATE":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}
	case "DATETIME", "TIMESTAMP":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}
	case "JSON", "JSONB":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}
	case "ARRAY":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY}
	case "MAP":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_MAP}
	case "STRUCT":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT}
	case "BINARY", "VARBINARY":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BYTES}
	default:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}
	}
}

// sqlRows wraps sqlx.Rows to implement drivers.Rows interface.
type sqlRows struct {
	*sqlx.Rows
}

// MapScan implements drivers.Rows.
func (r *sqlRows) MapScan(dest map[string]any) error {
	return r.Rows.MapScan(dest)
}

// contextWithoutDeadline removes the deadline from the context but preserves cancellation.
// This prevents queries from being cancelled due to tight client timeouts while still
// respecting explicit cancellation signals.
func contextWithoutDeadline(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(context.WithoutCancel(parent))
	go func() {
		<-parent.Done()
		cancel()
	}()
	return ctx
}
