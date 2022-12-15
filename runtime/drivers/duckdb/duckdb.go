package duckdb

import (
	"github.com/jmoiron/sqlx"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/priorityworker"

	// Load duckdb driver
	_ "github.com/marcboeker/go-duckdb"
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

	bootQueries := []string{
		"INSTALL 'json'",
		"LOAD 'json'",
		"INSTALL 'parquet'",
		"LOAD 'parquet'",
		"INSTALL 'httpfs'",
		"LOAD 'httpfs'",
		"SET max_expression_depth TO 250",
	}
	connectionPool := NewConnectionPool(cfg.PoolSize)
	for i := 0; i < cfg.PoolSize; i++ {
		db, err := sqlx.Open("duckdb", cfg.DSN)
		if err != nil {
			return nil, err
		}
		// database/sql has a built-in connection pool, but DuckDB loads extensions on a per-connection basis.
		// So we allow only one open connection at a time. In the future, we may instead consider using db.Conn()
		// and building our own connection pool to work around DuckDB's idiosyncracies.
		db.SetMaxOpenConns(1)
		for _, qry := range bootQueries {
			_, err = db.Exec(qry)
			if err != nil {
				return nil, err
			}
		}
		connectionPool.enqueue(db)
	}

	conn := &connection{connectionPool: connectionPool}
	conn.worker = priorityworker.New(conn.executeQuery, cfg.PoolSize)

	return conn, nil
}

type connection struct {
	connectionPool *ConnectionPool
	worker         *priorityworker.PriorityWorker[*job]
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	c.worker.Stop()
	close(c.connectionPool.dbChan)
	for db := range c.connectionPool.dbChan {
		err := db.Close()
		if err != nil {
			return err
		}
	}
	return nil
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

type ConnectionPool struct {
	dbChan chan *sqlx.DB
}

func NewConnectionPool(numConnections int) *ConnectionPool {
	dbChan := make(chan *sqlx.DB, numConnections)
	return &ConnectionPool{dbChan: dbChan}
}

// enqueue adds a DB handle to the buffered channel, it will block if the channel is full which should not happen in
// normal scenarios. Make sure to enqueue() after dequeue() to ensure the DB handle is returned to the pool.
func (p *ConnectionPool) enqueue(db *sqlx.DB) {
	if db != nil {
		p.dbChan <- db
	}
}

// dequeue removes a DB handle from the buffered channel, it will block if the channel is empty.
// Make sure to enqueue() after dequeue() to ensure the DB handle is returned to the pool.
func (p *ConnectionPool) dequeue() (*sqlx.DB, error) {
	db, ok := <-p.dbChan
	if !ok {
		return nil, drivers.ErrClosed
	}
	return db, nil
}
