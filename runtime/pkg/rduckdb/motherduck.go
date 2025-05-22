package rduckdb

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb/v2"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

type motherduck struct {
	db       *sqlx.DB
	logger   *zap.Logger
	writeSem *semaphore.Weighted

	opts *MotherDuckDBOptions
}

type MotherDuckDBOptions struct {
	// Database is the name of motherduck database to connect to. Accepts both with md: prefix and without.
	Database string
	// Token is the token to connect to motherduck.
	Token string

	// LocalPath is the path to the local DuckDB database file.
	LocalPath string
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

func (d *MotherDuckDBOptions) bareDB() string {
	return strings.TrimPrefix(d.Database, "md:")
}

func (d *MotherDuckDBOptions) dsn() string {
	if strings.HasPrefix(d.Database, "md:") {
		return d.Database
	}
	return "md:" + d.Database
}

func (d *MotherDuckDBOptions) setDefaultSettings() {
	if d.Settings == nil {
		d.Settings = make(map[string]string)
	}
	if d.LocalMemoryLimitGB > 0 {
		d.Settings["memory_limit"] = fmt.Sprintf("%d bytes", d.LocalMemoryLimitGB*1000*1000*1000)
	}
	if d.LocalCPU > 0 {
		d.Settings["threads"] = strconv.Itoa(d.LocalCPU)
	}
}

// NewMotherDuck creates a duckdb database connection with the given motherduck db attached to it.
// Operations like CreateTableAsSelect, RenameTable, MutateTable are not atomic and can fail in the middle.
func NewMotherDuck(ctx context.Context, opts *MotherDuckDBOptions) (res DB, dbErr error) {
	opts.Logger.Debug("open motherduck db", observability.ZapCtx(ctx))
	// open the db
	opts.setDefaultSettings()
	dsn, err := url.Parse(filepath.Join(opts.LocalPath, "main.db"))
	if err != nil {
		return nil, err
	}
	query := dsn.Query()
	for k, v := range opts.Settings {
		query.Set(k, v)
	}
	// Rebuild DuckDB DSN (which should be "path?key=val&...")
	// this is required since spaces and other special characters are valid in db file path but invalid and hence encoded in URL
	connector, err := duckdb.NewConnector(generateDSN(dsn.Path, query.Encode()), func(execer driver.ExecerContext) error {
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

	// run init queries and attach motherduck
	opts.DBInitQueries = append(opts.DBInitQueries, fmt.Sprintf("INSTALL 'motherduck'; LOAD 'motherduck'; SET motherduck_token=%s; ATTACH %s;", safeSQLString(opts.Token), safeSQLString(opts.dsn())))
	for _, qry := range opts.DBInitQueries {
		_, err := db.ExecContext(ctx, qry)
		if err != nil {
			return nil, err
		}
	}

	err = otelsql.RegisterDBStatsMetrics(db.DB, otelsql.WithAttributes(opts.OtelAttributes...))
	if err != nil {
		return nil, fmt.Errorf("registering db stats metrics: %w", err)
	}
	return &motherduck{
		db:       db,
		logger:   opts.Logger,
		writeSem: semaphore.NewWeighted(1),
		opts:     opts,
	}, nil
}

// AcquireReadConnection implements DB. In practice this does not enforce that the returned connection is read-only.
func (m *motherduck) AcquireReadConnection(ctx context.Context) (conn *sqlx.Conn, release func() error, err error) {
	conn, err = m.acquireConn(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("acquire read connection failed: %w", err)
	}
	return conn, conn.Close, nil
}

// Close implements DB.
func (m *motherduck) Close() error {
	return m.db.Close()
}

// CreateTableAsSelect implements DB.
func (m *motherduck) CreateTableAsSelect(ctx context.Context, name, query string, opts *CreateTableOptions) (res *TableWriteMetrics, err error) {
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

	m.logger.Debug("create: create motherduck table", zap.String("name", name), zap.Bool("view", opts.View), observability.ZapCtx(ctx))

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
func (m *motherduck) DropTable(ctx context.Context, name string) (resErr error) {
	m.logger.Debug("drop motherduck table", zap.String("name", name), observability.ZapCtx(ctx))
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

func (m *motherduck) dropTableUnsafe(ctx context.Context, name string, conn *sqlx.Conn) (resErr error) {
	var typ string
	tbl, err := m.schemaUsingConn(ctx, "", name, conn)
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
func (m *motherduck) MutateTable(ctx context.Context, name string, initQueries []string, mutateFn func(ctx context.Context, conn *sqlx.Conn) error) (res *TableWriteMetrics, resErr error) {
	ctx, span := tracer.Start(ctx, "MutateTable", trace.WithAttributes(attribute.String("name", name)))
	defer func() {
		if resErr != nil {
			span.SetStatus(codes.Error, resErr.Error())
		}
		span.End()
	}()

	m.logger.Debug("mutate table", zap.String("name", name), observability.ZapCtx(ctx))
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
func (m *motherduck) RenameTable(ctx context.Context, oldName, newName string) (resErr error) {
	ctx, span := tracer.Start(ctx, "RenameTable", trace.WithAttributes(attribute.String("old_name", oldName), attribute.String("new_name", newName)))
	defer func() {
		if resErr != nil {
			span.SetStatus(codes.Error, resErr.Error())
		}
		span.End()
	}()

	m.logger.Debug("rename table", zap.String("from", oldName), zap.String("to", newName), observability.ZapCtx(ctx))
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
	tbl, err := m.Schema(ctx, "", oldName)
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
func (m *motherduck) Schema(ctx context.Context, ilike, name string) ([]*Table, error) {
	conn, err := m.acquireConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()
	return m.schemaUsingConn(ctx, ilike, name, conn)
}

func (m *motherduck) schemaUsingConn(ctx context.Context, ilike, name string, conn *sqlx.Conn) ([]*Table, error) {
	if ilike != "" && name != "" {
		return nil, fmt.Errorf("cannot specify both `ilike` and `name`")
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

	q := fmt.Sprintf(`
		SELECT
			coalesce(t.table_catalog, current_database()) AS "database",
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
		GROUP BY 1, 2, 3
		ORDER BY 1, 2, 3
	`, whereClause)

	var res []*Table
	err := conn.SelectContext(ctx, &res, q, args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Size implements DB.
func (m *motherduck) Size() int64 {
	// todo: What is the size of the motherduck database? How to ignore tables which were not created within this project ?
	// Do we even need to track this for motherduck?
	return 0
}

func (m *motherduck) acquireConn(ctx context.Context) (*sqlx.Conn, error) {
	conn, err := m.db.Connx(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection failed: %w", err)
	}
	_, err = conn.ExecContext(ctx, fmt.Sprintf("USE %s", safeSQLString(m.opts.bareDB())))
	if err != nil {
		return nil, fmt.Errorf("acquire connection failed: %w", err)
	}
	return conn, nil
}
