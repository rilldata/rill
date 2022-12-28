package duckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	"go.uber.org/zap"

	"github.com/rilldata/rill/runtime/drivers"

	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
)

func init() {
	drivers.Register("duckdb", Driver{})
}

type Driver struct{}

func (d Driver) Open(dsn string, logger *zap.Logger) (drivers.Connection, error) {
	cfg, err := newConfig(dsn)
	if err != nil {
		return nil, err
	}
	connector, err := duckdb.NewConnector(cfg.DSN,
		// nolint:staticcheck // TODO: remove when go-duckdb implements the driver.ExecerContext interface
		func(execer driver.Execer) error {
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
		db:     db,
		sem:    priorityqueue.NewSemaphore(cfg.PoolSize),
		logger: logger,
	}

	return c, nil
}

type connection struct {
	db     *sqlx.DB
	sem    *priorityqueue.Semaphore
	logger *zap.Logger
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
func (c *connection) getConn(ctx context.Context) (conn *sqlx.Conn, release func(), err error) {
	// Try to get conn from context
	conn = connFromContext(ctx)
	if conn != nil {
		return conn, func() {}, nil
	}

	conn, err = c.db.Connx(ctx)
	if err != nil {
		return nil, nil, err
	}
	release = func() {
		// call release in a goroutine as it will block until rows.close() is called so if a method returns rows
		// and the caller is responsible for closing them, the method will not return and cause deadlock
		go func() {
			err := conn.Close()
			if err != nil {
				c.logger.Error("error releasing connection", zap.Error(err))
			}
		}()
	}
	return conn, release, nil
}
