package duckdb

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/c2h5oh/datasize"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

func init() {
	drivers.Register("duckdb", Driver{name: "duckdb"})
	drivers.Register("motherduck", Driver{name: "motherduck"})
	drivers.RegisterAsConnector("duckdb", Driver{name: "duckdb"})
	drivers.RegisterAsConnector("motherduck", Driver{name: "motherduck"})
}

var spec = drivers.Spec{
	DisplayName: "DuckDB",
	Description: "DuckDB SQL connector.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/motherduck",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:  "path",
			Type: drivers.StringPropertyType,
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "DuckDB SQL query.",
			Placeholder: "select * from read_csv('data/file.csv', header=true);",
		},
		{
			Key:         "db",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "DB",
			Description: "Path to external DuckDB database. Use md:<dbname> for motherduckb.",
			Placeholder: "/path/to/main.db or md:main.db(for motherduck)",
		},
	},
	ImplementsCatalog: true,
	ImplementsOLAP:    true,
}

var motherduckSpec = drivers.Spec{
	DisplayName: "MotherDuck",
	Description: "MotherDuck SQL connector.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/motherduck",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "token",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
	},
	ImplementsOLAP: true,
}

type Driver struct {
	name string
}

func (d Driver) Open(instanceID string, cfgMap map[string]any, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("duckdb driver can't be shared")
	}

	cfg, err := newConfig(cfgMap)
	if err != nil {
		return nil, err
	}
	logger.Debug("opening duckdb handle", zap.String("dsn", cfg.DSN))

	// We've seen the DuckDB .wal and .tmp files grow to 100s of GBs in some cases.
	// This prevents recovery after restarts since DuckDB hangs while trying to reprocess the files.
	// This is a hacky solution that deletes the files (if they exist) before re-opening the DB.
	// Generally, this should not lead to data loss since reconcile will bring the database back to the correct state.
	if cfg.DBFilePath != "" {
		// Always drop the .tmp directory
		tmpPath := cfg.DBFilePath + ".tmp"
		_ = os.RemoveAll(tmpPath)

		// Drop the .wal file if it's bigger than 100MB
		walPath := cfg.DBFilePath + ".wal"
		if stat, err := os.Stat(walPath); err == nil {
			if stat.Size() >= 100*int64(datasize.MB) {
				_ = os.Remove(walPath)
			}
		}
	}

	if cfg.DBStoragePath != "" {
		if err := os.MkdirAll(cfg.DBStoragePath, fs.ModePerm); err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
	}

	// See note in connection struct
	olapSemSize := cfg.PoolSize - 1
	if olapSemSize < 1 {
		olapSemSize = 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &connection{
		instanceID:     instanceID,
		config:         cfg,
		logger:         logger,
		activity:       ac,
		metaSem:        semaphore.NewWeighted(1),
		olapSem:        priorityqueue.NewSemaphore(olapSemSize),
		longRunningSem: semaphore.NewWeighted(1), // Currently hard-coded to 1
		dbCond:         sync.NewCond(&sync.Mutex{}),
		driverConfig:   cfgMap,
		driverName:     d.name,
		connTimes:      make(map[int]time.Time),
		ctx:            ctx,
		cancel:         cancel,
	}

	// register a callback to add a gauge on number of connections in use per db
	attrs := []attribute.KeyValue{attribute.String("db", c.config.DBFilePath)}
	c.registration = observability.Must(meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		observer.ObserveInt64(connectionsInUse, int64(c.dbConnCount), metric.WithAttributes(attrs...))
		return nil
	}, connectionsInUse))

	// Open the DB
	err = c.reopenDB()
	if err != nil {
		if c.config.ErrorOnIncompatibleVersion || !strings.Contains(err.Error(), "created with an older, incompatible version of Rill") {
			return nil, err
		}

		c.logger.Debug("Resetting .db file because it was created with an older, incompatible version of Rill")

		tmpPath := cfg.DBFilePath + ".tmp"
		_ = os.RemoveAll(tmpPath)
		walPath := cfg.DBFilePath + ".wal"
		_ = os.Remove(walPath)
		_ = os.Remove(cfg.DBFilePath)

		// reopen connection again
		if err := c.reopenDB(); err != nil {
			return nil, err
		}
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

	go c.periodicallyCheckConnDurations(time.Minute)

	return c, nil
}

