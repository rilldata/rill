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
			c.activity.Emit(ctx, "duckdb_queue_latency_ms", float64(queueLatency), attrs...)
			c.activity.Emit(ctx, "duckdb_total_latency_ms", float64(totalLatency), attrs...)
			if acquired {
				c.activity.Emit(ctx, "duckdb_query_latency_ms", float64(totalLatency-queueLatency), attrs...)
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

	schema, err := rowsToSchema(rows)
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
	var paths []string
	path := c.config.DBFilePath
	if path == "" {
		return 0, true
	}

	// Add .wal file path (e.g final size will be sum of *.db and *.db.wal)
	dbWalPath := fmt.Sprintf("%s.wal", path)
	paths = append(paths, path, dbWalPath)
	return fileSize(paths), true
}

// AddTableColumn implements drivers.OLAPStore.
func (c *connection) AddTableColumn(ctx context.Context, tableName, columnName, typ string) error {
	if !c.config.ExtTableStorage {
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", safeSQLName(tableName), safeSQLName(columnName), typ),
			Priority: 1,
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
	if !c.config.ExtTableStorage {
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("ALTER TABLE %s ALTER %s TYPE %s", safeSQLName(tableName), safeSQLName(columnName), newType),
			Priority: 1,
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
func (c *connection) CreateTableAsSelect(ctx context.Context, name string, view bool, sql string) error {
	if view {
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("CREATE OR REPLACE VIEW %s AS (%s)", safeSQLName(name), sql),
			Priority:    1,
			LongRunning: true,
		})
	}
	if !c.config.ExtTableStorage {
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s)", safeSQLName(name), sql),
			Priority:    1,
			LongRunning: true,
		})
	}
	// create a new db file in /<instanceid>/<name> directory
	sourceDir := filepath.Join(c.config.ExtStoragePath, name)
	if err := os.Mkdir(sourceDir, fs.ModePerm); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
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
		return err
	}

	if err := c.Exec(ctx, &drivers.Statement{
		Query:       fmt.Sprintf("CREATE OR REPLACE TABLE %s.default AS (%s)", safeSQLName(db), sql),
		Priority:    1,
		LongRunning: true,
	}); err != nil {
		c.detachAndRemoveFile(ctx, db, dbFile)
		return err
	}

	return c.WithConnection(ctx, 1, true, false, func(ctx, ensuredCtx context.Context, _ *dbsql.Conn) error {
		// success update version
		err = c.updateVersion(name, newVersion)
		if err != nil {
			// extreme bad luck
			c.detachAndRemoveFile(ensuredCtx, db, dbFile)
			return err
		}

		// create view query
		err = c.Exec(ctx, &drivers.Statement{
			Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s AS SELECT * FROM %s.default", safeSQLName(name), safeSQLName(db)),
		})
		if err != nil {
			c.detachAndRemoveFile(ensuredCtx, db, dbFile)
			return err
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
	if view {
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("DROP VIEW IF EXISTS %s", safeSQLName(name)),
			Priority:    100,
			LongRunning: true,
		})
	}
	if !c.config.ExtTableStorage {
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("DROP TABLE IF EXISTS %s", safeSQLName(name)),
			Priority:    100,
			LongRunning: true,
		})
	}

	version, exist, err := c.tableVersion(name)
	if err != nil {
		return err
	}

	if !exist {
		return nil
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
	return os.RemoveAll(filepath.Join(c.config.ExtStoragePath, name))
}

// InsertTableAsSelect implements drivers.OLAPStore.
func (c *connection) InsertTableAsSelect(ctx context.Context, name string, byName bool, sql string) error {
	var insertByNameClause string
	if byName {
		insertByNameClause = "BY NAME"
	} else {
		insertByNameClause = ""
	}

	if !c.config.ExtTableStorage {
		return c.Exec(ctx, &drivers.Statement{
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
		return fmt.Errorf("table %q does not exist", name)
	}
	return c.Exec(ctx, &drivers.Statement{
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
	if strings.EqualFold(oldName, newName) {
		return fmt.Errorf("old and new name are same case insensitive strings")
	}
	if view || !c.config.ExtTableStorage {
		return c.dropAndReplace(ctx, oldName, newName, view)
	}

	oldVersion, exist, err := c.tableVersion(oldName)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("table %q does not exist", oldName)
	}

	oldVersionInNewDir, replaceInNewTable, err := c.tableVersion(newName)
	if err != nil {
		return err
	}

	newSrcDir := filepath.Join(c.config.ExtStoragePath, newName)
	err = os.Mkdir(newSrcDir, fs.ModePerm)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}

	return c.WithConnection(ctx, 100, true, false, func(currentCtx, ctx context.Context, conn *dbsql.Conn) error {
		// drop old view
		err = c.Exec(currentCtx, &drivers.Statement{Query: fmt.Sprintf("DROP VIEW IF EXISTS %s", safeSQLName(oldName))})
		if err != nil {
			return err
		}

		// detach old db
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DETACH %s", safeSQLName(dbName(oldName, oldVersion)))})
		if err != nil {
			return err
		}

		// move old file as a new file in source directory
		newVersion := fmt.Sprint(time.Now().UnixMilli())
		oldFile := filepath.Join(c.config.ExtStoragePath, oldName, fmt.Sprintf("%s.db", oldVersion))
		newFile := filepath.Join(newSrcDir, fmt.Sprintf("%s.db", newVersion))
		err = os.Rename(oldFile, newFile)
		if err != nil {
			return err
		}

		err = c.updateVersion(newName, newVersion)
		if err != nil {
			return err
		}
		_ = os.RemoveAll(filepath.Join(c.config.ExtStoragePath, oldName))

		newDB := dbName(newName, newVersion)
		// attach new db
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("ATTACH %s AS %s", safeSQLString(newFile), safeSQLName(newDB))})
		if err != nil {
			return err
		}

		// change view query
		err = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s AS SELECT * FROM %s.default", safeSQLName(newName), safeSQLName(newDB))})
		if err != nil {
			return err
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
	existingTo, _ := c.InformationSchema().Lookup(ctx, newName)
	if existingTo != nil {
		return c.Exec(ctx, &drivers.Statement{
			Query:       fmt.Sprintf("ALTER %s %s RENAME TO %s", typ, safeSQLName(oldName), safeSQLName(newName)),
			Priority:    100,
			LongRunning: true,
		})
	}

	return c.WithConnection(ctx, 100, true, true, func(ctx, ensuredCtx context.Context, conn *dbsql.Conn) error {
		err := c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP %s IF EXIST %s", typ, newName)})
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
	pathToFile := filepath.Join(c.config.ExtStoragePath, name, "version.txt")
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
	pathToFile := filepath.Join(c.config.ExtStoragePath, name, "version.txt")
	file, err := os.Create(pathToFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(version)
	return err
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

func fileSize(paths []string) int64 {
	var size int64
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil { // ignoring error since only error possible is *PathError
			size += info.Size()
		}
	}
	return size
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
