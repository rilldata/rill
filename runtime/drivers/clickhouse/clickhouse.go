package clickhouse

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

var (
	modeReadOnly  = "read"
	modeReadWrite = "readwrite"
)

func init() {
	drivers.Register("clickhouse", driver{})
	drivers.RegisterAsConnector("clickhouse", driver{})
}

var spec = drivers.Spec{
	DisplayName: "ClickHouse",
	Description: "Connect to ClickHouse.",
	DocsURL:     "https://docs.rilldata.com/build/connectors/olap/clickhouse",
	// Important: Any edits to the below properties must be accompanied by changes to the client-side form validation schemas.
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "managed",
			Type:        drivers.BooleanPropertyType,
			Required:    false,
			DisplayName: "Managed",
			Description: "Use a managed ClickHouse instance. This will start an embedded ClickHouse server in development.",
			Placeholder: "false",
			Default:     "false",
		},
		{
			Key:         "mode",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Mode",
			Description: "Set the mode for the ClickHouse connection. By default, it is set to 'read' which allows only read operations. Set to 'readwrite' to enable model creation and table mutations.",
			Placeholder: modeReadOnly,
			Default:     modeReadOnly,
			NoPrompt:    true,
		},
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Connection string",
			Placeholder: "clickhouse://localhost:9000?username=default&password=password",
			Secret:      true,
			NoPrompt:    true,
		},
		{
			Key:         "host",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Host",
			Description: "Hostname or IP address of the ClickHouse server",
			Placeholder: "your-instance.clickhouse.cloud or your.clickhouse.server.com",
			Hint:        "Your ClickHouse hostname (e.g., your-instance.clickhouse.cloud or your-server.com)",
		},
		{
			Key:         "port",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Port",
			Description: "Port number of the ClickHouse server",
			Placeholder: "9000",
			Hint:        "Default port is 9000 for native protocol. Also commonly used: 8443 for ClickHouse Cloud (HTTPS), 8123 for HTTP",
			Default:     "9000",
		},
		{
			Key:         "username",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Username",
			Description: "Username to connect to the ClickHouse server",
			Placeholder: "default",
			Hint:        "Username for authentication",
			Default:     "default",
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Password",
			Description: "Password to connect to the ClickHouse server",
			Placeholder: "Database password",
			Secret:      true,
			Hint:        "Password to your database",
		},
		{
			Key:         "database",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Database",
			Description: "Name of the ClickHouse database to connect to",
			Placeholder: "default",
			Hint:        "Database name (default is 'default')",
			Default:     "default",
		},
		{
			Key:         "cluster",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Cluster",
			Description: "Cluster name. If set, Rill will create all models in the cluster as distributed tables.",
			Placeholder: "Cluster name",
			Hint:        "Cluster name (required for some self-hosted ClickHouse setups)",
		},
		{
			Key:         "ssl",
			Type:        drivers.BooleanPropertyType,
			Required:    false,
			DisplayName: "SSL",
			Description: "Use SSL to connect to the ClickHouse server",
			Hint:        "Enable SSL for secure connections. For ClickHouse Cloud, SSL is always enabled.",
			Default:     "true",
		},
	},
	ImplementsOLAP: true,
}

type driver struct{}