func (d Driver) Drop(cfgMap map[string]any, logger *zap.Logger) error {
	cfg, err := newConfig(cfgMap)
	if err != nil {
		return err
	}
	if cfg.DBStoragePath != "" {
		return os.RemoveAll(cfg.DBStoragePath)
	}
	if cfg.DBFilePath != "" {
		err = os.Remove(cfg.DBFilePath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		// Hacky approach to remove the wal file
		_ = os.Remove(cfg.DBFilePath + ".wal")
		// also temove the temp dir
		_ = os.RemoveAll(cfg.DBFilePath + ".tmp")
	}

	return nil
}

func (d Driver) Spec() drivers.Spec {
	if d.name == "motherduck" {
		return motherduckSpec
	}
	return spec
}

func (d Driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d Driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	// The "sql" property of a DuckDB source can reference other connectors like S3.
	// We try to extract those and return them here.
	// We will in most error cases just return nil and let errors be handled during source ingestion.

	sql, ok := src["sql"].(string)
	if !ok {
		return nil, nil
	}

	ast, err := duckdbsql.Parse(sql)
	if err != nil {
		return nil, nil
	}

	res := make([]string, 0)

	refs := ast.GetTableRefs()
	for _, ref := range refs {
		if len(ref.Paths) == 0 {
			continue
		}

		uri, err := url.Parse(ref.Paths[0])
		if err != nil {
			return nil, err
		}

		switch uri.Scheme {
		case "s3", "azure":
			res = append(res, uri.Scheme)
		case "gs":
			res = append(res, "gcs")
		default:
			// Ignore
		}
	}

	return res, nil
}

type connection struct {
	instanceID string
	db         *sqlx.DB
	// driverConfig is input config passed during Open
	driverConfig map[string]any
	driverName   string
	// config is parsed configs
	config   *config
	logger   *zap.Logger
	activity *activity.Client
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
	// State for maintaining connection acquire times, which enables periodically checking for hanging DuckDB queries (we have previously seen deadlocks in DuckDB).
	connTimesMu sync.Mutex
	nextConnID  int
	connTimes   map[int]time.Time
	// Cancellable context to control internal processes like emitting the stats
	ctx    context.Context
	cancel context.CancelFunc
	// registration should be unregistered on close
	registration metric.Registration
}

var _ drivers.OLAPStore = &connection{}

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
	_ = c.registration.Unregister()
	return c.db.Close()
}

// AsRegistry Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return c, true
}

