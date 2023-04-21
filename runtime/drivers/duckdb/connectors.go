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

	from, err := sourceReader(localPaths, source.Properties)
	if err != nil {
		return err
	}

	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %q AS (SELECT * FROM %s)", source.Name, from)

	return c.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1})
}

func sourceReader(paths []string, properties map[string]interface{}) (string, error) {
	format, ok := properties["format"].(string)
	if !ok {
		format = fileutil.FullExt(paths[0])
	} else {
		format = fmt.Sprintf(".%s", format)
	}
	ingestionPropPrefix := "duckdb."
	ingestionProps := make(map[string]interface{})
	for key, value := range properties {
		if strings.HasPrefix(key, ingestionPropPrefix) {
			ingestionProps[strings.TrimPrefix(key, ingestionPropPrefix)] = value
		}
	}
	pathsStr := strings.Join(paths, "','")
	if containsAny(format, []string{".csv", ".tsv", ".txt"}) {
		// auto_detect is true by default
		return fmt.Sprintf("read_csv_auto(['%s']%s)", pathsStr, convertToFunctionParamsStr(ingestionProps)), nil
	} else if strings.Contains(format, ".parquet") {
		return fmt.Sprintf("read_parquet(['%s']%s)", pathsStr, convertToFunctionParamsStr(ingestionProps)), nil
	} else if containsAny(format, []string{".json", ".ndjson"}) {
		// auto_detect is false by default so setting it to true simplifies the ingestion
		// if columns are defined then DuckDB turns the auto-detection off so no need to check this case here
		if _, autoDetectDefined := ingestionProps["auto_detect"]; !autoDetectDefined {
			ingestionProps["auto_detect"] = true
		}
		return fmt.Sprintf("read_json(['%s']%s)", pathsStr, convertToFunctionParamsStr(ingestionProps)), nil
	} else {
		return "", fmt.Errorf("file type not supported : %s", format)
	}
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

func containsAny(s string, targets []string) bool {
	source := strings.ToLower(s)
	for _, target := range targets {
		if strings.Contains(source, target) {
			return true
		}
	}
	return false
}

func convertToFunctionParamsStr(ingestionProperties map[string]interface{}) string {
	ingestionParamsStr := make([]string, 0, len(ingestionProperties))
	for key, value := range ingestionProperties {
		ingestionParamsStr = append(ingestionParamsStr, fmt.Sprintf("%s=%v", key, value))
	}
	paramsJoined := strings.Join(ingestionParamsStr, ",")
	if paramsJoined != "" {
		paramsJoined = "," + paramsJoined
	}
	return paramsJoined
}
