package druid

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
	// Load calcite avatica driver for druid
	_ "github.com/apache/calcite-avatica-go/v5"
)

var spec = drivers.Spec{
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    true,
			Description: "Druid connection string (using the Avatica protobuf endpoint)",
			Secret:      true,
		},
	},
}

func init() {
	drivers.Register("druid", driver{})
	drivers.RegisterAsConnector("druid", driver{})
}

type driver struct{}

type configProperties struct {
	// DSN is the connection string
	DSN string `mapstructure:"dsn"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.Execute.
	LogQueries bool `mapstructure:"log_queries"`
}

// Open connects to Druid using Avatica.
// Note that the Druid connection string must have the form "http://host/druid/v2/sql/avatica-protobuf/".
func (d driver) Open(confMap map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("druid driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(confMap, conf)
	if err != nil {
		return nil, err
	}

	if conf.DSN == "" {
		return nil, fmt.Errorf("no DSN provided to open the connection")
	}

	db, err := sqlx.Open("avatica", conf.DSN)
	if err != nil {
		return nil, err
	}

	// very roughly approximating num queries required for a typical page load
	db.SetMaxOpenConns(20)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("druid: %w", err)
	}

	conn := &connection{
		db:     db,
		config: conf,
		logger: logger,
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
	db     *sqlx.DB
	config *configProperties
	logger *zap.Logger
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "druid"
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, m)
	return m
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return c.db.Close()
}

// Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
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

// OLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
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

// AsSQLStore implements drivers.Connection.
// Use OLAPStore instead.
func (c *connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

func (c *connection) EstimateSize() (int64, bool) {
	return 0, false
}

func (c *connection) AcquireLongRunning(ctx context.Context) (func(), error) {
	return func() {}, nil
}
