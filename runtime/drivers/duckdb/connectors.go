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
	"go.uber.org/zap"
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
		c.logger.Error("ingestion timeout exceeded", zap.String("source", source.Name), zap.Duration("timeout", timeout))
		return nil, fmt.Errorf("ingestion timeout exceeded (source=%q, timeout=%s)", source.Name, timeout.String())
	}

	return summary, err
}

func (c *connection) ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) (*drivers.IngestionSummary, error) {
	// Driver-specific overrides
	if source.Connector == "local_file" {
		return c.ingestLocalFiles(ctx, env, source)
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

		st := time.Now()
		c.logger.Info("ingesting files", zap.Strings("files", files))
		if err := c.ingestIteratorFiles(ctx, source, files, appendToTable); err != nil {
			return nil, err
		}

		size := fileSize(files)
		summary.BytesIngested += size
		c.logger.Info("ingested files", zap.Strings("files", files), zap.Int64("bytes_ingested", size), zap.Float64("time_taken", time.Since(st).Seconds()))
		appendToTable = true
	}
	return summary, nil
}

// for files downloaded locally from remote sources
func (c *connection) ingestIteratorFiles(ctx context.Context, source *connectors.Source, filenames []string, appendToTable bool) error {
	from, err := sourceReader(filenames, source.Properties)
	if err != nil {
		return err
	}

	var query string
	if appendToTable {
		query = fmt.Sprintf("INSERT INTO %q (SELECT * FROM %s);", source.Name, from)
	} else {
		query = fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", source.Name, from)
	}
	c.logger.Debug("generated query", zap.String("query", query))
	return c.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
}

// local files
func (c *connection) ingestLocalFiles(ctx context.Context, env *connectors.Env, source *connectors.Source) (*drivers.IngestionSummary, error) {
	conf, err := localfile.ParseConfig(source.Properties)
	if err != nil {
		return nil, err
	}

	path, err := resolveLocalPath(env, conf.Path, source.Name)
	if err != nil {
		return nil, err
	}

	// get all files in case glob passed
	localPaths, err := doublestar.FilepathGlob(path)
	if err != nil {
		return nil, err
	}
	if len(localPaths) == 0 {
		return nil, fmt.Errorf("file does not exist at %s", conf.Path)
	}

	// Ingest data
	from, err := sourceReader(localPaths, source.Properties)
	if err != nil {
		return nil, err
	}
	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %q AS (SELECT * FROM %s)", source.Name, from)
	err = c.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1})
	if err != nil {
		return nil, err
	}

	bytesIngested := fileSize(localPaths)
	return &drivers.IngestionSummary{BytesIngested: bytesIngested}, nil
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

func sourceReader(paths []string, properties map[string]any) (string, error) {
	format, formatDefined := properties["format"].(string)
	if formatDefined {
		format = fmt.Sprintf(".%s", format)
	} else {
		format = fileutil.FullExt(paths[0])
	}

	var ingestionProps map[string]any
	if duckDBProps, ok := properties["duckdb"].(map[string]any); ok {
		ingestionProps = duckDBProps
	} else {
		ingestionProps = map[string]any{}
	}

	// Generate a "read" statement
	if containsAny(format, []string{".csv", ".tsv", ".txt"}) {
		// CSV reader
		return generateReadCsvStatement(paths, ingestionProps)
	} else if strings.Contains(format, ".parquet") {
		// Parquet reader
		return generateReadParquetStatement(paths, ingestionProps)
	} else if containsAny(format, []string{".json", ".ndjson"}) {
		// JSON reader
		return generateReadJSONStatement(paths, ingestionProps)
	} else {
		return "", fmt.Errorf("file type not supported : %s", format)
	}
}

func containsAny(s string, targets []string) bool {
	source := strings.ToLower(s)
	for _, target := range targets {
		if strings.Contains(source, target) {
			return true
		}
	}
	return false
}

func generateReadCsvStatement(paths []string, properties map[string]any) (string, error) {
	ingestionProps := copyMap(properties)
	// set sample_size to 200000 by default
	if _, sampleSizeDefined := ingestionProps["sample_size"]; !sampleSizeDefined {
		ingestionProps["sample_size"] = 200000
	}
	// auto_detect (enables auto-detection of parameters) is true by default, it takes care of params/schema
	return fmt.Sprintf("read_csv_auto(%s)", convertToStatementParamsStr(paths, ingestionProps)), nil
}

func generateReadParquetStatement(paths []string, properties map[string]any) (string, error) {
	ingestionProps := copyMap(properties)
	// set hive_partitioning to true by default
	if _, hivePartitioningDefined := ingestionProps["hive_partitioning"]; !hivePartitioningDefined {
		ingestionProps["hive_partitioning"] = true
	}
	return fmt.Sprintf("read_parquet(%s)", convertToStatementParamsStr(paths, ingestionProps)), nil
}

func generateReadJSONStatement(paths []string, properties map[string]any) (string, error) {
	ingestionProps := copyMap(properties)
	// auto_detect is false by default so setting it to true simplifies the ingestion
	// if columns are defined then DuckDB turns the auto-detection off so no need to check this case here
	if _, autoDetectDefined := ingestionProps["auto_detect"]; !autoDetectDefined {
		ingestionProps["auto_detect"] = true
	}
	// set sample_size to 200000 by default
	if _, sampleSizeDefined := ingestionProps["sample_size"]; !sampleSizeDefined {
		ingestionProps["sample_size"] = 200000
	}
	return fmt.Sprintf("read_json(%s)", convertToStatementParamsStr(paths, ingestionProps)), nil
}

func copyMap(originalMap map[string]any) map[string]any {
	newMap := make(map[string]any, len(originalMap))
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}

func convertToStatementParamsStr(paths []string, properties map[string]any) string {
	ingestionParamsStr := make([]string, 0, len(properties)+1)
	// The first parameter is a source path
	ingestionParamsStr = append(ingestionParamsStr, fmt.Sprintf("['%s']", strings.Join(paths, "','")))
	for key, value := range properties {
		ingestionParamsStr = append(ingestionParamsStr, fmt.Sprintf("%s=%v", key, value))
	}
	return strings.Join(ingestionParamsStr, ",")
}
