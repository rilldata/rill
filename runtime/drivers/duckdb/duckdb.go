package duckdb

import (
	"github.com/jmoiron/sqlx"
	// Register some standard stuff.
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

	// database/sql has a built-in connection pool, but DuckDB loads extensions on a per-connection basis.
	// So we allow only one open connection at a time. In the future, we may instead consider using db.Conn()
	// and building our own connection pool to work around DuckDB's idiosyncracies.
	db.SetMaxOpenConns(1)

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
		_, err = db.Exec(qry)
		if err != nil {
			return nil, err
		}
	}

	conn := &connection{db: db}
	conn.worker = priorityworker.New(conn.executeQuery)

	return conn, nil
}

type connection struct {
	db     *sqlx.DB
	worker *priorityworker.PriorityWorker[*job]
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	c.worker.Stop()
	return c.db.Close()
}

// Registry implements drivers.Connection.
func (c *connection) RegistryStore() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) CatalogStore() (drivers.CatalogStore, bool) {
	return c, true
}

// Repo implements drivers.Connection.
func (c *connection) RepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *connection) OLAPStore() (drivers.OLAPStore, bool) {
	return c, true
}
