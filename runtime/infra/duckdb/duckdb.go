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
	worker *priorityworker.PriorityWorker[*job]
}

type job struct {
	stmt   *infra.Statement
	result *sqlx.Rows
}

func (c *connection) Execute(ctx context.Context, stmt *infra.Statement) (*sqlx.Rows, error) {
	j := &job{
		stmt: stmt,
	}

	err := c.worker.Process(ctx, stmt.Priority, j)
	if err != nil {
		if err == priorityworker.ErrStopped {
			return nil, infra.ErrClosed
		}
		return nil, err
	}

	return j.result, nil
}

func (c *connection) executeQuery(ctx context.Context, j *job) error {
	if j.stmt.DryRun {
		// TODO: Find way to validate with args
		prepared, err := c.db.PrepareContext(ctx, j.stmt.Query)
		if err != nil {
			return err
		}
		prepared.Close()
		return nil
	}

	rows, err := c.db.QueryxContext(ctx, j.stmt.Query, j.stmt.Args...)
	j.result = rows
	return err
}

func (c *connection) InformationSchema() string {
	return ""
}

func (c *connection) Close() error {
	c.worker.Stop()
	return c.db.Close()
}
