package sqlite

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"

	// Load sqlite driver
	_ "modernc.org/sqlite"
)

func init() {
	drivers.Register("sqlite", driver{})
	drivers.RegisterAsConnector("sqlite", driver{})
}

type driver struct{}

type configProperties struct {
	// DSN is the connection string for the SQLite database.
	DSN string `mapstructure:"dsn"`
	// ID is an optional globally unique ID for the SQLite database.
	// If provided, we'll run periodic backups of the SQLite file to object storage.
	// See connection.startBackups() for details.
	ID string `mapstructure:"id"`
}

func (d driver) Open(_ string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	// Parse config
	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}
	if conf.DSN == "" {
		return nil, fmt.Errorf("require dsn to open sqlite connection")
	}

	// The sqlite driver requires the DSN to contain "_time_format=sqlite" to support TIMESTAMP types in all timezones.
	if !strings.Contains(conf.DSN, "_time_format") {
		if strings.Contains(conf.DSN, "?") {
			conf.DSN += "&_time_format=sqlite"
		} else {
			conf.DSN += "?_time_format=sqlite"
		}
	}

	// Open DB handle
	db, err := otelsql.Open("sqlite", conf.DSN)
	if err != nil {
		return nil, err
	}
	dbx := sqlx.NewDb(db, "sqlite")
	db.SetMaxOpenConns(1)

	// Create the handle
	ctx, cancel := context.WithCancel(context.Background())
	h := &connection{
		db:       dbx,
		logger:   logger,
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		storage:  st,
		backupID: conf.ID,
	}

	// Start backups in the background (no-op if backups are not configured)
	go h.startBackups()

	return h, nil
}

func (d driver) Spec() drivers.Spec {
	return drivers.Spec{
		DisplayName: "SQLite",
		Description: "Import data from SQLite into DuckDB.",
		DocsURL:     "https://docs.rilldata.com/build/connectors/data-source/sqlite",
		// Important: Any edits to the below properties must be accompanied by changes to the client-side form validation schemas.
		SourceProperties: []*drivers.PropertySpec{
			{
				Key:         "db",
				Type:        drivers.StringPropertyType,
				Required:    true,
				DisplayName: "DB",
				Description: "Path to SQLite db file",
				Placeholder: "/path/to/sqlite.db",
			},
			{
				Key:         "table",
				Type:        drivers.StringPropertyType,
				Required:    true,
				DisplayName: "Table",
				Description: "SQLite table name",
				Placeholder: "table",
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
		ImplementsRegistry: true,
		ImplementsCatalog:  true,
	}
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return true, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type connection struct {
	db     *sqlx.DB
	logger *zap.Logger
	config map[string]any

	// Backup management.
	// See c.startBackups() for details.
	ctx      context.Context
	cancel   context.CancelFunc
	storage  *storage.Client
	backupID string
}

var _ drivers.Handle = &connection{}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "sqlite"
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	return maps.Clone(c.config)
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	c.cancel()
	return c.db.Close()
}

// AsRegistry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return c, true
}

// AsCatalogStore implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return &catalogStore{connection: c, instanceID: instanceID}, true
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

// AsInformationSchema implements drivers.Connection.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
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
