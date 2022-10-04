package duckdb

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/priorityworker"
)

func init() {
	drivers.Register("duckdb", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (drivers.Connection, error) {
	db, err := sqlx.Open("duckdb", dsn)
	if err != nil {
		return nil, err
	}

	conn := &connection{db: db}
	conn.worker = priorityworker.New(conn.executeQuery)

	return conn, nil
}

type connection struct {
	db     *sqlx.DB
	worker *priorityworker.PriorityWorker[*job]
}

// Close implements drivers.Connection
func (c *connection) Close() error {
	c.worker.Stop()
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
	return c, true
}

// Migrate implements drivers.Connection
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection
func (c *connection) MigrationStatus(ctx context.Context) (current int, desired int, err error) {
	return 0, 0, nil
}
