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
	var tableSchema map[string]string

	// parse required properties from source.Properties
	allowColAddition, allowColRelaxation, err := schemaRelaxationProperties(source.Properties)
	if err != nil {
		return nil, err
	}

	format, formatDefined := source.Properties["format"].(string)
	if formatDefined {
		format = fmt.Sprintf(".%s", format)
	}

	var ingestionProps map[string]any
	if duckDBProps, ok := source.Properties["duckdb"].(map[string]any); ok {
		ingestionProps = duckDBProps
	} else {
		ingestionProps = map[string]any{}
	}

	for iterator.HasNext() {
		files, err := iterator.NextBatch(_iteratorBatch)
		if err != nil {
			return nil, err
		}

		if !formatDefined {
			format = fileutil.FullExt(files[0])
			formatDefined = true
		}

		from, err := sourceReader(files, format, ingestionProps)
		if err != nil {
			return nil, err
		}

		var query string
		if appendToTable {
			srcSchema, newSchema, err := c.updateSchema(ctx, from, files, source, tableSchema, allowColAddition, allowColRelaxation)
			if err != nil {
				return nil, fmt.Errorf("failed to update schema %w", err)
			}

			tableSchema = newSchema
			if !hasKey(ingestionProps, "columns", "types", "dtypes") && format != ".parquet" {
				// add columns and their datatypes to ensure the datatypes are not inferred again
				from, err = sourceReader(files, format, addSchemaInference(ingestionProps, srcSchema))
				if err != nil {
					return nil, err
				}
			}

			colNames := strings.Join(keys(srcSchema), ",")
			query = fmt.Sprintf("INSERT INTO %q (%s) (SELECT %s FROM %s);", source.Name, colNames, colNames, from)
		} else {
			query = fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", source.Name, from)
		}
		if err := c.Exec(ctx, &drivers.Statement{Query: query, Priority: 1}); err != nil {
			return nil, err
		}

		summary.BytesIngested += fileSize(files)
		appendToTable = true
	}
	return summary, nil
}

// updateSchema updates the schema of the table in case new file adds a new column or
// updates the datatypes of an existing columns with a wider datatype.
func (c *connection) updateSchema(ctx context.Context, from string, fileNames []string, source *connectors.Source, oldSchema map[string]string,
	allowAddition, allowRelaxation bool,
) (srcSchema, currentSchema map[string]string, err error) {
	// schema of new files
	if srcSchema, err = c.scanSchemaFromQuery(ctx, fmt.Sprintf("DESCRIBE (SELECT * FROM %s LIMIT 0);", from)); err != nil {
		return
	}

	// combined schema
	qry := fmt.Sprintf("DESCRIBE ((SELECT * FROM %s limit 0) UNION ALL BY NAME (SELECT * FROM %s limit 0));", source.Name, from)
	unionSchema, err := c.scanSchemaFromQuery(ctx, qry)
	if err != nil {
		return nil, nil, err
	}

	// current schema
	currentSchema = oldSchema
	if currentSchema == nil {
		currentSchema, err = c.scanSchemaFromQuery(ctx, fmt.Sprintf("DESCRIBE %s;", source.Name))
		if err != nil {
			return nil, nil, err
		}
	}

	newCols := make(map[string]string)
	colTypeChanged := make(map[string]string)
	for colName, colType := range unionSchema {
		oldType, ok := currentSchema[colName]
		if !ok {
			newCols[colName] = colType
		} else if oldType != colType {
			colTypeChanged[colName] = colType
		}
	}

	if !allowRelaxation {
		if len(srcSchema) < len(unionSchema) {
			c.logger.Error("new files are missing columns and column relaxation not allowed",
				zap.String("files", strings.Join(names(fileNames), ",")),
				zap.String("columns", strings.Join(missingMapKeys(unionSchema, srcSchema), ",")))
			return nil, nil, errors.New("new files are missing columns and schema relaxation not allowed")
		}

		if len(colTypeChanged) != 0 {
			c.logger.Error("new files change datatypes of some columns and column relaxation not allowed",
				zap.String("files", strings.Join(names(fileNames), ",")),
				zap.String("columns", strings.Join(keys(colTypeChanged), ",")))
			return nil, nil, errors.New("new files change datatypes of some columns and column relaxation not allowed")
		}
	}

	if len(newCols) != 0 && !allowAddition {
		c.logger.Error("new files have new columns and column addition not allowed",
			zap.String("files", strings.Join(names(fileNames), ",")),
			zap.String("columns", strings.Join(keys(newCols), ",")))
		return nil, nil, errors.New("new files have new columns and column addition not allowed")
	}

	for colName, colType := range newCols {
		currentSchema[colName] = colType
		qry := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", source.Name, colName, colType)
		if err := c.Exec(ctx, &drivers.Statement{Query: qry}); err != nil {
			return nil, nil, err
		}
	}

	for colName, colType := range colTypeChanged {
		currentSchema[colName] = colType
		qry := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DATA TYPE %s", source.Name, colName, colType)
		if err := c.Exec(ctx, &drivers.Statement{Query: qry}); err != nil {
			return nil, nil, err
		}
	}
	return srcSchema, currentSchema, nil
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

	var format string
	if conf.Format != "" {
		format = fmt.Sprintf(".%s", conf.Format)
	} else {
		format = fileutil.FullExt(localPaths[0])
	}

	var ingestionProps map[string]any
	if duckDBProps, ok := source.Properties["duckdb"].(map[string]any); ok {
		ingestionProps = duckDBProps
	} else {
		ingestionProps = map[string]any{}
	}

	// Ingest data
	from, err := sourceReader(localPaths, format, ingestionProps)
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

func (c *connection) scanSchemaFromQuery(ctx context.Context, qry string) (map[string]string, error) {
	result, err := c.Execute(ctx, &drivers.Statement{Query: qry, Priority: 1})
	if err != nil {
		return nil, err
	}
	defer result.Close()

	schema := make(map[string]string)
	for i := 0; result.Next(); i++ {
		var s duckDBTableSchemaResult
		if err := result.StructScan(&s); err != nil {
			return nil, err
		}
		schema[s.ColumnName] = s.ColumnType
	}
	return schema, nil
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
	} else {
		return "", fmt.Errorf("file type not supported : %s", format)
	}
}

