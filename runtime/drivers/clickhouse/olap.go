package clickhouse

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/graceful"
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

var _ drivers.OLAPStore = &connection{}

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectClickHouse
}

func (c *connection) WithConnection(ctx context.Context, priority int, longRunning bool, fn drivers.WithConnectionFunc) error {
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
	wrappedCtx := c.sessionAwareContext(contextWithConn(ctx, conn))
	ensuredCtx := c.sessionAwareContext(contextWithConn(context.Background(), conn))
	return fn(wrappedCtx, ensuredCtx, conn.Conn)
}

func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	// Log query if enabled (usually disabled)
	if c.config.LogQueries {
		c.logger.Info("clickhouse query", zap.String("sql", stmt.Query), zap.Any("args", stmt.Args))
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
	// Log query if enabled (usually disabled)
	if c.config.LogQueries {
		c.logger.Info("clickhouse query", zap.String("sql", stmt.Query), zap.Any("args", stmt.Args))
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

	if c.config.SettingsOverride != "" {
		stmt.Query += "\n SETTINGS " + c.config.SettingsOverride
	} else {
		stmt.Query += "\n SETTINGS cast_keep_nullable = 1, join_use_nulls = 1, session_timezone = 'UTC', prefer_global_in_and_join = 1, insert_distributed_sync = 1"
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
func (c *connection) CreateTableAsSelect(ctx context.Context, name, sql string, opts *drivers.CreateTableOptions) error {
	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.TableOpts, outputProps); err != nil {
		return fmt.Errorf("failed to parse output properties: %w", err)
	}
	var onClusterClause string
	if c.config.Cluster != "" {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}

	if outputProps.Typ == "VIEW" {
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("CREATE OR REPLACE VIEW %s %s AS %s", safeSQLName(name), onClusterClause, sql),
			Priority: 100,
		})
	} else if outputProps.Typ == "DICTIONARY" {
		return c.createDictionary(ctx, name, sql, outputProps)
	}
	// on replicated databases `create table t as select * from ...` is prohibited
	// so we need to create a table first and then insert data into it
	if err := c.createTable(ctx, name, sql, outputProps); err != nil {
		return err
	}
	// insert into table
	return c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(name), sql),
		Priority: 100,
	})
}

// InsertTableAsSelect implements drivers.OLAPStore.
func (c *connection) InsertTableAsSelect(ctx context.Context, name, sql string, opts *drivers.InsertTableOptions) error {
	if !opts.InPlace {
		return fmt.Errorf("clickhouse: inserts does not support inPlace=false")
	}
	if opts.Strategy == drivers.IncrementalStrategyAppend {
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("INSERT INTO %s %s", safeSQLName(name), sql),
			Priority:    1,
			LongRunning: true,
		})
	}

	if opts.Strategy == drivers.IncrementalStrategyPartitionOverwrite {
		_, onCluster, err := informationSchema{c: c}.entityType(ctx, c.config.Database, name)
		if err != nil {
			return err
		}
		onClusterClause := ""
		if onCluster {
			onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
		}
		// Get the engine info of the given table
		engine, err := c.getTableEngine(ctx, name)
		if err != nil {
			return err
		}
		// Distributed table cannot be altered directly, so we need to alter the local table
		if engine == "Distributed" {
			name = localTableName(name)
		}
		// create temp table with the same schema using a deterministic name
		tempName := fmt.Sprintf("__rill_temp_%s_%x", name, md5.Sum([]byte(sql)))
		// clean up the temp table
		defer func() {
			// cleanup using a different ctx to prevent cleanups being impacted by the main ctx cancellation
			// this is a best effort cleanup and query can still timeout and we don't want to wait forever due to blocked calls
			// this is triggered before the table is even created to handle situations
			// where before the client can trigger query cancel the query succeeds and the view is created but the driver stil reports query cancelled
			ctx, cancel := graceful.WithMinimumDuration(ctx, 15*time.Second)
			defer cancel()
			err = c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s %s", safeSQLName(tempName), onClusterClause),
				Priority: 1,
			})
			if err != nil {
				c.logger.Warn("clickhouse: failed to drop temp table", zap.String("name", tempName), zap.Error(err))
			}
		}()
		err = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s %s AS %s", safeSQLName(tempName), onClusterClause, name),
			Priority: 1,
		})
		if err != nil {
			return err
		}
		// insert into temp table
		err = c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("INSERT INTO %s %s", safeSQLName(tempName), sql),
			Priority:    1,
			LongRunning: true,
		})
		if err != nil {
			return err
		}
		// list partitions from the temp table
		partitions, err := c.getTablePartitions(ctx, tempName)
		if err != nil {
			return err
		}
		// iterate over partitions and replace them in the main table
		for _, part := range partitions {
			// alter the main table to replace the partition
			err = c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("ALTER TABLE %s %s REPLACE PARTITION ? FROM %s", safeSQLName(name), onClusterClause, safeSQLName(tempName)),
				Args:     []any{part},
				Priority: 1,
			})
			if err != nil {
				return err
			}
		}
		return nil
	}

	if opts.Strategy == drivers.IncrementalStrategyMerge {
		_, onCluster, err := informationSchema{c: c}.entityType(ctx, c.config.Database, name)
		if err != nil {
			return err
		}
		onClusterClause := ""
		if onCluster {
			onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
		}
		// get the engine info of the given table
		engine, err := c.getTableEngine(ctx, name)
		if err != nil {
			return err
		}
		if !strings.Contains(engine, "ReplacingMergeTree") {
			return fmt.Errorf("clickhouse: merge strategy requires ReplacingMergeTree engine")
		}

		// insert into table using the merge strategy
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("INSERT INTO %s %s %s", safeSQLName(name), onClusterClause, sql),
			Priority:    1,
			LongRunning: true,
		})
	}
	return fmt.Errorf("incremental insert strategy %q not supported", opts.Strategy)
}

