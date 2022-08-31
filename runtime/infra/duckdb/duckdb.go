package duckdb

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/infra"
	"github.com/rilldata/rill/runtime/pkg/priorityworker"
)

func init() {
	infra.Register("duckdb", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (infra.Connection, error) {
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
	worker *priorityworker.PriorityWorker[*query]
}

type query struct {
	sql  string
	args []any
	rows *sqlx.Rows
}

func (c *connection) Execute(ctx context.Context, priority int, sql string, args ...any) (*sqlx.Rows, error) {
	q := &query{
		sql:  sql,
		args: args,
	}

	err := c.worker.Process(ctx, priority, q)
	if err != nil {
		if err == priorityworker.ErrStopped {
			return nil, infra.ErrClosed
		}
		return nil, err
	}

	return q.rows, nil
}

func (c *connection) executeQuery(ctx context.Context, q *query) error {
	rows, err := c.db.QueryxContext(ctx, q.sql, q.args...)
	q.rows = rows
	return err
}

func (c *connection) InformationSchema() string {
	return ""
}

func (c *connection) Close() error {
	c.worker.Stop()
	return c.db.Close()
}
