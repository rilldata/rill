package starrocks

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

func init() {
	drivers.Register("starrocks", driver{})
	drivers.RegisterAsConnector("starrocks", driver{})
}

var spec = drivers.Spec{
	DisplayName: "StarRocks",
	Description: "Connect to StarRocks.",
	DocsURL:     "https://docs.rilldata.com/build/connectors/olap/starrocks",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "StarRocks Connection String",
			Required:    false,
			Placeholder: "starrocks://user:password@host:9030/database",
			Hint:        "Either provide a connection string or fill in the individual fields below",
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
			Required:    false,
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
	},
	ImplementsOLAP: true,
}

type driver struct{}

// ConfigProperties defines the configuration for StarRocks connection.
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
}

// Validate checks the configuration for errors.
func (c *ConfigProperties) Validate() error {
	// If DSN is provided, it takes precedence and we don't validate individual fields
	if c.DSN != "" {
		return nil
	}

	// If DSN is not provided, host is required
	if c.Host == "" {
		return errors.New("invalid config: host is required when DSN is not provided")
	}

	return nil
}

// ResolveDSN builds a connection string from individual fields if DSN is not set.
// For external catalogs, the database is NOT included in DSN because it doesn't exist in default_catalog.
// The database will be set after connection using SET CATALOG and USE database.
func (c *ConfigProperties) ResolveDSN() (string, error) {
	if c.DSN != "" {
		// Check if DSN uses starrocks:// protocol and convert to MySQL format
		dsn, err := c.parseDSN(c.DSN)
		if err != nil {
			return "", err
		}
		return dsn, nil
	}

	// Use mysql.Config to build DSN properly
	cfg := mysql.NewConfig()
	cfg.User = c.Username
	cfg.Passwd = c.Password
	cfg.Net = "tcp"

	// Set address
	if strings.Contains(c.Host, ":") {
		cfg.Addr = c.Host
	} else {
		port := c.Port
		if port == 0 {
			port = 9030 // StarRocks default MySQL protocol port
		}
		cfg.Addr = fmt.Sprintf("%s:%d", c.Host, port)
	}

	// For external catalogs (non-defaultCatalog), don't include database in DSN
	// because the database exists in the external catalog, not in defaultCatalog.
	// MySQL driver would fail to connect if database doesn't exist in defaultCatalog.
	// The database will be set after connection using SET CATALOG and USE database.
	if c.Catalog == "" || c.Catalog == defaultCatalog {
		cfg.DBName = c.Database
	}
	// For external catalogs: cfg.DBName remains empty

	// Enable parseTime for DATE/DATETIME conversion
	// Custom driver will handle edge cases where parsing fails
	cfg.ParseTime = true

	// Format DSN
	return cfg.FormatDSN(), nil
}

// parseDSN parses a StarRocks DSN and converts it to MySQL format.
// Supports both starrocks:// and standard MySQL DSN formats.
// starrocks://user:password@host:port/database -> user:password@tcp(host:port)/database?parseTime=true
// Also extracts database name and stores it in ConfigProperties.Database for later use.
func (c *ConfigProperties) parseDSN(dsn string) (string, error) {
	// If DSN doesn't start with starrocks://, assume it's already in MySQL format
	if !strings.HasPrefix(dsn, "starrocks://") {
		// Parse MySQL DSN to extract database name
		cfg, err := mysql.ParseDSN(dsn)
		if err == nil && cfg.DBName != "" && c.Database == "" {
			c.Database = cfg.DBName
		}
		return dsn, nil
	}

	// Remove starrocks:// prefix
	rest := strings.TrimPrefix(dsn, "starrocks://")

	// Split into user:password@host:port/database parts
	// Format: [user[:password]@]host[:port]/database
	var username, password, host, database string
	port := 9030

	// Find @ to separate credentials from host
	atIdx := strings.Index(rest, "@")
	var hostPart string
	if atIdx >= 0 {
		// Has credentials
		creds := rest[:atIdx]
		hostPart = rest[atIdx+1:]

		// Parse credentials
		colonIdx := strings.Index(creds, ":")
		if colonIdx >= 0 {
			username = creds[:colonIdx]
			password = creds[colonIdx+1:]
		} else {
			username = creds
		}
	} else {
		hostPart = rest
	}

	// Parse host:port/database
	slashIdx := strings.Index(hostPart, "/")
	var hostPortPart string
	if slashIdx >= 0 {
		hostPortPart = hostPart[:slashIdx]
		database = hostPart[slashIdx+1:]
	} else {
		hostPortPart = hostPart
	}

	// Parse host:port
	colonIdx := strings.LastIndex(hostPortPart, ":")
	if colonIdx >= 0 {
		host = hostPortPart[:colonIdx]
		portStr := hostPortPart[colonIdx+1:]
		if p, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
			return "", fmt.Errorf("invalid port in StarRocks DSN: %q is not a valid number: %w", portStr, err)
		} else if p != 1 {
			return "", fmt.Errorf("invalid port in StarRocks DSN: expected numeric port, got %q", portStr)
		}
	} else {
		host = hostPortPart
	}

	// Store database in ConfigProperties for later use (e.g., GetTable, Lookup)
	if database != "" && c.Database == "" {
		c.Database = database
	}

	// Build MySQL DSN using mysql.Config
	cfg := mysql.NewConfig()
	cfg.User = username
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%d", host, port)
	cfg.DBName = database
	cfg.ParseTime = true // Enable date/time parsing

	return cfg.FormatDSN(), nil
}

