package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
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

func (d driver) Open(_ string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	dsn, ok := config["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("require dsn to open sqlite connection")
	}

	// The sqlite driver requires the DSN to contain "_time_format=sqlite" to support TIMESTAMP types in all timezones.
	if !strings.Contains(dsn, "_time_format") {
		if strings.Contains(dsn, "?") {
			dsn += "&_time_format=sqlite"
		} else {
			dsn += "?_time_format=sqlite"
		}
	}

	// Open DB handle
	db, err := otelsql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	dbx := sqlx.NewDb(db, "sqlite")
	db.SetMaxOpenConns(1)
	return &connection{
		db:     dbx,
		config: config,
	}, nil
}

func (d driver) Spec() drivers.Spec {
	return drivers.Spec{
		DisplayName: "SQLite",
		Description: "Import data from SQLite into DuckDB.",
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
	config map[string]any
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
	return c.config
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
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

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
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
