package duckdb

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
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

func (c *connection) WithConnection(ctx context.Context, priority int, longRunning, tx bool, fn drivers.WithConnectionFunc) error {
	// Check not nested
	if connFromContext(ctx) != nil {
		panic("nested WithConnection")
	}

	// Acquire connection
	conn, release, err := c.acquireOLAPConn(ctx, priority, longRunning, tx)
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
	conn, release, err := c.acquireOLAPConn(ctx, stmt.Priority, stmt.LongRunning, false)
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

func (c *connection) EstimateSize() (int64, bool) {
	path := c.config.DBFilePath
	if path == "" {
		return 0, true
	}

	// Add .wal file path (e.g final size will be sum of *.db and *.db.wal)
	dbWalPath := fmt.Sprintf("%s.wal", path)
	paths := []string{path, dbWalPath}
	if c.config.ExtTableStorage {
		entries, err := os.ReadDir(c.config.DBStoragePath)
		if err == nil { // ignore error
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				path := filepath.Join(c.config.DBStoragePath, entry.Name())
				version, exist, err := c.tableVersion(entry.Name())
				if err != nil || !exist {
					continue
				}
				paths = append(paths, filepath.Join(path, fmt.Sprintf("%s.db", version)), filepath.Join(path, fmt.Sprintf("%s.db.wal", version)))
			}
		}
	}
	return fileSize(paths), true
}

// AddTableColumn implements drivers.OLAPStore.
func (c *connection) AddTableColumn(ctx context.Context, tableName, columnName, typ string) error {
	c.logger.Debug("add table column", zap.String("tableName", tableName), zap.String("columnName", columnName), zap.String("typ", typ))
	if !c.config.ExtTableStorage {
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", safeSQLName(tableName), safeSQLName(columnName), typ),
			Priority:    1,
			LongRunning: true,
		})
	}

	version, exist, err := c.tableVersion(tableName)
	if err != nil {
		return err
	}

	if !exist {
		return fmt.Errorf("table %q does not exist", tableName)
	}
	dbName := dbName(tableName, version)
	return c.WithConnection(ctx, 1, true, false, func(ctx, ensuredCtx context.Context, conn *dbsql.Conn) error {
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("ALTER TABLE %s.default ADD COLUMN %s %s", safeSQLName(dbName), safeSQLName(columnName), typ)})
		if err != nil {
			return err
		}
		// recreate view to propagate schema changes
		return c.Exec(ensuredCtx, &drivers.Statement{Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s AS SELECT * FROM %s.default", safeSQLName(tableName), safeSQLName(dbName))})
	})
}

// AlterTableColumn implements drivers.OLAPStore.
func (c *connection) AlterTableColumn(ctx context.Context, tableName, columnName, newType string) error {
	c.logger.Debug("alter table column", zap.String("tableName", tableName), zap.String("columnName", columnName), zap.String("newType", newType))
	if !c.config.ExtTableStorage {
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("ALTER TABLE %s ALTER %s TYPE %s", safeSQLName(tableName), safeSQLName(columnName), newType),
			Priority:    1,
			LongRunning: true,
		})
	}

	version, exist, err := c.tableVersion(tableName)
	if err != nil {
		return err
	}

	if !exist {
		return fmt.Errorf("table %q does not exist", tableName)
	}
	dbName := dbName(tableName, version)
	return c.WithConnection(ctx, 1, true, false, func(ctx, ensuredCtx context.Context, conn *dbsql.Conn) error {
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("ALTER TABLE %s.default ALTER %s TYPE %s", safeSQLName(dbName), safeSQLName(columnName), newType)})
		if err != nil {
			return err
		}

		// recreate view to propagate schema changes
		return c.Exec(ensuredCtx, &drivers.Statement{Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s AS SELECT * FROM %s.default", safeSQLName(tableName), safeSQLName(dbName))})
	})
}

