package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"

	// Load postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
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
			Key:         "pg_database_url",
			DisplayName: "Postgress Connection String",
			Type:        drivers.StringPropertyType,
			Required:    false,
			Href:        "https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING",
			Placeholder: "postgresql://user:password@localhost/mydb",
			Hint:        "Either set this or configure PG_DATABASE_URL env variable.",
		},
	},
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:    "pg_database_url",
			Secret: true,
		},
	},
}

type driver struct{}

func (d driver) Open(config map[string]any, shared bool, client activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	dsn, ok := config["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("require dsn to open sqlite connection")
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &connection{
		db:     db,
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

type connection struct {
	db     *sqlx.DB
	config map[string]any
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (*connection) MigrationStatus(ctx context.Context) (current int, desired int, err error) {
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

// OLAP implements drivers.Connection.
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
