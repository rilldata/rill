package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

var useCache = false

// Create instruments
var (
	meter                 = otel.Meter("github.com/rilldata/rill/runtime/drivers/clickhouse")
	queriesCounter        = observability.Must(meter.Int64Counter("queries"))
	queueLatencyHistogram = observability.Must(meter.Int64Histogram("queue_latency", metric.WithUnit("ms")))
	queryLatencyHistogram = observability.Must(meter.Int64Histogram("query_latency", metric.WithUnit("ms")))
	totalLatencyHistogram = observability.Must(meter.Int64Histogram("total_latency", metric.WithUnit("ms")))
)

var _ drivers.OLAPStore = &connection{}

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectClickHouse
}

func (c *connection) WithConnection(ctx context.Context, priority int, longRunning, tx bool, fn drivers.WithConnectionFunc) error {
	// Check not nested
	if connFromContext(ctx) != nil {
		panic("nested WithConnection")
	}

	// Acquire connection
	conn, release, err := c.acquireOLAPConn(ctx, priority)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	// Call fn with connection embedded in context
	wrappedCtx := contextWithConn(ctx, conn)
	ensuredCtx := contextWithConn(context.Background(), conn)
	return fn(wrappedCtx, ensuredCtx, conn.Conn)
}

func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	conn, release, err := c.acquireOLAPConn(ctx, stmt.Priority)
	if err != nil {
		return err
	}

	// TODO: should we use timeout to acquire connection as well ?
	var cancelFunc context.CancelFunc
	if stmt.ExecutionTimeout != 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, stmt.ExecutionTimeout)
	}
	defer func() {
		if cancelFunc != nil {
			cancelFunc()
		}
		_ = release()
	}()
	_, err = conn.ExecContext(ctx, stmt.Query, stmt.Args...)
	return err
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (res *drivers.Result, outErr error) {
	// We use the meta conn for dry run queries
	if stmt.DryRun {
		conn, release, err := c.acquireMetaConn(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { _ = release() }()

		// TODO: Find way to validate with args

		name := uuid.NewString()
		_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE TEMPORARY VIEW %q AS %s", name, stmt.Query))
		if err != nil {
			return nil, err
		}

		_, err = conn.ExecContext(context.Background(), fmt.Sprintf("DROP VIEW %q", name))
		return nil, err
	}

	stmt.Query += "\n SETTINGS cast_keep_nullable = 1, join_use_nulls = 1"
	if useCache {
		stmt.Query += ", use_query_cache = 1"
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
			c.activity.Emit(ctx, "clickhouse_queue_latency_ms", float64(queueLatency), attrs...)
			c.activity.Emit(ctx, "clickhouse_total_latency_ms", float64(totalLatency), attrs...)
			if acquired {
				c.activity.Emit(ctx, "clickhouse_query_latency_ms", float64(totalLatency-queueLatency), attrs...)
			}
		}
	}()

	// Acquire connection
	conn, release, err := c.acquireOLAPConn(ctx, stmt.Priority)
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
			Query:    fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", safeSQLName(name), sql),
			Priority: 100,
		})
	}
	return c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s ENGINE = MergeTree ORDER BY tuple() AS %s", safeSQLName(name), sql),
		Priority: 100,
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
		Query:    fmt.Sprintf("DROP %s %s", typ, safeSQLName(name)),
		Priority: 100,
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
			Query:    fmt.Sprintf("RENAME TABLE %s TO %s", safeSQLName(name), safeSQLName(newName)),
			Priority: 100,
		})
	}

	// clickhouse does not support renaming views so we capture the OLD view DDL and use it to create new view
	res, err := c.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("SHOW CREATE VIEW %s", safeSQLName(name)),
		Priority: 100,
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

	// create new view
	sql = strings.Replace(sql, name, safeSQLName(newName), 1)
	err = c.Exec(ctx, &drivers.Statement{
		Query:    sql,
		Priority: 100,
	})
	if err != nil {
		return err
	}

	// drop old view
	err = c.Exec(context.Background(), &drivers.Statement{
		Query:    fmt.Sprintf("DROP VIEW %s", safeSQLName(name)),
		Priority: 100,
	})
	if err != nil {
		c.logger.Error("clickhouse: failed to drop old view", zap.String("name", name), zap.Error(err))
	}
	return nil
}

func (c *connection) DropDB() error {
	return fmt.Errorf("dropping database not supported")
}

// acquireMetaConn gets a connection from the pool for "meta" queries like information schema (i.e. fast queries).
// It returns a function that puts the connection back in the pool (if applicable).
func (c *connection) acquireMetaConn(ctx context.Context) (*sqlx.Conn, func() error, error) {
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
// It returns a function that puts the connection back in the pool (if applicable).
func (c *connection) acquireOLAPConn(ctx context.Context, priority int) (*sqlx.Conn, func() error, error) {
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
	conn, releaseConn, err := c.acquireConn(ctx)
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

// acquireConn returns a DuckDB connection. It should only be used internally in acquireMetaConn and acquireOLAPConn.
func (c *connection) acquireConn(ctx context.Context) (*sqlx.Conn, func() error, error) {
	conn, err := c.db.Connx(ctx)
	if err != nil {
		return nil, nil, err
	}

	release := func() error {
		return conn.Close()
	}
	return conn, release, nil
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

// databaseTypeToPB converts clikchouse types to rill internal types
// refer the list of types here : https://clickhouse.com/docs/en/sql-reference/data-types
// excludes mapping for Aggregation function types, Nested data structures, Tuples, Geo types, Special data types
func databaseTypeToPB(dbt string, nullable bool) (*runtimev1.Type, error) {
	dbt = strings.ToUpper(dbt)
	// for nullable the datatype is Nullable(X)
	if strings.HasPrefix(dbt, "NULLABLE(") {
		dbt = dbt[9 : len(dbt)-1]
		nullable = true
	}
	// for LowCardinality the datatype is LowCardinality(X)
	if strings.HasPrefix(dbt, "LOWCARDINALITY(") {
		dbt = dbt[15 : len(dbt)-1]
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
	default:
		match = false
	}
	if match {
		return t, nil
	}

	// All other complex types have details in parentheses after the type name.
	base, args, ok := splitBaseAndArgs(dbt)
	if !ok {
		return nil, fmt.Errorf("encountered unsupported clickhouse type '%s'", dbt)
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
			return nil, fmt.Errorf("encountered unsupported clickhouse type '%s'", dbt)
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
	case "ENUM":
		t.Code = runtimev1.Type_CODE_STRING // representing enums as strings for now
	default:
		return nil, fmt.Errorf("encountered unsupported clickhouse type '%s'", dbt)
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
