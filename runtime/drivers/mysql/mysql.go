package mysql

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
	drivers.Register("mysql", driver{})
	drivers.RegisterAsConnector("mysql", driver{})
}

var spec = drivers.Spec{
	DisplayName: "MySQL",
	Description: "Connect to MySQL.",
	DocsURL:     "https://docs.rilldata.com/build/connectors/data-source/mysql",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "MySQL Connection String",
			Required:    true,
			DocsURL:     "https://dev.mysql.com/doc/refman/8.4/en/connecting-using-uri-or-key-value-pairs.html#connecting-using-uri",
			Placeholder: "mysql://user:password@host:3306/my-db",
			Hint:        "Can be configured here or by setting the 'connector.mysql.dsn' environment variable (using '.env' or '--env')",
			Secret:      true,
		},
		{
			Key:         "user",
			Type:        drivers.StringPropertyType,
			DisplayName: "Username",
			Placeholder: "mysql",
			Required:    true,
			Hint:        "MySQL username for authentication",
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			DisplayName: "Password",
			Placeholder: "your_password",
			Hint:        "MySQL password for authentication",
			Secret:      true,
		},
		{
			Key:         "host",
			Type:        drivers.StringPropertyType,
			DisplayName: "Host",
			Placeholder: "localhost",
			Required:    true,
			Hint:        "MySQL server hostname or IP address",
		},
		{
			Key:         "port",
			Type:        drivers.StringPropertyType,
			DisplayName: "Port",
			Placeholder: "3306",
			Default:     "3306",
			Hint:        "MySQL server port (default is 3306)",
		},

		{
			Key:         "database",
			Type:        drivers.StringPropertyType,
			DisplayName: "Database",
			Placeholder: "your_database",
			Required:    true,
			Hint:        "Name of the MySQL database to connect to",
		},
		{
			Key:         "ssl-mode",
			Type:        drivers.StringPropertyType,
			DisplayName: "SSL Mode",
			Placeholder: "require",
			Hint:        "Options include disabled, preferred or required",
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
	SSLMode    string `mapstructure:"ssl-mode"`
	LogQueries bool   `mapstructure:"log_queries"`
}

func (c *ConfigProperties) ResolveDSN() (string, error) {
	if c.DSN != "" {
		if c.Host != "" || c.Port != 0 || c.Database != "" || c.User != "" || c.Password != "" || c.SSLMode != "" {
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
	if c.Port != 0 {
		host = fmt.Sprintf("%s:%d", host, c.Port)
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

	dsn := enocodeExtra(u.String())
	query := url.Values{}
	if c.SSLMode != "" {
		query.Set("ssl-mode", c.SSLMode)
		dsn = fmt.Sprintf("%s?%s", dsn, query.Encode())
	}
	return dsn, nil
}

// enocodeExtra convert & and = in the string to it's hex code. Required because duckdb does not handle properly.
func enocodeExtra(s string) string {
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

func (c *ConfigProperties) resolveGoFormatDSN() (string, error) {
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

	q := u.Query()
	sslMode := strings.ToUpper(q.Get("ssl-mode"))

	var tlsConfig string
	switch sslMode {
	case "":
		// If no ssl-mode provided, use default (no TLSConfig set)
	case "DISABLED":
		tlsConfig = "false"
	case "PREFERRED":
		tlsConfig = "preferred"
	case "REQUIRED":
		tlsConfig = "skip-verify"
	default:
		return "", fmt.Errorf("unsupported ssl-mode: %s", sslMode)
	}

	cfg := mysql.NewConfig()
	cfg.User = user
	cfg.Passwd = pass
	cfg.Net = "tcp"
	cfg.Addr = addr
	cfg.DBName = dbName
	cfg.TLSConfig = tlsConfig
	return cfg.FormatDSN(), nil
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("mysql driver can't be shared")
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
		return fmt.Errorf("failed to open MySQL connection: %w", err)
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
	return "mysql"
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
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
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

	c.db, c.dbErr = sqlx.Open("mysql", dsn)
	if c.dbErr != nil {
		return nil, c.dbErr
	}
	c.db.SetConnMaxIdleTime(time.Minute)

	return c.db, nil
}
