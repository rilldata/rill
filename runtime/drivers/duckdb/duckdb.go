package duckdb

import (
	"github.com/jmoiron/sqlx"

	"github.com/rilldata/rill/runtime/drivers"

	// Load duckdb driver
	_ "github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
)

func init() {
	drivers.Register("duckdb", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (drivers.Connection, error) {
	cfg, err := newConfig(dsn)
	if err != nil {
		return nil, err
	}

	conn := &connection{
		pool: make(chan *sqlx.DB, cfg.PoolSize),
		sem:  priorityqueue.NewSemaphore(cfg.PoolSize),
	}

	// database/sql has a built-in connection pool, but DuckDB loads extensions on a per-connection basis,
	// which means we need to manually initialize each connection before it's used.
	// database/sql doesn't give us that flexibility, so we implement our own (very simple) pool.

	bootQueries := []string{
		"INSTALL 'json'",
		"LOAD 'json'",
		"INSTALL 'parquet'",
		"LOAD 'parquet'",
		"INSTALL 'httpfs'",
		"LOAD 'httpfs'",
		"SET max_expression_depth TO 250",
	}

	for i := 0; i < cfg.PoolSize; i++ {
		db, err := sqlx.Open("duckdb", cfg.DSN)
		if err != nil {
			return nil, err
		}

		// This effectively disables the built-in pool in database/sql
		db.SetMaxOpenConns(1)

		for _, qry := range bootQueries {
			_, err = db.Exec(qry)
			if err != nil {
				return nil, err
			}
		}

		conn.pool <- db
	}

	return conn, nil
}

type connection struct {
	pool chan *sqlx.DB
	sem  *priorityqueue.Semaphore
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	close(c.pool)
	var firstErr error
	for db := range c.pool {
		err := db.Close()
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// RegistryStore Registry implements drivers.Connection.
func (c *connection) RegistryStore() (drivers.RegistryStore, bool) {
	return nil, false
}

// CatalogStore Catalog implements drivers.Connection.
func (c *connection) CatalogStore() (drivers.CatalogStore, bool) {
	return c, true
}

// RepoStore Repo implements drivers.Connection.
func (c *connection) RepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAPStore OLAP implements drivers.Connection.
func (c *connection) OLAPStore() (drivers.OLAPStore, bool) {
	return c, true
}
