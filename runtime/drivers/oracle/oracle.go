package oracle

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	_ "github.com/sijms/go-ora/v2"
)

func init() {
	drivers.Register("oracle", driver{})
	drivers.RegisterAsConnector("oracle", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Oracle",
	Description: "Connect to Oracle.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connectors/data-source/oracle",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "Oracle Connection String",
			Required:    true,
			Placeholder: "oracle://user:password@host:1521/service_name",
			Hint:        "Can be configured here or by setting the 'connector.oracle.dsn' environment variable (using '.env' or '--env')",
			Secret:      true,
		},
		{
			Key:         "user",
			Type:        drivers.StringPropertyType,
			DisplayName: "Username",
			Placeholder: "system",
			Required:    true,
			Hint:        "Oracle username for authentication",
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			DisplayName: "Password",
			Placeholder: "your_password",
			Hint:        "Oracle password for authentication",
			Secret:      true,
		},
		{
			Key:         "host",
			Type:        drivers.StringPropertyType,
			DisplayName: "Host",
			Placeholder: "localhost",
			Required:    true,
			Hint:        "Oracle server hostname or IP address",
		},
		{
			Key:         "port",
			Type:        drivers.StringPropertyType,
			DisplayName: "Port",
			Placeholder: "1521",
			Default:     "1521",
			Hint:        "Oracle listener port (default is 1521)",
		},
		{
			Key:         "service_name",
			Type:        drivers.StringPropertyType,
			DisplayName: "Service Name",
			Placeholder: "ORCLPDB1",
			Required:    true,
			Hint:        "Oracle service name or SID",
		},
		{
			Key:         "log_queries",
			Type:        drivers.BooleanPropertyType,
			DisplayName: "Log Queries",
			Default:     "false",
			Hint:        "Enable logging of all SQL queries (useful for debugging)",
		},
	},
	ImplementsSQLStore:  true,
	ImplementsOLAP:      true,
	ImplementsWarehouse: true,
}

type driver struct{}

// ConfigProperties holds Oracle connection configuration.
type ConfigProperties struct {
	DSN         string `mapstructure:"dsn"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	ServiceName string `mapstructure:"service_name"`
	LogQueries  bool   `mapstructure:"log_queries"`
}

// ResolveDSN builds an Oracle connection string from the config properties.
func (c *ConfigProperties) ResolveDSN() (string, error) {
	if c.DSN != "" {
		if c.Host != "" || c.Port != 0 || c.User != "" || c.Password != "" || c.ServiceName != "" {
			return "", fmt.Errorf("invalid config: DSN is set but other connection fields are also set")
		}
		return c.DSN, nil
	}

	host := c.Host
	if host == "" {
		host = "localhost"
	}
	port := c.Port
	if port == 0 {
		port = 1521
	}

	var userInfo *url.Userinfo
	if c.User != "" {
		if c.Password != "" {
			userInfo = url.UserPassword(c.User, c.Password)
		} else {
			userInfo = url.User(c.User)
		}
	}

	u := &url.URL{
		Scheme: "oracle",
		User:   userInfo,
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/" + c.ServiceName,
	}
	return u.String(), nil
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("oracle driver can't be shared")
	}

	conf := &ConfigProperties{}
	if err := mapstructure.WeakDecode(config, conf); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &connection{
		config:     config,
		storage:    st,
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
	storage    *storage.Client
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
		return fmt.Errorf("failed to open Oracle connection: %w", err)
	}
	return db.PingContext(ctx)
}

// Migrate implements drivers.Handle.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Driver implements drivers.Handle.
func (c *connection) Driver() string {
	return "oracle"
}

// Config implements drivers.Handle.
func (c *connection) Config() map[string]any {
	return maps.Clone(c.config)
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
	return c, true
}

// AsObjectStore implements drivers.Handle.
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

// AsFileStore implements drivers.Handle.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return c, true
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

	conf := &ConfigProperties{}
	if err := mapstructure.WeakDecode(c.config, conf); err != nil {
		c.dbErr = fmt.Errorf("failed to decode config: %w", err)
		return nil, c.dbErr
	}

	dsn, err := conf.ResolveDSN()
	if err != nil {
		c.dbErr = err
		return nil, c.dbErr
	}

	c.db, c.dbErr = sqlx.Open("oracle", dsn)
	if c.dbErr != nil {
		return nil, c.dbErr
	}
	c.db.SetConnMaxIdleTime(time.Minute)

	return c.db, nil
}
