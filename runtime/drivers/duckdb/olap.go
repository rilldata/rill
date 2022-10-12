package duckdb

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/file"
	"github.com/rilldata/rill/runtime/connectors/s3"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/priorityworker"
)

type job struct {
	stmt   *drivers.Statement
	result *sqlx.Rows
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (*sqlx.Rows, error) {
	j := &job{
		stmt: stmt,
	}

	err := c.worker.Process(ctx, stmt.Priority, j)
	if err != nil {
		if err == priorityworker.ErrStopped {
			return nil, drivers.ErrClosed
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

type informationSchema struct {
	c *connection
}

func (c *connection) InformationSchema() drivers.InformationSchema {
	return &informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context) ([]*drivers.Table, error) {
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

func (i informationSchema) Lookup(ctx context.Context, name string) (*drivers.Table, error) {
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
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (i informationSchema) scanTables(rows *sqlx.Rows) ([]*drivers.Table, error) {
	var res []*drivers.Table

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

		t := &drivers.Table{
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
			t.Columns = append(t.Columns, drivers.Column{
				Name:     colName.(string),
				Type:     columnTypes[idx].(string),
				Nullable: columnNullable[idx].(bool),
			})
		}

		res = append(res, t)
	}

	return res, nil
}

func (c *connection) Ingest(ctx context.Context, source *connectors.Source) error {
	err := source.Validate()
	if err != nil {
		return err
	}

	// Driver-specific overrides
	switch source.Connector {
	case "file":
		return c.ingestFile(ctx, source)
	case "s3":
		return c.ingestS3(ctx, source)
	}

	// TODO: Use generic connectors.Consume when it's implemented
	return drivers.ErrUnsupportedConnector
}

func (c *connection) ingestFile(ctx context.Context, source *connectors.Source) error {
	conf, err := file.ParseConfig(source.Properties)
	if err != nil {
		return err
	}

	// Not using query args since not quite sure about behaviour of injecting table names that way.
	// Also, it's a source, so the caller can be trusted.

	from := fmt.Sprintf("'%s'", conf.Path)
	if conf.Format == "csv" && conf.CSVDelimiter != "" {
		from = fmt.Sprintf("read_csv_auto('%s', delim='%s')", conf.Path, conf.CSVDelimiter)
	}

	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s)", source.Name, from)

	rows, err := c.Execute(ctx, &drivers.Statement{Query: qry, Priority: 1})
	if err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}

	return nil
}

func (c *connection) ingestS3(ctx context.Context, source *connectors.Source) error {
	conf, err := s3.ParseConfig(source.Properties)
	if err != nil {
		return err
	}

	// TODO: set AWS settings for the transaction only

	args := []any{conf.AWSRegion}
	qry := "SET s3_region=?;"

	if conf.AWSKey != "" && conf.AWSSecret != "" {
		qry += "SET s3_access_key_id=?;SET s3_secret_access_key=?;"
		args = append(args, conf.AWSKey, conf.AWSSecret)
	} else if conf.AWSSession != "" {
		qry += "SET s3_session_token=?;"
		args = append(args, conf.AWSSession)
	}

	qry += fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM '%s');", source.Name, conf.Path)

	rows, err := c.Execute(ctx, &drivers.Statement{
		Query:    qry,
		Args:     args,
		Priority: 1,
	})
	if err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}

	return nil
}
