package starrocks

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/apache/arrow-go/v18/arrow/flight/flightsql"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("starrocks", driver{})
	drivers.RegisterAsConnector("starrocks", driver{})
}

var spec = drivers.Spec{
	DisplayName: "StarRocks",
	Description: "Connect to StarRocks.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connectors/olap/starrocks",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "StarRocks Connection String",
			Required:    false,
			Placeholder: "user:password@tcp(host:9030)/?timeout=30s&readTimeout=300s&parseTime=true",
			Hint:        "MySQL DSN format only. If provided, do not set host/port/username/password. Catalog and database should be set separately for external catalogs.",
			Description: "Complete MySQL connection string. Use either DSN or individual fields below, not both.",
			Secret:      true,
		},
		{
			Key:         "host",
			Type:        drivers.StringPropertyType,
			DisplayName: "Host",
			Required:    true,
			Placeholder: "localhost",
			Description: "Hostname or IP address of the StarRocks FE node",
			Hint:        "StarRocks FE (Frontend) server hostname",
		},
		{
			Key:         "port",
			Type:        drivers.NumberPropertyType,
			DisplayName: "Port",
			Required:    false,
			Placeholder: "9030",
			Default:     "9030",
			Description: "MySQL protocol port of the StarRocks FE node",
			Hint:        "Default MySQL protocol port is 9030",
		},
		{
			Key:         "username",
			Type:        drivers.StringPropertyType,
			DisplayName: "Username",
			Required:    true,
			Placeholder: "root",
			Default:     "root",
			Description: "Username to connect to StarRocks",
			Hint:        "StarRocks username for authentication",
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			DisplayName: "Password",
			Required:    false,
			Placeholder: "your_password",
			Description: "Password to connect to StarRocks",
			Hint:        "StarRocks password for authentication",
			Secret:      true,
		},
		{
			Key:         "catalog",
			Type:        drivers.StringPropertyType,
			DisplayName: "Catalog",
			Required:    true,
			Placeholder: "default_catalog",
			Default:     "default_catalog",
			Description: "Name of the StarRocks catalog (for external catalogs like Iceberg, Hive)",
			Hint:        "Use default_catalog for internal tables, or specify external catalog name",
		},
		{
			Key:         "database",
			Type:        drivers.StringPropertyType,
			DisplayName: "Database",
			Required:    false,
			Placeholder: "default",
			Description: "Name of the StarRocks database to connect to",
			Hint:        "Database name to use as default",
		},
		{
			Key:         "ssl",
			Type:        drivers.BooleanPropertyType,
			DisplayName: "SSL",
			Required:    false,
			Default:     "false",
			Description: "Enable SSL for secure connections",
			Hint:        "Enable SSL/TLS encryption for the connection",
		},
		{
			Key:         "log_queries",
			Type:        drivers.BooleanPropertyType,
			DisplayName: "Log Queries",
			Required:    false,
			Default:     "false",
			Description: "Enable logging of all SQL queries",
			Hint:        "Useful for debugging (logs all SQL statements)",
		},
		{
			Key:         "transport",
			Type:        drivers.StringPropertyType,
			DisplayName: "Transport Protocol",
			Required:    false,
			Default:     "mysql",
			Description: "Query transport protocol: 'mysql' (default) or 'flight_sql' (Arrow Flight SQL for better performance)",
			Hint:        "Arrow Flight SQL requires StarRocks 3.0+ with arrow_flight_sql_port enabled",
		},
		{
			Key:         "flight_sql_port",
			Type:        drivers.NumberPropertyType,
			DisplayName: "Arrow Flight SQL Port",
			Required:    false,
			Default:     "9408",
			Placeholder: "9408",
			Description: "Arrow Flight SQL port on the StarRocks FE node (arrow_flight_port in fe.conf)",
			Hint:        "Only used when transport is 'flight_sql'. Connect to FE, which handles load balancing to BE nodes.",
		},
		{
			Key:         "flight_sql_be_addr",
			Type:        drivers.StringPropertyType,
			DisplayName: "Arrow Flight SQL BE Address Override",
			Required:    false,
			Description: "Override BE address (host:port) for Flight SQL DoGet calls. Use when BE nodes are behind NAT/Docker.",
			Hint:        "Only needed if BE nodes advertise internal addresses that are not directly reachable.",
		},
	},
	ImplementsOLAP: true,
}

type driver struct{}

