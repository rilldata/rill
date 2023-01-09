package duckdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/localfile"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
)

func (c *connection) Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	err := source.Validate()
	if err != nil {
		return err
	}

	// todo :: check if this exceptional handling can be merged
	if source.Connector == "local_file" {
		conf, err := localfile.ParseConfig(source.Properties)
		if err != nil {
			return err
		}
		// locally uploaded file
		if !fileutil.IsGlob(conf.Path) {
			return c.ingestFile(ctx, env, source)
		} else {
			return c.ingestLocalGlob(ctx, source, conf.Path)
		}
	}

	localPaths, err := connectors.ConsumeAsFile(ctx, env, source)
	if err != nil {
		return err
	}
	c.logger.Info(fmt.Sprintf("ingesting files %v", localPaths))
	defer os.RemoveAll(filepath.Dir(localPaths[0]))
	// mutliple parquet files can be loaded in single sql
	// this seems to be performing very fast as compared to appending individual files
	return c.ingestMulti(ctx, source, localPaths)
}

func (c *connection) ingestLocalGlob(ctx context.Context, source *connectors.Source, glob string) error {
	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM '%s');", source.Name, glob)
	c.logger.Info("running query %v", zap.String("query", query))
	return c.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
}

func (c *connection) ingestMulti(ctx context.Context, source *connectors.Source, filenames []string) error {
	from, err := getMultiSourceReader(filenames)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", source.Name, from)
	c.logger.Info("running query %v", zap.String("query", query))
	return c.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
}

func (c *connection) ingestFile(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	conf, err := localfile.ParseConfig(source.Properties)
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

	return c.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1})
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

func getMultiSourceReader(paths []string) (string, error) {
	ext := fileutil.FullExt(paths[0])
	dir := filepath.Dir(paths[0])
	if ext == "" {
		return "", fmt.Errorf("invalid file")
	} else if strings.Contains(ext, ".csv") || strings.Contains(ext, ".tsv") || strings.Contains(ext, ".txt") {
		return fmt.Sprintf("read_csv_auto('%s/*%s')", dir, ext), nil
	} else if strings.Contains(ext, ".parquet") {
		return fmt.Sprintf("read_parquet('%s/*%s')", dir, ext), nil
	} else {
		return "", fmt.Errorf("file type not supported : %s", ext)
	}
}
