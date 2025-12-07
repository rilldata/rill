package starrocks

import (
	"context"
	"errors"
	"fmt"
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
	if c.DSN == "" && c.Host == "" {
		return errors.New("either DSN or Host must be provided")
	}
	return nil
}

const (
	defaultCatalog = "default_catalog"
	defaultPort    = 9030
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

	conn := &connection{
		configProp: cfg,
		logger:     logger,
		activity:   ac,
		// Single concurrent query allowed for connection affinity
		querySem: semaphore.NewWeighted(10),
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

	// db is lazily initialized
	db *sqlx.DB
	// querySem limits concurrent queries
	querySem *semaphore.Weighted
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
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
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
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// getDB returns the database connection, creating it if necessary.
func (c *connection) getDB(ctx context.Context) (*sqlx.DB, error) {
	if c.db != nil {
		return c.db, nil
	}

	dsn, err := c.buildDSN()
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN: %w", err)
	}

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(20)
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
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	c.db = db
	return db, nil
}

// buildDSN constructs the MySQL DSN from configuration.
func (c *connection) buildDSN() (string, error) {
	// If DSN is provided, use it directly (with potential conversion)
	if c.configProp.DSN != "" {
		return c.convertDSN(c.configProp.DSN)
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
	// writeTimeout: write timeout (30 seconds)
	cfg.Timeout = 30 * time.Second
	cfg.ReadTimeout = 300 * time.Second
	cfg.WriteTimeout = 30 * time.Second

	if c.configProp.SSL {
		cfg.TLSConfig = "true"
	}

	return cfg.FormatDSN(), nil
}

// convertDSN converts StarRocks URL format to MySQL DSN format if needed.
// StarRocks URL: starrocks://user:password@host:port/database
// MySQL DSN: user:password@tcp(host:port)/
// Note: Database name is stripped from DSN because external catalogs require
// SET CATALOG before accessing databases. All queries use fully qualified names.
func (c *connection) convertDSN(dsn string) (string, error) {
	// Default timeout parameters
	timeoutParams := "timeout=30s&readTimeout=300s&writeTimeout=30s&parseTime=true"

	// If already in MySQL format, strip database name if present
	if !strings.HasPrefix(dsn, "starrocks://") {
		// For MySQL format, strip database name after the closing parenthesis
		// e.g., "user:pass@tcp(host:port)/dbname" -> "user:pass@tcp(host:port)/"
		if idx := strings.LastIndex(dsn, ")/"); idx != -1 {
			// Keep everything up to ")/" and add "?" for params
			rest := dsn[idx+2:]
			if qIdx := strings.Index(rest, "?"); qIdx != -1 {
				// Preserve existing parameters and add timeout parameters if not present
				existingParams := rest[qIdx+1:]
				if !strings.Contains(existingParams, "timeout=") {
					if strings.Contains(existingParams, "parseTime") {
						// Replace parseTime if present
						existingParams = strings.ReplaceAll(existingParams, "parseTime=true", "")
						existingParams = strings.Trim(existingParams, "&")
						if existingParams != "" {
							dsn = dsn[:idx+2] + "?" + timeoutParams + "&" + existingParams
						} else {
							dsn = dsn[:idx+2] + "?" + timeoutParams
						}
					} else {
						dsn = dsn[:idx+2] + "?" + timeoutParams + "&" + existingParams
					}
				} else {
					// Timeout already present, keep as is
					dsn = dsn[:idx+2] + rest[qIdx:]
				}
			} else {
				dsn = dsn[:idx+2] + "?" + timeoutParams
			}
		}
		return dsn, nil
	}

	// Parse StarRocks URL format
	dsn = strings.TrimPrefix(dsn, "starrocks://")

	// Split user:pass@host:port/db
	var userPass, hostPortDB string
	if atIdx := strings.Index(dsn, "@"); atIdx != -1 {
		userPass = dsn[:atIdx]
		hostPortDB = dsn[atIdx+1:]
	} else {
		hostPortDB = dsn
	}

	// Split host:port/db (we ignore the database name)
	var hostPort string
	if slashIdx := strings.Index(hostPortDB, "/"); slashIdx != -1 {
		hostPort = hostPortDB[:slashIdx]
		// Ignore database name - we use fully qualified table names
	} else {
		hostPort = hostPortDB
	}

	// Build MySQL DSN without database name and with timeout parameters
	var result string
	if userPass != "" {
		result = fmt.Sprintf("%s@tcp(%s)/?%s", userPass, hostPort, timeoutParams)
	} else {
		result = fmt.Sprintf("tcp(%s)/?%s", hostPort, timeoutParams)
	}

	return result, nil
}
