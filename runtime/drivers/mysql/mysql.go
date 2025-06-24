package mysql

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

	// Load mysql driver
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	drivers.Register("mysql", driver{})
	drivers.RegisterAsConnector("mysql", driver{})
}

var spec = drivers.Spec{
	DisplayName: "MySQL",
	Description: "Connect to MySQL.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/mysql",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Placeholder: "username:password@tcp(example.com:3306)/my-db",
			Secret:      true,
		},
	},
	// Important: Any edits to the below properties must be accompanied by changes to the client-side form validation schemas.
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to extract data from MySQL.",
			Placeholder: "select * from table;",
		},
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "MySQL Connection String",
			Required:    false,
			DocsURL:     "https://github.com/go-sql-driver/mysql?tab=readme-ov-file#dsn-data-source-name",
			Placeholder: "username:password@tcp(example.com:3306)/my-db",
			Hint:        "Can be configured here or by setting the 'connector.mysql.dsn' environment variable (using '.env' or '--env')",
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
	DSN string `mapstructure:"dsn"`
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("mysql driver can't be shared")
	}
	return &connection{
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

type connection struct {
	config map[string]any
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	db, err := c.getDB()
	if err != nil {
		return err
	}
	defer db.Close()
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

// InformationSchema implements drivers.Handle.
func (c *connection) InformationSchema() drivers.InformationSchema {
	return c
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
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
	return nil, false
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	return nil, false
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

// getDB opens a new sqlx.DB connection using the config.
func (c *connection) getDB() (*sqlx.DB, error) {
	conf := &ConfigProperties{}
	if err := mapstructure.WeakDecode(c.config, conf); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	if conf.DSN == "" {
		return nil, fmt.Errorf("dsn not provided")
	}
	db, err := sqlx.Open("mysql", conf.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}
	return db, nil
}