// ConfigProperties defines the configuration for StarRocks connection.
// NOTE: Timezone configuration is not supported for StarRocks.
// StarRocks uses server-side timezone settings and queries use UTC by default.
// Unlike some other drivers, there is no client-side timezone parameter in the DSN.
type ConfigProperties struct {
	// DSN is the complete connection string. Either DSN or individual fields should be set.
	DSN string `mapstructure:"dsn"`
	// Host is the StarRocks FE hostname or IP.
	Host string `mapstructure:"host"`
	// Port is the MySQL protocol port (default: 9030).
	Port int `mapstructure:"port"`
	// Username for authentication.
	Username string `mapstructure:"username"`
	// Password for authentication.
	Password string `mapstructure:"password"`
	// Catalog is the StarRocks catalog (for external catalogs like Iceberg, Hive).
	Catalog string `mapstructure:"catalog"`
	// Database is the default database to use.
	Database string `mapstructure:"database"`
	// SSL enables TLS encryption.
	SSL bool `mapstructure:"ssl"`
	// LogQueries enables SQL query logging.
	LogQueries bool `mapstructure:"log_queries"`
	// Transport selects the query transport protocol: "mysql" (default) or "flight_sql".
	Transport string `mapstructure:"transport"`
	// FlightSQLPort is the Arrow Flight SQL port on FE (default: 9408).
	// Only used when Transport is "flight_sql".
	FlightSQLPort int `mapstructure:"flight_sql_port"`
	// FlightSQLBEAddr overrides the BE address (host:port) for Flight SQL DoGet calls.
	// When set, all DoGet calls are routed to this address instead of the BE
	// addresses from FlightInfo endpoint Locations. Useful when BE nodes are behind
	// NAT/Docker and their advertised addresses are not directly reachable.
	FlightSQLBEAddr string `mapstructure:"flight_sql_be_addr"`
}

// Validate checks the configuration for errors.
func (c *ConfigProperties) Validate() error {
	// Either DSN or individual connection parameters must be provided
	if c.DSN == "" && c.Host == "" {
		return errors.New("either DSN or Host must be provided")
	}

	// If DSN is provided, other connection parameters should not be set
	// Exception: catalog and database can be set for external catalog configuration
	if c.DSN != "" {
		if c.Host != "" || c.Port != 0 || c.Username != "" || c.Password != "" {
			return errors.New("when DSN is provided, individual connection parameters (host, port, username, password) should not be set")
		}
	}

	// Validate transport
	if c.Transport != "" && c.Transport != "mysql" && c.Transport != "flight_sql" {
		return fmt.Errorf("invalid transport %q: must be 'mysql' or 'flight_sql'", c.Transport)
	}

	return nil
}

const (
	defaultCatalog       = "default_catalog"
	defaultPort          = 9030
	defaultFlightSQLPort = 9408
)

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("starrocks driver: instance ID is required")
	}

	cfg := &ConfigProperties{}
	if err := mapstructure.WeakDecode(config, cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Apply defaults
	if cfg.Catalog == "" {
		cfg.Catalog = defaultCatalog
	}
	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}
	if cfg.FlightSQLPort == 0 {
		cfg.FlightSQLPort = defaultFlightSQLPort
	}

	conn := &connection{
		configProp: cfg,
		logger:     logger,
		activity:   ac,
	}

	// Open database connection immediately in drivers.Open
	// This ensures the connection is established and validated upfront
	if err := conn.initDB(); err != nil {
		return nil, fmt.Errorf("failed to initialize database connection: %w", err)
	}

	return conn, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type connection struct {
	configProp *ConfigProperties
	logger     *zap.Logger
	activity   *activity.Client

	// db is initialized in drivers.Open (always available for Exec/Ping/DDL)
	db *sqlx.DB
	// flightClient is initialized when transport is "flight_sql"
	flightClient *flightsql.Client
	// beClients caches Flight SQL clients for BE nodes, keyed by "host:port".
	// Reusing clients avoids per-query gRPC connection overhead.
	beClients   map[string]*flightsql.Client
	beClientsMu sync.Mutex
}

var _ drivers.Handle = (*connection)(nil)

// Driver implements drivers.Handle.
func (c *connection) Driver() string {
	return "starrocks"
}

