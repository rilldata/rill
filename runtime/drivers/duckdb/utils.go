package duckdb

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

// rawConn is similar to *sql.Conn.Raw, but additionally unwraps otelsql (which we use for instrumentation).
func rawConn(conn *sql.Conn, f func(driver.Conn) error) error {
	return conn.Raw(func(raw any) error {
		// For details, see: https://github.com/XSAM/otelsql/issues/98
		if c, ok := raw.(interface{ Raw() driver.Conn }); ok {
			raw = c.Raw()
		}

		// This is currently guaranteed, but adding check to be safe
		driverConn, ok := raw.(driver.Conn)
		if !ok {
			return fmt.Errorf("internal: did not obtain a driver.Conn")
		}

		return f(driverConn)
	})
}

type sinkProperties struct {
	Table string `mapstructure:"table"`
}

func parseSinkProperties(props map[string]any) (*sinkProperties, error) {
	cfg := &sinkProperties{}
	if err := mapstructure.Decode(props, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse sink properties: %w", err)
	}
	return cfg, nil
}

type dbSourceProperties struct {
	Database string `mapstructure:"db"`
	SQL      string `mapstructure:"sql"`
}

func parseDBSourceProperties(props map[string]any) (*dbSourceProperties, error) {
	cfg := &dbSourceProperties{}
	if err := mapstructure.Decode(props, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse source properties: %w", err)
	}
	if cfg.SQL == "" {
		return nil, fmt.Errorf("property 'sql' is mandatory")
	}
	return cfg, nil
}

type fileSourceProperties struct {
	SQL                   string         `mapstructure:"sql"`
	DuckDB                map[string]any `mapstructure:"duckdb"`
	Format                string         `mapstructure:"format"`
	AllowSchemaRelaxation bool           `mapstructure:"allow_schema_relaxation"`
	BatchSize             string         `mapstructure:"batch_size"`
	CastToENUM            []string       `mapstructure:"cast_to_enum"`

	// Backwards compatibility
	HivePartitioning            *bool  `mapstructure:"hive_partitioning"`
	CSVDelimiter                string `mapstructure:"csv.delimiter"`
	IngestAllowSchemaRelaxation *bool  `mapstructure:"ingest.allow_schema_relaxation"`
}

func parseFileSourceProperties(props map[string]any) (*fileSourceProperties, error) {
	cfg := &fileSourceProperties{}
	if err := mapstructure.Decode(props, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse source properties: %w", err)
	}

	if cfg.DuckDB == nil {
		cfg.DuckDB = map[string]any{}
	}

	if cfg.HivePartitioning != nil {
		cfg.DuckDB["hive_partitioning"] = *cfg.HivePartitioning
		cfg.HivePartitioning = nil
	}

	if cfg.CSVDelimiter != "" {
		cfg.DuckDB["delim"] = fmt.Sprintf("'%v'", cfg.CSVDelimiter)
		cfg.CSVDelimiter = ""
	}

	if cfg.IngestAllowSchemaRelaxation != nil {
		cfg.AllowSchemaRelaxation = *cfg.IngestAllowSchemaRelaxation
		cfg.IngestAllowSchemaRelaxation = nil
	}

	if cfg.AllowSchemaRelaxation {
		if val, ok := cfg.DuckDB["union_by_name"].(bool); ok && !val {
			return nil, fmt.Errorf("can't set `union_by_name` and `allow_schema_relaxation` at the same time")
		}

		if hasKey(cfg.DuckDB, "columns", "types", "dtypes") {
			return nil, fmt.Errorf("if any of `columns`,`types`,`dtypes` is set `allow_schema_relaxation` must be disabled")
		}
	}
	return cfg, nil
}

func sourceReader(paths []string, format string, ingestionProps map[string]any) (string, error) {
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
	}
	return "", fmt.Errorf("file type not supported : %s", format)
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
	// set format to auto by default
	if _, formatDefined := ingestionProps["format"]; !formatDefined {
		ingestionProps["format"] = "auto"
	}
	return fmt.Sprintf("read_json(%s)", convertToStatementParamsStr(paths, ingestionProps)), nil
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

type duckDBTableSchemaResult struct {
	ColumnName string  `db:"column_name"`
	ColumnType string  `db:"column_type"`
	Nullable   *string `db:"null"`
	Key        *string `db:"key"`
	Default    *string `db:"default"`
	Extra      *string `db:"extra"`
}

// utility functions
func hasKey(m map[string]interface{}, key ...string) bool {
	for _, k := range key {
		if _, ok := m[k]; ok {
			return true
		}
	}
	return false
}

func missingMapKeys(src, lookup map[string]string) []string {
	keys := make([]string, 0)
	for k := range src {
		if _, ok := lookup[k]; !ok {
			keys = append(keys, k)
		}
	}
	return keys
}

func keys(src map[string]string) []string {
	keys := make([]string, 0, len(src))
	for k := range src {
		keys = append(keys, k)
	}
	return keys
}

func names(filePaths []string) []string {
	names := make([]string, len(filePaths))
	for i, f := range filePaths {
		names[i] = filepath.Base(f)
	}
	return names
}

// copyMap does a shallow copy of the map
func copyMap(originalMap map[string]any) map[string]any {
	newMap := make(map[string]any, len(originalMap))
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
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

func fileSize(paths []string) int64 {
	var size int64
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil { // ignoring error since only error possible is *PathError
			size += info.Size()
		}
	}
	return size
}

func quoteName(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}

func escapeDoubleQuotes(column string) string {
	return strings.ReplaceAll(column, "\"", "\"\"")
}

func safeName(name string) string {
	if name == "" {
		return name
	}
	return quoteName(escapeDoubleQuotes(name))
}

func sizeWithinStorageLimits(olap drivers.OLAPStore, size int64) bool {
	limit, ok := olap.(drivers.Handle).Config()["storage_limit_bytes"].(int64)
	if !ok || limit <= 0 { // no limit
		return true
	}

	dbSizeInBytes, ok := olap.EstimateSize()
	if ok && dbSizeInBytes+size > limit {
		return false
	}
	return true
}