// CreateTableAsSelect implements drivers.OLAPStore.
// We add a \n at the end of the any user query to ensure any comment at the end of model doesn't make the query incomplete.
func (c *connection) CreateTableAsSelect(ctx context.Context, name string, view bool, sql string) error {
	c.logger.Debug("create table", zap.String("name", name), zap.Bool("view", view))
	if view {
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("CREATE OR REPLACE VIEW %s AS (%s\n)", safeSQLName(name), sql),
			Priority:    1,
			LongRunning: true,
		})
	}
	if !c.config.ExtTableStorage {
		return c.execWithLimits(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s\n)", safeSQLName(name), sql),
			Priority:    1,
			LongRunning: true,
		})
	}

	return c.WithConnection(ctx, 1, true, false, func(ctx, ensuredCtx context.Context, _ *dbsql.Conn) error {
		// NOTE: Running mkdir while holding the connection to avoid directory getting cleaned up when concurrent calls to RenameTable cause reopenDB to be called.

		// create a new db file in /<instanceid>/<name> directory
		sourceDir := filepath.Join(c.config.DBStoragePath, name)
		if err := os.Mkdir(sourceDir, fs.ModePerm); err != nil && !errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("create: unable to create dir %q: %w", sourceDir, err)
		}

		// check if some older version existed previously to detach it later
		oldVersion, oldVersionExists, _ := c.tableVersion(name)

		newVersion := fmt.Sprint(time.Now().UnixMilli())
		dbFile := filepath.Join(sourceDir, fmt.Sprintf("%s.db", newVersion))
		db := dbName(name, newVersion)

		// attach new db
		err := c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("ATTACH %s AS %s", safeSQLString(dbFile), safeSQLName(db))})
		if err != nil {
			removeDBFile(dbFile)
			return fmt.Errorf("create: attach %q db failed: %w", dbFile, err)
		}

		// Enforce storage limits
		if err := c.execWithLimits(ctx, &drivers.Statement{Query: fmt.Sprintf("CREATE OR REPLACE TABLE %s.default AS (%s\n)", safeSQLName(db), sql)}); err != nil {
			c.detachAndRemoveFile(ensuredCtx, db, dbFile)
			return fmt.Errorf("create: create %q.default table failed: %w", db, err)
		}

		// success update version
		err = c.updateVersion(name, newVersion)
		if err != nil {
			// extreme bad luck
			c.detachAndRemoveFile(ensuredCtx, db, dbFile)
			return fmt.Errorf("create: update version %q failed: %w", newVersion, err)
		}

		qry, err := c.generateSelectQuery(ctx, db)
		if err != nil {
			return err
		}

		// create view query
		err = c.Exec(ctx, &drivers.Statement{
			Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", safeSQLName(name), qry),
		})
		if err != nil {
			c.detachAndRemoveFile(ensuredCtx, db, dbFile)
			return fmt.Errorf("create: create view %q failed: %w", name, err)
		}

		if oldVersionExists {
			oldDB := dbName(name, oldVersion)
			// ignore these errors since source has been correctly ingested and attached
			c.detachAndRemoveFile(ensuredCtx, oldDB, filepath.Join(sourceDir, fmt.Sprintf("%s.db", oldVersion)))
		}
		return nil
	})
}