// AsRepoStore Repo implements drivers.Connection.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP OLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
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

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor() (drivers.ModelExecutor, bool) {
	return c, true
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	olap, _ := to.(*connection)
	if c == to {
		if from == to {
			return NewDuckDBToDuckDB(olap, c.logger), true
		}
		if from.Driver() == "motherduck" {
			return NewMotherduckToDuckDB(from, olap, c.logger), true
		}
		if store, ok := from.AsSQLStore(); ok {
			return NewSQLStoreToDuckDB(store, olap, c.logger), true
		}
		if store, ok := from.AsObjectStore(); ok { // objectstore to duckdb transfer
			return NewObjectStoreToDuckDB(store, olap, c.logger), true
		}
		if store, ok := from.AsFileStore(); ok {
			return NewFileStoreToDuckDB(store, olap, c.logger), true
		}
	}
	return nil, false
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
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
	var bootQueries []string

	// Add custom boot queries before any other (e.g. to override the extensions repository)
	if c.config.BootQueries != "" {
		bootQueries = append(bootQueries, c.config.BootQueries)
	}

	// Add required boot queries
	bootQueries = append(bootQueries,
		"INSTALL 'json'",
		"LOAD 'json'",
		"INSTALL 'icu'",
		"LOAD 'icu'",
		"INSTALL 'parquet'",
		"LOAD 'parquet'",
		"INSTALL 'httpfs'",
		"LOAD 'httpfs'",
		"INSTALL 'sqlite'",
		"LOAD 'sqlite'",
		"SET max_expression_depth TO 250",
		"SET timezone='UTC'",
		"SET old_implicit_casting = true", // Implicit Cast to VARCHAR
	)

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
			if err != nil && strings.Contains(err.Error(), "Failed to download extension") {
				// Retry using another mirror. Based on: https://github.com/duckdb/duckdb/issues/9378
				_, err = execer.ExecContext(context.Background(), qry+" FROM 'http://nightly-extensions.duckdb.org'", nil)
			}
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

	if !c.config.ExtTableStorage {
		return nil
	}

	conn, err := db.Connx(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()

	c.logLimits(conn)

	// 2023-12-11: Hail mary for solving this issue: https://github.com/duckdblabs/rilldata/issues/6.
	// Forces DuckDB to create catalog entries for the information schema up front (they are normally created lazily).
	// Can be removed if the issue persists.
	_, err = conn.ExecContext(context.Background(), `
		select
			coalesce(t.table_catalog, current_database()) as "database",
			t.table_schema as "schema",
			t.table_name as "name",
			t.table_type as "type", 
			array_agg(c.column_name order by c.ordinal_position) as "column_names",
			array_agg(c.data_type order by c.ordinal_position) as "column_types",
			array_agg(c.is_nullable = 'YES' order by c.ordinal_position) as "column_nullable"
		from information_schema.tables t
		join information_schema.columns c on t.table_schema = c.table_schema and t.table_name = c.table_name
		group by 1, 2, 3, 4
		order by 1, 2, 3, 4
	`)
	if err != nil {
		return err
	}

	// List the directories directly in the external storage directory
	// Load the version.txt from each sub-directory
	// If version.txt is found, attach only the .db file matching the version.txt.
	// If attach fails, log the error and delete the version.txt and .db file (e.g. might be DuckDB version change)
	entries, err := os.ReadDir(c.config.DBStoragePath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(c.config.DBStoragePath, entry.Name())
		version, exist, err := c.tableVersion(entry.Name())
		if err != nil {
			c.logger.Error("error in fetching db version", zap.String("table", entry.Name()), zap.Error(err))
			_ = os.RemoveAll(path)
			continue
		}
		if !exist {
			_ = os.RemoveAll(path)
			continue
		}

		dbFile := filepath.Join(path, fmt.Sprintf("%s.db", version))
		db := dbName(entry.Name(), version)
		_, err = conn.ExecContext(context.Background(), fmt.Sprintf("ATTACH %s AS %s", safeSQLString(dbFile), safeSQLName(db)))
		if err != nil {
			c.logger.Error("attach failed clearing db file", zap.String("db", dbFile), zap.Error(err))
			_, _ = conn.ExecContext(context.Background(), fmt.Sprintf("DROP VIEW IF EXISTS %s", safeSQLName(entry.Name())))
			_ = os.RemoveAll(path)
		}
	}
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
	conn, releaseConn, err := c.acquireConn(ctx, false)
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

	// Get new conn
	conn, releaseConn, err := c.acquireConn(ctx, tx)
	if err != nil {
		c.olapSem.Release()
		if longRunning {
			c.longRunningSem.Release(1)
		}
		return nil, nil, err
	}

	// Build release func
	release := func() error {
		err := releaseConn()
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
func (c *connection) acquireConn(ctx context.Context, tx bool) (*sqlx.Conn, func() error, error) {
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

	// Poor man's transaction support – see struct docstring for details.
	if tx {
		c.txMu.Lock()

		// When tx is true, and the database is backed by a file, we reopen the database to ensure only one DuckDB connection is open.
		// This avoids the following issue: https://github.com/duckdb/duckdb/issues/9150
		if c.config.DBFilePath != "" {
			err := c.reopenDB()
			if err != nil {
				c.txMu.Unlock()
				return nil, nil, err
			}
		}
	} else {
		c.txMu.RLock()
	}
	releaseTx := func() {
		if tx {
			c.txMu.Unlock()
		} else {
			c.txMu.RUnlock()
		}
	}

	conn, err := c.db.Connx(ctx)
	if err != nil {
		releaseTx()
		return nil, nil, err
	}

	c.connTimesMu.Lock()
	connID := c.nextConnID
	c.nextConnID++
	c.connTimes[connID] = time.Now()
	c.connTimesMu.Unlock()

	release := func() error {
		err := conn.Close()
		c.connTimesMu.Lock()
		delete(c.connTimes, connID)
		c.connTimesMu.Unlock()
		releaseTx()
		c.dbCond.L.Lock()
		c.dbConnCount--
		if c.dbConnCount == 0 && c.dbReopen {
			c.dbReopen = false
			err = c.reopenDB()
			if err == nil {
				c.logger.Debug("reopened DuckDB successfully")
			} else {
				c.logger.Debug("reopen of DuckDB failed - the handle is now permanently locked", zap.Error(err))
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
// nolint
func (c *connection) periodicallyEmitStats(d time.Duration) {
	if c.activity == nil {
		// Activity client isn't set, there is no need to report stats
		return
	}

	statTicker := time.NewTicker(d)
	for {
		select {
		case <-statTicker.C:
			estimatedDBSize, _ := c.EstimateSize()
			c.activity.RecordMetric(c.ctx, "duckdb_estimated_size_bytes", float64(estimatedDBSize))

			// NOTE :: running CALL pragma_database_size() while duckdb is ingesting data is causing the WAL file to explode.
			// Commenting the below code for now. Verify with next duckdb release

			// // Motherduck driver doesn't provide pragma stats
			// if c.driverName == "motherduck" {
			// 	continue
			// }

			// var stat dbStat
			// // Obtain a connection, query, release
			// err := func() error {
			// 	conn, release, err := c.acquireMetaConn(c.ctx)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	defer func() { _ = release() }()
			// 	err = conn.GetContext(c.ctx, &stat, "CALL pragma_database_size()")
			// 	return err
			// }()
			// if err != nil {
			// 	c.logger.Error("couldn't query DuckDB stats", zap.Error(err))
			// 	continue
			// }

			// // Emit collected stats as activity events
			// commonDims := []attribute.KeyValue{
			// 	attribute.String("duckdb.name", stat.DatabaseName),
			// }

			// dbSize, err := humanReadableSizeToBytes(stat.DatabaseSize)
			// if err != nil {
			// 	c.logger.Error("couldn't convert duckdb size to bytes", zap.Error(err))
			// } else {
			// 	c.activity.RecordMetric(c.ctx, "duckdb_size_bytes", dbSize, commonDims...)
			// }

			// walSize, err := humanReadableSizeToBytes(stat.WalSize)
			// if err != nil {
			// 	c.logger.Error("couldn't convert duckdb wal size to bytes", zap.Error(err))
			// } else {
			// 	c.activity.RecordMetric(c.ctx, "duckdb_wal_size_bytes", walSize, commonDims...)
			// }

			// memoryUsage, err := humanReadableSizeToBytes(stat.MemoryUsage)
			// if err != nil {
			// 	c.logger.Error("couldn't convert duckdb memory usage to bytes", zap.Error(err))
			// } else {
			// 	c.activity.RecordMetric(c.ctx, "duckdb_memory_usage_bytes", memoryUsage, commonDims...)
			// }

			// memoryLimit, err := humanReadableSizeToBytes(stat.MemoryLimit)
			// if err != nil {
			// 	c.logger.Error("couldn't convert duckdb memory limit to bytes", zap.Error(err))
			// } else {
			// 	c.activity.RecordMetric(c.ctx, "duckdb_memory_limit_bytes", memoryLimit, commonDims...)
			// }

			// c.activity.RecordMetric(c.ctx, "duckdb_block_size_bytes", float64(stat.BlockSize), commonDims...)
			// c.activity.RecordMetric(c.ctx, "duckdb_total_blocks", float64(stat.TotalBlocks), commonDims...)
			// c.activity.RecordMetric(c.ctx, "duckdb_free_blocks", float64(stat.FreeBlocks), commonDims...)
			// c.activity.RecordMetric(c.ctx, "duckdb_used_blocks", float64(stat.UsedBlocks), commonDims...)

		case <-c.ctx.Done():
			statTicker.Stop()
			return
		}
	}
}

// maxAcquiredConnDuration is the maximum duration a connection can be held for before we consider it potentially hanging/deadlocked.
const maxAcquiredConnDuration = 1 * time.Hour

// periodicallyCheckConnDurations periodically checks the durations of all acquired connections and logs a warning if any have been held for longer than maxAcquiredConnDuration.
func (c *connection) periodicallyCheckConnDurations(d time.Duration) {
	connDurationTicker := time.NewTicker(d)
	defer connDurationTicker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-connDurationTicker.C:
			c.connTimesMu.Lock()
			for connID, connTime := range c.connTimes {
				if time.Since(connTime) > maxAcquiredConnDuration {
					c.logger.Error("duckdb: a connection has been held for longer than the maximum allowed duration", zap.Int("conn_id", connID), zap.Duration("duration", time.Since(connTime)))
				}
			}
			c.connTimesMu.Unlock()
		}
	}
}

func (c *connection) logLimits(conn *sqlx.Conn) {
	row := conn.QueryRowContext(context.Background(), "SELECT value FROM duckdb_settings() WHERE name='max_memory'")
	var memory string
	_ = row.Scan(&memory)

	row = conn.QueryRowContext(context.Background(), "SELECT value FROM duckdb_settings() WHERE name='threads'")
	var threads string
	_ = row.Scan(&threads)

	c.logger.Debug("duckdb limits", zap.String("memory", memory), zap.String("threads", threads))
}

// fatalInternalError logs a critical internal error and exits the process.
// This is used for errors that are completely unrecoverable.
// Ideally, we should refactor to cleanup/reopen/rebuild so that we don't need this.
func (c *connection) fatalInternalError(err error) {
	c.logger.Fatal("duckdb: critical internal error", zap.Error(err))
}

// Regex to parse human-readable size returned by DuckDB
// nolint
var humanReadableSizeRegex = regexp.MustCompile(`^([\d.]+)\s*(\S+)$`)

// Reversed logic of StringUtil::BytesToHumanReadableString
// see https://github.com/cran/duckdb/blob/master/src/duckdb/src/common/string_util.cpp#L157
// Examples: 1 bytes, 2 bytes, 1KB, 1MB, 1TB, 1PB
// nolint
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

// nolint
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
