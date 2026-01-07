package rduckdb

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/duckdb/duckdb-go/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

type generic struct {
	db            *sqlx.DB
	localFileName string // localFileName is the name of the local DuckDB file, used for cleanup.
	logger        *zap.Logger
	writeSem      *semaphore.Weighted

	opts *GenericOptions
}

type GenericOptions struct {
	// Path to the external DuckDB database.
	Path string
	// Attach string allows user to directly pass a DuckDB attach string.
	//  Example syntax : "'ducklake:metadata.ducklake' AS my_ducklake(DATA_PATH 'datafiles1')"
	Attach string
	// DBName is set to the name of the database identified by the Path.
	DBName string
	// SchemaName switches the default schema.
	SchemaName string
	// ReadOnlyMode is set to true if the connection is read-only.
	ReadOnlyMode bool

	// LocalDataDir is the path to the local DuckDB database file.
	LocalDataDir string
	// LocalCPU cores available for the local instance of DB.
	LocalCPU int
	// LocalMemoryLimitGB is the amount of memory available for the local instance of DB.
	LocalMemoryLimitGB int
	// Settings are additional query parameters to be passed when creating local DuckDB instance.
	Settings map[string]string

	// DBInitQueries are run when the database is first created. These are typically global duckdb configurations.
	DBInitQueries []string
	// ConnInitQueries are run when a new connection is created. These are typically local duckdb configurations.
	ConnInitQueries []string

	Logger         *zap.Logger
	OtelAttributes []attribute.KeyValue
}

func (d *GenericOptions) validateAndApplyDefaults() error {
	if d.Path != "" && d.Attach != "" {
		return fmt.Errorf("cannot specify both `path` and `attach`")
	}
	if d.Settings == nil {
		d.Settings = make(map[string]string)
	}
	if d.LocalMemoryLimitGB > 0 {
		d.Settings["memory_limit"] = fmt.Sprintf("%d bytes", d.LocalMemoryLimitGB*1000*1000*1000)
	}
	if d.LocalCPU > 0 {
		d.Settings["threads"] = strconv.Itoa(d.LocalCPU)
	}
	return nil
}

