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
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Connection string",
			Placeholder: "clickhouse://localhost:9000?username=default&password=",
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
			Key:         "ssl",
			Type:        drivers.BooleanPropertyType,
			Required:    true,
			DisplayName: "SSL",
			Description: "Use SSL to connect to the ClickHouse server",
		},
		{
			Key:         "database",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Database",
			Description: "Specify the database within the ClickHouse server",
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
	DSN      string `mapstructure:"dsn"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	// Database specifies the name of the ClickHouse database within the cluster.
	Database string `mapstructure:"database"`
	// SSL determines whether secured connection need to be established. To be set when setting individual fields.
	SSL bool `mapstructure:"ssl"`
	// Cluster name. Required for running distributed queries.
	Cluster string `mapstructure:"cluster"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.Execute.
	LogQueries bool `mapstructure:"log_queries"`
	// SettingsOverride override the default settings used in queries. One use case is to disable settings and set `readonly = 1` when using read-only user.
	SettingsOverride string `mapstructure:"settings_override"`
	// EmbedPort is the port to run Clickhouse locally (0 is random port).
	EmbedPort      int  `mapstructure:"embed_port"`
	CanScaleToZero bool `mapstructure:"can_scale_to_zero"`
	// MaxOpenConns is the maximum number of open connections to the database.
	// https://github.com/ClickHouse/clickhouse-go/blob/main/clickhouse_options.go
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
	maxOpenConnections := 20 // Very roughly approximating the number of queries required for a typical page load.
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
		if conf.Password != "" {
			opts.Auth.Username = conf.Username
			opts.Auth.Password = conf.Password
		} else if conf.Username != "" {
			opts.Auth.Username = conf.Username
		}

		// database
		if conf.Database != "" {
			opts.Auth.Database = conf.Database
		}

		// max_open_conns
		if conf.MaxOpenConns != 0 {
			maxOpenConnections = conf.MaxOpenConns
		}

		// max_idle_conns
		if conf.MaxIdleConns != 0 {
			opts.MaxIdleConns = conf.MaxIdleConns
		}

		// dial_timeout
		if conf.DialTimeout != "" {
			d, err := time.ParseDuration(conf.DialTimeout)
			if err != nil {
				return nil, fmt.Errorf("failed to parse dial_timeout: %w", err)
			}
			opts.DialTimeout = d
		}

		// conn_max_lifetime
		if conf.ConnMaxLifetime != "" {
			d, err := time.ParseDuration(conf.ConnMaxLifetime)
			if err != nil {
				return nil, fmt.Errorf("failed to parse conn_max_lifetime: %w", err)
			}
			opts.ConnMaxLifetime = d
		}

		// read_timeout
		if conf.ReadTimeout != "" {
			d, err := time.ParseDuration(conf.ReadTimeout)
			if err != nil {
				return nil, fmt.Errorf("failed to parse read_timeout: %w", err)
			}
			opts.ReadTimeout = d
		}
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

	// Apply our own defaults for the options.
	// We increase timeouts to decrease the chance of dropped connections with scaled-to-zero ClickHouse.
	if opts.DialTimeout == 0 {
		opts.DialTimeout = time.Second * 60
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = time.Second * 300
	}
	if opts.ConnMaxLifetime == 0 {
		opts.ConnMaxLifetime = time.Hour
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
		logger.Warn("clickHouse connection is established with HTTP protocol. Use native port for better performance")
	}
	db.SetMaxOpenConns(maxOpenConnections)

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
	c := &connection{
		db:         db,
		config:     conf,
		logger:     logger,
		activity:   ac,
		instanceID: instanceID,
		ctx:        ctx,
		cancel:     cancel,
		metaSem:    semaphore.NewWeighted(1),
		olapSem:    priorityqueue.NewSemaphore(maxOpenConnections - 1),
		opts:       opts,
		embed:      embed,
	}

	c.used()
	go c.periodicallyEmitStats(time.Minute)

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

type connection struct {
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
func (c *connection) Ping(ctx context.Context) error {
	err := c.db.PingContext(ctx)
	c.used()
	return err
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "clickhouse"
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, &m)
	return m
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	c.cancel()

	errDB := c.db.Close()

	var errEmbed error
	if c.embed != nil {
		errEmbed = c.embed.stop()
	}

	return errors.Join(errDB, errEmbed)
}

// Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
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

// OLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	c.instanceID = instanceID
	return c, true
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	if opts.OutputHandle != c {
		return nil, false
	}
	if opts.InputHandle == c {
		return &selfToSelfExecutor{c}, true
	}
	if opts.InputHandle.Driver() == "s3" {
		return &s3ToSelfExecutor{opts.InputHandle, c}, true
	}
	if opts.InputHandle.Driver() == "local_file" {
		return &localFileToSelfExecutor{opts.InputHandle, c}, true
	}
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return c, true
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

func (c *connection) AcquireLongRunning(ctx context.Context) (func(), error) {
	return nil, fmt.Errorf("not implemented")
}

// used should be called after a query to the database completes.
// It bumps the result of lastUsedOn(), which can be used to guess if the DB may currently be scaled to zero.
//
// Periodic background jobs that rely on lastUsedOn should not call this function since it will lead to the database never scaling to zero.
func (c *connection) used() {
	c.lastUsedUnixTime.Store(time.Now().Unix())
}

// lastUsedOn returns the time we last queried the connection.
// This can be used to guess if the DB may currently be scaled to zero.
func (c *connection) lastUsedOn() time.Time {
	return time.Unix(c.lastUsedUnixTime.Load(), 0)
}

// Periodically collects stats about the database and emit them as activity events.
func (c *connection) periodicallyEmitStats(d time.Duration) {
	if c.activity == nil {
		// Activity client isn't set, there is no need to report stats
		return
	}

	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Skip if it hasn't been used recently and may be scaled to zero.
			if c.config.CanScaleToZero && time.Since(c.lastUsedOn()) > 2*d {
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
}

// estimateSize returns the estimated combined disk size of all resources in the database in bytes.
func (c *connection) estimateSize(ctx context.Context) (int64, error) {
	var size int64
	err := c.db.QueryRowxContext(ctx, `SELECT sum(bytes_on_disk) AS size FROM system.parts WHERE (active = 1) AND lower(database) NOT IN ('information_schema', 'system')`).Scan(&size)
	if err != nil {
		return 0, err
	}

	return size, nil
}
