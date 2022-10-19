package duckdb

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/file"
	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) Ingest(ctx context.Context, source *connectors.Source) error {
	err := source.Validate()
	if err != nil {
		return err
	}

	// Driver-specific overrides
	switch source.Connector {
	case "file":
		return c.ingestFile(ctx, source)
	}

	// TODO: Use generic connectors.Consume when it's implemented
	return connectors.ConsumeAsFile(ctx, source, func(filename string) error {
		return c.ingestFromRawFile(ctx, source, filename)
	})
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

func (c *connection) ingestFromRawFile(ctx context.Context, source *connectors.Source, filename string) error {
	rows, err := c.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM '%s');", source.Name, filename),
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
