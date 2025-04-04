package duckdb

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

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

func safeName(name string) string {
	return drivers.DialectDuckDB.EscapeIdentifier(name)
}
