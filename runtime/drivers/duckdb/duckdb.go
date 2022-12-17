package duckdb

import (
	"context"

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

	// database/sql has a built-in connection pool, but DuckDB loads extensions on a per-connection basis,
	// which means we need to manually initialize each connection before it's used.
	// database/sql doesn't give us that flexibility, so we implement our own (very simple) pool.

	db, err := sqlx.Open("duckdb", cfg.DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.PoolSize)

	c := &connection{
		db:   db,
		pool: make(chan *sqlx.Conn, cfg.PoolSize),
		sem:  priorityqueue.NewSemaphore(cfg.PoolSize),
	}

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
		conn, err := db.Connx(context.Background())
		if err != nil {
			return nil, err
		}

		for _, qry := range bootQueries {
			_, err = conn.ExecContext(context.Background(), qry)
			if err != nil {
				return nil, err
			}
		}

		c.pool <- conn
	}

	return c, nil
}

type connection struct {
	db   *sqlx.DB
	pool chan *sqlx.Conn
	sem  *priorityqueue.Semaphore
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	close(c.pool)
	var firstErr error
	for conn := range c.pool {
		err := conn.Close()
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return firstErr
	}
	return c.db.Close()
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

// getConn gets a connection from the pool.
// It returns a function that puts the connection back in the pool if applicable.
func (c *connection) getConn(ctx context.Context) (conn *sqlx.Conn, release func(), err error) {
	// Try to get conn from context
	conn = connFromContext(ctx)
	if conn != nil {
		return conn, func() {}, nil
	}

	conn, ok := <-c.pool
	if !ok {
		return nil, nil, drivers.ErrClosed
	}
	fn := func() { c.pool <- conn }
	return conn, fn, nil
}
