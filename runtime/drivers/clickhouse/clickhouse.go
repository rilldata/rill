package clickhouse

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
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
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

func init() {
	drivers.Register("clickhouse", driver{})
	drivers.RegisterAsConnector("clickhouse", driver{})
}

var spec = drivers.Spec{
	DisplayName: "ClickHouse",
	Description: "Connect to ClickHouse.",
	DocsURL:     "https://docs.rilldata.com/reference/olap-engines/clickhouse",
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
			NoPrompt:    true,
		},
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    false,
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
			Placeholder: "localhost",
		},
		{
			Key:         "port",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Port",
			Description: "Port number of the ClickHouse server",
			Placeholder: "9000",
		},
		{
			Key:         "username",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Username",
			Description: "Username to connect to the ClickHouse server",
			Placeholder: "default",
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Password",
			Description: "Password to connect to the ClickHouse server",
			Placeholder: "password",
			Secret:      true,
		},
		{
			Key:         "database",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Database",
			Description: "Name of the ClickHouse database to connect to",
			Placeholder: "default",
		},
		{
			Key:         "ssl",
			Type:        drivers.BooleanPropertyType,
			Required:    true,
			DisplayName: "SSL",
			Description: "Use SSL to connect to the ClickHouse server",
		},
	},
	ImplementsOLAP: true,
}

type driver struct{}

