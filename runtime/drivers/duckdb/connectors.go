package duckdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/blob"
	"github.com/rilldata/rill/runtime/connectors/localfile"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// increasing this limit can increase speed ingestion
// but may increase bottleneck at duckdb or network/db IO
// set without any benchamarks
const CONCURRENT_BLOB_DOWNLOAD_LIMIT = 32

func (c *connection) Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	err := source.Validate()
	if err != nil {
		return err
	}

	// todo :: check if this exceptional handling can be merged
	// locally uploaded file
	if source.Connector == "local_file" {
		conf, err := localfile.ParseConfig(source.Properties)
		if err != nil {
			return err
		}
		if !fileutil.HasMeta(conf.Path) {
			return c.ingestFile(ctx, env, source)
		}
	}

	blobHandler, err := connectors.PrepareBlob(ctx, source)
	if err != nil {
		return err
	}
	defer blobHandler.Close()
	if len(blobHandler.FileNames) == 0 {
		return fmt.Errorf("no filenames matching glob pattern")
	}
	c.logger.Info(fmt.Sprintf("matching files %v", blobHandler.FileNames))

	// downloading first file and creating new table
	if err := c.downloadAndIngest(ctx, source, blobHandler, blobHandler.FileNames[0], true); err != nil {
		return err
	}
	// downloading other files in batch and appending to previoulsy created table
	remainingFiles := blobHandler.FileNames[1:]
	g, errCtx := errgroup.WithContext(ctx)
	for i, file := range remainingFiles {
		localFile := file
		g.Go(func() error {
			return c.downloadAndIngest(errCtx, source, blobHandler, localFile, false)
		})
		if (i+1)%CONCURRENT_BLOB_DOWNLOAD_LIMIT == 0 {
			if err := g.Wait(); err != nil {
				return err
			}
		}
	}
	return g.Wait()
}

func (c *connection) downloadAndIngest(ctx context.Context, source *connectors.Source, blobHandler *blob.BlobHandler, fileName string, createNewTable bool) error {
	c.logger.Debug("started ingesting ", zap.String("filename", fileName))
	path, err := blobHandler.DownloadObject(ctx, fileName)
	if blobHandler.BlobType != blob.File {
		defer os.Remove(path)
	}
	if err != nil {
		err = fmt.Errorf("file %s download failed with error %w", fileName, err)
		return err
	}
	if err = c.ingestFromRawFile(ctx, source, path, createNewTable); err != nil {
		return err
	}
	c.logger.Debug("finished ingesting ", zap.String("filename", fileName))
	return nil
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

	query := fmt.Sprintf("%s (SELECT * FROM %s);", insertStatement, from)
	return c.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
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
