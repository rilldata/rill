package duckdb

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/duckdb/extensions"
	"github.com/rilldata/rill/runtime/drivers/file"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
	"github.com/rilldata/rill/runtime/pkg/rduckdb"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"gocloud.dev/blob"
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
	DocsURL:     "https://docs.rilldata.com/build/connectors/olap/duckdb",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Path",
			Description: "Path to external DuckDB database.",
			Placeholder: "/path/to/main.db",
		},
		{
			Key:         "attach",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Attach",
			Description: "Attach to an existing DuckDB database. This is an alternative to `path` that supports attach options.",
			Placeholder: "'ducklake:metadata.ducklake' AS my_ducklake(DATA_PATH 'datafiles')",
		},
		{
			Key:         "mode",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Mode",
			Description: "Set the mode for the DuckDB connection. By default, it is set to 'read' which allows only read operations. Set to 'readwrite' to enable model creation and table mutations.",
			Placeholder: modeReadOnly,
			Default:     modeReadOnly,
			NoPrompt:    true,
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to run on DuckDB.",
			Placeholder: "select * from table;",
		},
	},
	ImplementsOLAP: true,
}

var motherduckSpec = drivers.Spec{
	DisplayName: "MotherDuck",
	Description: "MotherDuck SQL connector.",
	DocsURL:     "https://docs.rilldata.com/build/connectors/olap/motherduck",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Path",
			Description: "Path to Motherduck database. Must be prefixed with `md:`",
			Placeholder: "md:my_db",
		},
		{
			Key:         "token",
			Type:        drivers.StringPropertyType,
			Secret:      true,
			Required:    true,
			DisplayName: "Token",
			Description: "MotherDuck token",
			Placeholder: "your_motherduck_token",
		},
		{
			Key:         "mode",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Mode",
			Description: "Set the mode for the DuckDB connection. By default, it is set to 'read' which allows only read operations. Set to 'readwrite' to enable model creation and table mutations.",
			Placeholder: modeReadOnly,
			Default:     modeReadOnly,
			NoPrompt:    true,
		},
		{
			Key:         "schema_name",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Schema name",
			Placeholder: "main",
			Hint:        "Set the default schema used by the MotherDuck database",
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to extract data from MotherDuck.",
			Placeholder: "select * from table;",
		},
	},
	ImplementsOLAP: true,
}

type Driver struct {
	name string
}

func (d Driver) Open(instanceID string, cfgMap map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("duckdb driver can't be shared")
	}

	err := extensions.InstallExtensionsOnce()
	if err != nil {
		logger.Warn("failed to install embedded DuckDB extensions, let DuckDB download them", zap.Error(err))
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

	// Open remote bucket for backups if configured
	var remote *blob.Bucket
	if cfg.EnableBackups {
		b, ok, err := st.OpenBucket(context.Background())
		if err != nil {
			return nil, err
		}
		if ok {
			remote = b
		}
	}

	// Create the handle
	ctx, cancel := context.WithCancel(context.Background())
	c := &connection{
		instanceID:     instanceID,
		config:         cfg,
		logger:         logger,
		activity:       ac,
		storage:        st,
		remote:         remote,
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
	attrs := []attribute.KeyValue{attribute.String("instance_id", instanceID)}
	c.registration = observability.Must(meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		observer.ObserveInt64(connectionsInUse, int64(c.dbConnCount), metric.WithAttributes(attrs...))
		return nil
	}, connectionsInUse))

	// Open the DB
	err = c.reopenDB(context.Background())
	if err != nil {
		if remote != nil {
			_ = remote.Close()
		}
		// Check for another process currently accessing the DB
		if strings.Contains(err.Error(), "Could not set lock on file") {
			return nil, fmt.Errorf("failed to open database (is Rill already running?): %w", err)
		}
		return nil, err
	}

	go c.periodicallyEmitStats(time.Minute)

	go c.periodicallyCheckConnDurations(time.Minute)

	return c, nil
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
	// do not use directly it can also be nil or closed
	// use acquireOLAPConn/acquireMetaConn for select and acquireDB for write queries
	db rduckdb.DB
	// driverConfig is input config passed during Open
	driverConfig map[string]any
	driverName   string
	// config is parsed configs
	config   *config
	logger   *zap.Logger
	activity *activity.Client
	storage  *storage.Client
	remote   *blob.Bucket
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
	// If DuckDB encounters a fatal error, all queries will fail until the DB has been reopened.
	// When dbReopen is set to true, dbCond will be used to stop acquisition of new connections,
	// and then when dbConnCount becomes 0, the DB will be reopened and dbReopen set to false again.
	// If the reopen fails, dbErr will be set and all subsequent connection acquires will return it.
	dbConnCount int
	dbCond      *sync.Cond
	dbReopen    bool
	dbErr       error
	// State for maintaining connection acquire times, which enables periodically checking for hanging DuckDB queries (we have previously seen deadlocks in DuckDB).
	connTimesMu    sync.Mutex
	nextConnID     int
	connTimes      map[int]time.Time
	hangingConnErr error
	// Cancellable context to control internal processes like emitting the stats
	ctx    context.Context
	cancel context.CancelFunc
	// registration should be unregistered on close
	registration metric.Registration
}

