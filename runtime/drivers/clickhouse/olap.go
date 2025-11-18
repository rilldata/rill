package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// Create instruments
var (
	meter                 = otel.Meter("github.com/rilldata/rill/runtime/drivers/clickhouse")
	queriesCounter        = observability.Must(meter.Int64Counter("queries"))
	queueLatencyHistogram = observability.Must(meter.Int64Histogram("queue_latency", metric.WithUnit("ms")))
	queryLatencyHistogram = observability.Must(meter.Int64Histogram("query_latency", metric.WithUnit("ms")))
	totalLatencyHistogram = observability.Must(meter.Int64Histogram("total_latency", metric.WithUnit("ms")))
)

var errUnsupportedType = errors.New("encountered unsupported clickhouse type")

var _ drivers.OLAPStore = &Connection{}

func (c *Connection) Dialect() drivers.Dialect {
	return drivers.DialectClickHouse
}

func (c *Connection) MayBeScaledToZero(ctx context.Context) bool {
	return c.config.CanScaleToZero
}

func (c *Connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	// Check not nested
	if connFromContext(ctx) != nil {
		panic("nested WithConnection")
	}

	// Acquire a connection from write pool, since this is meant to be used for operations that may write (e.g. creating temp tables).
	// Beware that this means that if later calls to acquireOLAPConn even with the write flag not set with the same context then they will get the same connection from the write pool.
	// But I think this is the expected behavior as we want to have a single connection for the whole WithConnection block.
	conn, release, err := c.acquireOLAPConn(ctx, priority, true)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	// Call fn with connection embedded in context
	wrappedCtx := c.sessionAwareContext(contextWithConn(ctx, conn))
	ensuredCtx := c.sessionAwareContext(contextWithConn(context.Background(), conn))
	return fn(wrappedCtx, ensuredCtx)
}

func (c *Connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	// Log query if enabled (usually disabled)
	if c.config.LogQueries {
		c.logger.Info("clickhouse query", zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)), zap.Any("args", stmt.Args), observability.ZapCtx(ctx))
	}

	// We can not directly append settings to the query as in Execute method because some queries like CREATE TABLE will not support it.
	// Instead, we set the settings in the context.
	// TODO: Fix query_settings_override not honoured here.
	ctx = contextWithQueryID(ctx)
	if c.supportSettings {
		settings := map[string]any{
			"cast_keep_nullable":        1,
			"insert_distributed_sync":   1,
			"prefer_global_in_and_join": 1,
			"session_timezone":          "UTC",
			"join_use_nulls":            1,
		}
		ctx = clickhouse.Context(ctx, clickhouse.WithSettings(settings))
	}

	// We use the meta conn for dry run queries
	if stmt.DryRun {
		conn, release, err := c.acquireMetaConn(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = release() }()

		_, err = conn.ExecContext(ctx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
		return err
	}

	// Use write connection for Exec operations
	conn, release, err := c.acquireOLAPConn(ctx, stmt.Priority, true)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	var cancelFunc context.CancelFunc
	if stmt.ExecutionTimeout != 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, stmt.ExecutionTimeout)
		defer cancelFunc()
	}

	_, err = conn.ExecContext(ctx, stmt.Query, stmt.Args...)
	return err
}