// DropTable implements drivers.OLAPStore.
func (c *connection) DropTable(ctx context.Context, name string, view bool) error {
	c.logger.Debug("drop table", zap.String("name", name), zap.Bool("view", view))
	if !c.config.ExtTableStorage {
		var typ string
		if view {
			typ = "VIEW"
		} else {
			typ = "TABLE"
		}
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("DROP %s IF EXISTS %s", typ, safeSQLName(name)),
			Priority:    100,
			LongRunning: true,
		})
	}
	// determine if it is a true view or view on externally stored table
	version, exist, err := c.tableVersion(name)
	if err != nil {
		return err
	}

	if !exist {
		if !view {
			return nil
		}
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("DROP VIEW IF EXISTS %s", safeSQLName(name)),
			Priority:    100,
			LongRunning: true,
		})
	}

	err = c.WithConnection(ctx, 100, true, false, func(ctx, ensuredCtx context.Context, _ *dbsql.Conn) error {
		// drop view
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP VIEW IF EXISTS %s", safeSQLName(name))})
		if err != nil {
			return err
		}

		oldDB := dbName(name, version)
		err = c.Exec(ensuredCtx, &drivers.Statement{Query: fmt.Sprintf("DETACH %s", safeSQLName(oldDB))})
		if err != nil && !strings.Contains(err.Error(), "database not found") { // ignore database not found errors for idempotency
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// delete source directory
	return os.RemoveAll(filepath.Join(c.config.DBStoragePath, name))
}

// InsertTableAsSelect implements drivers.OLAPStore.
func (c *connection) InsertTableAsSelect(ctx context.Context, name string, byName bool, sql string) error {
	c.logger.Debug("insert into table", zap.String("name", name), zap.Bool("byName", byName))
	var insertByNameClause string
	if byName {
		insertByNameClause = "BY NAME"
	} else {
		insertByNameClause = ""
	}

	if !c.config.ExtTableStorage {
		// Enforce storage limits
		return c.execWithLimits(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("INSERT INTO %s %s (%s)", safeSQLName(name), insertByNameClause, sql),
			Priority:    1,
			LongRunning: true,
		})
	}

	version, exist, err := c.tableVersion(name)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("InsertTableAsSelect: table %q does not exist", name)
	}
	return c.execWithLimits(ctx, &drivers.Statement{
		Query:       fmt.Sprintf("INSERT INTO %s.default %s (%s)", safeSQLName(dbName(name, version)), insertByNameClause, sql),
		Priority:    1,
		LongRunning: true,
	})
}

// RenameTable implements drivers.OLAPStore.
// For drop and replace (when running `RenameTable("__tmp_foo", "foo")`):
// `DROP VIEW __tmp_foo`
// `DETACH __tmp_foo__1`
// `mv __tmp_foo/1.db foo/2.db`
// `echo 2 > version.txt`
// `rm __tmp_foo`
// `ATTACH 'foo/2.db' AS foo__2`
// `CREATE OR REPLACE VIEW foo AS SELECT * FROM foo_2`
// `DETACH foo__1`
// `rm foo/1.db`
func (c *connection) RenameTable(ctx context.Context, oldName, newName string, view bool) error {
	c.logger.Debug("rename table", zap.String("from", oldName), zap.String("to", newName), zap.Bool("view", view), zap.Bool("ext", c.config.ExtTableStorage))
	if strings.EqualFold(oldName, newName) {
		return fmt.Errorf("rename: old and new name are same case insensitive strings")
	}
	if !c.config.ExtTableStorage {
		return c.dropAndReplace(ctx, oldName, newName, view)
	}
	// determine if it is a true view or a view on externally stored table
	oldVersion, exist, err := c.tableVersion(oldName)
	if err != nil {
		return err
	}
	if !exist {
		return c.dropAndReplace(ctx, oldName, newName, view)
	}

	// reopen duckdb connections which should delete any temporary files built up during ingestion
	// making an empty call so that stop the world call with tx=true is very fast and only blocks for the duration of close and open db hanle call
	err = c.WithConnection(ctx, 100, false, true, func(_, _ context.Context, _ *dbsql.Conn) error { return nil })
	if err != nil {
		return err
	}

	oldVersionInNewDir, replaceInNewTable, err := c.tableVersion(newName)
	if err != nil {
		return err
	}

	newSrcDir := filepath.Join(c.config.DBStoragePath, newName)
	oldSrcDir := filepath.Join(c.config.DBStoragePath, oldName)

	return c.WithConnection(ctx, 100, true, false, func(currentCtx, ctx context.Context, conn *dbsql.Conn) error {
		err = os.Mkdir(newSrcDir, fs.ModePerm)
		if err != nil && !errors.Is(err, fs.ErrExist) {
			return err
		}

		// drop old view
		err = c.Exec(currentCtx, &drivers.Statement{Query: fmt.Sprintf("DROP VIEW IF EXISTS %s", safeSQLName(oldName))})
		if err != nil {
			return fmt.Errorf("rename: drop %q view failed: %w", oldName, err)
		}

		// detach old db
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DETACH %s", safeSQLName(dbName(oldName, oldVersion)))})
		if err != nil {
			return fmt.Errorf("rename: detach %q db failed: %w", dbName(oldName, oldVersion), err)
		}

		// move old file as a new file in source directory
		newVersion := fmt.Sprint(time.Now().UnixMilli())
		newFile := filepath.Join(newSrcDir, fmt.Sprintf("%s.db", newVersion))
		err = os.Rename(filepath.Join(oldSrcDir, fmt.Sprintf("%s.db", oldVersion)), newFile)
		if err != nil {
			return fmt.Errorf("rename: rename file failed: %w", err)
		}
		// also move .db.wal file in case checkpointing was not completed
		_ = os.Rename(filepath.Join(oldSrcDir, fmt.Sprintf("%s.db.wal", oldVersion)),
			filepath.Join(newSrcDir, fmt.Sprintf("%s.db.wal", newVersion)))

		err = c.updateVersion(newName, newVersion)
		if err != nil {
			return fmt.Errorf("rename: update version failed: %w", err)
		}
		err = os.RemoveAll(filepath.Join(c.config.DBStoragePath, oldName))
		if err != nil {
			c.logger.Error("rename: unable to delete old path", zap.Error(err))
		}

		newDB := dbName(newName, newVersion)
		// attach new db
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("ATTACH %s AS %s", safeSQLString(newFile), safeSQLName(newDB))})
		if err != nil {
			return fmt.Errorf("rename: attach %q db failed: %w", newDB, err)
		}

		qry, err := c.generateSelectQuery(ctx, newDB)
		if err != nil {
			return err
		}

		// change view query
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", safeSQLName(newName), qry)})
		if err != nil {
			return fmt.Errorf("rename: create %q view failed: %w", newName, err)
		}

		if replaceInNewTable { // new table had some other file previously
			c.detachAndRemoveFile(ctx, dbName(newName, oldVersionInNewDir), filepath.Join(newSrcDir, fmt.Sprintf("%s.db", oldVersionInNewDir)))
		}
		return nil
	})
}

