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
	result *sqlx.Rows
}

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectDuckDB
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	j := &job{
		stmt: stmt,
	}

	err := c.worker.Process(ctx, stmt.Priority, j)
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
	db, err := c.connectionPool.dequeue()
	defer c.connectionPool.enqueue(db)
	if err != nil {
		return err
	}
	if j.stmt.DryRun {
		// TODO: Find way to validate with args
		prepared, err := db.PrepareContext(ctx, j.stmt.Query)
		if err != nil {
			return err
		}
		err = prepared.Close()
		if err != nil {
			return err
		}
		return nil
	}
	rows, err := db.QueryxContext(ctx, j.stmt.Query, j.stmt.Args...)
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
