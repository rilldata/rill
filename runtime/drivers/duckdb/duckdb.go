package duckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"

	"github.com/rilldata/rill/runtime/drivers"

	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
)

func init() {
	drivers.Register("duckdb", Driver{})
}

type Driver struct{}

func (d Driver) Open(dsn string) (drivers.Connection, error) {
	cfg, err := newConfig(dsn)
	if err != nil {
		return nil, err
	}
	connector, err := duckdb.NewConnector(cfg.DSN, func(execer driver.Execer) error {
		bootQueries := []string{
			"INSTALL 'json'",
			"LOAD 'json'",
			"INSTALL 'parquet'",
			"LOAD 'parquet'",
			"INSTALL 'httpfs'",
			"LOAD 'httpfs'",
			"SET max_expression_depth TO 250",
		}

		for _, qry := range bootQueries {
			_, err = execer.Exec(qry, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sqlDB := sql.OpenDB(connector)
	db := sqlx.NewDb(sqlDB, "duckdb")
	db.SetMaxOpenConns(cfg.PoolSize)

	c := &connection{
		db:  db,
		sem: priorityqueue.NewSemaphore(cfg.PoolSize),
	}

	return c, nil
}

type connection struct {
	db  *sqlx.DB
	sem *priorityqueue.Semaphore
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
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
func (c *connection) getConn(ctx context.Context) (conn *sqlx.Conn, release func() error, err error) {
	// Try to get conn from context
	conn = connFromContext(ctx)
	if conn != nil {
		return conn, func() error { return nil }, nil
	}

	conn, err = c.db.Connx(ctx)
	if err != nil {
		return nil, nil, err
	}
	return conn, conn.Close, nil
}