func (c *connection) dropAndReplace(ctx context.Context, oldName, newName string, view bool) error {
	var typ string
	if view {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	existing, err := c.InformationSchema().Lookup(ctx, newName)
	if err != nil {
		if !errors.Is(err, drivers.ErrNotFound) {
			return err
		}
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("ALTER %s %s RENAME TO %s", typ, safeSQLName(oldName), safeSQLName(newName)),
			Priority:    100,
			LongRunning: true,
		})
	}

	return c.WithConnection(ctx, 100, true, true, func(ctx, ensuredCtx context.Context, conn *dbsql.Conn) error {
		// The newName may currently be occupied by a name of another type than oldName.
		var existingTyp string
		if existing.View {
			existingTyp = "VIEW"
		} else {
			existingTyp = "TABLE"
		}

		err := c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP %s IF EXISTS %s", existingTyp, newName)})
		if err != nil {
			return err
		}

		return c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("ALTER %s %s RENAME TO %s", typ, safeSQLName(oldName), safeSQLName(newName))})
	})
}

func (c *connection) detachAndRemoveFile(ctx context.Context, db, dbFile string) {
	err := c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DETACH %s", safeSQLName(db)), Priority: 100})
	if err != nil {
		c.logger.Error("detach failed", zap.String("db", db), zap.Error(err))
	}
	removeDBFile(dbFile)
}

