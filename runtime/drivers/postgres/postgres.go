package postgres

import (
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	drivers.Register("postgres", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (drivers.Connection, error) {
	db, err := sqlx.Connect("pgx", dsn)
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
	return nil, false
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
