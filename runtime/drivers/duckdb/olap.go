package duckdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	duckdbreplicator "github.com/rilldata/duckdb-replicator"
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
	meter                 = otel.Meter("github.com/rilldata/rill/runtime/drivers/duckdb")
	queriesCounter        = observability.Must(meter.Int64Counter("queries"))
	queueLatencyHistogram = observability.Must(meter.Int64Histogram("queue_latency", metric.WithUnit("ms")))
	queryLatencyHistogram = observability.Must(meter.Int64Histogram("query_latency", metric.WithUnit("ms")))
	totalLatencyHistogram = observability.Must(meter.Int64Histogram("total_latency", metric.WithUnit("ms")))
	connectionsInUse      = observability.Must(meter.Int64ObservableGauge("connections_in_use"))
)

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectDuckDB
}

func (c *connection) WithConnection(ctx context.Context, priority int, longRunning bool, fn drivers.WithConnectionFunc) error {
	// Check not nested
	if connFromContext(ctx) != nil {
		panic("nested WithConnection")
	}

	// Acquire connection
	conn, release, err := c.acquireOLAPConn(ctx, priority, longRunning)
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
	res, err := c.Execute(ctx, stmt)
	if err != nil {
		return err
	}
	if stmt.DryRun {
		return nil
	}
	err = res.Close()
	return c.checkErr(err)
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (res *drivers.Result, outErr error) {
	// Log query if enabled (usually disabled)
	if c.config.LogQueries {
		c.logger.Info("duckdb query", zap.String("sql", stmt.Query), zap.Any("args", stmt.Args))
	}

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
			return nil, c.checkErr(err)
		}

		_, err = conn.ExecContext(context.Background(), fmt.Sprintf("DROP VIEW %q", name))
		return nil, c.checkErr(err)
	}

	// Gather metrics only for actual queries
	var acquiredTime time.Time
	acquired := false
	start := time.Now()
	defer func() {
		totalLatency := time.Since(start).Milliseconds()
		queueLatency := acquiredTime.Sub(start).Milliseconds()

		attrs := []attribute.KeyValue{
			attribute.String("db", c.config.DBFilePath),
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
			c.activity.RecordMetric(ctx, "duckdb_queue_latency_ms", float64(queueLatency), attrs...)
			c.activity.RecordMetric(ctx, "duckdb_total_latency_ms", float64(totalLatency), attrs...)
			if acquired {
				c.activity.RecordMetric(ctx, "duckdb_query_latency_ms", float64(totalLatency-queueLatency), attrs...)
			}
		}
	}()

	// Acquire connection
	conn, release, err := c.acquireOLAPConn(ctx, stmt.Priority, stmt.LongRunning)
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

		// err must be checked before release
		err = c.checkErr(err)
		_ = release()
		return nil, err
	}

	schema, err := RowsToSchema(rows)
	if err != nil {
		if cancelFunc != nil {
			cancelFunc()
		}

		// err must be checked before release
		err = c.checkErr(err)
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

// TODO :: implement estimate size
func (c *connection) estimateSize() int64 {
	return 0
}

// AddTableColumn implements drivers.OLAPStore.
func (c *connection) AddTableColumn(ctx context.Context, tableName, columnName, typ string) error {
	return c.db.AddTableColumn(ctx, tableName, columnName, typ)
}

// AlterTableColumn implements drivers.OLAPStore.
func (c *connection) AlterTableColumn(ctx context.Context, tableName, columnName, newType string) error {
	return c.db.AlterTableColumn(ctx, tableName, columnName, newType)
}

// CreateTableAsSelect implements drivers.OLAPStore.
// We add a \n at the end of the any user query to ensure any comment at the end of model doesn't make the query incomplete.
func (c *connection) CreateTableAsSelect(ctx context.Context, name string, view bool, sql string, tableOpts map[string]any) error {
	return c.db.CreateTableAsSelect(ctx, name, sql, &duckdbreplicator.CreateTableOptions{View: view})
}

// InsertTableAsSelect implements drivers.OLAPStore.
func (c *connection) InsertTableAsSelect(ctx context.Context, name, sql string, byName, inPlace bool, strategy drivers.IncrementalStrategy, uniqueKey []string) error {
	var st duckdbreplicator.IncrementalStrategy
	switch strategy {
	case drivers.IncrementalStrategyAppend:
		st = duckdbreplicator.IncrementalStrategyAppend
	case drivers.IncrementalStrategyMerge:
		st = duckdbreplicator.IncrementalStrategyMerge
	default:
		return fmt.Errorf("incremental insert strategy %q not supported", strategy)
	}
	return c.db.InsertTableAsSelect(ctx, name, sql, &duckdbreplicator.InsertTableOptions{
		ByName:    byName,
		Strategy:  st,
		UniqueKey: uniqueKey,
	})
}

// DropTable implements drivers.OLAPStore.
func (c *connection) DropTable(ctx context.Context, name string, view bool) error {
	return c.db.DropTable(ctx, name)
}

// RenameTable implements drivers.OLAPStore.
func (c *connection) RenameTable(ctx context.Context, oldName, newName string, view bool) error {
	return c.db.RenameTable(ctx, oldName, newName)
}

func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return false
}

func RowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
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

// safeSQLName returns a quoted SQL identifier.
func safeSQLName(name string) string {
	return safeName(name)
}

func safeSQLString(name string) string {
	return drivers.DialectDuckDB.EscapeStringValue(name)
}