func generateReadCsvStatement(paths []string, properties map[string]any) (string, error) {
	ingestionProps := copyMap(properties)
	// set sample_size to 200000 by default
	if _, sampleSizeDefined := ingestionProps["sample_size"]; !sampleSizeDefined {
		ingestionProps["sample_size"] = 200000
	}
	// set union_by_name to unify the schema of the files
	if _, defined := ingestionProps["union_by_name"]; !defined {
		ingestionProps["union_by_name"] = true
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
	// set union_by_name to unify the schema of the files
	if _, defined := ingestionProps["union_by_name"]; !defined {
		ingestionProps["union_by_name"] = true
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

func convertToStatementParamsStr(paths []string, properties map[string]any) string {
	ingestionParamsStr := make([]string, 0, len(properties)+1)
	// The first parameter is a source path
	ingestionParamsStr = append(ingestionParamsStr, fmt.Sprintf("['%s']", strings.Join(paths, "','")))
	for key, value := range properties {
		ingestionParamsStr = append(ingestionParamsStr, fmt.Sprintf("%s=%v", key, value))
	}
	return strings.Join(ingestionParamsStr, ",")
}

func schemaToDuckDBColumnsProp(schema map[string]string) string {
	var typeStr strings.Builder
	typeStr.WriteString("{")
	i := 0
	for name, dtype := range schema {
		typeStr.WriteString(fmt.Sprintf("'%s':'%s'", name, dtype))
		i++
		if i != len(schema) {
			typeStr.WriteString(",")
		}
	}
	typeStr.WriteString("}")
	return typeStr.String()
}

type duckDBTableSchemaResult struct {
	ColumnName string  `db:"column_name"`
	ColumnType string  `db:"column_type"`
	Nullable   *string `db:"null"`
	Key        *string `db:"key"`
	Default    *string `db:"default"`
	Extra      *string `db:"extra"`
}

func schemaRelaxationProperties(prop map[string]interface{}) (allowAddition, allowRelaxation bool, err error) {
	allowAddition, additionDefined := prop["allow_field_addition"].(bool)
	allowRelaxation, relaxationDefined := prop["allow_field_relaxation"].(bool)

	val, ok := prop["union_by_name"].(bool)
	if ok && !val && allowAddition {
		// if union_by_name is set as false addition can't be done
		return false, false, fmt.Errorf("if `union_by_name` is set `allow_field_addition` must be disabled")
	}

	if hasKey(prop, "columns", "types", "dtypes") && allowRelaxation {
		return false, false, fmt.Errorf("if any of `columns`,`types`,`dtypes` is set `allow_field_relaxation` must be disabled")
	}

	// set default values
	if !additionDefined {
		allowAddition = true
	}

	if !relaxationDefined {
		allowRelaxation = true
	}

	return allowAddition, allowRelaxation, nil
}

func addSchemaInference(duckDBProps map[string]interface{}, schema map[string]string) map[string]interface{} {
	// add columns and their datatypes to ensure the datatypes are not inferred again
	ingestionProps := copyMap(duckDBProps)
	ingestionProps["columns"] = schemaToDuckDBColumnsProp(schema)
	return ingestionProps
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
