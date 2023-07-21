package transporter

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
)

func sourceReader(paths []string, format, query string, ingestionProps map[string]any) (string, error) {
	if query != "" {
		return sourceReaderFromQuery(paths, format, query, ingestionProps)
	}

	// Generate a "read" statement
	if containsAny(format, []string{".csv", ".tsv", ".txt"}) {
		// CSV reader
		return generateReadCsvStatement(paths, getCSVIngestionProps(ingestionProps))
	} else if strings.Contains(format, ".parquet") {
		// Parquet reader
		return generateReadParquetStatement(paths, getParquetIngestionProps(ingestionProps))
	} else if containsAny(format, []string{".json", ".ndjson"}) {
		// JSON reader
		return generateReadJSONStatement(paths, getJSONIngestionProps(ingestionProps))
	} else {
		return "", fmt.Errorf("file type not supported : %s", format)
	}
}

func sourceReaderFromQuery(paths []string, format, query string, ingestionProps map[string]any) (string, error) {
	var props map[string]any
	var fn string
	if containsAny(format, []string{".csv", ".tsv", ".txt"}) {
		props = getCSVIngestionProps(ingestionProps)
		fn = "read_csv_auto"
	} else if strings.Contains(format, ".parquet") {
		props = getParquetIngestionProps(ingestionProps)
		fn = "read_parquet"
	} else if containsAny(format, []string{".json", ".ndjson"}) {
		props = getJSONIngestionProps(ingestionProps)
		fn = "read_json"
	}

	colsVal, ok := props["columns"]
	if ok {
		cols, err := parseColumns(colsVal)
		if err != nil {
			return "", err
		}
		props["columns"] = cols
	}

	// TODO: get the cached object
	ast, err := duckdbsql.Parse(query)
	if err != nil {
		return "", err
	}

	err = ast.RewriteTableRefs(func(table *duckdbsql.TableRef) (*duckdbsql.TableRef, bool) {
		if table.Function != "" {
			fn = table.Function
		}
		return &duckdbsql.TableRef{
			Function:   fn,
			Paths:      paths,
			Properties: props,
		}, true
	})
	if err != nil {
		return "", err
	}

	return ast.Format()
}

func getCSVIngestionProps(properties map[string]any) map[string]any {
	ingestionProps := copyMap(properties)
	// set sample_size to 200000 by default
	if _, sampleSizeDefined := ingestionProps["sample_size"]; !sampleSizeDefined {
		ingestionProps["sample_size"] = 200000
	}
	return ingestionProps
}

func generateReadCsvStatement(paths []string, properties map[string]any) (string, error) {
	// auto_detect (enables auto-detection of parameters) is true by default, it takes care of params/schema
	return fmt.Sprintf("select * from read_csv_auto(%s)", convertToStatementParamsStr(paths, properties)), nil
}

func getParquetIngestionProps(properties map[string]any) map[string]any {
	ingestionProps := copyMap(properties)
	// set hive_partitioning to true by default
	if _, hivePartitioningDefined := ingestionProps["hive_partitioning"]; !hivePartitioningDefined {
		ingestionProps["hive_partitioning"] = true
	}
	return ingestionProps
}

func generateReadParquetStatement(paths []string, properties map[string]any) (string, error) {
	return fmt.Sprintf("select * from read_parquet(%s)", convertToStatementParamsStr(paths, properties)), nil
}

func getJSONIngestionProps(properties map[string]any) map[string]any {
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
	return ingestionProps
}

func generateReadJSONStatement(paths []string, properties map[string]any) (string, error) {
	return fmt.Sprintf("select * from read_json(%s)", convertToStatementParamsStr(paths, properties)), nil
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

func schemaRelaxationProperty(prop map[string]interface{}) (bool, error) {
	allowSchemaRelaxation, defined := prop["allow_schema_relaxation"].(bool)
	val, ok := prop["union_by_name"].(bool)
	if ok && !val && allowSchemaRelaxation {
		// if union_by_name is set as false addition can't be done
		return false, fmt.Errorf("if `union_by_name` is set `allow_schema_relaxation` must be disabled")
	}

	if hasKey(prop, "columns", "types", "dtypes") && allowSchemaRelaxation {
		return false, fmt.Errorf("if any of `columns`,`types`,`dtypes` is set `allow_schema_relaxation` must be disabled")
	}

	// set default values
	if !defined {
		allowSchemaRelaxation = true
	}

	return allowSchemaRelaxation, nil
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

var (
	jsonKeyCorrection   = regexp.MustCompile(`(?m)([a-zA-Z0-9_]+?):`)
	jsonQuoteCorrection = regexp.MustCompile(`(?m)'`)
)

func parseColumns(colsVal any) (any, error) {
	if cv, ok := colsVal.(string); ok {
		var c map[string]any
		cv = jsonKeyCorrection.ReplaceAllString(cv, `"$1":`)
		cv = jsonQuoteCorrection.ReplaceAllString(cv, `"`)
		fmt.Println(cv)
		err := json.Unmarshal([]byte(cv), &c)
		if err != nil {
			return nil, errors.Join(errors.New("error parsing columns"), err)
		}
		return c, nil
	}
	return colsVal, nil
}
