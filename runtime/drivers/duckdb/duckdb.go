package duckdb

import (
	"context"
	"database/sql/driver"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/duckdb/transporter"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
	"go.opentelemetry.io/otel/attribute"
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
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
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

func (d Driver) Open(cfgMap map[string]any, shared bool, ac activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("duckdb driver can't be shared")
	}

	cfg, err := newConfig(cfgMap)
	if err != nil {
		return nil, err
	}

	// See note in connection struct
	olapSemSize := cfg.PoolSize - 1
	if olapSemSize < 1 {
		olapSemSize = 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &connection{
		config:         cfg,
		logger:         logger,
		activity:       ac,
		metaSem:        semaphore.NewWeighted(1),
		olapSem:        priorityqueue.NewSemaphore(olapSemSize),
		longRunningSem: semaphore.NewWeighted(1), // Currently hard-coded to 1
		dbCond:         sync.NewCond(&sync.Mutex{}),
		driverConfig:   cfgMap,
		driverName:     d.name,
		shared:         shared,
		ctx:            ctx,
		cancel:         cancel,
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

	go c.periodicallyEmitStats(time.Minute)

	return c, nil
}

func (d Driver) Drop(cfgMap map[string]any, logger *zap.Logger) error {
	cfg, err := newConfig(cfgMap)
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

func (d Driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

type connection struct {
	db *sqlx.DB
	// driverConfig is input config passed during Open
	driverConfig map[string]any
	driverName   string
	// config is parsed configs
	config   *config
	logger   *zap.Logger
	activity activity.Client
	// This driver may issue both OLAP and "meta" queries (like catalog info) against DuckDB.
	// Meta queries are usually fast, but OLAP queries may take a long time. To enable predictable parallel performance,
	// we gate queries with semaphores that limits the number of concurrent queries of each type.
	// The metaSem allows 1 query at a time and the olapSem allows cfg.PoolSize-1 queries at a time.
	// When cfg.PoolSize is 1, we set olapSem to still allow 1 query at a time.
	// This creates contention for the same connection in database/sql's pool, but its locks will handle that.
	metaSem *semaphore.Weighted
	olapSem *priorityqueue.Semaphore
	// The OLAP interface additionally provides an option to limit the number of long-running queries, as designated by the caller.
	// longRunningSem enforces this limitation.
	longRunningSem *semaphore.Weighted
	// The OLAP interface also provides an option to acquire a connection "transactionally".
	// We've run into issues with DuckDB freezing up on transactions, so we just use a lock for now to serialize them (inconsistency in case of crashes is acceptable).
	txMu sync.RWMutex
	// If DuckDB encounters a fatal error, all queries will fail until the DB has been reopened.
	// When dbReopen is set to true, dbCond will be used to stop acquisition of new connections,
	// and then when dbConnCount becomes 0, the DB will be reopened and dbReopen set to false again.
	// If the reopen fails, dbErr will be set and all subsequent connection acquires will return it.
	dbConnCount int
	dbCond      *sync.Cond
	dbReopen    bool
	dbErr       error
	shared      bool
	// Cancellable context to control internal processes like emitting the stats
	ctx    context.Context
	cancel context.CancelFunc
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
	c.cancel()
	return c.db.Close()
}

// AsRegistry Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	if c.shared {
		// duckdb catalog is instance specific
		return nil, false
	}
	return c, true
}

// AsRepoStore Repo implements drivers.Connection.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsOLAP OLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	if c.shared {
		// duckdb olap is instance specific
		return nil, false
	}
	return c, true
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Connection.
// Use OLAPStore instead.
func (c *connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	olap, _ := to.AsOLAP("") // if c == to, connection is instance specific
	if c == to {
		if from == to {
			return transporter.NewDuckDBToDuckDB(olap, c.logger), true
		}
		if from.Driver() == "motherduck" {
			return transporter.NewMotherduckToDuckDB(from, olap, c.logger), true
		}
		if store, ok := from.AsSQLStore(); ok {
			return transporter.NewSQLStoreToDuckDB(store, olap, c.logger), true
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

	// We want to set preserve_insertion_order=false in hosted environments only (where source data is never viewed directly). Setting it reduces batch data ingestion time by ~40%.
	// Hack: Using AllowHostAccess as a proxy indicator for a hosted environment.
	if !c.config.AllowHostAccess {
		bootQueries = append(bootQueries, "SET preserve_insertion_order TO false")
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
func (c *connection) acquireOLAPConn(ctx context.Context, priority int, longRunning, tx bool) (*sqlx.Conn, func() error, error) {
	// Try to get conn from context (means the call is wrapped in WithConnection)
	conn := connFromContext(ctx)
	if conn != nil {
		return conn, func() error { return nil }, nil
	}

	// Acquire long-running semaphore if applicable
	if longRunning {
		err := c.longRunningSem.Acquire(ctx, 1)
		if err != nil {
			return nil, nil, err
		}
	}

	// Acquire semaphore
	err := c.olapSem.Acquire(ctx, priority)
	if err != nil {
		if longRunning {
			c.longRunningSem.Release(1)
		}
		return nil, nil, err
	}

	// Poor man's transaction support – see struct docstring for details
	if tx {
		c.txMu.Lock()
	} else {
		c.txMu.RLock()
	}

	// Get new conn
	conn, releaseConn, err := c.acquireConn(ctx)
	if err != nil {
		if tx {
			c.txMu.Unlock()
		} else {
			c.txMu.RUnlock()
		}
		c.olapSem.Release()
		if longRunning {
			c.longRunningSem.Release(1)
		}
		return nil, nil, err
	}

	// Build release func
	release := func() error {
		err := releaseConn()
		if tx {
			c.txMu.Unlock()
		} else {
			c.txMu.RUnlock()
		}
		c.olapSem.Release()
		if longRunning {
			c.longRunningSem.Release(1)
		}
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

// Periodically collects stats using pragma_database_size() and emits as activity events
func (c *connection) periodicallyEmitStats(d time.Duration) {
	if c.activity == nil {
		// Activity client isn't set, there is no need to report stats
		return
	}

	statTicker := time.NewTicker(d)
	for {
		select {
		case <-statTicker.C:
			var stat dbStat
			// Obtain a connection, query, release
			err := func() error {
				conn, release, err := c.acquireMetaConn(c.ctx)
				if err != nil {
					return err
				}
				defer func() { _ = release() }()
				err = conn.GetContext(c.ctx, &stat, "CALL pragma_database_size()")
				return err
			}()
			if err != nil {
				c.logger.Error("couldn't query DuckDB stats", zap.Error(err))
				continue
			}

			// Emit collected stats as activity events
			commonDims := []attribute.KeyValue{
				attribute.String("duckdb.name", stat.DatabaseName),
			}

			dbSize, err := humanReadableSizeToBytes(stat.DatabaseSize)
			if err != nil {
				c.logger.Error("couldn't convert duckdb size to bytes", zap.Error(err))
			} else {
				c.activity.Emit(c.ctx, "duckdb_size_bytes", dbSize, commonDims...)
			}

			walSize, err := humanReadableSizeToBytes(stat.WalSize)
			if err != nil {
				c.logger.Error("couldn't convert duckdb wal size to bytes", zap.Error(err))
			} else {
				c.activity.Emit(c.ctx, "duckdb_wal_size_bytes", walSize, commonDims...)
			}

			memoryUsage, err := humanReadableSizeToBytes(stat.MemoryUsage)
			if err != nil {
				c.logger.Error("couldn't convert duckdb memory usage to bytes", zap.Error(err))
			} else {
				c.activity.Emit(c.ctx, "duckdb_memory_usage_bytes", memoryUsage, commonDims...)
			}

			memoryLimit, err := humanReadableSizeToBytes(stat.MemoryLimit)
			if err != nil {
				c.logger.Error("couldn't convert duckdb memory limit to bytes", zap.Error(err))
			} else {
				c.activity.Emit(c.ctx, "duckdb_memory_limit_bytes", memoryLimit, commonDims...)
			}

			c.activity.Emit(c.ctx, "duckdb_block_size_bytes", float64(stat.BlockSize), commonDims...)
			c.activity.Emit(c.ctx, "duckdb_total_blocks", float64(stat.TotalBlocks), commonDims...)
			c.activity.Emit(c.ctx, "duckdb_free_blocks", float64(stat.FreeBlocks), commonDims...)
			c.activity.Emit(c.ctx, "duckdb_used_blocks", float64(stat.UsedBlocks), commonDims...)

			estimatedDBSize, _ := c.EstimateSize()
			c.activity.Emit(c.ctx, "duckdb_estimated_size_bytes", float64(estimatedDBSize))

		case <-c.ctx.Done():
			statTicker.Stop()
			return
		}
	}
}

// Regex to parse human-readable size returned by DuckDB
var humanReadableSizeRegex = regexp.MustCompile(`^([\d.]+)\s*(\S+)$`)

// Reversed logic of StringUtil::BytesToHumanReadableString
// see https://github.com/cran/duckdb/blob/master/src/duckdb/src/common/string_util.cpp#L157
// Examples: 1 bytes, 2 bytes, 1KB, 1MB, 1TB, 1PB
func humanReadableSizeToBytes(sizeStr string) (float64, error) {
	var multiplier float64

	match := humanReadableSizeRegex.FindStringSubmatch(sizeStr)

	if match == nil {
		return 0, fmt.Errorf("invalid size format: '%s'", sizeStr)
	}

	sizeFloat, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}

	switch match[2] {
	case "byte", "bytes":
		multiplier = 1
	case "KB":
		multiplier = 1000
	case "MB":
		multiplier = 1000 * 1000
	case "GB":
		multiplier = 1000 * 1000 * 1000
	case "TB":
		multiplier = 1000 * 1000 * 1000 * 1000
	case "PB":
		multiplier = 1000 * 1000 * 1000 * 1000 * 1000
	default:
		return 0, fmt.Errorf("unknown size unit '%s' in '%s'", match[2], sizeStr)
	}

	return sizeFloat * multiplier, nil
}

type dbStat struct {
	DatabaseName string `db:"database_name"`
	DatabaseSize string `db:"database_size"`
	BlockSize    int64  `db:"block_size"`
	TotalBlocks  int64  `db:"total_blocks"`
	UsedBlocks   int64  `db:"used_blocks"`
	FreeBlocks   int64  `db:"free_blocks"`
	WalSize      string `db:"wal_size"`
	MemoryUsage  string `db:"memory_usage"`
	MemoryLimit  string `db:"memory_limit"`
}
