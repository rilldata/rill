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
	"golang.org/x/sync/errgroup"
)

func (c *connection) Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	err := source.Validate()
	if err != nil {
		return err
	}

	// Driver-specific overrides
	// switch source.Connector {
	// case "local_file":
	// 	return c.ingestFile(ctx, env, source)
	// }

	// if source.Connector == "local_file" {
	// 	return c.ingestFile(ctx, env, source)
	// }

	result, err := connectors.FetchFileNamesForGlob(ctx, source)
	if err != nil {
		return err
	}
	defer result.Bucket.Close()
	if len(result.FileNames) == 0 {
		return fmt.Errorf("no filenames matching glob pattern")
	}

	fmt.Println(result.FileNames)

	path, err := result.DownloadObject(ctx, result.FileNames[0])
	if err != nil {
		return err
	}
	if err := c.ingestFromRawFile(ctx, source, path, true); err != nil {
		return err
	}

	batchSize := 10

	for i, file := range result.FileNames[1:] {
		g, errCtx := errgroup.WithContext(ctx)
		localFile := file
		ingest := func() error {
			fmt.Printf("started ingesting %s\n", localFile)
			path, err := result.DownloadObject(errCtx, localFile)
			if source.Connector != "local_file" {
				defer os.Remove(path)
			}
			if err != nil {
				err = fmt.Errorf("file %s download failed with error %w", localFile, err)
				fmt.Println(err)
				return err
			}
			if err = c.ingestFromRawFile(errCtx, source, path, false); err != nil {
				fmt.Printf("%s\n", err.Error())
				return err
			}
			fmt.Printf("finished ingesting %s\n", localFile)
			return nil
		}
		g.Go(ingest)
		if i%batchSize == 0 || i == len(result.FileNames)-1 {
			if err := g.Wait(); err != nil {
				return err
			}
		}
	}
	return nil
	// path, err := connectors.ConsumeAsFile(ctx, env, source)
	// if err != nil {
	// 	return err
	// }
	// defer os.Remove(path)

	// return c.ingestFromRawFile(ctx, source, path)
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

	rows, err := c.Execute(ctx, &drivers.Statement{Query: qry, Priority: 1})
	if err != nil {
		return err
	}
	err = rows.Close()
	return err
}

func (c *connection) ingestFromRawFile(ctx context.Context, source *connectors.Source, path string, createNewTable bool) error {
	from, err := getSourceReader(path)
	if err != nil {
		return err
	}
	insertStatement := ""
	if createNewTable {
		insertStatement = fmt.Sprintf("CREATE OR REPLACE TABLE %s AS", source.Name)
	} else {
		insertStatement = fmt.Sprintf("insert into %s", source.Name)
	}

	rows, err := c.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("%s (SELECT * FROM %s);", insertStatement, from),
		Priority: 1,
	})
	if err != nil {
		return err
	}
	err = rows.Close()
	return err
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
