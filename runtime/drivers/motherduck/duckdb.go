package motherduck

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"strings"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

func init() {
	drivers.Register("motherduck", Driver{})
}

type Driver struct{}

func (d Driver) Open(dsn string, logger *zap.Logger) (drivers.Connection, error) {
	cfg, err := newConfig(dsn)
	if err != nil {
		return nil, err
	}

	bootQueries := []string{
		"INSTALL 'json'",
		"LOAD 'json'",
		"INSTALL 'icu'",
		"LOAD 'icu'",
		"INSTALL 'parquet'",
		"LOAD 'parquet'",
		"INSTALL 'httpfs'",
		"LOAD 'httpfs'",
		"SET max_expression_depth TO 250",
	}

	// DuckDB extensions need to be loaded separately on each connection, but the built-in connection pool in database/sql doesn't enable that.
	// So we use go-duckdb's custom connector to pass a callback that it invokes for each new connection.
	// nolint:staticcheck // TODO: remove when go-duckdb implements the driver.ExecerContext interface
	connector, err := duckdb.NewConnector(cfg.DSN, func(execer driver.ExecerContext) error {
		for _, qry := range bootQueries {
			_, err = execer.ExecContext(context.Background(), qry, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		// Adding check if using old database files with upgraded version
		if strings.Contains(err.Error(), "Trying to read a database file with version number") {
			return nil, fmt.Errorf("database file %q was created with an older, incompatible version of Rill (please remove it and try again)", cfg.DSN)
		}

		if strings.Contains(err.Error(), "Could not set lock on file") {
			return nil, fmt.Errorf("failed to open database (is Rill already running?): %w", err)
		}

		return nil, err
	}

	// This driver may issue both OLAP and "meta" queries (like catalog info) against DuckDB.
	// Meta queries are usually fast, but OLAP queries may take a long time. To enable predictable parallel performance,
	// we gate queries with semaphores that limits the number of concurrent queries of each type.
	// The metaSem allows 1 query at a time and the olapSem allows cfg.PoolSize-1 queries at a time.
	//
	// When cfg.PoolSize is 1, we set olapSem to still allow 1 query at a time.
	// This creates contention for the same connection in database/sql's pool, but its locks will handle that.

	sqlDB := otelsql.OpenDB(connector)
	db := sqlx.NewDb(sqlDB, "duckdb")
	db.SetMaxOpenConns(cfg.PoolSize)

	// We want to use all except one connection for OLAP queries.
	olapSemSize := cfg.PoolSize - 1
	if olapSemSize < 1 {
		olapSemSize = 1
	}

	c := &connection{
		db:      db,
		metaSem: semaphore.NewWeighted(1),
		olapSem: priorityqueue.NewSemaphore(olapSemSize),
		logger:  logger,
		config:  cfg,
	}

	// Return nice error for old macOS versions
	conn, err := c.db.Connx(context.Background())
	if err != nil && strings.Contains(err.Error(), "Symbol not found") {
		fmt.Printf("Your version of macOS is not supported. Please upgrade to the latest major release of macOS. See this link for details: https://support.apple.com/en-in/macos/upgrade")
		os.Exit(1)
	} else if err == nil {
		conn.Close()
	} else {
		return nil, err
	}

	return c, nil
}

func (d Driver) Drop(dsn string, logger *zap.Logger) error {
	cfg, err := newConfig(dsn)
	if err != nil {
		return err
	}

	db, err := sql.Open("duckdb", "md:")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %q", cfg.DB))
	return err
}

type connection struct {
	db     *sqlx.DB
	logger *zap.Logger
	// metaSem gates meta queries (like catalog and information schema)
	metaSem *semaphore.Weighted
	// olapSem gates OLAP queries
	olapSem *priorityqueue.Semaphore
	config  *config
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return c.db.Close()
}

// RegistryStore Registry implements drivers.Connection.
func (c *connection) RegistryStore() (drivers.RegistryStore, bool) {
	return nil, false
}

// CatalogStore Catalog implements drivers.Connection.
func (c *connection) CatalogStore() (drivers.CatalogStore, bool) {
	return c, true
}

// RepoStore Repo implements drivers.Connection.
func (c *connection) RepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAPStore OLAP implements drivers.Connection.
func (c *connection) OLAPStore() (drivers.OLAPStore, bool) {
	return c, true
}

// acquireMetaConn gets a connection from the pool for "meta" queries like catalog and information schema (i.e. fast queries).
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
	conn, err = c.db.Connx(ctx)
	if err != nil {
		c.metaSem.Release(1)
		return nil, nil, err
	}

	// Build release func
	release := func() error {
		err := conn.Close()
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
	conn, err = c.db.Connx(ctx)
	if err != nil {
		c.olapSem.Release()
		return nil, nil, err
	}

	// Build release func
	release := func() error {
		err := conn.Close()
		c.olapSem.Release()
		return err
	}

	return conn, release, nil
}