// DropTable implements drivers.OLAPStore.
func (c *connection) DropTable(ctx context.Context, name string) error {
	typ, onCluster, err := informationSchema{c: c}.entityType(ctx, c.config.Database, name)
	if err != nil {
		return err
	}
	var onClusterClause string
	if onCluster {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}
	switch typ {
	case "VIEW":
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP VIEW %s %s", safeSQLName(name), onClusterClause),
			Priority: 100,
		})
	case "DICTIONARY":
		// first drop the dictionary
		err := c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP DICTIONARY %s %s", safeSQLName(name), onClusterClause),
			Priority: 100,
		})
		// then drop the temp table
		_ = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE %s %s", safeSQLName(tempTableForDictionary(name)), onClusterClause),
			Priority: 100,
		})
		return err
	case "TABLE":
		// drop the main table
		err := c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE %s %s", safeSQLName(name), onClusterClause),
			Priority: 100,
		})
		if err != nil {
			return err
		}
		// then drop the local table in case of cluster
		if onCluster && !strings.HasSuffix(name, "_local") {
			return c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("DROP TABLE %s %s", safelocalTableName(name), onClusterClause),
				Priority: 100,
			})
		}
		return nil
	default:
		return fmt.Errorf("clickhouse: unknown entity type %q", typ)
	}
}

func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return c.config.CanScaleToZero
}

// RenameTable implements drivers.OLAPStore.
func (c *connection) RenameTable(ctx context.Context, oldName, newName string) error {
	typ, onCluster, err := informationSchema{c: c}.entityType(ctx, c.config.Database, oldName)
	if err != nil {
		return err
	}
	var onClusterClause string
	if onCluster {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}

	switch typ {
	case "VIEW":
		return c.renameView(ctx, oldName, newName, onClusterClause)
	case "DICTIONARY":
		return c.renameTable(ctx, oldName, newName, onClusterClause)
	case "TABLE":
		if !onCluster {
			return c.renameTable(ctx, oldName, newName, onClusterClause)
		}
		// capture the full engine of the old distributed table
		args := []any{c.config.Database, oldName}
		if c.config.Database == "" {
			args = []any{nil, oldName}
		}
		var engineFull string
		res, err := c.Execute(ctx, &drivers.Statement{
			Query:    "SELECT engine_full FROM system.tables WHERE database = coalesce(?, currentDatabase()) AND name = ?",
			Args:     args,
			Priority: 100,
		})
		if err != nil {
			return err
		}

		for res.Next() {
			if err := res.Scan(&engineFull); err != nil {
				res.Close()
				return err
			}
		}
		res.Close()
		engineFull = strings.ReplaceAll(engineFull, localTableName(oldName), safelocalTableName(newName))

		// build the column type clause
		columnClause, err := c.columnClause(ctx, oldName)
		if err != nil {
			return err
		}

		// rename the local table
		err = c.renameTable(ctx, localTableName(oldName), localTableName(newName), onClusterClause)
		if err != nil {
			return err
		}

		// recreate the distributed table
		err = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s %s %s Engine = %s", safeSQLName(newName), onClusterClause, columnClause, engineFull),
			Priority: 100,
		})
		if err != nil {
			return err
		}

		// drop the old table
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE %s %s", safeSQLName(oldName), onClusterClause),
			Priority: 100,
		})
	default:
		return fmt.Errorf("clickhouse: unknown entity type %q", typ)
	}
}

