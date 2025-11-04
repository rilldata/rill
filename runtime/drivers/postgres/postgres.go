package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	// Load postgres driver
	_ "github.com/jackc/pgx/v5/stdlib"
)

func init() {
	drivers.Register("postgres", driver{})
	drivers.RegisterAsConnector("postgres", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Postgres",
	Description: "Connect to Postgres.",
	DocsURL:     "https://docs.rilldata.com/build/connectors/data-source/postgres",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "Postgres Connection String",
			DocsURL:     "https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING",
			Placeholder: "postgresql://postgres:postgres@localhost:5432/postgres",
			Hint:        "Can be configured here or by setting the 'connector.postgres.dsn' environment variable (using '.env' or '--env').",
			Secret:      true,
		},
		{
			Key:         "host",
			Type:        drivers.StringPropertyType,
			DisplayName: "Host",
			Placeholder: "localhost",
			Required:    true,
			Hint:        "Postgres server hostname or IP address",
		},
		{
			Key:         "port",
			Type:        drivers.StringPropertyType,
			DisplayName: "Port",
			Placeholder: "5432",
			Default:     "5432",
			Hint:        "Postgres server port (default is 5432)",
		},
		{
			Key:         "user",
			Type:        drivers.StringPropertyType,
			DisplayName: "Username",
			Placeholder: "postgres",
			Required:    true,
			Hint:        "Postgres username for authentication",
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			DisplayName: "Password",
			Placeholder: "your_password",
			Hint:        "Postgres password for authentication",
			Secret:      true,
		},
		{
			Key:         "dbname",
			Type:        drivers.StringPropertyType,
			DisplayName: "Database",
			Placeholder: "postgres",
			Required:    true,
			Hint:        "Name of the Postgres database to connect to",
		},
		{
			Key:         "sslmode",
			Type:        drivers.StringPropertyType,
			DisplayName: "SSL Mode",
			Placeholder: "require",
			Hint:        "Options include disable, allow, prefer, require",
		},
	},
	ImplementsSQLStore: true,
}

type driver struct{}

type ConfigProperties struct {
	DatabaseURL string `mapstructure:"database_url"`
	DSN         string `mapstructure:"dsn"`
	Host        string `mapstructure:"host"`
	Port        string `mapstructure:"port"`
	DBname      string `mapstructure:"dbname"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	SSLMode     string `mapstructure:"sslmode"`
}

func (c *ConfigProperties) Validate() error {
	var dsn string
	if c.DSN != "" {
		dsn = c.DSN
	} else {
		dsn = c.DatabaseURL
	}

	var set []string
	if c.Host != "" {
		set = append(set, "host")
	}
	if c.Port != "" {
		set = append(set, "port")
	}
	if c.User != "" {
		set = append(set, "user")
	}
	if c.Password != "" {
		set = append(set, "password")
	}
	if c.DBname != "" {
		set = append(set, "dbname")
	}
	if c.SSLMode != "" {
		set = append(set, "sslmode")
	}
	if dsn != "" && len(set) > 0 {
		return fmt.Errorf("postgres: Only one of 'dsn' or [%s] can be set", strings.Join(set, ", "))
	}
	return nil
}

func (c *ConfigProperties) ResolveDSN() string {
	if c.DSN != "" {
		return c.DSN
	}
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	var parts []string
	if c.Host != "" {
		parts = append(parts, "host="+quotedValue(c.Host))
	}
	if c.Port != "" {
		parts = append(parts, "port="+quotedValue(c.Port))
	}
	if c.User != "" {
		parts = append(parts, "user="+quotedValue(c.User))
	}
	if c.Password != "" {
		parts = append(parts, "password="+quotedValue(c.Password))
	}
	if c.DBname != "" {
		parts = append(parts, "dbname="+quotedValue(c.DBname))
	}
	if c.SSLMode != "" {
		parts = append(parts, "sslmode="+quotedValue(c.SSLMode))
	}
	return strings.Join(parts, " ")
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("postgres driver can't be shared")
	}

	conf := &ConfigProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return &connection{
		config: conf,
		logger: logger,
		dbMu:   semaphore.NewWeighted(1),
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
	config *ConfigProperties
	logger *zap.Logger

	db    *sqlx.DB // lazily populated using getDB
	dbErr error
	dbMu  *semaphore.Weighted
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	// Open DB handle
	db, err := c.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
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
	return "postgres"
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	var m map[string]any
	_ = mapstructure.WeakDecode(c.config, &m)
	return m
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	if c.db != nil {
		c.db.Close()
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

func (c *connection) getDB(ctx context.Context) (*sqlx.DB, error) {
	err := c.dbMu.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer c.dbMu.Release(1)
	if c.db != nil || c.dbErr != nil {
		return c.db, c.dbErr
	}

	c.db, c.dbErr = sqlx.Connect("pgx", c.config.ResolveDSN())
	if c.dbErr != nil {
		return nil, c.dbErr
	}
	c.db.SetConnMaxIdleTime(time.Minute)
	return c.db, c.dbErr
}

func quotedValue(val string) string {
	// Quote if it contains special characters
	if strings.ContainsAny(val, " \t\r\n'\\=") {
		val = strings.ReplaceAll(val, `\`, `\\`)
		val = strings.ReplaceAll(val, `'`, `\'`)
		return fmt.Sprintf("'%s'", val)
	}
	return val
}