func (c *Connection) Query(ctx context.Context, stmt *drivers.Statement) (res *drivers.Result, outErr error) {
	ctx = contextWithQueryID(ctx)

	// Log query if enabled (usually disabled)
	if c.config.LogQueries {
		c.logger.Info("clickhouse query", zap.String("sql", c.Dialect().SanitizeQueryForLogging(stmt.Query)), zap.Any("args", stmt.Args))
	}

	// We use the meta conn for dry run queries
	if stmt.DryRun {
		conn, release, err := c.acquireMetaConn(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { _ = release() }()

		_, err = conn.ExecContext(ctx, fmt.Sprintf("EXPLAIN %s", stmt.Query), stmt.Args...)
		return nil, err
	}

	if c.supportSettings {
		if c.config.QuerySettingsOverride != "" {
			stmt.Query += "\n SETTINGS " + c.config.QuerySettingsOverride
		} else {
			stmt.Query += "\n SETTINGS cast_keep_nullable = 1, join_use_nulls = 1, session_timezone = 'UTC', prefer_global_in_and_join = 1, insert_distributed_sync = 1"
			if c.config.QuerySettings != "" {
				stmt.Query += ", " + c.config.QuerySettings
			}
		}
	}

	// Gather metrics only for actual queries
	var acquiredTime time.Time
	acquired := false
	start := time.Now()
	defer func() {
		totalLatency := time.Since(start).Milliseconds()
		queueLatency := acquiredTime.Sub(start).Milliseconds()

		attrs := []attribute.KeyValue{
			attribute.Bool("cancelled", errors.Is(outErr, context.Canceled)),
			attribute.Bool("failed", outErr != nil),
			attribute.String("instance_id", c.instanceID),
		}

		attrSet := attribute.NewSet(attrs...)

		queriesCounter.Add(ctx, 1, metric.WithAttributeSet(attrSet))
		queueLatencyHistogram.Record(ctx, queueLatency, metric.WithAttributeSet(attrSet))
		totalLatencyHistogram.Record(ctx, totalLatency, metric.WithAttributeSet(attrSet))
		if acquired {
			// Only track query latency when not cancelled in queue
			queryLatencyHistogram.Record(ctx, totalLatency-queueLatency, metric.WithAttributeSet(attrSet))
		}

		if c.activity != nil {
			c.activity.RecordMetric(ctx, "clickhouse_queue_latency_ms", float64(queueLatency), attrs...)
			c.activity.RecordMetric(ctx, "clickhouse_total_latency_ms", float64(totalLatency), attrs...)
			if acquired {
				c.activity.RecordMetric(ctx, "clickhouse_query_latency_ms", float64(totalLatency-queueLatency), attrs...)
			}
		}
	}()

	// Acquire connection
	conn, release, err := c.acquireOLAPConn(ctx, stmt.Priority, false)
	acquiredTime = time.Now()
	if err != nil {
		return nil, err
	}
	acquired = true

	// NOTE: We can't just "defer release()" because release() will block until rows.Close() is called.
	// We must be careful to make sure release() is called on all code paths.

	var cancelFunc context.CancelFunc
	if stmt.ExecutionTimeout != 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, stmt.ExecutionTimeout)
	}

	rows, err := conn.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		if cancelFunc != nil {
			cancelFunc()
		}
		_ = release()
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		if cancelFunc != nil {
			cancelFunc()
		}
		_ = rows.Close()
		_ = release()
		return nil, err
	}

	res = &drivers.Result{Rows: rows, Schema: schema}
	res.SetCleanupFunc(func() error {
		if cancelFunc != nil {
			cancelFunc()
		}
		return release()
	})

	return res, nil
}

