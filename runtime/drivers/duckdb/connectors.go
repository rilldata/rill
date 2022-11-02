package duckdb

import (
	"context"
	"fmt"
	"os"
	"strings"

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

	path, err := connectors.ConsumeAsFile(ctx, source)
	if err != nil {
		return err
	}
	return c.ingestFromRawFile(ctx, source, path)
}

func (c *connection) ingestFile(ctx context.Context, source *connectors.Source) error {
	conf, err := file.ParseConfig(source.Properties)
	if err != nil {
		return err
	}

	// Not using query args since not quite sure about behaviour of injecting table names that way.
	// Also, it's a source, so the caller can be trusted.

	var from string
	if conf.Format == "csv" && conf.CSVDelimiter != "" {
		from = fmt.Sprintf("read_csv_auto('%s', delim='%s')", conf.Path, conf.CSVDelimiter)
	} else {
		from = getSourceReader(conf.Path)
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

func (c *connection) ingestFromRawFile(ctx context.Context, source *connectors.Source, path string) error {
	defer os.Remove(path)
	rows, err := c.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", source.Name, getSourceReader(path)),
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

func getSourceReader(path string) string {
	_, extension := connectors.SplitFileRecursive(path)
	if strings.Contains(extension, "parquet") {
		return fmt.Sprintf("read_parquet('%s')", path)
	} else {
		return fmt.Sprintf("'%s'", path)
	}
}
