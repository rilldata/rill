package duckdb

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectDuckDB
}

func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	// Check not nested
	if connFromContext(ctx) != nil {
		panic("nested WithConnection")
	}

	// Acquire connection
	conn, release, err := c.acquireOLAPConn(ctx, priority)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

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
	if stmt.DryRun {
		return nil
	}
	return res.Close()
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	// We use the meta conn for dry run queries
	if stmt.DryRun {
		conn, release, err := c.acquireMetaConn(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { _ = release() }()

		// TODO: Find way to validate with args

		name := uuid.NewString()
		_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE TEMPORARY VIEW %q AS %s", name, stmt.Query))
		if err != nil {
			return nil, err
		}

		_, err = conn.ExecContext(context.Background(), fmt.Sprintf("DROP VIEW %q", name))
		return nil, err
	}

	// Acquire connection
	conn, release, err := c.acquireOLAPConn(ctx, stmt.Priority)
	if err != nil {
		return nil, err
	}
	// NOTE: We can't just "defer release()" because release() will block until rows.Close() is called.
	// We must be careful to make sure release() is called on all code paths.

	rows, err := conn.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		_ = release()
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		_ = rows.Close()
		_ = release()
		return nil, err
	}

	res := &drivers.Result{Rows: rows, Schema: schema}
	res.SetCleanupFunc(release) // Will call release when res.Close() is called.

	return res, nil
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

func (c *connection) DropDB() error {
	// ignoring close error
	c.Close()
	return os.Remove(c.config.DBFilePath)
}