type configProperties struct {
	// Managed is set internally if the connector has `managed: true`.
	Managed bool `mapstructure:"managed"`
	// Provision is set when Managed is true and provisioning should be handled by this driver.
	// (In practice, this gets set on local and means we should start an embedded Clickhouse server).
	Provision bool `mapstructure:"provision"`
	// DSN is the connection string. Either DSN can be passed or the individual properties below can be set.
	DSN string `mapstructure:"dsn"`
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
	// SSL determines whether secured connection need to be established. Should not be set if DSN is set.
	SSL bool `mapstructure:"ssl"`
	// Cluster name. If a cluster is configured, Rill will create all models in the cluster as distributed tables.
	Cluster string `mapstructure:"cluster"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.Execute.
	LogQueries bool `mapstructure:"log_queries"`
	// QuerySettingsOverride overrides the default query settings used for OLAP SELECT queries.
	// Use cases include disabling settings or setting `readonly = 1` when using read-only user.
	QuerySettingsOverride string `mapstructure:"query_settings_override"`
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

// Open connects to Clickhouse using std API.
// Connection string format : https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#dsn
func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("clickhouse driver can't be shared")
	}

	conf := &configProperties{
		CanScaleToZero: true,
	}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	// build clickhouse options
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

	// max_idle_conns
	if conf.MaxIdleConns != 0 {
		opts.MaxIdleConns = conf.MaxIdleConns
	}

	// conn_max_lifetime
	if conf.ConnMaxLifetime != "" {
		d, err := time.ParseDuration(conf.ConnMaxLifetime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conn_max_lifetime: %w", err)
		}
		opts.ConnMaxLifetime = d
	}

	// dial_timeout
	if conf.DialTimeout != "" {
		d, err := time.ParseDuration(conf.DialTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dial_timeout: %w", err)
		}
		opts.DialTimeout = d
	}
	if opts.DialTimeout == 0 { // Apply an increased default to reduce the chance of dropped connections with scaled-to-zero ClickHouse.
		opts.DialTimeout = time.Second * 60
	}

	// read_timeout
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

	db := sqlx.NewDb(otelsql.OpenDB(clickhouse.Connector(opts)), "clickhouse")
	err = db.Ping()
	if err != nil {
		if !strings.Contains(err.Error(), "unexpected packet") && !strings.Contains(err.Error(), "i/o timeout") {
			return nil, err
		}
		if conf.DSN != "" {
			return nil, err
		}
		// may be the port is http, also try with http protocol if DSN is not provided
		opts.Protocol = clickhouse.HTTP
		db = sqlx.NewDb(otelsql.OpenDB(clickhouse.Connector(opts)), "clickhouse")
		err := db.Ping()
		if err != nil {
			return nil, err
		}
		// connection with http protocol is successful
		logger.Warn("ClickHouse connection was established with the HTTP protocol, consider using the native port for better performance")
	}

	// Limit the number of concurrent connections
	if conf.MaxOpenConns == 0 {
		conf.MaxOpenConns = 20 // based on observations
	}
	db.SetMaxOpenConns(conf.MaxOpenConns)

	// Capture database stats with OpenTelemetry
	err = otelsql.RegisterDBStatsMetrics(db.DB, otelsql.WithAttributes(semconv.DBSystemClickhouse, attribute.String("instance_id", instanceID)))
	if err != nil {
		return nil, fmt.Errorf("registering db stats metrics: %w", err)
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

	ctx, cancel := context.WithCancel(context.Background())
	c := &Connection{
		db:         db,
		config:     conf,
		logger:     logger,
		activity:   ac,
		instanceID: instanceID,
		ctx:        ctx,
		cancel:     cancel,
		metaSem:    semaphore.NewWeighted(1),
		olapSem:    priorityqueue.NewSemaphore(conf.MaxOpenConns - 1),
		opts:       opts,
		embed:      embed,
	}

	c.used()
	go c.periodicallyEmitStats(time.Minute, 10*time.Minute)

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
	db         *sqlx.DB
	config     *configProperties
	logger     *zap.Logger
	activity   *activity.Client
	instanceID string

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
}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	err := c.db.PingContext(ctx)
	c.used()
	return err
}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "clickhouse"
}

// Config used to open the Connection
func (c *Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, &m)
	return m
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	c.cancel()

	errDB := c.db.Close()

	var errEmbed error
	if c.embed != nil {
		errEmbed = c.embed.stop()
	}

	return errors.Join(errDB, errEmbed)
}

// Registry implements drivers.Connection.
func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *Connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
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

// OLAP implements drivers.Connection.
func (c *Connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// Migrate implements drivers.Connection.
func (c *Connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *Connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *Connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *Connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	if opts.OutputHandle != c {
		return nil, false
	}
	if opts.InputHandle == c {
		return &selfToSelfExecutor{c}, true
	}
	if opts.InputHandle.Driver() == "s3" || opts.InputHandle.Driver() == "gcs" {
		return &objectStoreToSelfExecutor{opts.InputHandle, c}, true
	}
	if opts.InputHandle.Driver() == "local_file" {
		return &localFileToSelfExecutor{opts.InputHandle, c}, true
	}
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (c *Connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return c, true
}

// AsFileStore implements drivers.Connection.
func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *Connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
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
func (c *Connection) periodicallyEmitStats(sensitive, regular time.Duration) {
	if c.activity == nil {
		// Activity client isn't set, there is no need to report stats
		return
	}

	// Sensitive ticker for sensitive stats
	sensitiveTicker := time.NewTicker(sensitive)
	defer sensitiveTicker.Stop()

	// Regular ticker for non-sensitive stats
	regularTicker := time.NewTicker(regular)
	defer regularTicker.Stop()

	// Emit non-sensitive stats periodically
	go func() {
		for {
			select {
			case <-regularTicker.C:
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
			case <-c.ctx.Done():
				return
			}
		}
	}()

	// Emit sensitive stats periodically
	go func() {
		for {
			select {
			case <-sensitiveTicker.C:
				// Skip if it hasn't been used recently and may be scaled to zero.
				if c.config.CanScaleToZero && time.Since(c.lastUsedOn()) > 2*sensitive {
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

					c.logger.Log(lvl, "failed to estimate clickhouse size", zap.Error(err), zap.Bool("managed", c.config.Managed))
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

// estimateSize returns the estimated combined disk size of all resources in the database in bytes.
func (c *Connection) estimateSize(ctx context.Context) (int64, error) {
	var size int64
	err := c.db.QueryRowxContext(ctx, `SELECT sum(bytes_on_disk) AS size FROM system.parts WHERE (active = 1) AND lower(database) NOT IN ('information_schema', 'system')`).Scan(&size)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// latestRCUPerService returns the sum latest RCU value reported for nodes in each service i.e. read/write.
func (c *Connection) latestRCUPerService(ctx context.Context) (map[string]float64, error) {
	var query string
	if c.config.Cluster == "" {
		query = "SELECT service as billing_service, anyLast(value) as latest_value from billing.events event_name = 'rcu' GROUP BY service"
	} else {
		query = fmt.Sprintf(`SELECT service as billing_service, sum(value) AS latest_value FROM (SELECT service, anyLast(value) as value FROM clusterAllReplicas('%s', billing.events) event_name = 'rcu' GROUP BY hostName(), service) GROUP BY billing_service`, c.config.Cluster)
	}
	rows, err := c.db.QueryxContext(ctx, query)
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
