package duckdb

import (
	"context"
	"database/sql/driver"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/duckdb/transporter"
	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

func init() {
	drivers.Register("duckdb", Driver{name: "duckdb"})
	drivers.Register("motherduck", Driver{name: "motherduck"})
	drivers.RegisterAsConnector("motherduck", Driver{name: "motherduck"})
}

// spec for duckdb as motherduck connector
var spec = drivers.Spec{
	DisplayName: "MotherDuck",
	Description: "Import data from MotherDuck.",
	SourceProperties: []drivers.PropertySchema{
		{
			Key:         "query",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Query",
			Description: "Query to extract data from MotherDuck.",
			Placeholder: "select * from my_db.my_table;",
		},
	},
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:    "token",
			Secret: true,
		},
	},
}

type Driver struct {
	name string
}

func (d Driver) Open(config map[string]any, logger *zap.Logger) (drivers.Connection, error) {
	dsn, ok := config["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("require dsn to open duckdb connection")
	}

	cfg, err := newConfig(dsn)
	if err != nil {
		return nil, err
	}

	// See note in connection struct
	olapSemSize := cfg.PoolSize - 1
	if olapSemSize < 1 {
		olapSemSize = 1
	}

	c := &connection{
		config:       cfg,
		logger:       logger,
		metaSem:      semaphore.NewWeighted(1),
		olapSem:      priorityqueue.NewSemaphore(olapSemSize),
		dbCond:       sync.NewCond(&sync.Mutex{}),
		driverConfig: config,
		driverName:   d.name,
	}

	// Open the DB
	err = c.reopenDB()
	if err != nil {
		return nil, err
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

func (d Driver) Drop(config map[string]any, logger *zap.Logger) error {
	dsn, ok := config["dsn"].(string)
	if !ok {
		return fmt.Errorf("require dsn to drop duckdb connection")
	}

	cfg, err := newConfig(dsn)
	if err != nil {
		return err
	}

	if cfg.DBFilePath != "" {
		err = os.Remove(cfg.DBFilePath)
		if err != nil {
			return err
		}
		// Hacky approach to remove the wal file
		_ = os.Remove(cfg.DBFilePath + ".wal")
	}

	return nil
}

func (d Driver) Spec() drivers.Spec {
	return spec
}

func (d Driver) HasAnonymousSourceAccess(ctx context.Context, src drivers.Source, logger *zap.Logger) (bool, error) {
	return false, nil
}

type connection struct {
	db *sqlx.DB
	// driverConfig is input config passed during Open
	driverConfig map[string]any
	driverName   string
	// config is parsed configs
	config *config
	logger *zap.Logger
	// This driver may issue both OLAP and "meta" queries (like catalog info) against DuckDB.
	// Meta queries are usually fast, but OLAP queries may take a long time. To enable predictable parallel performance,
	// we gate queries with semaphores that limits the number of concurrent queries of each type.
	// The metaSem allows 1 query at a time and the olapSem allows cfg.PoolSize-1 queries at a time.
	// When cfg.PoolSize is 1, we set olapSem to still allow 1 query at a time.
	// This creates contention for the same connection in database/sql's pool, but its locks will handle that.
	metaSem *semaphore.Weighted
	olapSem *priorityqueue.Semaphore
	// If DuckDB encounters a fatal error, all queries will fail until the DB has been reopened.
	// When dbReopen is set to true, dbCond will be used to stop acquisition of new connections,
	// and then when dbConnCount becomes 0, the DB will be reopened and dbReopen set to false again.
	// If the reopen fails, dbErr will be set and all subsequent connection acquires will return it.
	dbConnCount int
	dbCond      *sync.Cond
	dbReopen    bool
	dbErr       error
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return c.driverName
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	return c.driverConfig
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return c.db.Close()
}

// AsRegistry Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore() (drivers.CatalogStore, bool) {
	return c, true
}

// AsRepoStore Repo implements drivers.Connection.
func (c *connection) AsRepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// AsOLAP OLAP implements drivers.Connection.
func (c *connection) AsOLAP() (drivers.OLAPStore, bool) {
	return c, true
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Connection) (drivers.Transporter, bool) {
	olap, _ := to.AsOLAP()
	if c == to {
		if from.Driver() == "motherduck" {
			return transporter.NewMotherduckToDuckDB(from, olap, c.logger), true
		}
		if store, ok := from.AsObjectStore(); ok { // objectstore to duckdb transfer
			return transporter.NewObjectStoreToDuckDB(store, olap, c.logger), true
		}
		if store, ok := from.AsFileStore(); ok {
			return transporter.NewFileStoreToDuckDB(store, olap, c.logger), true
		}
	}
	return nil, false
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// reopenDB opens the DuckDB handle anew. If c.db is already set, it closes the existing handle first.
func (c *connection) reopenDB() error {
	// If c.db is already open, close it first
	if c.db != nil {
		err := c.db.Close()
		if err != nil {
			return err
		}
		c.db = nil
	}

	// Queries to run when a new DuckDB connection is opened.
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
		"SET timezone='UTC'",
	}

	// DuckDB extensions need to be loaded separately on each connection, but the built-in connection pool in database/sql doesn't enable that.
	// So we use go-duckdb's custom connector to pass a callback that it invokes for each new connection.
	connector, err := duckdb.NewConnector(c.config.DSN, func(execer driver.ExecerContext) error {
		for _, qry := range bootQueries {
			_, err := execer.ExecContext(context.Background(), qry, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		// Check for using incompatible database files
		if strings.Contains(err.Error(), "Trying to read a database file with version number") {
			return fmt.Errorf("database file %q was created with an older, incompatible version of Rill (please remove it and try again)", c.config.DSN)
		}

		// Check for another process currently accessing the DB
		if strings.Contains(err.Error(), "Could not set lock on file") {
			return fmt.Errorf("failed to open database (is Rill already running?): %w", err)
		}

		return err
	}

	// Create new DB
	sqlDB := otelsql.OpenDB(connector)
	db := sqlx.NewDb(sqlDB, "duckdb")
	db.SetMaxOpenConns(c.config.PoolSize)
	c.db = db

	return nil
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
// acquireConn implements the connection tracking and DB reopening logic described in the struct definition for connection.
func (c *connection) acquireConn(ctx context.Context) (*sqlx.Conn, func() error, error) {
	c.dbCond.L.Lock()
	for {
		if c.dbErr != nil {
			c.dbCond.L.Unlock()
			return nil, nil, c.dbErr
		}
		if !c.dbReopen {
			break
		}
		c.dbCond.Wait()
	}

	c.dbConnCount++
	c.dbCond.L.Unlock()

	conn, err := c.db.Connx(ctx)
	if err != nil {
		return nil, nil, err
	}

	release := func() error {
		err := conn.Close()
		c.dbCond.L.Lock()
		c.dbConnCount--
		if c.dbConnCount == 0 && c.dbReopen {
			c.dbReopen = false
			err = c.reopenDB()
			if err == nil {
				c.logger.Info("reopened DuckDB successfully")
			} else {
				c.logger.Error("reopen of DuckDB failed - the handle is now permanently locked", zap.Error(err))
			}
			c.dbErr = err
			c.dbCond.Broadcast()
		}
		c.dbCond.L.Unlock()
		return err
	}

	return conn, release, nil
}

// checkErr marks the DB for reopening if the error is an internal DuckDB error.
// In all other cases, it just proxies the err.
// It should be wrapped around errors returned from DuckDB queries. **It must be called while still holding an acquired DuckDB connection.**
func (c *connection) checkErr(err error) error {
	if err != nil {
		if strings.HasPrefix(err.Error(), "INTERNAL Error:") || strings.HasPrefix(err.Error(), "FATAL Error") {
			c.dbCond.L.Lock()
			defer c.dbCond.L.Unlock()
			c.dbReopen = true
			c.logger.Error("encountered internal DuckDB error - scheduling reopen of DuckDB", zap.Error(err))
		}
	}
	return err
}
