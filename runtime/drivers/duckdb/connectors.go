package duckdb

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/localfile"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

// Ingest data from a source with a timeout
func (c *connection) Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	cancellableCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	timeout := 30
	if value, ok := source.Properties["timeout"]; ok {
		timeout = int(value.(float64))
	}

	channel := make(chan error, 1)
	go func() {
		err := c.ingestWithoutTimeout(cancellableCtx, env, source) //relies on duck db query cancellation to cancel the ingestion 
		channel <- err
	}()

	select {
	case result := <-channel:
		return result

	case <-time.After(time.Duration(timeout) * time.Second):
		return context.DeadlineExceeded
	}
}

func (c *connection) ingestWithoutTimeout(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	err := source.Validate()
	if err != nil {
		return err
	}

	// Driver-specific overrides
	// switch source.Connector {
	// case "local_file":
	// 	return c.ingestFile(ctx, env, source)
	// }
	if source.Connector == "local_file" {
		return c.ingestLocalFiles(ctx, env, source)
	}

	iterator, err := connectors.ConsumeAsIterator(ctx, env, source)
	if err != nil {
		return err
	}
	defer iterator.Close()

	appendToTable := false
	for iterator.HasNext() {
		files, err := iterator.NextBatch(ctx, 1)
		if err != nil {
			return err
		}

		ingestBatch := func() error {
			defer fileutil.ForceRemoveFiles(files)
			return c.ingestFiles(ctx, source, files, appendToTable)
		}

		if err := ingestBatch(); err != nil {
			return err
		}

		appendToTable = true
	}
	return nil
}

// for files downloaded locally from remote sources
func (c *connection) ingestFiles(ctx context.Context, source *connectors.Source, filenames []string, appendToTable bool) error {
	format := ""
	if value, ok := source.Properties["format"]; ok {
		format = value.(string)
	}

	delimiter := ""
	if value, ok := source.Properties["csv.delimiter"]; ok {
		format = value.(string)
	}

	hivePartition := 1
	if value, ok := source.Properties["hive_partitioning"]; ok {
		if !value.(bool) {
			hivePartition = 0
		}
	}

	from, err := sourceReader(filenames, delimiter, format, hivePartition)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", source.Name, from)
	if appendToTable {
		query = fmt.Sprintf("INSERT INTO %s (SELECT * FROM %s);", source.Name, from)
	}

	return c.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
}

// local files
func (c *connection) ingestLocalFiles(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	conf, err := localfile.ParseConfig(source.Properties)
	if err != nil {
		return err
	}

	path, err := fileutil.ExpandHome(conf.Path)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(path) {
		// If the path is relative, it's relative to the repo root
		if env.RepoDriver != "file" || env.RepoDSN == "" {
			return fmt.Errorf("file connector cannot ingest source '%s': path is relative, but repo is not available", source.Name)
		}
		path = filepath.Join(env.RepoDSN, path)
	}

	// get all files in case glob passed
	localPaths, err := doublestar.FilepathGlob(path)
	if err != nil {
		return err
	}
	if len(localPaths) == 0 {
		return fmt.Errorf("file does not exist at %s", conf.Path)
	}

	from, err := sourceReader(localPaths, conf.CSVDelimiter, conf.Format, 0)
	if err != nil {
		return err
	}

	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s)", source.Name, from)

	return c.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1})
}

func sourceReader(paths []string, csvDelimiter, format string, hivePartition int) (string, error) {
	if format == "" {
		format = fileutil.FullExt(paths[0])
	}

	if format == "" {
		return "", fmt.Errorf("invalid file")
	} else if strings.Contains(format, ".csv") || strings.Contains(format, ".tsv") || strings.Contains(format, ".txt") {
		return sourceReaderWithDelimiter(paths, csvDelimiter), nil
	} else if strings.Contains(format, ".parquet") {
		return fmt.Sprintf("read_parquet(['%s'], HIVE_PARTITIONING=%v)", strings.Join(paths, "','"), hivePartition), nil
	} else {
		return "", fmt.Errorf("file type not supported : %s", format)
	}
}

func sourceReaderWithDelimiter(paths []string, delimiter string) string {
	if delimiter == "" {
		return fmt.Sprintf("read_csv_auto(['%s'])", strings.Join(paths, "','"))
	}
	return fmt.Sprintf("read_csv_auto(['%s'], delim='%s')", strings.Join(paths, "','"), delimiter)
}