func (c *connection) renameView(ctx context.Context, oldName, newName, onCluster string) error {
	// clickhouse does not support renaming views so we capture the OLD view's select statement and use it to create new view
	args := []any{c.config.Database, oldName}
	if c.config.Database == "" {
		args = []any{nil, oldName}
	}
	res, err := c.Execute(ctx, &drivers.Statement{
		Query:    "SELECT as_select FROM system.tables WHERE database = coalesce(?, currentDatabase()) AND name = ?",
		Args:     args,
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
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE VIEW %s %s AS %s", safeSQLName(newName), onCluster, sql),
		Priority: 100,
	})
	if err != nil {
		return err
	}

	// drop old view
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP VIEW %s %s", safeSQLName(oldName), onCluster),
		Priority: 100,
	})
	if err != nil {
		c.logger.Error("clickhouse: failed to drop old view", zap.String("name", oldName), zap.Error(err))
	}
	return nil
}

func (c *connection) renameTable(ctx context.Context, oldName, newName, onCluster string) error {
	var exists bool
	err := c.db.QueryRowContext(ctx, fmt.Sprintf("EXISTS %s", safeSQLName(newName))).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("RENAME TABLE %s TO %s %s", safeSQLName(oldName), safeSQLName(newName), onCluster),
			Priority: 100,
		})
	}
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("EXCHANGE TABLES %s AND %s %s", safeSQLName(oldName), safeSQLName(newName), onCluster),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	// drop the old table
	return c.DropTable(context.Background(), oldName)
}

func (c *connection) createTable(ctx context.Context, name, sql string, outputProps *ModelOutputProperties) error {
	var onClusterClause string
	if c.config.Cluster != "" {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}
	var create strings.Builder
	create.WriteString("CREATE OR REPLACE TABLE ")
	if c.config.Cluster != "" {
		// need to create a local table on the cluster first
		fmt.Fprintf(&create, "%s %s", safelocalTableName(name), onClusterClause)
	} else {
		create.WriteString(safeSQLName(name))
	}

	if outputProps.Columns == "" {
		if sql == "" {
			return fmt.Errorf("clickhouse: no columns specified for table %q", name)
		}
		// infer columns
		v := fmt.Sprintf("__rill_temp_%s_%x", name, md5.Sum([]byte(sql)))
		defer func() {
			// cleanup using a different ctx to prevent cleanups being impacted by the main ctx cancellation
			// this is a best effort cleanup and query can still timeout and we don't want to wait forever due to blocked calls
			// this is triggered before the view is even created to handle situations
			// where before the client can trigger query cancel the query succeeds and the view is created but the driver stil reports query cancelled
			ctx, cancel := graceful.WithMinimumDuration(ctx, 15*time.Second)
			defer cancel()
			_ = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP VIEW IF EXISTS %s %s", v, onClusterClause)})
		}()
		err := c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s %s AS %s", v, onClusterClause, sql)})
		if err != nil {
			return err
		}
		// create table with same schema as view
		fmt.Fprintf(&create, " AS %s ", v)
	} else {
		fmt.Fprintf(&create, " %s ", outputProps.Columns)
	}

	tableConfig := outputProps.tblConfig()
	create.WriteString(tableConfig)

	// validate incremental strategy
	if outputProps.IncrementalStrategy == drivers.IncrementalStrategyPartitionOverwrite &&
		!strings.Contains(strings.ToUpper(tableConfig), "PARTITION BY") {
		return fmt.Errorf("clickhouse: incremental strategy partition_overwrite requires a partition key")
	}

	// create table
	err := c.Exec(ctx, &drivers.Statement{Query: create.String(), Priority: 100})
	if err != nil {
		return err
	}

	if c.config.Cluster == "" {
		return nil
	}
	// create the distributed table
	var distributed strings.Builder
	database := "currentDatabase()"
	if c.config.Database != "" {
		database = safeSQLString(c.config.Database)
	}
	fmt.Fprintf(&distributed, "CREATE OR REPLACE TABLE %s %s AS %s", safeSQLName(name), onClusterClause, safelocalTableName(name))
	fmt.Fprintf(&distributed, " ENGINE = Distributed(%s, %s, %s", safeSQLString(c.config.Cluster), database, safeSQLString(localTableName(name)))
	if outputProps.DistributedShardingKey != "" {
		fmt.Fprintf(&distributed, ", %s", outputProps.DistributedShardingKey)
	} else {
		fmt.Fprintf(&distributed, ", rand()")
	}
	distributed.WriteString(")")
	if outputProps.DistributedSettings != "" {
		fmt.Fprintf(&distributed, " SETTINGS %s", outputProps.DistributedSettings)
	}
	return c.Exec(ctx, &drivers.Statement{Query: distributed.String(), Priority: 100})
}

