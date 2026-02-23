package databricks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	dbsql "github.com/databricks/databricks-sql-go"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

func init() {
	drivers.Register("databricks", driver{})
	drivers.RegisterAsConnector("databricks", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Databricks",
	Description: "Connect to Databricks.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connectors/data-source/databricks",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "host",
			Type:        drivers.StringPropertyType,
			DisplayName: "Host",
			Required:    true,
			Placeholder: "adb-12345.azuredatabricks.net",
			Hint:        "Your Databricks workspace hostname.",
		},
		{
			Key:         "http_path",
			Type:        drivers.StringPropertyType,
			DisplayName: "HTTP Path",
			Required:    true,
			Placeholder: "/sql/1.0/warehouses/abc123",
			Hint:        "The HTTP path to your SQL warehouse or cluster.",
		},
		{
			Key:         "token",
			Type:        drivers.StringPropertyType,
			DisplayName: "Access Token",
			Required:    true,
			Placeholder: "dapi...",
			Hint:        "A Databricks personal access token.",
			Secret:      true,
		},
		{
			Key:         "catalog",
			Type:        drivers.StringPropertyType,
			DisplayName: "Catalog",
			Placeholder: "main",
			Hint:        "Unity Catalog catalog name. If not set, the workspace default is used.",
		},
		{
			Key:         "schema",
			Type:        drivers.StringPropertyType,
			DisplayName: "Schema",
			Placeholder: "default",
			Hint:        "Default schema within the catalog.",
		},
	},
	ImplementsWarehouse: true,
}

type driver struct{}

type configProperties struct {
	Host     string `mapstructure:"host"`
	HTTPPath string `mapstructure:"http_path"`
	Token    string `mapstructure:"token"`
	Catalog  string `mapstructure:"catalog"`
	Schema   string `mapstructure:"schema"`

	// LogQueries controls whether to log the raw SQL passed to OLAP.
	LogQueries bool `mapstructure:"log_queries"`
}

func (c *configProperties) validate() error {
	if c.Host == "" {
		return errors.New("databricks: property 'host' is required")
	}
	if c.HTTPPath == "" {
		return errors.New("databricks: property 'http_path' is required")
	}
	if c.Token == "" {
		return errors.New("databricks: property 'token' is required")
	}
	return nil
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("databricks driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}
	if err := conf.validate(); err != nil {
		return nil, err
	}

	return &connection{
		config:  conf,
		storage: st,
		logger:  logger,
		dbMu:    semaphore.NewWeighted(1),
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
	config  *configProperties
	storage *storage.Client
	logger  *zap.Logger

	db    *sqlx.DB // lazily populated via getDB
	dbErr error
	dbMu  *semaphore.Weighted
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to open databricks connection: %w", err)
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
	return "databricks"
}

// Config implements drivers.Handle.
func (c *connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, &m)
	return m
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

// getDB returns a lazily-initialized sqlx.DB for the Databricks connection.
func (c *connection) getDB(ctx context.Context) (*sqlx.DB, error) {
	err := c.dbMu.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer c.dbMu.Release(1)
	if c.db != nil || c.dbErr != nil {
		return c.db, c.dbErr
	}

	c.db, c.dbErr = c.openDB()
	return c.db, c.dbErr
}

// openDB creates a new *sqlx.DB using the Databricks connector API.
func (c *connection) openDB() (*sqlx.DB, error) {
	opts := []dbsql.ConnOption{
		dbsql.WithServerHostname(c.config.Host),
		dbsql.WithHTTPPath(c.config.HTTPPath),
		dbsql.WithAccessToken(c.config.Token),
		dbsql.WithPort(443),
	}
	if c.config.Catalog != "" || c.config.Schema != "" {
		opts = append(opts, dbsql.WithInitialNamespace(c.config.Catalog, c.config.Schema))
	}

	connector, err := dbsql.NewConnector(opts...)
	if err != nil {
		return nil, fmt.Errorf("databricks: failed to create connector: %w", err)
	}

	db := sqlx.NewDb(sql.OpenDB(connector), "databricks")
	db.SetConnMaxIdleTime(time.Minute)
	return db, nil
}

// openRawDB creates a plain *sql.DB (not wrapped in sqlx) for use in QueryAsFiles.
func (c *connection) openRawDB() (*sql.DB, error) {
	opts := []dbsql.ConnOption{
		dbsql.WithServerHostname(c.config.Host),
		dbsql.WithHTTPPath(c.config.HTTPPath),
		dbsql.WithAccessToken(c.config.Token),
		dbsql.WithPort(443),
	}
	if c.config.Catalog != "" || c.config.Schema != "" {
		opts = append(opts, dbsql.WithInitialNamespace(c.config.Catalog, c.config.Schema))
	}

	connector, err := dbsql.NewConnector(opts...)
	if err != nil {
		return nil, fmt.Errorf("databricks: failed to create connector: %w", err)
	}

	db := sql.OpenDB(connector)
	db.SetConnMaxIdleTime(time.Minute)
	return db, nil
}