func (c *connection) tableVersion(name string) (string, bool, error) {
	pathToFile := filepath.Join(c.config.DBStoragePath, name, "version.txt")
	contents, err := os.ReadFile(pathToFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(string(contents)), true, nil
}

func (c *connection) updateVersion(name, version string) error {
	pathToFile := filepath.Join(c.config.DBStoragePath, name, "version.txt")
	file, err := os.Create(pathToFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(version)
	return err
}

func (c *connection) execWithLimits(parentCtx context.Context, stmt *drivers.Statement) error {
	storageLimit := c.config.StorageLimitBytes
	if storageLimit <= 0 { // no limit
		return c.Exec(parentCtx, stmt)
	}

	// check current size
	currentSize, _ := c.EstimateSize()
	storageLimit -= currentSize
	// current size already exceeds limit
	if storageLimit <= 0 {
		return drivers.ErrStorageLimitExceeded
	}

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()
	limitExceeded := atomic.Bool{}
	// Start background goroutine to check size is not exceeded during query execution
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if size, ok := c.EstimateSize(); ok && size > storageLimit {
					limitExceeded.Store(true)
					cancel()
					return
				}
			}
		}
	}()

	err := c.Exec(ctx, stmt)
	if limitExceeded.Load() {
		return drivers.ErrStorageLimitExceeded
	}
	return err
}

// convertToEnum converts a varchar col in table to an enum type.
// Generally to be used for low cardinality varchar columns although not enforced here.
func (c *connection) convertToEnum(ctx context.Context, table string, cols []string) error {
	if len(cols) == 0 {
		return fmt.Errorf("empty list")
	}
	if !c.config.ExtTableStorage {
		return fmt.Errorf("`cast_to_enum` is only supported when `external_table_storage` is enabled")
	}
	c.logger.Debug("convert column to enum", zap.String("table", table), zap.Strings("col", cols))

	oldVersion, exist, err := c.tableVersion(table)
	if err != nil {
		return err
	}

	if !exist {
		return fmt.Errorf("table %q does not exist", table)
	}

	// scan main db and main schema
	res, err := c.Execute(ctx, &drivers.Statement{
		Query:    "SELECT current_database(), current_schema()",
		Priority: 100,
	})
	if err != nil {
		return err
	}

	var mainDB, mainSchema string
	if res.Next() {
		if err := res.Scan(&mainDB, &mainSchema); err != nil {
			_ = res.Close()
			return err
		}
	}
	_ = res.Close()

	sourceDir := filepath.Join(c.config.DBStoragePath, table)
	newVersion := fmt.Sprint(time.Now().UnixMilli())
	newDBFile := filepath.Join(sourceDir, fmt.Sprintf("%s.db", newVersion))
	newDB := dbName(table, newVersion)
	return c.WithConnection(ctx, 100, true, false, func(ctx, ensuredCtx context.Context, _ *dbsql.Conn) error {
		// attach new db
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("ATTACH %s AS %s", safeSQLString(newDBFile), safeSQLName(newDB))})
		if err != nil {
			removeDBFile(newDBFile)
			return fmt.Errorf("create: attach %q db failed: %w", newDBFile, err)
		}

		// switch to new db
		// this is only required since duckdb has bugs around db scoped custom types
		// TODO: remove this when https://github.com/duckdb/duckdb/pull/9622 is released
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("USE %s", safeSQLName(newDB))})
		if err != nil {
			c.detachAndRemoveFile(ctx, newDB, newDBFile)
			return fmt.Errorf("failed switch db %q: %w", newDB, err)
		}
		defer func() {
			// switch to original db, notice `db.schema` just doing USE db switches context to `main` schema in the current db if doing `USE main`
			// we want to switch to original db and schema
			err = c.Exec(ensuredCtx, &drivers.Statement{Query: fmt.Sprintf("USE %s.%s", safeSQLName(mainDB), safeSQLName(mainSchema))})
			if err != nil {
				c.detachAndRemoveFile(ctx, newDB, newDBFile)
				// This should NEVER happen
				c.fatalInternalError(fmt.Errorf("failed to switch back from db %q: %w", mainDB, err))
			}
		}()

		oldDB := dbName(table, oldVersion)
		for _, col := range cols {
			enum := fmt.Sprintf("%s_enum", col)
			if err = c.Exec(ensuredCtx, &drivers.Statement{Query: fmt.Sprintf("CREATE TYPE %s AS ENUM (SELECT DISTINCT %s FROM %s.default WHERE %s IS NOT NULL)", safeSQLName(enum), safeSQLName(col), safeSQLName(oldDB), safeSQLName(col))}); err != nil {
				c.detachAndRemoveFile(ctx, newDB, newDBFile)
				return fmt.Errorf("failed to create enum %q: %w", enum, err)
			}
		}

		var selectQry string
		for _, col := range cols {
			enum := fmt.Sprintf("%s_enum", col)
			selectQry += fmt.Sprintf("CAST(%s AS %s) AS %s,", safeSQLName(col), safeSQLName(enum), safeSQLName(col))
		}
		selectQry += fmt.Sprintf("* EXCLUDE(%s)", strings.Join(cols, ","))

		if err := c.Exec(ensuredCtx, &drivers.Statement{Query: fmt.Sprintf("CREATE OR REPLACE TABLE \"default\" AS SELECT %s FROM %s.default", selectQry, safeSQLName(oldDB))}); err != nil {
			c.detachAndRemoveFile(ctx, newDB, newDBFile)
			return fmt.Errorf("failed to create table with enum values: %w", err)
		}

		// recreate view to propagate schema changes
		selectQry, err := c.generateSelectQuery(ctx, newDB)
		if err != nil {
			return err
		}

		// NOTE :: db name need to be appened in the view query else query fails when switching to main db
		if err := c.Exec(ensuredCtx, &drivers.Statement{Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s.%s.%s AS %s", safeSQLName(mainDB), safeSQLName(mainSchema), safeSQLName(table), selectQry)}); err != nil {
			c.detachAndRemoveFile(ctx, newDB, newDBFile)
			return fmt.Errorf("failed to create view %q: %w", table, err)
		}

		// update version and detach old db
		if err := c.updateVersion(table, newVersion); err != nil {
			c.detachAndRemoveFile(ctx, newDB, newDBFile)
			return fmt.Errorf("failed to update version: %w", err)
		}

		c.detachAndRemoveFile(ensuredCtx, oldDB, filepath.Join(sourceDir, fmt.Sprintf("%s.db", oldVersion)))
		return nil
	})
}

