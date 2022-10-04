package sqlite

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	drivers.Register("sqlite", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (drivers.Connection, error) {
	db, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return &connection{db: db}, nil
}

type connection struct {
	db *sqlx.DB
}

// Close implements drivers.Connection
func (c *connection) Close() error {
	return c.db.Close()
}

// Registry implements drivers.Connection
func (c *connection) Registry() (drivers.Registry, bool) {
	return c, true
}

// Catalog implements drivers.Connection
func (c *connection) Catalog() (drivers.Catalog, bool) {
	return nil, false
}

// Repo implements drivers.Connection
func (c *connection) Repo() (drivers.Repo, bool) {
	return nil, false
}

// OLAP implements drivers.Connection
func (c *connection) OLAP() (drivers.OLAP, bool) {
	return nil, false
}
