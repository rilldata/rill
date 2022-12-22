package duckdb

import (
	"context"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectDuckDB
}

func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	// Get priority
	err := c.sem.Acquire(ctx, priority)
	if err != nil {
		return err
	}
	defer c.sem.Release()

	// Take connection from pool
	conn, release, err := c.getConn(ctx)
	if err != nil {
		return err
	}
	defer release()

	// Call fn with connection embedded in context
	wrappedCtx := contextWithConn(ctx, conn)
	ensuredCtx := contextWithConn(context.Background(), conn)
	return fn(wrappedCtx, ensuredCtx)
}

func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Execute(ctx, stmt)
	if err != nil {
		return err
	}
	return res.Close()
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	// If the call is wrapped in WithConnection, we disregard priority and execute immediately.
	// Otherwise, we go through the priority semaphore
	conn := connFromContext(ctx)
	if conn == nil {
		// Get priority
		err := c.sem.Acquire(ctx, stmt.Priority)
		if err != nil {
			return nil, err
		}
		defer c.sem.Release()

		// Take connection from pool
		connx, release, err := c.getConn(ctx)
		if err != nil {
			return nil, err
		}
		defer release()
		conn = connx
	}

	if stmt.DryRun {
		// TODO: Find way to validate with args
		prepared, err := c.metaConn.PrepareContext(ctx, stmt.Query)
		if err != nil {
			return nil, err
		}
		return nil, prepared.Close()
	}

	rows, err := conn.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		return nil, err
	}

	return &drivers.Result{Rows: rows, Schema: schema}, nil
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
