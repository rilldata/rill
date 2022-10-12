package duckdb

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	aws_s3 "github.com/rilldata/rill/runtime/connectors/aws-s3"
	local_file "github.com/rilldata/rill/runtime/connectors/local-file"
	"github.com/rilldata/rill/runtime/connectors/sources"
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

func (c *connection) Ingest(ctx context.Context, source sources.Source) (*sqlx.Rows, error) {
	var rows *sqlx.Rows
	var err error

	switch source.Connector {
	case sources.LocalFileConnectorName:
		rows, err = c.ingestFromFile(ctx, source)
	case sources.AWSS3ConnectorName:
		rows, err = c.ingestFromS3Bucket(ctx, source)
	default:
		err = drivers.ErrUnsupportedConnector
	}

	return rows, err
}

func (c *connection) ingestFromFile(ctx context.Context, source sources.Source) (*sqlx.Rows, error) {
	var conf local_file.LocalFileConfig
	err := mapstructure.Decode(source.Properties, &conf)
	if err != nil {
		return nil, err
	}

	return c.Execute(ctx, &drivers.Statement{
		Query:    "CREATE OR REPLACE TABLE ? AS (SELECT * FROM ?)",
		Args:     []any{source.Name, conf.Path},
		Priority: 0,
	})
}

func (c *connection) ingestFromS3Bucket(ctx context.Context, source sources.Source) (*sqlx.Rows, error) {
	var conf aws_s3.AWSS3Config
	err := mapstructure.Decode(source.Properties, conf)
	if err != nil {
		return nil, err
	}

	// TODO: set aws settings these for the transaction only
	query := "SET s3_region=?;"
	args := []any{conf.AwsRegion}

	if conf.AwsKey != "" && conf.AwsSecret != "" {
		query += "SET s3_access_key_id=?;SET s3_secret_access_key=?;"
		args = append(args, conf.AwsKey, conf.AwsSecret)
	} else if conf.AwsSession != "" {
		query += "SET s3_session_token=?;"
		args = append(args, conf.AwsSession)
	}

	query += "CREATE OR REPLACE TABLE ? AS (SELECT * FROM ?);"
	args = append(args, source.Name, conf.Path)

	return c.Execute(ctx, &drivers.Statement{
		Query: query,
		Args:  args,
	})
}