// Config implements drivers.Handle.
func (c *connection) Config() map[string]any {
	m := make(map[string]any)
	_ = mapstructure.Decode(c.configProp, &m)
	return m
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// Migrate implements drivers.Handle.
func (c *connection) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Close implements drivers.Handle.
func (c *connection) Close() error {
	var errs []error
	c.beClientsMu.Lock()
	for addr, client := range c.beClients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close BE flight client %s: %w", addr, err))
		}
	}
	c.beClients = nil
	c.beClientsMu.Unlock()
	if c.flightClient != nil {
		if err := c.flightClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close flight client: %w", err))
		}
	}
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close db: %w", err))
		}
	}
	return errors.Join(errs...)
}

// AsRegistry implements drivers.Handle.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
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

// AsOLAP implements drivers.Handle.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// AsInformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return &informationSchemaImpl{c: c}, true
}

// AsObjectStore implements drivers.Handle.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
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

// AsModelExecutor implements drivers.Handle.
// StarRocks is a read-only OLAP connector, model execution is not supported.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
// StarRocks is a read-only OLAP connector, model management is not supported.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// initDB initializes the database connection.
// Called during drivers.Open to establish connection upfront.
func (c *connection) initDB() error {
	dsn := c.buildDSN()

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	// Rely on database driver's internal connection pooling (uses driver default for MaxOpenConns)
	// MaxOpenConns set to 20 to default limit concurrent connections
	db.SetMaxOpenConns(20) // 0 means unlimited
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection with an independent context to prevent premature cancellation
	// Use a context with sufficient timeout (30 seconds) instead of the request context
	// This prevents 499 errors when the frontend request is cancelled quickly
	pingCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Validate catalog exists (only if specified and not default_catalog)
	// This validation happens once during initialization, not on every Ping()
	// Use catalog's information_schema to verify the catalog is accessible
	if c.configProp.Catalog != "" && c.configProp.Catalog != defaultCatalog {
		var dummy int
		catalogQuery := fmt.Sprintf("SELECT 1 FROM %s.information_schema.schemata LIMIT 1", c.configProp.Catalog)
		if err := db.QueryRowContext(pingCtx, catalogQuery).Scan(&dummy); err != nil {
			db.Close()
			return fmt.Errorf("failed to validate catalog %q: %w", c.configProp.Catalog, err)
		}
	}

	// Validate database exists (only if specified)
	// Use fully qualified path: <catalog>.information_schema.schemata
	// This validation happens once during initialization, not on every Ping()
	if c.configProp.Database != "" {
		catalog := c.configProp.Catalog
		if catalog == "" {
			catalog = defaultCatalog
		}

		var dbCount int
		dbQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s.information_schema.schemata WHERE schema_name = ?", catalog)
		if err := db.QueryRowContext(pingCtx, dbQuery, c.configProp.Database).Scan(&dbCount); err != nil {
			db.Close()
			return fmt.Errorf("failed to validate database: %w", err)
		}
		if dbCount == 0 {
			db.Close()
			return fmt.Errorf("database %q does not exist in catalog %q", c.configProp.Database, catalog)
		}
	}

	c.db = db

	// Initialize Arrow Flight SQL client if transport is "flight_sql"
	if c.configProp.Transport == "flight_sql" {
		if err := c.initFlightClient(); err != nil {
			db.Close()
			return fmt.Errorf("failed to initialize Arrow Flight SQL client: %w", err)
		}
	}

	return nil
}

// buildDSN constructs the MySQL DSN from configuration.
func (c *connection) buildDSN() string {
	// If DSN is provided, use it as-is (MySQL format expected)
	if c.configProp.DSN != "" {
		return c.configProp.DSN
	}

	// Build DSN from individual fields
	// Note: We don't set DBName because external catalogs (Iceberg, Hive, etc.)
	// require SET CATALOG before accessing databases. All queries use fully
	// qualified table names (catalog.database.table) instead.
	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%d", c.configProp.Host, c.configProp.Port)
	cfg.User = c.configProp.Username
	cfg.Passwd = c.configProp.Password
	cfg.ParseTime = true
	cfg.Loc = time.UTC

	// Set timeouts to prevent connection issues
	// timeout: connection timeout (30 seconds)
	// readTimeout: read timeout (300 seconds for long-running queries)
	// writeTimeout: write timeout (300 seconds, matching readTimeout for consistency)
	cfg.Timeout = 30 * time.Second
	cfg.ReadTimeout = 300 * time.Second
	cfg.WriteTimeout = 300 * time.Second

	if c.configProp.SSL {
		cfg.TLSConfig = "true"
	}

	return cfg.FormatDSN()
}