var _ drivers.OLAPStore = &connection{}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	conn, rel, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	err = conn.PingContext(ctx)
	_ = rel()
	c.connTimesMu.Lock()
	defer c.connTimesMu.Unlock()
	return errors.Join(err, c.hangingConnErr)
}

// Driver implements drivers.Handle.
func (c *connection) Driver() string {
	return c.driverName
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	return maps.Clone(c.driverConfig)
}

// Close implements drivers.Handle.
func (c *connection) Close() error {
	c.cancel()
	_ = c.registration.Unregister()
	if c.remote != nil {
		_ = c.remote.Close()
	}
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// AsRegistry Registry implements drivers.Handle.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore Catalog implements drivers.Handle.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore Repo implements drivers.Handle.
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

// AsOLAP OLAP implements drivers.Handle.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// AsInformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return c, true
}

// AsObjectStore implements drivers.Handle.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	if opts.OutputHandle == c && c.config.Mode != modeReadWrite {
		return nil, fmt.Errorf("model execution is disabled. To enable modeling on this database, set 'mode: readwrite' in your connector configuration. WARNING: This will allow Rill to create and overwrite tables in your database")
	}
	if opts.InputHandle == c && opts.OutputHandle == c {
		return &selfToSelfExecutor{c}, nil
	}
	if opts.OutputHandle == c {
		if w, ok := opts.InputHandle.AsWarehouse(); ok {
			return &warehouseToSelfExecutor{c, w}, nil
		}
		if f, ok := opts.InputHandle.AsFileStore(); ok && opts.InputConnector == "local_file" {
			return &localFileToSelfExecutor{c, f}, nil
		}
		switch opts.InputHandle.Driver() {
		case "mysql", "postgres":
			return &sqlStoreToSelfExecutor{c}, nil
		case "https":
			return &httpsToSelfExecutor{c}, nil
		case "motherduck":
			return &mdToSelfExecutor{c}, nil
		}
		if _, ok := opts.InputHandle.AsObjectStore(); ok {
			return &objectStoreToSelfExecutor{c}, nil
		}
	}
	if opts.InputHandle == c {
		if opts.OutputHandle.Driver() == "file" {
			outputProps := &file.ModelOutputProperties{}
			if err := mapstructure.WeakDecode(opts.PreliminaryOutputProperties, outputProps); err != nil {
				return nil, drivers.ErrNotImplemented
			}
			if supportsExportFormat(outputProps.Format, outputProps.Headers) {
				return &selfToFileExecutor{c}, nil
			}
		}
	}
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	if c.config.Mode != modeReadWrite {
		c.logger.Warn("Model execution is disabled. To enable modeling on this DuckDB database, set 'mode: readwrite' in your connector configuration. WARNING: This will allow Rill to create and overwrite tables in your database.")
		return nil, false
	}
	return c, true
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Handle.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// Migrate implements drivers.Handle.
func (c *connection) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (int, int, error) {
	return 0, 0, nil
}