func (c *Connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	// ClickHouse does not return schema with LIMIT 0, so we need to wrap query inside DESCRIBE to explicitly get the schema
	query = fmt.Sprintf("DESCRIBE (%s)", query)

	if c.config.LogQueries {
		c.logger.Info("clickhouse query", zap.String("sql", c.Dialect().SanitizeQueryForLogging(query)), zap.Any("args", args))
	}

	conn, release, err := c.acquireOLAPConn(ctx, 0, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	ctx, cancelFunc := context.WithTimeout(ctx, drivers.DefaultQuerySchemaTimeout)
	defer cancelFunc()

	rows, err := conn.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schema := &runtimev1.StructType{}
	m := make(map[string]any)
	for rows.Next() {
		if err = rows.MapScan(m); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		// Convert ClickHouse data type to runtimev1.StructType_Field_Type
		cType, ok := m["type"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse clickHouse type from schema")
		}
		t, err := databaseTypeToPB(cType, false)
		if err != nil {
			return nil, fmt.Errorf("failed to convert clickHouse type %q: %w", cType, err)
		}
		name, ok := m["name"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse column name from schema")
		}
		schema.Fields = append(schema.Fields, &runtimev1.StructType_Field{
			Name: name,
			Type: t,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error scanning schema: %w", err)
	}
	return schema, nil
}

func (c *Connection) InformationSchema() drivers.OLAPInformationSchema {
	return c
}

// acquireMetaConn gets a connection from the pool for "meta" queries like information schema (i.e. fast queries).
// It returns a function that puts the connection back in the pool (if applicable).
func (c *Connection) acquireMetaConn(ctx context.Context) (*SQLConn, func() error, error) {
	// Try to get conn from context (means the call is wrapped in WithConnection)
	conn := connFromContext(ctx)
	if conn != nil {
		return conn, func() error { return nil }, nil
	}

	// Acquire semaphore
	err := c.metaSem.Acquire(ctx, 1)
	if err != nil {
		return nil, nil, err
	}

	// Get new conn
	conn, releaseConn, err := c.acquireConn(ctx)
	if err != nil {
		c.metaSem.Release(1)
		return nil, nil, err
	}

	// Build release func
	release := func() error {
		err := releaseConn()
		c.metaSem.Release(1)
		return err
	}

	return conn, release, nil
}

// acquireOLAPConn gets a connection from the pool for OLAP queries (i.e. slow queries).
// It returns a function that puts the connection back in the pool (if applicable). write bool indicates if the connection is for an exec query.
func (c *Connection) acquireOLAPConn(ctx context.Context, priority int, write bool) (*SQLConn, func() error, error) {
	// Try to get conn from context (means the call is wrapped in WithConnection)
	conn := connFromContext(ctx)
	if conn != nil {
		return conn, func() error { return nil }, nil
	}

	// Acquire semaphore
	err := c.olapSem.Acquire(ctx, priority)
	if err != nil {
		return nil, nil, err
	}

	// Get new conn
	var releaseConn func() error
	if write {
		conn, releaseConn, err = c.acquireWriteConn(ctx)
	} else {
		conn, releaseConn, err = c.acquireConn(ctx)
	}
	if err != nil {
		c.olapSem.Release()
		return nil, nil, err
	}

	// Build release func
	release := func() error {
		err := releaseConn()
		c.olapSem.Release()
		return err
	}

	return conn, release, nil
}

// acquireConn returns a ClickHouse connection. It should only be used internally in acquireMetaConn and acquireOLAPConn.
func (c *Connection) acquireConn(ctx context.Context) (*SQLConn, func() error, error) {
	conn, err := c.readDB.Connx(ctx)
	if err != nil {
		return nil, nil, err
	}

	c.used()
	release := func() error {
		c.used()
		return conn.Close()
	}
	return &SQLConn{Conn: conn, supportSettings: c.supportSettings}, release, nil
}

// acquireWriteConn returns a ClickHouse write connection for write operations. It should only be used internally in acquireOLAPConn.
func (c *Connection) acquireWriteConn(ctx context.Context) (*SQLConn, func() error, error) {
	conn, err := c.writeDB.Connx(ctx)
	if err != nil {
		return nil, nil, err
	}

	c.used()
	release := func() error {
		c.used()
		return conn.Close()
	}
	return &SQLConn{Conn: conn, supportSettings: c.supportSettings}, release, nil
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

		ct.ScanType()

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

// When supportSettings is false, the cluster is in readonly mode and does not allow
// modifying any settings. The clickhouse-go driver automatically sets 'max_execution_time'
// if a context has a deadline, which would cause errors. To avoid this, we override the
// connection methods to remove the deadline from the context before executing any query.
type SQLConn struct {
	*sqlx.Conn
	supportSettings bool
}

func (sc *SQLConn) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	if sc.supportSettings {
		return sc.Conn.QueryxContext(ctx, query, args...)
	}
	ctx2 := contextWithoutDeadline(ctx)
	return sc.Conn.QueryxContext(ctx2, query, args...)
}

func (sc *SQLConn) QueryRowContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	if sc.supportSettings {
		return sc.Conn.QueryRowxContext(ctx, query, args...)
	}
	ctx2 := contextWithoutDeadline(ctx)
	return sc.Conn.QueryRowxContext(ctx2, query, args...)
}

func contextWithoutDeadline(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(context.WithoutCancel(parent))
	go func() {
		<-parent.Done()
		cancel()
	}()
	return ctx
}

// databaseTypeToPB converts Clickhouse types to Rill's generic schema type.
// Refer the list of types here: https://clickhouse.com/docs/en/sql-reference/data-types
func databaseTypeToPB(dbt string, nullable bool) (*runtimev1.Type, error) {
	dbt = strings.ToUpper(dbt)

	// For nullable the datatype is Nullable(X)
	if strings.HasPrefix(dbt, "NULLABLE(") {
		dbt = dbt[9 : len(dbt)-1]
		return databaseTypeToPB(dbt, true)
	}

	// For LowCardinality the datatype is LowCardinality(X)
	if strings.HasPrefix(dbt, "LOWCARDINALITY(") {
		dbt = dbt[15 : len(dbt)-1]
		return databaseTypeToPB(dbt, nullable)
	}

	match := true
	t := &runtimev1.Type{Nullable: nullable}
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
	case "INT256":
		t.Code = runtimev1.Type_CODE_INT256
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
	case "UINT256":
		t.Code = runtimev1.Type_CODE_UINT256
	case "FLOAT32":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "FLOAT64":
		t.Code = runtimev1.Type_CODE_FLOAT64
	// can be DECIMAL or DECIMAL(...) which is covered below
	case "DECIMAL":
		t.Code = runtimev1.Type_CODE_FLOAT64
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
	case "INTERVALNANOSECOND", "INTERVALMICROSECOND", "INTERVALMILLISECOND", "INTERVALSECOND", "INTERVALMINUTE", "INTERVALHOUR", "INTERVALDAY", "INTERVALWEEK", "INTERVALMONTH", "INTERVALQUARTER", "INTERVALYEAR":
		t.Code = runtimev1.Type_CODE_INTERVAL
	case "JSON":
		t.Code = runtimev1.Type_CODE_JSON
	case "UUID":
		t.Code = runtimev1.Type_CODE_UUID
	case "IPV4":
		t.Code = runtimev1.Type_CODE_STRING
	case "IPV6":
		t.Code = runtimev1.Type_CODE_STRING
	case "OTHER":
		t.Code = runtimev1.Type_CODE_JSON
	case "NOTHING":
		t.Code = runtimev1.Type_CODE_STRING
	case "POINT":
		return databaseTypeToPB("Array(Float64)", nullable)
	case "RING":
		return databaseTypeToPB("Array(Point)", nullable)
	case "LINESTRING":
		return databaseTypeToPB("Array(Point)", nullable)
	case "MULTILINESTRING":
		return databaseTypeToPB("Array(LineString)", nullable)
	case "POLYGON":
		return databaseTypeToPB("Array(Ring)", nullable)
	case "MULTIPOLYGON":
		return databaseTypeToPB("Array(Polygon)", nullable)
	default:
		match = false
	}
	if match {
		return t, nil
	}

	// All other complex types have details in parentheses after the type name.
	base, args, ok := splitBaseAndArgs(dbt)
	if !ok {
		return nil, errUnsupportedType
	}

	switch base {
	case "DATETIME":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "DATETIME64":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	// Example: "DECIMAL(10,20)", "DECIMAL(10)"
	case "DECIMAL":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "DECIMAL32":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "DECIMAL64":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "DECIMAL128":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "DECIMAL256":
		t.Code = runtimev1.Type_CODE_DECIMAL
	case "FIXEDSTRING":
		t.Code = runtimev1.Type_CODE_STRING
	case "ARRAY":
		t.Code = runtimev1.Type_CODE_ARRAY
		var err error
		t.ArrayElementType, err = databaseTypeToPB(dbt[6:len(dbt)-1], true)
		if err != nil {
			return nil, err
		}
	// Example: "MAP(VARCHAR, INT)"
	case "MAP":
		fieldStrs := strings.Split(args, ",")
		if len(fieldStrs) != 2 {
			return nil, errUnsupportedType
		}

		keyType, err := databaseTypeToPB(strings.TrimSpace(fieldStrs[0]), true)
		if err != nil {
			return nil, err
		}

		valType, err := databaseTypeToPB(strings.TrimSpace(fieldStrs[1]), true)
		if err != nil {
			return nil, err
		}

		t.Code = runtimev1.Type_CODE_MAP
		t.MapType = &runtimev1.MapType{
			KeyType:   keyType,
			ValueType: valType,
		}
	case "ENUM", "ENUM8", "ENUM16":
		// Representing enums as strings
		t.Code = runtimev1.Type_CODE_STRING
	case "TUPLE":
		t.Code = runtimev1.Type_CODE_STRUCT
		t.StructType = &runtimev1.StructType{}
		fields := splitCommasUnlessQuotedOrNestedInParens(args)
		if len(fields) == 0 {
			return nil, errUnsupportedType
		}
		_, _, isNamed := splitStructFieldStr(fields[0])
		for i, fieldStr := range fields {
			if isNamed {
				name, typ, ok := splitStructFieldStr(fieldStr)
				if !ok {
					return nil, errUnsupportedType
				}
				fieldType, err := databaseTypeToPB(typ, false)
				if err != nil {
					return nil, err
				}
				t.StructType.Fields = append(t.StructType.Fields, &runtimev1.StructType_Field{
					Name: name,
					Type: fieldType,
				})
			} else {
				fieldType, err := databaseTypeToPB(fieldStr, true)
				if err != nil {
					return nil, err
				}
				t.StructType.Fields = append(t.StructType.Fields, &runtimev1.StructType_Field{
					Name: fmt.Sprintf("%d", i),
					Type: fieldType,
				})
			}
		}
	default:
		return nil, errUnsupportedType
	}

	return t, nil
}

// Splits a type with args in parentheses, for example:
//
//	`Nullable(UInt64)` -> (`Nullable`, `UInt64`, true)
func splitBaseAndArgs(s string) (string, string, bool) {
	// Split on opening parenthesis
	base, rest, found := strings.Cut(s, "(")
	if !found {
		return "", "", false
	}

	// Remove closing parenthesis
	rest = rest[0 : len(rest)-1]

	return base, rest, true
}

// Splits a comma-separated list, but ignores commas inside strings or nested in parentheses.
// (NOTE: DuckDB escapes strings by replacing `"` with `""`. Example: hello "world" -> "hello ""world""".)
//
// Examples:
//
//	`10,20` -> [`10`, `20`]
//	`VARCHAR, INT` -> [`VARCHAR`, `INT`]
//	`"foo "",""" INT, "bar" STRUCT("a" INT, "b" INT)` -> [`"foo "",""" INT`, `"bar" STRUCT("a" INT, "b" INT)`]
func splitCommasUnlessQuotedOrNestedInParens(s string) []string {
	// Result slice
	splits := []string{}
	// Starting idx of current split
	fromIdx := 0
	// True if quote level is unmatched (this is sufficient for escaped quotes since they will immediately flip again)
	quoted := false
	// Nesting level
	nestCount := 0

	// Consume input character-by-character
	for idx, char := range s {
		// Toggle quoted
		if char == '"' {
			quoted = !quoted
			continue
		}
		// If quoted, don't parse for nesting or commas
		if quoted {
			continue
		}
		// Increase nesting on opening paren
		if char == '(' {
			nestCount++
			continue
		}
		// Decrease nesting on closing paren
		if char == ')' {
			nestCount--
			continue
		}
		// If nested, don't parse for commas
		if nestCount != 0 {
			continue
		}
		// If not nested and there's a comma, add split to result
		if char == ',' {
			splits = append(splits, s[fromIdx:idx])
			fromIdx = idx + 1
			continue
		}
		// If not nested, and there's a space at the start of the split, skip it
		if fromIdx == idx && char == ' ' {
			fromIdx++
			continue
		}
	}

	// Add last split to result and return
	splits = append(splits, s[fromIdx:])
	return splits
}

// splitStructFieldStr splits a single struct name/type pair.
// It expects fieldStr to have the format `name TYPE` or `"name" TYPE`.
// If the name string is quoted and contains escaped quotes `""`, they'll be replaced by `"`.
// For example: splitStructFieldStr(`"hello "" world" VARCHAR`) -> (`hello " world`, `VARCHAR`, true).
func splitStructFieldStr(fieldStr string) (string, string, bool) {
	// If the string DOES NOT start with a `"`, we can just split on the first space.
	if fieldStr == "" || fieldStr[0] != '"' {
		return strings.Cut(fieldStr, " ")
	}

	// Find end of quoted string (skipping `""` since they're escaped quotes)
	idx := 1
	found := false
	for !found && idx < len(fieldStr) {
		// Continue if not a quote
		if fieldStr[idx] != '"' {
			idx++
			continue
		}

		// Skip two ahead if it's two quotes in a row (i.e. an escaped quote)
		if len(fieldStr) > idx+1 && fieldStr[idx+1] == '"' {
			idx += 2
			continue
		}

		// It's the last quote of the string. We're done.
		idx++
		found = true
	}

	// If not found, format was unexpected
	if !found {
		return "", "", false
	}

	// Remove surrounding `"` and replace escaped quotes `""` with `"`
	nameStr := strings.ReplaceAll(fieldStr[1:idx-1], `""`, `"`)

	// The rest of the string is the type, minus the initial space
	typeStr := strings.TrimLeft(fieldStr[idx:], " ")

	return nameStr, typeStr, true
}
