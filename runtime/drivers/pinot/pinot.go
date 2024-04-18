package pinot

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/pinot/sqldriver"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"

	// Load Pinot sql driver
	_ "github.com/rilldata/rill/runtime/drivers/pinot/sqldriver"
)

func init() {
	drivers.Register("pinot", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Pinot",
	Description: "Connect to Apache Pinot.",
	DocsURL:     "https://docs.rilldata.com/reference/olap-engines/pinot",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Connection string",
			Placeholder: "http(s)://username:password@localhost:9000",
			Secret:      true,
		},
	},
	SourceProperties: nil,
	ImplementsOLAP:   true,
}

type driver struct{}

// Open a connection to Apache Pinot using HTTP API.
func (d driver) Open(config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("pinot driver can't be shared")
	}
	dsn, ok := config["dsn"].(string)
	if !ok || dsn == "" {
		return nil, fmt.Errorf("require dsn to open pinot connection")
	}

	db, err := sqlx.Open("pinot", dsn)
	if err != nil {
		return nil, err
	}

	// very roughly approximating num queries required for a typical page load
	db.SetMaxOpenConns(20)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("pinot: %w", err)
	}

	controller, headers, err := sqldriver.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	conn := &connection{
		db:      db,
		config:  config,
		baseURL: controller,
		headers: headers,
	}
	return conn, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type connection struct {
	db      *sqlx.DB
	config  map[string]any
	baseURL string
	headers map[string]string
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "pinot"
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	return c.config
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return c.db.Close()
}

func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

func (c *connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}
