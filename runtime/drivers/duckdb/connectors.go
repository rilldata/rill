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

const _iteratorBatch = 8

// Ingest data from a source with a timeout
func (c *connection) Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	timeoutInSeconds := 30
	if source.Timeout > 0 {
		timeoutInSeconds = int(source.Timeout)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)
	defer cancel()

	// Driver-specific overrides
	// switch source.Connector {
	// case "local_file":
	// 	return c.ingestFile(ctx, env, source)
	// }
	if source.Connector == "local_file" {
		return c.ingestLocalFiles(ctxWithTimeout, env, source)
	}

	iterator, err := connectors.ConsumeAsIterator(ctxWithTimeout, env, source)
	if err != nil {
		return err
	}
	defer iterator.Close()

	appendToTable := false
	for iterator.HasNext() {
		files, err := iterator.NextBatch(_iteratorBatch)
		if err != nil {
			return err
		}

		if err := c.ingestFiles(ctxWithTimeout, source, files, appendToTable); err != nil {
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
	} else if strings.Contains(format, ".json") {
		return fmt.Sprintf("read_json_objects(['%s'])", strings.Join(paths, "','")), nil
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