// duckDB raises Contents of view were altered: types don't match! error even when number of columns are same but sequence of column changes in underlying table.
// This causes temporary query failures till the model view is not updated to reflect the new column sequence.
// We ensure that view for external table storage is always generated using a stable order of columns of underlying table.
// Additionally we want to keep the same order as the underlying table locally so that we can show columns in the same order as they appear in source data.
// Using `AllowHostAccess` as proxy to check if we are running in local/cloud mode.
func (c *connection) generateSelectQuery(ctx context.Context, db string) (string, error) {
	if c.config.AllowHostAccess {
		return fmt.Sprintf("SELECT * FROM %s.default", safeSQLName(db)), nil
	}

	rows, err := c.Execute(ctx, &drivers.Statement{
		Query: fmt.Sprintf(`
			SELECT column_name AS name
			FROM information_schema.columns
			WHERE table_catalog = %s AND table_name = 'default'
			ORDER BY name ASC`, safeSQLString(db)),
	})
	if err != nil {
		return "", err
	}
	defer rows.Close()

	cols := make([]string, 0)
	var col string
	for rows.Next() {
		if err := rows.Scan(&col); err != nil {
			return "", err
		}
		cols = append(cols, safeName(col))
	}

	return fmt.Sprintf("SELECT %s FROM %s.default", strings.Join(cols, ", "), safeSQLName(db)), nil
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

func dbName(name, version string) string {
	return fmt.Sprintf("%s_%s", name, version)
}

func removeDBFile(dbFile string) {
	_ = os.Remove(dbFile)
	// Hacky approach to remove the wal and tmp file
	_ = os.Remove(dbFile + ".wal")
	_ = os.RemoveAll(dbFile + ".tmp")
}

// safeSQLName returns a quoted SQL identifier.
func safeSQLName(name string) string {
	if name == "" {
		return name
	}
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(name, "\"", "\"\""))
}

func safeSQLString(name string) string {
	if name == "" {
		return name
	}
	return fmt.Sprintf("'%s'", strings.ReplaceAll(name, "'", "''"))
}
