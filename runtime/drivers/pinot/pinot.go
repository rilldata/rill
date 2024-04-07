package pinot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
)

var spec = drivers.Spec{
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    true,
			Description: "Apache Pinot connection string",
			Secret:      true,
		},
	},
}

func init() {
	drivers.Register("pinot", driver{})
}

type driver struct{}

// Open a connection to Apache Pinot using HTTP API.
func (d driver) Open(config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("pinot driver can't be shared")
	}
	dsn, ok := config["dsn"].(string)
	if !ok {
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

	conn := &connection{
		db:         db,
		config:     config,
		metaClient: &http.Client{},
		baseURL:    "http://localhost:9000", // TODO parse from dsn
	}
	return conn, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
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
	db         *sqlx.DB
	config     map[string]any
	metaClient *http.Client // client for metadata operations
	baseURL    string
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
