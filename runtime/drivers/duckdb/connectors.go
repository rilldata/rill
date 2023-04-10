package duckdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/localfile"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

const (
	_iteratorBatch        = 8
	_defaultIngestTimeout = 60 * time.Minute
)

// Ingest data from a source with a timeout
func (c *connection) Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) (*drivers.IngestionSummary, error) {
	// Wraps c.ingest with timeout handling

	timeout := _defaultIngestTimeout
	if source.Timeout > 0 {
		timeout = time.Duration(source.Timeout) * time.Second
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	summary, err := c.ingest(ctxWithTimeout, env, source)
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		return nil, fmt.Errorf("ingestion timeout exceeded (source=%q, timeout=%s)", source.Name, timeout.String())
	}

	return summary, err
}

func (c *connection) ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) (*drivers.IngestionSummary, error) {
	// Driver-specific overrides
	if source.Connector == "local_file" {
		err := c.ingestLocalFiles(ctx, env, source)
		if err != nil {
			return nil, err
		}
		return &drivers.IngestionSummary{}, nil
	}

	iterator, err := connectors.ConsumeAsIterator(ctx, env, source)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	appendToTable := false
	summary := &drivers.IngestionSummary{}
	for iterator.HasNext() {
		files, err := iterator.NextBatch(_iteratorBatch)
		if err != nil {
			return nil, err
		}

		if err := c.ingestIteratorFiles(ctx, source, files, appendToTable); err != nil {
			return nil, err
		}

		summary.BytesIngested += fileSize(files)
		appendToTable = true
	}
	return summary, nil
}

// for files downloaded locally from remote sources
func (c *connection) ingestIteratorFiles(ctx context.Context, source *connectors.Source, filenames []string, appendToTable bool) error {
	format := ""
	if value, ok := source.Properties["format"]; ok {
		format = value.(string)
	}

	delimiter := ""
	if value, ok := source.Properties["csv.delimiter"]; ok {
		delimiter = value.(string)
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

	var query string
	if appendToTable {
		query = fmt.Sprintf("INSERT INTO %q (SELECT * FROM %s);", source.Name, from)
	} else {
		query = fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", source.Name, from)
	}
	return c.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
}

// local files
func (c *connection) ingestLocalFiles(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	conf, err := localfile.ParseConfig(source.Properties)
	if err != nil {
		return err
	}

	path, err := resolveLocalPath(env, conf.Path, source.Name)
	if err != nil {
		return err
	}

	// get all files in case glob passed
	localPaths, err := doublestar.FilepathGlob(path)
	if err != nil {
		return err
	}
	if len(localPaths) == 0 {
		return fmt.Errorf("file does not exist at %s", conf.Path)
	}

	hivePartition := 1
	if conf.HivePartition != nil && !*conf.HivePartition {
		hivePartition = 0
	}

	from, err := sourceReader(localPaths, conf.CSVDelimiter, conf.Format, hivePartition)
	if err != nil {
		return err
	}

	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %q AS (SELECT * FROM %s)", source.Name, from)

	return c.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1})
}

func sourceReader(paths []string, csvDelimiter, format string, hivePartition int) (string, error) {
	if format == "" {
		format = fileutil.FullExt(paths[0])
	} else {
		// users will set format like csv, tsv, parquet
		// while infering format from file name extensions its better to rely on .csv, .parquet
		format = fmt.Sprintf(".%s", format)
	}

	if format == "" {
		return "", fmt.Errorf("invalid file")
	} else if strings.Contains(format, ".csv") || strings.Contains(format, ".tsv") || strings.Contains(format, ".txt") {
		return sourceReaderWithDelimiter(paths, csvDelimiter), nil
	} else if strings.Contains(format, ".parquet") {
		return fmt.Sprintf("read_parquet(['%s'], HIVE_PARTITIONING=%v)", strings.Join(paths, "','"), hivePartition), nil
	} else if strings.Contains(format, ".json") || strings.Contains(format, ".ndjson") {
		return fmt.Sprintf("read_json_auto(['%s'])", strings.Join(paths, "','")), nil
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

func fileSize(paths []string) int64 {
	var size int64
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil { // ignoring error since only error possible is *PathError
			size += info.Size()
		}
	}
	return size
}

func resolveLocalPath(env *connectors.Env, path, sourceName string) (string, error) {
	path, err := fileutil.ExpandHome(path)
	if err != nil {
		return "", err
	}

	repoRoot := env.RepoRoot
	finalPath := path
	if !filepath.IsAbs(path) {
		finalPath = filepath.Join(repoRoot, path)
	}

	if !env.AllowHostAccess && !strings.HasPrefix(finalPath, repoRoot) {
		// path is outside the repo root
		return "", fmt.Errorf("file connector cannot ingest source '%s': path is outside repo root", sourceName)
	}
	return finalPath, nil
}