// NewGeneric creates a duckdb database connection with the given Path.
// It can be used to run OLAP queries on an external local DuckDB database or a duckdb service like MotherDuck.
// Operations like CreateTableAsSelect, RenameTable, MutateTable are not atomic and can fail in the middle.
func NewGeneric(ctx context.Context, opts *GenericOptions) (res DB, dbErr error) {
	opts.Logger.Debug("duckdb: open generic db", observability.ZapCtx(ctx))
	// open the db
	//
	// we create a ephemeral local DuckDB instance and then attach the external database if Path is set.
	// This is to control where wal and tmp files are created.
	// An ephemeral local DuckDB instance is used since the go-duckdb driver caches some state wrt same database path
	// and attaching same motherduck instance leads to issues.
	err := opts.validateAndApplyDefaults()
	if err != nil {
		return nil, err
	}

	localFileName := filepath.Join(opts.LocalDataDir, "main"+uuid.NewString()[:8]+".db")
	connector, err := duckdb.NewConnector(dsnForLocalDuckDB(localFileName, opts.Settings), func(execer driver.ExecerContext) error {
		for _, qry := range opts.ConnInitQueries {
			_, err := execer.ExecContext(ctx, qry, nil)
			if err != nil && strings.Contains(err.Error(), "Failed to download extension") {
				// Retry using another mirror. Based on: https://github.com/duckdb/duckdb/issues/9378
				_, err = execer.ExecContext(ctx, qry+" FROM 'http://nightly-extensions.duckdb.org'", nil)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	spanOptions := otelsql.SpanOptions{
		SpanFilter: func(ctx context.Context, method otelsql.Method, query string, args []driver.NamedValue) bool {
			// log all queries except create secret which can contain sensitive data
			return !createSecretRegex.MatchString(query)
		},
	}
	db := sqlx.NewDb(otelsql.OpenDB(connector, otelsql.WithSpanOptions(spanOptions)), "duckdb")
	defer func() {
		if dbErr != nil && db != nil {
			_ = db.Close()
		}
	}()

	// run init queries
	for _, qry := range opts.DBInitQueries {
		_, err := db.ExecContext(ctx, qry)
		if err != nil {
			return nil, err
		}
	}

	// attach the passed external db
	if opts.Path != "" {
		if opts.ReadOnlyMode {
			_, err = db.ExecContext(ctx, fmt.Sprintf("ATTACH %s (READ_ONLY)", safeSQLString(opts.Path)))
		} else {
			_, err = db.ExecContext(ctx, fmt.Sprintf("ATTACH %s", safeSQLString(opts.Path)))
		}
		if err != nil {
			return nil, fmt.Errorf("error attaching external db: %w", err)
		}
	} else if opts.Attach != "" {
		// attach the database using the attach string
		_, err = db.ExecContext(ctx, fmt.Sprintf("ATTACH %s", opts.Attach))
		if err != nil {
			return nil, fmt.Errorf("error attaching external db using attach string: %w", err)
		}
	}
	if opts.DBName == "" {
		// find the attached database name
		err = db.QueryRowxContext(
			ctx,
			`SELECT database_name
			 FROM duckdb_databases()
			 WHERE internal = false -- ignore internal information_schema databases
			   AND (path IS NOT NULL OR database_name = 'memory') -- all databases except the in-memory one should have a path 
			   AND database_name != current_database()`,
		).Scan(&opts.DBName)
		if err != nil {
			return nil, fmt.Errorf("error getting attached database name: %w. Set property `db_name` in the corresponding connector.yaml", err)
		}
	}

	err = otelsql.RegisterDBStatsMetrics(db.DB, otelsql.WithAttributes(opts.OtelAttributes...))
	if err != nil {
		return nil, fmt.Errorf("registering db stats metrics: %w", err)
	}
	return &generic{
		db:            db,
		localFileName: localFileName,
		logger:        opts.Logger,
		writeSem:      semaphore.NewWeighted(1),
		opts:          opts,
	}, nil
}

// AcquireReadConnection implements DB. In practice this does not enforce that the returned connection is read-only.
func (m *generic) AcquireReadConnection(ctx context.Context) (conn *sqlx.Conn, release func() error, err error) {
	conn, err = m.acquireConn(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("acquire read connection failed: %w", err)
	}
	return conn, conn.Close, nil
}

// Close implements DB.
func (m *generic) Close() error {
	_ = os.RemoveAll(m.localFileName)
	return m.db.Close()
}

// CreateTableAsSelect implements DB.
func (m *generic) CreateTableAsSelect(ctx context.Context, name, query string, opts *CreateTableOptions) (res *TableWriteMetrics, err error) {
	ctx, span := tracer.Start(ctx, "CreateTableAsSelect", trace.WithAttributes(
		attribute.String("name", name),
		attribute.String("query", query),
		attribute.Bool("view", opts.View),
	))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	err = m.writeSem.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer m.writeSem.Release(1)

	t := time.Now()
	var typ string
	if opts.View {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	conn, err := m.acquireConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("create: acquire connection failed: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	if err := m.dropTableUnsafe(ctx, name, conn); err != nil && !errors.Is(err, errNotFound) {
		return nil, fmt.Errorf("create: drop table %q failed: %w", name, err)
	}

	// run user queries
	if opts.BeforeCreateFn != nil {
		err = opts.BeforeCreateFn(ctx, conn)
		if err != nil {
			return nil, fmt.Errorf("create: BeforeCreateFn returned error: %w", err)
		}
	}
	execAfterCreate := func() error {
		if opts.AfterCreateFn == nil {
			return nil
		}
		err = opts.AfterCreateFn(ctx, conn)
		if err != nil {
			return fmt.Errorf("create: AfterCreateFn returned error: %w", err)
		}
		return nil
	}
	_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE %s %s AS (%s\n)", typ, safeSQLName(name), query), nil)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("create: create %s %q failed: %w", typ, name, err), execAfterCreate())
	}
	err = execAfterCreate()
	if err != nil {
		return nil, err
	}
	duration := time.Since(t)
	span.SetAttributes(attribute.Float64("query_duration", duration.Seconds()))
	return &TableWriteMetrics{Duration: duration}, nil
}

// DropTable implements DB.
func (m *generic) DropTable(ctx context.Context, name string) (resErr error) {
	ctx, span := tracer.Start(ctx, "DropTable", trace.WithAttributes(attribute.String("name", name)))
	defer func() {
		if resErr != nil {
			span.SetStatus(codes.Error, resErr.Error())
		}
		span.End()
	}()
	err := m.writeSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer m.writeSem.Release(1)

	conn, err := m.acquireConn(ctx)
	if err != nil {
		return fmt.Errorf("drop: acquire connection failed: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()
	return m.dropTableUnsafe(ctx, name, conn)
}

func (m *generic) dropTableUnsafe(ctx context.Context, name string, conn *sqlx.Conn) (resErr error) {
	var typ string
	tbl, _, err := m.schemaUsingConn(ctx, "", name, 0, "", conn)
	if err != nil {
		return err
	}
	if len(tbl) == 0 {
		return errNotFound
	}
	if tbl[0].View {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}
	_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP %s IF EXISTS %s", typ, safeSQLName(name)), nil)
	return err
}

// MutateTable implements DB.
func (m *generic) MutateTable(ctx context.Context, name string, initQueries []string, mutateFn func(ctx context.Context, conn *sqlx.Conn) error) (res *TableWriteMetrics, resErr error) {
	ctx, span := tracer.Start(ctx, "MutateTable", trace.WithAttributes(attribute.String("name", name)))
	defer func() {
		if resErr != nil {
			span.SetStatus(codes.Error, resErr.Error())
		}
		span.End()
	}()

	err := m.writeSem.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer m.writeSem.Release(1)

	t := time.Now()
	conn, err := m.acquireConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("mutate: acquire connection failed: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	err = mutateFn(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("mutate: mutate failed: %w", err)
	}

	duration := time.Since(t)
	span.SetAttributes(attribute.Float64("query_duration", duration.Seconds()))
	return &TableWriteMetrics{Duration: duration}, nil
}

// RenameTable implements DB.
func (m *generic) RenameTable(ctx context.Context, oldName, newName string) (resErr error) {
	ctx, span := tracer.Start(ctx, "RenameTable", trace.WithAttributes(attribute.String("old_name", oldName), attribute.String("new_name", newName)))
	defer func() {
		if resErr != nil {
			span.SetStatus(codes.Error, resErr.Error())
		}
		span.End()
	}()

	if strings.EqualFold(oldName, newName) {
		return fmt.Errorf("rename: Table with name %q already exists", newName)
	}
	err := m.writeSem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer m.writeSem.Release(1)

	// check the current type, if it is a view, rename it to a view
	// if it is a table, rename it to a table
	tbl, _, err := m.Schema(ctx, "", oldName, 0, "")
	if err != nil {
		return err
	}
	if len(tbl) == 0 {
		return fmt.Errorf("rename: Table with name %q does not exist", oldName)
	}

	// acquire a connection
	conn, err := m.acquireConn(ctx)
	if err != nil {
		return fmt.Errorf("rename: acquire connection failed: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// drop new table
	err = m.dropTableUnsafe(ctx, newName, conn)
	if err != nil && !errors.Is(err, errNotFound) {
		return fmt.Errorf("rename: Drop new table %q failed: %w", newName, err)
	}

	var typ string
	if tbl[0].View {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("ALTER %s %s RENAME TO %s", typ, oldName, newName), nil)
	return err
}

// Schema implements DB.
func (m *generic) Schema(ctx context.Context, ilike, name string, pageSize uint32, pageToken string) ([]*Table, string, error) {
	conn, err := m.acquireConn(ctx)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = conn.Close()
	}()
	return m.schemaUsingConn(ctx, ilike, name, pageSize, pageToken, conn)
}

func (m *generic) schemaUsingConn(ctx context.Context, ilike, name string, pageSize uint32, pageToken string, conn *sqlx.Conn) ([]*Table, string, error) {
	if ilike != "" && name != "" {
		return nil, "", fmt.Errorf("cannot specify both `ilike` and `name`")
	}

	var whereClause string
	var args []any
	if ilike != "" {
		whereClause = " AND t.table_name ilike ?"
		args = []any{ilike}
	} else if name != "" {
		whereClause = " AND t.table_name = ?"
		args = []any{name}
	}

	// Add pagination filter
	if pageToken != "" {
		var startAfterName string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfterName); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		whereClause += " AND t.table_name > ?"
		args = append(args, startAfterName)
	}

	q := fmt.Sprintf(`
		SELECT
			coalesce(t.table_catalog, current_database()) AS "database",
			current_schema() AS "schema",
			t.table_name AS "name",
			t.table_type = 'VIEW' AS "view", 
			array_agg(c.column_name ORDER BY c.ordinal_position) AS "column_names",
			array_agg(c.data_type ORDER BY c.ordinal_position) AS "column_types",
			array_agg(c.is_nullable = 'YES' ORDER BY c.ordinal_position) AS "column_nullable"
		FROM information_schema.tables t
		JOIN information_schema.columns c 
			ON t.table_schema = c.table_schema 
			AND t.table_name = c.table_name
		WHERE database = current_database() 
			AND t.table_schema = current_schema()
			%s
		GROUP BY ALL
		ORDER BY t.table_name
		LIMIT ?
	`, whereClause)

	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	args = append(args, limit+1)

	var res []*Table
	err := conn.SelectContext(ctx, &res, q, args...)
	if err != nil {
		return nil, "", err
	}

	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].Name)
	}

	return res, next, nil
}

// Size implements DB.
func (m *generic) Size() int64 {
	// todo: What is the size of the generic database? How to ignore tables which were not created within this project ?
	// Do we even need to track this for generic?
	return 0
}

func (m *generic) acquireConn(ctx context.Context) (*sqlx.Conn, error) {
	conn, err := m.db.Connx(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection failed: %w", err)
	}
	if m.opts.DBName != "" {
		_, err = conn.ExecContext(ctx, fmt.Sprintf("USE %s", safeSQLString(m.opts.DBName)))
		if err != nil {
			return nil, fmt.Errorf("acquire connection failed: %w", err)
		}
	}
	if m.opts.SchemaName != "" {
		_, err = conn.ExecContext(ctx, fmt.Sprintf("USE %s", safeSQLString(m.opts.SchemaName)))
		if err != nil {
			return nil, fmt.Errorf("acquire connection failed: %w", err)
		}
	}
	return conn, nil
}

func dsnForLocalDuckDB(path string, settings map[string]string) string {
	if len(settings) == 0 {
		return path
	}
	// Build DuckDB DSN (which should be "path?key=val&...")
	qry := make(url.Values)
	for k, v := range settings {
		qry.Set(k, v)
	}
	return path + "?" + qry.Encode()
}