func (c *connection) createDictionary(ctx context.Context, name, sql string, outputProps *ModelOutputProperties) error {
	var onClusterClause string
	if c.config.Cluster != "" {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}
	if sql == "" {
		if outputProps.Columns == "" {
			return fmt.Errorf("clickhouse: no columns specified for dictionary %q", name)
		}
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("CREATE OR REPLACE DICTIONARY %s %s %s %s", safeSQLName(name), onClusterClause, outputProps.Columns, outputProps.EngineFull),
			Priority: 100,
		})
	}

	// create a temp table first
	// NOTE :: this can only be dropped when the dictionary is dropped
	tempTable := tempTableForDictionary(name)
	err := c.createTable(ctx, tempTable, sql, outputProps)
	if err != nil {
		return err
	}
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(tempTable), sql),
		Priority: 100,
	})
	if err != nil {
		return err
	}

	if outputProps.Columns == "" {
		// infer columns
		outputProps.Columns, err = c.columnClause(ctx, tempTable)
		if err != nil {
			return err
		}
	}

	if outputProps.PrimaryKey == "" {
		return fmt.Errorf("clickhouse: no primary key specified for dictionary %q", name)
	}

	// create dictionary
	return c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf(`CREATE OR REPLACE DICTIONARY %s %s %s PRIMARY KEY %s SOURCE(CLICKHOUSE(TABLE %s)) LAYOUT(HASHED()) LIFETIME(0)`, safeSQLName(name), onClusterClause, outputProps.Columns, outputProps.PrimaryKey, c.Dialect().EscapeStringValue(tempTable)),
		Priority: 100,
	})
}

func (c *connection) columnClause(ctx context.Context, table string) (string, error) {
	var columnClause strings.Builder
	args := []any{c.config.Database, table}
	if c.config.Database == "" {
		args = []any{nil, table}
	}
	res, err := c.Execute(ctx, &drivers.Statement{
		Query:    "SELECT name, type FROM system.columns WHERE database = coalesce(?, currentDatabase()) AND table = ?",
		Args:     args,
		Priority: 100,
	})
	if err != nil {
		return "", err
	}
	defer res.Close()

	columnClause.WriteRune('(')
	var col, typ string
	for res.Next() {
		if err := res.Scan(&col, &typ); err != nil {
			return "", err
		}
		if columnClause.Len() > 1 {
			columnClause.WriteString(", ")
		}
		columnClause.WriteString(safeSQLName(col))
		columnClause.WriteString(" ")
		columnClause.WriteString(typ)
	}
	columnClause.WriteRune(')')
	return columnClause.String(), nil
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

	c.used()
	release := func() error {
		c.used()
		return conn.Close()
	}
	return conn, release, nil
}

func (c *connection) getTableEngine(ctx context.Context, name string) (string, error) {
	var engine string
	args := []any{c.config.Database, name}
	if c.config.Database == "" {
		args = []any{nil, name}
	}
	res, err := c.Execute(ctx, &drivers.Statement{
		Query:    "SELECT engine FROM system.tables WHERE database = coalesce(?, currentDatabase()) AND name = ?",
		Args:     args,
		Priority: 1,
	})
	if err != nil {
		return "", err
	}
	defer res.Close()
	if res.Next() {
		if err := res.Scan(&engine); err != nil {
			return "", err
		}
	}
	return engine, nil
}

func (c *connection) getTablePartitions(ctx context.Context, name string) ([]string, error) {
	res, err := c.Execute(ctx, &drivers.Statement{
		Query:    "SELECT DISTINCT partition FROM system.parts WHERE table = ?",
		Args:     []any{name},
		Priority: 1,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()
	// collect partitions
	var partitions []string
	for res.Next() {
		var part string
		if err := res.Scan(&part); err != nil {
			return nil, err
		}
		partitions = append(partitions, part)
	}
	return partitions, nil
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

var errUnsupportedType = errors.New("encountered unsupported clickhouse type")

func safelocalTableName(name string) string {
	return safeSQLName(name + "_local")
}

func localTableName(name string) string {
	return name + "_local"
}

func tempTableForDictionary(name string) string {
	return name + "_dict_temp_"
}

func safeSQLString(name string) string {
	return drivers.DialectClickHouse.EscapeStringValue(name)
}
