package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"

	// Load postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

func init() {
	drivers.Register("postgres", driver{})
}

type driver struct{}

func (d driver) Open(config map[string]any, logger *zap.Logger) (drivers.Connection, error) {
	dsnConfig, ok := config["dsn"]
	if !ok {
		return nil, fmt.Errorf("require dsn to open sqlite connection")
	}

	dsn := dsnConfig.(string)
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

type connection struct {
	db     *sqlx.DB
	config map[string]any
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
func (c *connection) RegistryStore() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) CatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *connection) RepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *connection) OLAPStore() (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from drivers.Connection, to drivers.Connection) (drivers.Transporter, bool) {
	return nil, false
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsConnector implements drivers.Connection.
func (c *connection) AsConnector() (drivers.Connector, bool) {
	return nil, false
}