// reopenDB opens the DuckDB handle anew. If c.db is already set, it closes the existing handle first.
func (c *connection) reopenDB(ctx context.Context) error {
	// If c.db is already open, close it first
	if c.db != nil {
		err := c.db.Close()
		if err != nil {
			return err
		}
		c.db = nil
	}

	var (
		dbInitQueries   []string
		connInitQueries []string
	)

	if c.driverName == "motherduck" || c.config.isMotherduck() {
		dbInitQueries = append(dbInitQueries,
			"INSTALL 'motherduck'",
			"LOAD 'motherduck'",
		)
		if c.config.Token != "" {
			dbInitQueries = append(dbInitQueries,
				fmt.Sprintf("SET motherduck_token = '%s'", c.config.Token),
			)
		}
	}

	// Add custom InitSQL queries before any other (e.g. to override the extensions repository)
	// BootQueries is deprecated. Use InitSQL instead. Retained for backward compatibility.
	if c.config.BootQueries != "" {
		dbInitQueries = append(dbInitQueries, c.config.BootQueries)
	}
	if c.config.InitSQL != "" {
		dbInitQueries = append(dbInitQueries, c.config.InitSQL)
	}

	dbInitQueries = append(dbInitQueries,
		"INSTALL 'json'",
		"INSTALL 'sqlite'",
		"INSTALL 'icu'",
		"INSTALL 'parquet'",
		"INSTALL 'httpfs'",
		"LOAD 'json'",
		"LOAD 'sqlite'",
		"LOAD 'icu'",
		"LOAD 'parquet'",
		"LOAD 'httpfs'",
		"SET GLOBAL timezone='UTC'",
		"SET GLOBAL old_implicit_casting = true", // Implicit Cast to VARCHAR
	)

	dataDir, err := c.storage.DataDir()
	if err != nil {
		return err
	}

	// We want to set preserve_insertion_order=false in hosted environments only (where source data is never viewed directly). Setting it reduces batch data ingestion time by ~40%.
	// Hack: Using AllowHostAccess as a proxy indicator for a hosted environment.
	if !c.config.AllowHostAccess {
		dbInitQueries = append(dbInitQueries,
			"SET GLOBAL preserve_insertion_order TO false",
		)
	}

	// Add init SQL if provided
	if c.config.ConnInitSQL != "" {
		connInitQueries = append(connInitQueries, c.config.ConnInitSQL)
	}
	connInitQueries = append(connInitQueries, "SET max_expression_depth TO 250")

	// Create new DB
	if c.config.Path != "" || c.config.Attach != "" {
		settings := make(map[string]string)
		maps.Copy(settings, c.config.readSettings())
		maps.Copy(settings, c.config.writeSettings())
		c.db, err = rduckdb.NewGeneric(ctx, &rduckdb.GenericOptions{
			Path:               c.config.Path,
			Attach:             c.config.Attach,
			DBName:             c.config.DatabaseName,
			SchemaName:         c.config.SchemaName,
			ReadOnlyMode:       c.config.Mode == modeReadOnly,
			LocalDataDir:       dataDir,
			LocalCPU:           c.config.CPU,
			LocalMemoryLimitGB: c.config.MemoryLimitGB,
			Settings:           settings,
			DBInitQueries:      dbInitQueries,
			ConnInitQueries:    connInitQueries,
			Logger:             c.logger,
			OtelAttributes:     []attribute.KeyValue{attribute.String("instance_id", c.instanceID)},
		})
		if err != nil {
			return err
		}
		return nil
	}
	c.db, err = rduckdb.NewDB(ctx, &rduckdb.DBOptions{
		LocalPath:       dataDir,
		Remote:          c.remote,
		CPU:             c.config.CPU,
		MemoryLimitGB:   c.config.MemoryLimitGB,
		ReadWriteRatio:  c.config.ReadWriteRatio,
		ReadSettings:    c.config.readSettings(),
		WriteSettings:   c.config.writeSettings(),
		DBInitQueries:   dbInitQueries,
		ConnInitQueries: connInitQueries,
		LogQueries:      c.config.LogQueries,
		Logger:          c.logger,
		OtelAttributes:  []attribute.KeyValue{attribute.String("instance_id", c.instanceID)},
	})
	return err
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
	conn, releaseConn, err := c.acquireReadConnection(ctx)
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
func (c *connection) acquireOLAPConn(ctx context.Context, priority int, longRunning bool) (*sqlx.Conn, func() error, error) {
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
	conn, releaseConn, err := c.acquireReadConnection(ctx)
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

// acquireReadConnection is a helper function to acquire a read connection from rduckdb.
// Do not use this function directly for OLAP queries. Use acquireOLAPConn, acquireMetaConn instead.
func (c *connection) acquireReadConnection(ctx context.Context) (*sqlx.Conn, func() error, error) {
	db, releaseDB, err := c.acquireDB()
	if err != nil {
		return nil, nil, err
	}

	conn, releaseConn, err := db.AcquireReadConnection(ctx)
	if err != nil {
		_ = releaseDB()
		return nil, nil, err
	}

	release := func() error {
		err := releaseConn()
		return errors.Join(err, releaseDB())
	}
	return conn, release, nil
}

// acquireDB returns rduckDB handle.
// acquireDB implements the connection tracking and DB reopening logic described in the struct definition for connection.
// It should not be used directly for select queries. For select queries use acquireOLAPConn and acquireMetaConn.
// It should only be used for write queries.
func (c *connection) acquireDB() (rduckdb.DB, func() error, error) {
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

	c.connTimesMu.Lock()
	connID := c.nextConnID
	c.nextConnID++
	c.connTimes[connID] = time.Now()
	c.connTimesMu.Unlock()

	release := func() error {
		c.connTimesMu.Lock()
		delete(c.connTimes, connID)
		c.connTimesMu.Unlock()
		c.dbCond.L.Lock()
		c.dbConnCount--
		if c.dbConnCount == 0 && c.dbReopen {
			c.triggerReopen()
		}
		c.dbCond.L.Unlock()
		return nil
	}
	return c.db, release, nil
}

func (c *connection) triggerReopen() {
	go func() {
		c.dbCond.L.Lock()
		defer c.dbCond.L.Unlock()
		if !c.dbReopen || c.dbConnCount != 0 {
			c.logger.Error("triggerReopen called but should not reopen", zap.Bool("dbReopen", c.dbReopen), zap.Int("dbConnCount", c.dbConnCount))
			return
		}
		c.dbReopen = false
		err := c.reopenDB(c.ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				c.logger.Error("reopen of DuckDB failed - the handle is now permanently locked", zap.Error(err))
			}
		}
		c.dbErr = err
		c.dbCond.Broadcast()
	}()
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
			estimatedDBSize := c.estimateSize()
			c.activity.RecordMetric(c.ctx, "duckdb_estimated_size_bytes", float64(estimatedDBSize))
		case <-c.ctx.Done():
			statTicker.Stop()
			return
		}
	}
}

// maxAcquiredConnDuration is the maximum duration a connection can be held for before we consider it potentially hanging/deadlocked.
const maxAcquiredConnDuration = 3 * time.Hour

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
			var connErr error
			for connID, connTime := range c.connTimes {
				if time.Since(connTime) > maxAcquiredConnDuration {
					connErr = fmt.Errorf("duckdb: a connection has been held for longer than the maximum allowed duration")
					c.logger.Error("duckdb: a connection has been held for longer than the maximum allowed duration", zap.Int("conn_id", connID), zap.Duration("duration", time.Since(connTime)))
				}
			}
			c.hangingConnErr = connErr
			c.connTimesMu.Unlock()
		}
	}
}
