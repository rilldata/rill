package snowflake

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"

	// Load database/sql driver
	_ "github.com/snowflakedb/gosnowflake"
)

func init() {
	drivers.Register("snowflake", driver{})
	drivers.RegisterAsConnector("snowflake", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Snowflake",
	Description: "Connect to Snowflake.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/snowflake",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "dsn",
			Type:   drivers.StringPropertyType,
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
			Description: "Query to extract data from Snowflake.",
			Placeholder: "select * from table",
		},
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "Snowflake Connection String",
			Required:    false,
			DocsURL:     "https://docs.rilldata.com/reference/connectors/snowflake",
			Placeholder: "<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_base64_url_encoded>",
			Hint:        "Can be configured here or by setting the 'connector.snowflake.dsn' environment variable (using '.env' or '--env')",
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
	ImplementsWarehouse: true,
}

type driver struct{}

type configProperties struct {
	DSN                string `mapstructure:"dsn"`
	ParallelFetchLimit int    `mapstructure:"parallel_fetch_limit"`
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("snowflake driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	if conf.DSN == "" {
		return nil, fmt.Errorf("dsn not provided")
	}

	// Open DB handle
	db, err := sqlx.Open("snowflake", conf.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	return &Connection{
		db:               db,
		configProperties: conf,
		storage:          st,
		logger:           logger,
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
	db               *sqlx.DB
	configProperties *configProperties
	storage          *storage.Client
	logger           *zap.Logger
}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
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
	return "snowflake"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.configProperties, &m)
	return m
}

// InformationSchema implements drivers.Handle.
func (c *Connection) InformationSchema() drivers.InformationSchema {
	return c
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	return c.db.Close()
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
	if opts.InputHandle == c {
		if store, ok := opts.OutputHandle.AsObjectStore(); ok {
			return &selfToObjectStoreExecutor{
				c:     c,
				store: store,
			}, true
		}
	}
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
	return c, true
}

// AsNotifier implements drivers.Connection.
func (c *Connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}
