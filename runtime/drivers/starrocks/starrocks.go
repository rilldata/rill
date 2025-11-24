package starrocks

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
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
			NoPrompt:    true,
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
	// Database is the default database to use.
	Database string `mapstructure:"database"`
	// SSL enables TLS encryption.
	SSL bool `mapstructure:"ssl"`
	// LogQueries enables SQL query logging.
	LogQueries bool `mapstructure:"log_queries"`
}

// Validate checks the configuration for errors.
func (c *ConfigProperties) Validate() error {
	if c.DSN != "" {
		if c.Host != "" || c.Port != 0 || c.Username != "" || c.Password != "" || c.Database != "" {
			return errors.New("invalid config: DSN is set but other connection fields are also set")
		}
		return nil
	}

	if c.Host == "" {
		return errors.New("invalid config: host is required when DSN is not provided")
	}

	return nil
}

// ResolveDSN builds a connection string from individual fields if DSN is not set.
func (c *ConfigProperties) ResolveDSN() (string, error) {
	if c.DSN != "" {
		return c.DSN, nil
	}

	var userInfo *url.Userinfo
	if c.Username != "" {
		if c.Password != "" {
			userInfo = url.UserPassword(c.Username, c.Password)
		} else {
			userInfo = url.User(c.Username)
		}
	}

	host := c.Host
	port := c.Port

	// Check if host already contains a port (e.g., "192.168.0.232:9030")
	// If not, append the port to the host
	if !strings.Contains(host, ":") {
		if port == 0 {
			port = 9030 // StarRocks default MySQL protocol port
		}
		host = fmt.Sprintf("%s:%d", host, port)
	}

	var path string
	if c.Database != "" {
		path = "/" + c.Database
	}

	u := &url.URL{
		Scheme: "mysql",
		User:   userInfo,
		Host:   host,
		Path:   path,
	}

	return encodeSpecialChars(u.String()), nil
}

// encodeSpecialChars encodes & and = characters to their hex codes.
// Required for proper DSN parsing.
func encodeSpecialChars(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == '&' || b == '=' {
			buf.WriteString(fmt.Sprintf("%%%02X", b))
		} else {
			buf.WriteByte(b)
		}
	}
	return buf.String()
}

// resolveGoMySQLDSN converts the DSN to Go MySQL driver format.
func (c *ConfigProperties) resolveGoMySQLDSN() (string, error) {
	dsn, err := c.ResolveDSN()
	if err != nil {
		return "", err
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("invalid DSN: %w", err)
	}

	var user, pass string
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	addr := u.Host
	dbName := strings.TrimPrefix(u.Path, "/")

	cfg := mysql.NewConfig()
	cfg.User = user
	cfg.Passwd = pass
	cfg.Net = "tcp"
	cfg.Addr = addr
	cfg.DBName = dbName

	// Configure SSL/TLS
	if c.SSL {
		cfg.TLSConfig = "skip-verify"
	}

	// StarRocks specific settings
	cfg.AllowNativePasswords = true
	cfg.InterpolateParams = true

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
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	if opts.InputHandle == c && opts.OutputHandle == c {
		// Self-to-self execution: both input and output are StarRocks
		return &selfToSelfExecutor{c: c}, nil
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

	dsn, err := c.configProp.resolveGoMySQLDSN()
	if err != nil {
		c.dbErr = err
		return nil, c.dbErr
	}

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