type configProperties struct {
	// Managed is set internally if the connector has `managed: true`.
	Managed bool `mapstructure:"managed"`
	// Mode is set automatically to readwrite if Managed is true.
	Mode string `mapstructure:"mode"`
	// Provision is set when Managed is true and provisioning should be handled by this driver.
	// (In practice, this gets set on local and means we should start an embedded Clickhouse server).
	Provision bool `mapstructure:"provision"`
	// DSN is the connection string. Either DSN can be passed or the individual properties below can be set.
	// Additionally, WriteDSN can optionally be used to use a different connection for mutations.
	DSN string `mapstructure:"dsn"`
	// WriteDSN is the connection string for write operations. When set, the normal connection config (DSN or host, etc.) is used for reads and this for writes.
	WriteDSN string `mapstructure:"write_dsn"`
	// Host configuration. Should not be set if DSN is set.
	Host string `mapstructure:"host"`
	// Port configuration. Should not be set if DSN is set.
	Port int `mapstructure:"port"`
	// Username configuration. Should not be set if DSN is set.
	Username string `mapstructure:"username"`
	// Password configuration. Should not be set if DSN is set.
	Password string `mapstructure:"password"`
	// Database configuration. Should not be set if DSN is set.
	Database string `mapstructure:"database"`
	// DatabaseWhitelist is a comma separated list of databases to fetch in information_schema all calls.
	// This is just a *quick hack* to avoid fetching all databases in the table list till we have a better solution.
	// This does not list queries to other databases.
	DatabaseWhitelist string `mapstructure:"database_whitelist"`
	// OptimizeTemporaryTablesBeforePartitionReplace determines whether to optimize temporary tables before partition replacement.
	OptimizeTemporaryTablesBeforePartitionReplace bool `mapstructure:"optimize_temporary_tables_before_partition_replace"`
	// SSL determines whether secured connection need to be established. Should not be set if DSN is set.
	SSL bool `mapstructure:"ssl"`
	// Cluster name. If a cluster is configured, Rill will create all models in the cluster as distributed tables.
	Cluster string `mapstructure:"cluster"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.Execute.
	LogQueries bool `mapstructure:"log_queries"`
	// QuerySettingsOverride overrides the default query settings used for OLAP SELECT queries.
	// Use cases include disabling settings or setting `readonly = 1` when using read-only user.
	QuerySettingsOverride string `mapstructure:"query_settings_override"`
	// QuerySettings are set on each read query. QuerySettingsOverride takes precedence over these settings and if set these are ignored./
	// Each setting must be separated by a comma. Example `max_threads = 8, max_memory_usage = 10000000000`
	QuerySettings string `mapstructure:"query_settings"`
	// EmbedPort is the port to run Clickhouse locally (0 is random port).
	EmbedPort int `mapstructure:"embed_port"`
	// CanScaleToZero indicates if the underlying Clickhouse service may scale to zero when idle.
	// When set to true, we try to avoid too frequent non-user queries to the database (such as alert checks and fetching metrics).
	CanScaleToZero bool `mapstructure:"can_scale_to_zero"`
	// MaxOpenConns is the maximum number of open connections to the database.
	// See https://github.com/ClickHouse/clickhouse-go/blob/main/clickhouse_options.go
	MaxOpenConns int `mapstructure:"max_open_conns"`
	// MaxIdleConns is the maximum number of connections in the idle connection pool. Default is 5s.
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	// DialTimeout is the timeout for dialing the Clickhouse server. Defaults to 60s.
	DialTimeout string `mapstructure:"dial_timeout"`
	// ConnMaxLifetime is the maximum amount of time a connection may be reused.
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
	// ReadTimeout is the maximum amount of time a connection may be reused. Default is 300s.
	ReadTimeout string `mapstructure:"read_timeout"`
}

func (c *configProperties) validate() error {
	if c.Managed {
		// In managed mode, clear connection properties but preserve provisioner DSN
		c.Username = ""
		c.Password = ""
		c.Host = ""
		c.Port = 0
		c.Database = ""
		c.SSL = false
	} else if c.DSN != "" && (c.Host != "" || c.Username != "" || c.Password != "" || c.Database != "" || c.Port != 0 || c.SSL) {
		// Only validate conflicts when not in managed mode
		return errors.New("only one of 'dsn' or [host, port, username, password, database, ssl] can be set")
	}

	return nil
}

// Open connects to Clickhouse using std API.
// Connection string format : https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#dsn
func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("clickhouse driver can't be shared")
	}

	// Parse config properties
	conf := &configProperties{
		CanScaleToZero: true,
		MaxOpenConns:   20,
		MaxIdleConns:   5,
	}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}
	if err := conf.validate(); err != nil {
		return nil, err
	}

	// Mode defaults to readwrite for managed connections, otherwise readonly.
	if conf.Managed {
		conf.Mode = modeReadWrite
	} else if conf.Mode == "" {
		conf.Mode = modeReadOnly
	}

	// Build connection options
	var opts *clickhouse.Options
	var embed *embedClickHouse
	if conf.DSN != "" {
		opts, err = clickhouse.ParseDSN(conf.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DSN: %w", err)
		}
	} else if conf.Host != "" {
		opts = &clickhouse.Options{}

		// address
		host := conf.Host
		if conf.Port != 0 {
			host = fmt.Sprintf("%s:%d", conf.Host, conf.Port)
		}
		opts.Addr = []string{host}
		opts.Protocol = clickhouse.Native
		if conf.Port == 8123 || conf.Port == 8443 { // Default HTTP ports
			opts.Protocol = clickhouse.HTTP
		}
		if conf.SSL {
			opts.TLS = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}

		// username password
		opts.Auth.Username = conf.Username
		opts.Auth.Password = conf.Password

		// database
		opts.Auth.Database = conf.Database
	} else if conf.Provision {
		// run clickhouse locally
		dataDir, err := st.DataDir(instanceID)
		if err != nil {
			return nil, err
		}
		tempDir, err := st.TempDir(instanceID)
		if err != nil {
			return nil, err
		}

		embed, err = newEmbedClickHouse(conf.EmbedPort, dataDir, tempDir, logger)
		if err != nil {
			return nil, err
		}
		opts, err = embed.start()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("no clickhouse connection configured: 'dsn', 'host' or 'managed: true' must be set")
	}

	// Open the main database connection
	db, err := openHandle(instanceID, conf, opts, logger)
	if err != nil {
		return nil, err
	}

	// If we have a separate write DSN, open the write connection.
	writeDB := db
	if conf.WriteDSN != "" {
		writeOpts, err := clickhouse.ParseDSN(conf.WriteDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to parse write DSN: %w", err)
		}
		writeDB, err = openHandle(instanceID, conf, writeOpts, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to open write connection: %w", err)
		}
	}
	// group by positional args are supported post 22.7 and we use them heavily in our queries
	row := db.QueryRow(`
        WITH
            splitByChar('.', version()) AS parts,
            toInt32(parts[1]) AS major,
            toInt32(parts[2]) AS minor
        SELECT (major > 22) OR ((major = 22) AND (minor >= 7)) AS is_supported
	`)

	var isSupported bool
	if err := row.Scan(&isSupported); err != nil {
		return nil, err
	}
	if !isSupported {
		return nil, fmt.Errorf("clickhouse version must be 22.7 or higher")
	}

	// Using the harmless, non–side-effecting setting
	// `show_table_uuid_in_table_create_query_if_not_nil` as a probe to check
	// whether the cluster mode supports modifying query settings. This setting
	// has no practical use for our purposes.
	supportSettings := true
	if _, err := db.Exec("SET show_table_uuid_in_table_create_query_if_not_nil = 1"); err != nil {
		if strings.Contains(err.Error(), "Cannot modify") && strings.Contains(err.Error(), "setting in readonly mode") {
			supportSettings = false
		}
	}

	// Compute OLAP queue size
	var olapSemSize int
	if conf.MaxOpenConns < 1 {
		// MaxOpenConns <= 0 means unlimited connections
		olapSemSize = math.MaxInt
	} else if conf.MaxOpenConns > 1 {
		// Leave one connection for meta queries. All others can be used for OLAP.
		olapSemSize = conf.MaxOpenConns - 1
	} else {
		// If there is only one connection, both meta and olap queries need to share it.
		// There will be contention at the database/sql layer, but it will work.
		olapSemSize = 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &Connection{
		readDB:          db,
		writeDB:         writeDB,
		config:          conf,
		logger:          logger,
		activity:        ac,
		instanceID:      instanceID,
		supportSettings: supportSettings,
		ctx:             ctx,
		cancel:          cancel,
		metaSem:         semaphore.NewWeighted(1),
		olapSem:         priorityqueue.NewSemaphore(olapSemSize),
		opts:            opts,
		embed:           embed,
	}

	c.used()
	go c.periodicallyEmitStats()

	return c, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type Connection struct {
	readDB          *sqlx.DB
	writeDB         *sqlx.DB
	config          *configProperties
	logger          *zap.Logger
	activity        *activity.Client
	instanceID      string
	supportSettings bool

	// context that is cancelled when the connection is closed
	ctx    context.Context
	cancel context.CancelFunc

	// lastUsedUnixTime stores the time we last queried the connection.
	// This is used to guess if the DB may currently be scaled to zero.
	lastUsedUnixTime atomic.Int64

	// logic around this copied from duckDB driver
	// This driver may issue both OLAP and "meta" queries (like catalog info) against DuckDB.
	// Meta queries are usually fast, but OLAP queries may take a long time. To enable predictable parallel performance,
	// we gate queries with semaphores that limits the number of concurrent queries of each type.
	// The metaSem allows 1 query at a time and the olapSem allows cfg.PoolSize-1 queries at a time.
	// When cfg.PoolSize is 1, we set olapSem to still allow 1 query at a time.
	// This creates contention for the same connection in database/sql's pool, but its locks will handle that.
	metaSem *semaphore.Weighted
	olapSem *priorityqueue.Semaphore

	// options used to open clickhouse connections
	opts *clickhouse.Options
	// embed is embedded clickhouse server for local run
	embed *embedClickHouse
	// billingTableExists cached state of whether the billing.events table exists in the database
	billingTableExists *bool
}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	// Check both connections
	err := c.readDB.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	if c.writeDB != c.readDB {
		err = c.writeDB.PingContext(ctx)
		if err != nil {
			return fmt.Errorf("write connection ping failed: %w", err)
		}
	}

	c.used()
	return nil
}

// Driver implements drivers.Handle.
func (c *Connection) Driver() string {
	return "clickhouse"
}

// Config used to open the Connection
func (c *Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, &m)
	return m
}

// Close implements drivers.Handle.
func (c *Connection) Close() error {
	c.cancel()

	var errReadDB error
	if err := c.readDB.Close(); err != nil {
		errReadDB = fmt.Errorf("closing connection: %w", err)
	}

	var errWriteDB error
	if c.writeDB != c.readDB {
		if err := c.writeDB.Close(); err != nil {
			errWriteDB = fmt.Errorf("closing write connection: %w", err)
		}
	}

	var errEmbed error
	if c.embed != nil {
		errEmbed = c.embed.stop()
	}

	return errors.Join(errReadDB, errWriteDB, errEmbed)
}

// Registry implements drivers.Handle.
func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Handle.
func (c *Connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Handle.
func (c *Connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *Connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *Connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// OLAP implements drivers.Handle.
func (c *Connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// AsInformationSchema implements drivers.Handle.
func (c *Connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return c, true
}

// Migrate implements drivers.Handle.
func (c *Connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *Connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Handle.
func (c *Connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *Connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	if opts.OutputHandle != c {
		return nil, drivers.ErrNotImplemented
	}
	if c.config.Mode != modeReadWrite {
		return nil, fmt.Errorf("model execution is disabled. To enable modeling on this ClickHouse database, set 'mode: readwrite' in your connector configuration. WARNING: This will allow Rill to create and overwrite tables in your database")
	}
	if opts.InputHandle == c {
		return &selfToSelfExecutor{c}, nil
	}
	if opts.InputHandle.Driver() == "s3" || opts.InputHandle.Driver() == "gcs" {
		return &objectStoreToSelfExecutor{opts.InputHandle, c}, nil
	}
	if opts.InputHandle.Driver() == "local_file" {
		return &localFileToSelfExecutor{opts.InputHandle, c}, nil
	}
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *Connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	if c.config.Mode != modeReadWrite {
		c.logger.Warn("Model execution is disabled. To enable modeling on this ClickHouse database, set 'mode: readwrite' in your connector configuration. WARNING: This will allow Rill to create and overwrite tables in your database.")
		return nil, false
	}
	return c, true
}

// AsFileStore implements drivers.Handle.
func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *Connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Handle.
func (c *Connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// used should be called after a query to the database completes.
// It bumps the result of lastUsedOn(), which can be used to guess if the DB may currently be scaled to zero.
//
// Periodic background jobs that rely on lastUsedOn should not call this function since it will lead to the database never scaling to zero.
func (c *Connection) used() {
	c.lastUsedUnixTime.Store(time.Now().Unix())
}

// lastUsedOn returns the time we last queried the connection.
// This can be used to guess if the DB may currently be scaled to zero.
func (c *Connection) lastUsedOn() time.Time {
	return time.Unix(c.lastUsedUnixTime.Load(), 0)
}

// Periodically collects stats about the database and emit them as activity events.
func (c *Connection) periodicallyEmitStats() {
	if c.activity == nil {
		// Activity client isn't set, there is no need to report stats
		return
	}

	// Sensitive ticker for sensitive stats
	sensitiveTicker := time.NewTicker(time.Minute)
	defer sensitiveTicker.Stop()

	// Regular ticker for non-sensitive stats
	regularTicker := time.NewTicker(10 * time.Minute)
	defer regularTicker.Stop()

	// Cache invalidation ticker to reset billing table existence cache
	cacheInvalidationTicker := time.NewTicker(60 * time.Minute)
	defer cacheInvalidationTicker.Stop()

	skipEstimatedSizeEmission := false
	for {
		select {
		case <-sensitiveTicker.C:
			// Skip if it hasn't been used recently and may be scaled to zero.
			if (c.config.CanScaleToZero && time.Since(c.lastUsedOn()) > 2*time.Minute) || skipEstimatedSizeEmission {
				continue
			}

			// Emit the estimated size of the database.
			size, err := c.estimateSize(c.ctx)
			if err == nil {
				c.activity.RecordMetric(c.ctx, "clickhouse_estimated_size_bytes", float64(size))
			} else if !errors.Is(err, c.ctx.Err()) {
				lvl := zap.WarnLevel
				if c.config.Managed {
					lvl = zap.ErrorLevel
				}

				var chErr *clickhouse.Exception
				if errors.As(err, &chErr) && chErr.Code == 497 {
					// Code 497 is "Not enough privileges" - downgrade to debug level and skip future emissions.
					lvl = zap.DebugLevel
					skipEstimatedSizeEmission = true
				}

				c.logger.Log(lvl, "failed to estimate clickhouse size", zap.Error(err), zap.Bool("managed", c.config.Managed))
			}
		case <-regularTicker.C:
			// Skip if it hasn't been used recently and may be scaled to zero.
			if !(c.config.CanScaleToZero && time.Since(c.lastUsedOn()) > 2*time.Minute) && !skipEstimatedSizeEmission {
				// Emit the estimated size per table.
				tableSizes, err := c.estimatePerTableSize(c.ctx)
				if err == nil {
					for _, ts := range tableSizes {
						c.activity.RecordMetric(c.ctx, "clickhouse_per_table_estimated_size_bytes", float64(ts.size), attribute.String("database", ts.database), attribute.String("table", ts.table))
					}
				} else if !errors.Is(err, c.ctx.Err()) {
					c.logger.Warn("failed to estimate clickhouse per-table sizes", zap.Error(err))
				}
			}

			// Check if billing.events table exists (with caching).
			billingTableExists, err := c.checkBillingTableExists(c.ctx, c.config.Cluster)
			if err != nil {
				if !errors.Is(err, c.ctx.Err()) {
					c.logger.Warn("failed to check if billing table exists", zap.Error(err))
				}
				continue
			}
			if !billingTableExists {
				c.logger.Debug("billing.events table does not exist in the database, RCU metrics will not be available", zap.String("clickhouse_host", c.config.Host), zap.String("clickhouse_dsn", c.config.DSN))
				continue
			}
			// Emit the latest RCU per service.
			latestRCU, err := c.latestRCUPerService(c.ctx)
			if err == nil {
				for service, value := range latestRCU {
					c.activity.RecordMetric(c.ctx, "clickhouse_rcu", value, attribute.String("billing_service", service))
				}
				if len(latestRCU) == 0 {
					c.logger.Warn("no RCU data found for any service", zap.String("clickhouse_host", c.config.Host))
				}
			} else if !errors.Is(err, c.ctx.Err()) {
				c.logger.Warn("failed to fetch latest RCU per service", zap.Error(err))
			}
		case <-cacheInvalidationTicker.C:
			// Invalidate the billing table existence cache and skip size emission flag.
			c.billingTableExists = nil
			skipEstimatedSizeEmission = false
		case <-c.ctx.Done():
			return
		}
	}
}

// estimateSize returns the estimated combined disk size of all resources in the database in bytes.
func (c *Connection) estimateSize(ctx context.Context) (int64, error) {
	var size int64
	var query string
	if c.config.Cluster == "" {
		query = `SELECT sum(bytes_on_disk) AS size FROM system.parts WHERE (active = 1) AND lower(database) NOT IN ('information_schema', 'system')`
	} else {
		query = fmt.Sprintf(`SELECT sum(bytes_on_disk) AS size FROM cluster('%s', system.parts) WHERE (active = 1) AND lower(database) NOT IN ('information_schema', 'system')`, c.config.Cluster)
	}
	err := c.readDB.QueryRowxContext(ctx, query).Scan(&size)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// estimatePerTableSize returns the estimated average disk size per table in bytes.
func (c *Connection) estimatePerTableSize(ctx context.Context) ([]*tableSize, error) {
	var query string
	if c.config.Cluster == "" {
		query = `SELECT database, table, sum(bytes_on_disk) AS size FROM system.parts WHERE (active = 1) AND lower(database) NOT IN ('information_schema', 'system') GROUP BY database, table`
	} else {
		query = fmt.Sprintf(`SELECT database, table, sum(bytes_on_disk) AS size FROM cluster('%s', system.parts) WHERE (active = 1) AND lower(database) NOT IN ('information_schema', 'system') GROUP BY database, table`, c.config.Cluster)
	}
	rows, err := c.readDB.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tableSizes []*tableSize
	for rows.Next() {
		var ts tableSize
		if err := rows.Scan(&ts.database, &ts.table, &ts.size); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		tableSizes = append(tableSizes, &ts)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return tableSizes, nil
}

// latestRCUPerService returns the sum latest RCU value reported for nodes in each service i.e. read/write.
func (c *Connection) latestRCUPerService(ctx context.Context) (map[string]float64, error) {
	var query string
	if c.config.Cluster == "" {
		query = "SELECT service as billing_service, anyLast(value) as latest_value from billing.events WHERE event_name = 'rcu' GROUP BY service"
	} else {
		query = fmt.Sprintf(`SELECT service as billing_service, sum(value) AS latest_value FROM (SELECT service, anyLast(value) as value FROM clusterAllReplicas('%s', billing.events) WHERE event_name = 'rcu' GROUP BY hostName(), service) GROUP BY billing_service`, c.config.Cluster)
	}
	rows, err := c.readDB.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	latestRCU := make(map[string]float64)
	for rows.Next() {
		var service string
		var value float64
		if err := rows.Scan(&service, &value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		latestRCU[service] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return latestRCU, nil
}

func (c *Connection) checkBillingTableExists(ctx context.Context, cluster string) (bool, error) {
	if c.billingTableExists != nil {
		return *c.billingTableExists, nil
	}
	var existsEverywhere bool
	var existsSomewhere bool
	if cluster == "" {
		err := c.readDB.QueryRowxContext(ctx, `SELECT count() > 0 as exists FROM system.tables WHERE database = 'billing' AND name = 'events'`).Scan(&existsEverywhere)
		if err != nil {
			return false, fmt.Errorf("failed to check if billing table exists: %w", err)
		}
	} else {
		err := c.readDB.QueryRowxContext(ctx, fmt.Sprintf(`
				SELECT countIf(found) = count() AS exists_everywhere, countIf(found) > 0 AS exists_somewhere
				FROM
				(
					SELECT hostName() AS host, max((database = 'billing') AND (name = 'events')) AS found
					FROM clusterAllReplicas('%s', system.tables)
					GROUP BY host
				)`, cluster)).Scan(&existsEverywhere, &existsSomewhere)
		if err != nil {
			return false, fmt.Errorf("failed to check if billing table exists in cluster %q: %w", cluster, err)
		}
		if existsSomewhere && !existsEverywhere {
			c.logger.Warn("billing.events table does not exists on all cluster nodes, RCU will not be reported", zap.String("clickhouse_host", c.config.Host), zap.String("clickhouse_dsn", c.config.DSN))
		}
	}
	c.billingTableExists = &existsEverywhere
	return existsEverywhere, nil
}

func openHandle(instanceID string, conf *configProperties, opts *clickhouse.Options, logger *zap.Logger) (*sqlx.DB, error) {
	// Apply certain options from conf that are not set in the DSN.
	if conf.MaxIdleConns != 0 {
		opts.MaxIdleConns = conf.MaxIdleConns
	}

	if conf.ConnMaxLifetime != "" {
		d, err := time.ParseDuration(conf.ConnMaxLifetime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conn_max_lifetime: %w", err)
		}
		opts.ConnMaxLifetime = d
	}

	if conf.DialTimeout != "" {
		d, err := time.ParseDuration(conf.DialTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dial_timeout: %w", err)
		}
		opts.DialTimeout = d
	}

	if conf.ReadTimeout != "" {
		d, err := time.ParseDuration(conf.ReadTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to parse read_timeout: %w", err)
		}
		opts.ReadTimeout = d
	}
	if opts.ReadTimeout == 0 { // Apply an increased default to reduce the chance of dropped connections with scaled-to-zero ClickHouse.
		opts.ReadTimeout = time.Second * 300
	}

	// NOTE: After https://github.com/ClickHouse/clickhouse-go/pull/1709, we can remove the manual TCP dial and
	// the default 60s DialTimeout.
	// The manual dial currently ensures that the host and port are reachable.
	// This check only verifies that the TCP socket is open — it will succeed even if the ClickHouse instance is scaled to zero.
	// It prevents invalid host/port combinations from proceeding to db.Ping, which uses a longer timeout to handle scale-to-zero scenarios.
	if conf.Host != "" && conf.Port != 0 {
		target := net.JoinHostPort(conf.Host, fmt.Sprintf("%d", conf.Port))
		conn, err := net.DialTimeout("tcp", target, 25*time.Second)
		if err != nil {
			return nil, fmt.Errorf("please check that the host and port are correct %s: %w", target, err)
		}
		conn.Close()
	}
	if opts.DialTimeout == 0 { // Apply an increased default to reduce the chance of dropped connections with scaled-to-zero ClickHouse.
		opts.DialTimeout = time.Second * 60
	}

	// Open the connection
	db := sqlx.NewDb(otelsql.OpenDB(clickhouse.Connector(opts)), "clickhouse")
	err := db.Ping()
	if err != nil {
		// Detect SSL/TLS mismatch (common causes: "read: EOF" or TLS Alert [21])
		if strings.Contains(err.Error(), "EOF") ||
			strings.Contains(err.Error(), "[handshake] unexpected packet [21]") ||
			(strings.Contains(err.Error(), "malformed HTTP response") && strings.Contains(err.Error(), "\\x15")) {
			return nil, fmt.Errorf("handshake failed (this usually happens due to SSL/TLS mismatch): %w", err)
		}
		// Return immediately without retrying in the following cases:
		//   1. The current protocol is already HTTP (no need to retry with HTTP again).
		//   2. A DSN was explicitly provided (respect the user’s configuration).
		//   3. The error is not "unexpected packet [72]" → The native protocol hit an HTTP endpoint.
		if opts.Protocol == clickhouse.HTTP || conf.DSN != "" || !strings.Contains(err.Error(), "[handshake] unexpected packet [72]") {
			return nil, err
		}
		// may be the port is http, also try with http protocol if DSN is not provided
		opts.Protocol = clickhouse.HTTP
		db = sqlx.NewDb(otelsql.OpenDB(clickhouse.Connector(opts)), "clickhouse")
		err := db.Ping()
		if err != nil {
			// Detect SSL/TLS mismatch (common causes: "read: EOF" or  \x15 means TLS Alert [21]"])
			if strings.Contains(err.Error(), "EOF") ||
				(strings.Contains(err.Error(), "malformed HTTP response") && strings.Contains(err.Error(), "\\x15")) {
				return nil, fmt.Errorf("handshake failed (this usually happens due to SSL/TLS mismatch): %w", err)
			}
			return nil, err
		}
		// connection with http protocol is successful
		logger.Warn("ClickHouse connection was established with the HTTP protocol, consider using the native port for better performance")
	}

	// Limit the number of concurrent connections
	if conf.MaxOpenConns != 0 {
		db.SetMaxOpenConns(conf.MaxOpenConns)
	}

	// Capture database stats with OpenTelemetry
	err = otelsql.RegisterDBStatsMetrics(db.DB, otelsql.WithAttributes(semconv.DBSystemClickhouse, attribute.String("instance_id", instanceID)))
	if err != nil {
		return nil, fmt.Errorf("registering db stats metrics: %w", err)
	}

	return db, nil
}

type tableSize struct {
	database string
	table    string
	size     int64
}
