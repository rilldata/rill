package duckdb

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/priorityworker"
)

type job struct {
	stmt   *drivers.Statement
	cb     func(conn *sqlx.Conn) error
	result *sqlx.Rows
}

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectDuckDB
}

func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	j := &job{
		cb: func(conn *sqlx.Conn) error {
			wrappedCtx := contextWithConn(ctx, conn)
			ensuredCtx := contextWithConn(context.Background(), conn)
			return fn(wrappedCtx, ensuredCtx)
		},
	}

	err := c.worker.Process(ctx, priority, j)
	if err != nil {
		if err == priorityworker.ErrStopped {
			return drivers.ErrClosed
		}
		return err
	}

	return nil
}

func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Execute(ctx, stmt)
	if err != nil {
		return err
	}
	return res.Close()
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	j := &job{
		stmt: stmt,
	}

	// If the call is wrapped in WithConnection, we disregard priority and execute immediately.
	// Otherwise, we use the priority worker.
	var err error
	if connFromContext(ctx) != nil {
		err = c.executeQuery(ctx, j)
	} else {
		err = c.worker.Process(ctx, stmt.Priority, j)
	}

	if err != nil {
		if errors.Is(err, priorityworker.ErrStopped) {
			return nil, drivers.ErrClosed
		}
		return nil, err
	}

	schema, err := rowsToSchema(j.result)
	if err != nil {
		return nil, err
	}

	return &drivers.Result{Rows: j.result, Schema: schema}, nil
}

func (c *connection) executeQuery(ctx context.Context, j *job) error {
	conn := connFromContext(ctx)
	if conn == nil {
		db, err := c.connectionPool.dequeue()
		if err != nil {
			return err
		}
		defer c.connectionPool.enqueue(db)

		conn, err = db.Connx(ctx)
		if err != nil {
			return err
		}
		// Note: Doesn't close the connection, just returns it to the pool.
		defer conn.Close()
	}

	if j.cb != nil {
		return j.cb(conn)
	}

	if j.stmt.DryRun {
		// TODO: Find way to validate with args
		prepared, err := conn.PrepareContext(ctx, j.stmt.Query)
		if err != nil {
			return err
		}
		return prepared.Close()
	}

	rows, err := conn.QueryxContext(ctx, j.stmt.Query, j.stmt.Args...)
	j.result = rows
	return err
}

func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	if r == nil {
		return nil, nil
	}

	cts, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(cts))
	for i, ct := range cts {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		t, err := databaseTypeToPB(ct.DatabaseTypeName(), nullable)
		if err != nil {
			return nil, err
		}

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: t,
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}
