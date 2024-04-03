package postgres

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	drivers.Register("postgres", driver{})
	drivers.RegisterAsConnector("postgres", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Postgres",
	Description: "Connect to Postgres.",
	SourceProperties: []drivers.PropertySchema{
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
			DisplayName: "Postgress Connection String",
			Type:        drivers.StringPropertyType,
			Required:    false,
			Href:        "https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING",
			Placeholder: "postgresql://postgres:postgres@localhost:5432/postgres",
			Hint:        "Either set this or pass --var connector.postgres.database_url=... to rill start",
		},
	},
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:    "database_url",
			Secret: true,
		},
	},
}

type driver struct{}

func (d driver) Open(config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	// actual db connection is opened during query
	return &connection{
		config: config,
	}, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
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
	return c.config
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

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Connection.
func (c *connection) AsSQLStore() (drivers.SQLStore, bool) {
	return c, true
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(*structpb.Struct) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}
