package duckdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/file"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

func (c *connection) Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	err := source.Validate()
	if err != nil {
		return err
	}

	// Driver-specific overrides
	switch source.Connector {
	case "file":
		return c.ingestFile(ctx, env, source)
	}

	path, err := connectors.ConsumeAsFile(ctx, env, source)
	if err != nil {
		return err
	}
	defer os.Remove(path)

	return c.ingestFromRawFile(ctx, source, path)
}

func (c *connection) ingestFile(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	conf, err := file.ParseConfig(source.Properties)
	if err != nil {
		return err
	}

	path := conf.Path
	if !filepath.IsAbs(path) {
		// If the path is relative, it's relative to the repo root
		if env.RepoDriver != "file" || env.RepoDSN == "" {
			return fmt.Errorf("file connector cannot ingest source '%s': path is relative, but repo is not available", source.Name)
		}
		path = filepath.Join(env.RepoDSN, path)
	}

	// Not using query args since not quite sure about behaviour of injecting table names that way.
	// Also, it's a source, so the caller can be trusted.

	var from string
	if conf.Format == ".csv" && conf.CSVDelimiter != "" {
		from = fmt.Sprintf("read_csv_auto('%s', delim='%s')", path, conf.CSVDelimiter)
	} else {
		from, err = getSourceReader(path)
		if err != nil {
			return err
		}
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
	from, err := getSourceReader(path)
	if err != nil {
		return err
	}
	rows, err := c.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", source.Name, from),
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

func getSourceReader(path string) (string, error) {
	ext := fileutil.FullExt(path)
	if ext == "" {
		return "", fmt.Errorf("invalid file")
	} else if strings.Contains(ext, ".csv") || strings.Contains(ext, ".tsv") || strings.Contains(ext, ".txt") {
		return fmt.Sprintf("read_csv_auto('%s')", path), nil
	} else if strings.Contains(ext, ".parquet") {
		return fmt.Sprintf("read_parquet('%s')", path), nil
	} else {
		return "", fmt.Errorf("file type not supported : %s", ext)
	}
}