// Open creates a new StarRocks connection handle.
func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("starrocks driver can't be shared")
	}

	conf := &ConfigProperties{}
	if err := mapstructure.WeakDecode(config, conf); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return &connection{
		config:     config,
		configProp: conf,
		logger:     logger,
		logQueries: conf.LogQueries,
		dbMu:       semaphore.NewWeighted(1),
	}, nil
}

// Spec returns the driver specification.
func (d driver) Spec() drivers.Spec {
	return spec
}

// HasAnonymousSourceAccess checks if the source can be accessed without credentials.
func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

// TertiarySourceConnectors returns additional connectors needed by this driver.
func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

// connection implements the drivers.Handle interface for StarRocks.
type connection struct {
	config     map[string]any
	configProp *ConfigProperties
	logger     *zap.Logger
	logQueries bool

	db    *sqlx.DB // lazily populated using getDB
	dbErr error
	dbMu  *semaphore.Weighted
}

// Ping tests the connection to StarRocks.
func (c *connection) Ping(ctx context.Context) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to open StarRocks connection: %w", err)
	}
	return db.PingContext(ctx)
}

// Driver returns the driver name.
func (c *connection) Driver() string {
	return "starrocks"
}

// Config returns the connection configuration.
func (c *connection) Config() map[string]any {
	return maps.Clone(c.config)
}

// Migrate runs database migrations (no-op for StarRocks).
func (c *connection) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus returns the migration status (no-op for StarRocks).
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Close closes the database connection.
func (c *connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
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
	return c, true
}

// AsObjectStore implements drivers.Handle.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
// Supports both same-connector and cross-connector (StarRocks→StarRocks) execution.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	// Output must be this connector (follows ClickHouse pattern)
	if opts.OutputHandle != c {
		return nil, drivers.ErrNotImplemented
	}

	// Case 1: Self-to-self execution (same connector instance)
	if opts.InputHandle == c {
		return &selfToSelfExecutor{c: c}, nil
	}

	// Case 2: StarRocks → StarRocks (different connector, e.g., external catalog → default catalog)
	if opts.InputHandle.Driver() == "starrocks" {
		inputConn, ok := opts.InputHandle.(*connection)
		if !ok {
			return nil, fmt.Errorf("invalid input handle type for StarRocks connector")
		}
		return &starrocksToSelfExecutor{
			inputConn:  inputConn,
			outputConn: c,
		}, nil
	}

	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return c, true
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

// getDB lazily initializes and returns a database connection.
func (c *connection) getDB(ctx context.Context) (*sqlx.DB, error) {
	err := c.dbMu.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer c.dbMu.Release(1)

	if c.db != nil || c.dbErr != nil {
		return c.db, c.dbErr
	}

	dsn, err := c.configProp.ResolveDSN()
	if err != nil {
		c.dbErr = err
		return nil, c.dbErr
	}

	// Use MySQL driver directly (StarRocks is MySQL-compatible)
	// Type conversions are handled in the OLAP layer
	c.db, c.dbErr = sqlx.Open("mysql", dsn)
	if c.dbErr != nil {
		return nil, c.dbErr
	}

	// Configure connection pool
	c.db.SetMaxOpenConns(10)
	c.db.SetMaxIdleConns(5)
	c.db.SetConnMaxIdleTime(time.Minute)
	c.db.SetConnMaxLifetime(5 * time.Minute)

	return c.db, nil
}
