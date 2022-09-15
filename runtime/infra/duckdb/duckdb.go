package duckdb

import (
	"context"
	"fmt"

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

func (c *connection) Close() error {
	c.worker.Stop()
	return c.db.Close()
}

type informationSchema struct {
	c *connection
}

func (c *connection) InformationSchema() infra.InformationSchema {
	return &informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context) ([]*infra.Table, error) {
	q := `
		select
			coalesce(t.table_catalog, '') as "database",
			t.table_schema as "schema",
			t.table_name as "name",
			t.table_type as "type", 
			array_agg(c.column_name order by c.ordinal_position) as "column_names",
			array_agg(c.data_type order by c.ordinal_position) as "column_types",
			array_agg(c.is_nullable = 'YES' order by c.ordinal_position) as "column_nullable"
		from information_schema.tables t
		join information_schema.columns c on t.table_schema = c.table_schema and t.table_name = c.table_name
		group by 1, 2, 3, 4
		order by 1, 2, 3, 4
	`

	rows, err := i.c.db.QueryxContext(ctx, q)
	if err != nil {
		return nil, err
	}

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

func (i informationSchema) Lookup(ctx context.Context, name string) (*infra.Table, error) {
	q := `
		select
			coalesce(t.table_catalog, '') as "database",
			t.table_schema as "schema",
			t.table_name as "name",
			t.table_type as "type", 
			array_agg(c.column_name order by c.ordinal_position) as "column_names",
			array_agg(c.data_type order by c.ordinal_position) as "column_types",
			array_agg(c.is_nullable = 'YES' order by c.ordinal_position) as "column_nullable"
		from information_schema.tables t
		join information_schema.columns c on t.table_schema = c.table_schema and t.table_name = c.table_name
		where t.table_name = ?
		group by 1, 2, 3, 4
		order by 1, 2, 3, 4
	`

	rows, err := i.c.db.QueryxContext(ctx, q, name)
	if err != nil {
		return nil, err
	}

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, infra.ErrNotFound
	}

	return tables[0], nil
}

func (i informationSchema) scanTables(rows *sqlx.Rows) ([]*infra.Table, error) {
	var res []*infra.Table

	for rows.Next() {
		var database string
		var schema string
		var name string
		var tableType string
		var columnNames []any
		var columnTypes []any
		var columnNullable []any

		err := rows.Scan(&database, &schema, &name, &tableType, &columnNames, &columnTypes, &columnNullable)
		if err != nil {
			return nil, err
		}

		t := &infra.Table{
			Database: database,
			Schema:   schema,
			Name:     name,
			Type:     tableType,
		}

		// should NEVER happen, but just to be safe
		if len(columnNames) != len(columnTypes) {
			panic(fmt.Errorf("duckdb: column slices have different length"))
		}

		for idx, colName := range columnNames {
			t.Columns = append(t.Columns, infra.Column{
				Name:     colName.(string),
				Type:     columnTypes[idx].(string),
				Nullable: columnNullable[idx].(bool),
			})
		}

		res = append(res, t)
	}

	return res, nil
}
