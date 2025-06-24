package postgres

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"

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
	DocsURL:     "https://docs.rilldata.com/reference/connectors/postgres",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "database_url",
			Secret: true,
		},
	},
	// Important: Any edits to the below properties must be accompanied by changes to the client-side form validation schemas.
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to extract data from Postgres.",
			Placeholder: "select * from table;",
		},
		{
			Key:         "database_url",
			Type:        drivers.StringPropertyType,
			DisplayName: "Postgres Connection String",
			Required:    false,
			DocsURL:     "https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING",
			Placeholder: "postgresql://postgres:postgres@localhost:5432/postgres",
			Hint:        "Can be configured here or by setting the 'connector.postgres.database_url' environment variable (using '.env' or '--env')",
			Secret:      true,
		},
		{
			Key:         "name",
			Type:        drivers.StringPropertyType,
			DisplayName: "Source name",
			Description: "The name of the source",
			Placeholder: "my_new_source",
			Required:    true,
		},
	},
	ImplementsSQLStore: true,
}

type driver struct{}

type ConfigProperties struct {
	DatabaseURL string `mapstructure:"database_url"`
	DSN         string `mapstructure:"dsn"`
}

func (c *ConfigProperties) ResolveDSN() string {
	if c.DSN != "" {
		return c.DSN
	}
	return c.DatabaseURL
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("postgres driver can't be shared")
	}

	return &Connection{
		config: config,
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

type Connection struct {
	config map[string]any
}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	db, err := c.getDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return db.PingContext(ctx)
}

// Migrate implements drivers.Connection.
func (c *Connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *Connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "postgres"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	return maps.Clone(c.config)
}

// InformationSchema implements drivers.Handle.
func (c *Connection) InformationSchema() drivers.InformationSchema {
	return c
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	return nil
}

// AsRegistry implements drivers.Connection.
func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Connection.
func (c *Connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Connection.
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

// AsOLAP implements drivers.Connection.
func (c *Connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Connection.
func (c *Connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *Connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (c *Connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
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

// getDB opens a new sqlx.DB connection using the config.
func (c *Connection) getDB() (*sqlx.DB, error) {
	conf := &ConfigProperties{}
	if err := mapstructure.WeakDecode(c.config, conf); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	dsn := conf.ResolveDSN()
	if dsn == "" {
		return nil, fmt.Errorf("database_url or dsn not provided")
	}
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}
	return db, nil
}
