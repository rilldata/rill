package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"

	// Load postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

func init() {
	drivers.Register("postgres", driver{})
}

type driver struct{}

func (d driver) Open(dsn string, logger *otelzap.Logger) (drivers.Connection, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &connection{db: db}, nil
}

type connection struct {
	db *sqlx.DB
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
