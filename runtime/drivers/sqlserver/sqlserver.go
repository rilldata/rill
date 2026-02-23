package sqlserver

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"time"

	_ "github.com/microsoft/go-mssqldb"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

func init() {
	drivers.Register("sqlserver", driver{})
	drivers.RegisterAsConnector("sqlserver", driver{})
}

var spec = drivers.Spec{
	DisplayName: "SQL Server",
	Description: "Connect to SQL Server.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connectors/data-source/sqlserver",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "SQL Server Connection String",
			Required:    true,
			Placeholder: "mssql://user:password@host:1433/my-db",
			Hint:        "Can be configured here or by setting the 'connector.sqlserver.dsn' environment variable (using '.env' or '--env')",
			Secret:      true,
		},
		{
			Key:         "user",
			Type:        drivers.StringPropertyType,
			DisplayName: "Username",
			Placeholder: "sa",
			Required:    true,
			Hint:        "SQL Server username for authentication",
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			DisplayName: "Password",
			Placeholder: "your_password",
			Hint:        "SQL Server password for authentication",
			Secret:      true,
		},
		{
			Key:         "host",
			Type:        drivers.StringPropertyType,
			DisplayName: "Host",
			Placeholder: "localhost",
			Required:    true,
			Hint:        "SQL Server hostname or IP address",
		},
		{
			Key:         "port",
			Type:        drivers.StringPropertyType,
			DisplayName: "Port",
			Placeholder: "1433",
			Default:     "1433",
			Hint:        "SQL Server port (default is 1433)",
		},
		{
			Key:         "database",
			Type:        drivers.StringPropertyType,
			DisplayName: "Database",
			Placeholder: "your_database",
			Required:    true,
			Hint:        "Name of the SQL Server database to connect to",
		},
		{
			Key:         "encrypt",
			Type:        drivers.BooleanPropertyType,
			DisplayName: "Encrypt",
			Default:     "false",
			Hint:        "Encrypt the connection using TLS",
		},
		{
			Key:         "log_queries",
			Type:        drivers.BooleanPropertyType,
			DisplayName: "Log Queries",
			Default:     "false",
			Hint:        "Enable logging of all SQL queries (useful for debugging)",
		},
	},
	ImplementsSQLStore: true,
	ImplementsOLAP:     true,
}

type driver struct{}

type ConfigProperties struct {
	DSN        string `mapstructure:"dsn"`
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Database   string `mapstructure:"database"`
	User       string `mapstructure:"user"`
	Password   string `mapstructure:"password"`
	Encrypt    bool   `mapstructure:"encrypt"`
	LogQueries bool   `mapstructure:"log_queries"`
}

// ResolveDSN returns a DSN in mssql:// format for DuckDB's mssql extension.
func (c *ConfigProperties) ResolveDSN() (string, error) {
	if c.DSN != "" {
		if c.Host != "" || c.Port != 0 || c.Database != "" || c.User != "" || c.Password != "" {
			return "", fmt.Errorf("invalid config: DSN is set but other connection fields are also set")
		}
		return c.DSN, nil
	}

	var userInfo *url.Userinfo
	if c.User != "" {
		if c.Password != "" {
			userInfo = url.UserPassword(c.User, c.Password)
		} else {
			userInfo = url.User(c.User)
		}
	}

	host := c.Host
	if host == "" {
		host = "localhost"
	}
	port := c.Port
	if port == 0 {
		port = 1433
	}
	host = fmt.Sprintf("%s:%d", host, port)

	var path string
	if c.Database != "" {
		path = "/" + c.Database
	}

	u := &url.URL{
		Scheme: "mssql",
		User:   userInfo,
		Host:   host,
		Path:   path,
	}

	dsn := u.String()
	if c.Encrypt {
		dsn = fmt.Sprintf("%s?encrypt=true", dsn)
	}
	return dsn, nil
}

// resolveGoFormatDSN returns a DSN in sqlserver:// format for the Go mssqldb driver.
func (c *ConfigProperties) resolveGoFormatDSN() (string, error) {
	if c.DSN != "" {
		// Rewrite mssql:// to sqlserver:// for the Go driver
		u, err := url.Parse(c.DSN)
		if err != nil {
			return "", fmt.Errorf("invalid DSN: %w", err)
		}
		u.Scheme = "sqlserver"
		// Move database from path to query param if present
		if u.Path != "" && u.Path != "/" {
			q := u.Query()
			q.Set("database", u.Path[1:]) // strip leading /
			u.RawQuery = q.Encode()
			u.Path = ""
		}
		return u.String(), nil
	}

	host := c.Host
	if host == "" {
		host = "localhost"
	}
	port := c.Port
	if port == 0 {
		port = 1433
	}

	var userInfo *url.Userinfo
	if c.User != "" {
		if c.Password != "" {
			userInfo = url.UserPassword(c.User, c.Password)
		} else {
			userInfo = url.User(c.User)
		}
	}

	q := url.Values{}
	if c.Database != "" {
		q.Set("database", c.Database)
	}
	if c.Encrypt {
		q.Set("encrypt", "true")
	}

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     userInfo,
		Host:     fmt.Sprintf("%s:%d", host, port),
		RawQuery: q.Encode(),
	}
	return u.String(), nil
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("sqlserver driver can't be shared")
	}

	conf := &ConfigProperties{}
	if err := mapstructure.WeakDecode(config, conf); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &connection{
		config:     config,
		logger:     logger,
		logQueries: conf.LogQueries,
		dbMu:       semaphore.NewWeighted(1),
	}, nil
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
	config     map[string]any
	logger     *zap.Logger
	logQueries bool

	db    *sqlx.DB // lazily populated using getDB
	dbErr error
	dbMu  *semaphore.Weighted
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to open SQL Server connection: %w", err)
	}
	return db.PingContext(ctx)
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "sqlserver"
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	return maps.Clone(c.config)
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// AsRegistry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Connection.
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

// AsOLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// AsInformationSchema implements drivers.Connection.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return c, true
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
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

	conf := &ConfigProperties{}
	if err := mapstructure.WeakDecode(c.config, conf); err != nil {
		c.dbErr = fmt.Errorf("failed to decode config: %w", err)
		return nil, c.dbErr
	}

	dsn, err := conf.resolveGoFormatDSN()
	if err != nil {
		c.dbErr = err
		return nil, c.dbErr
	}

	c.db, c.dbErr = sqlx.Open("sqlserver", dsn)
	if c.dbErr != nil {
		return nil, c.dbErr
	}
	c.db.SetConnMaxIdleTime(time.Minute)

	return c.db, nil
}
