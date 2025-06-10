package clickhouse

import (
	"context"
	"database/sql"
	sqldriver "database/sql/driver"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcboeker/go-duckdb/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

type localFileInputProps struct {
	Format string `mapstructure:"format"`
}

type localFileToSelfExecutor struct {
	fileStore drivers.Handle
	c         *Connection
}

var _ drivers.ModelExecutor = &selfToSelfExecutor{}

func (e *localFileToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return desired, true
	}
	return _defaultConcurrentInserts, true
}

func (e *localFileToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Parse the input and output properties
	inputProps := &localFileInputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}

	// We need to check a few things from the output properties, so we parse them here.
	// However, we don't do full validation yet, which is delayed until we call into the self executor at the end of this function.
	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if outputProps.Materialize != nil && !*outputProps.Materialize {
		return nil, fmt.Errorf("models with input connector `local_file` must be materialized")
	}
	// NOTE: We don't need to set the default materialize here because the self executor materializes as a table by default when InsertSQLs are used.

	// Get the local file path(s)
	from, ok := e.fileStore.AsFileStore()
	if !ok {
		return nil, fmt.Errorf("input handle %q does not implement filestore", opts.InputHandle.Driver())
	}
	localPaths, err := from.FilePaths(ctx, opts.InputProperties)
	if err != nil {
		return nil, err
	}
	if len(localPaths) == 0 {
		return nil, fmt.Errorf("no files to ingest")
	}

	// Infer the format if not provided
	if inputProps.Format == "" {
		inputProps.Format, err = fileExtToFormat(fileutil.FullExt(localPaths[0]))
		if err != nil {
			return nil, fmt.Errorf("failed to infer format: %w", err)
		}
	}

	// Infer the output schema if not provided.
	// The self executor does not support using InsertSQLs without an explicit schema.
	if outputProps.Columns == "" {
		outputProps.Columns, err = e.inferColumns(ctx, opts, inputProps.Format, localPaths)
		if err != nil {
			return nil, fmt.Errorf("failed to infer columns: %w", err)
		}
	}

	// Create insert statements
	insertSQLs := make([]string, len(localPaths))
	for i, path := range localPaths {
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", path, err)
		}

		sql := fmt.Sprintf("FORMAT %s\n", inputProps.Format) + string(contents)
		insertSQLs[i] = sql
	}

	// Build input properties for the self executor
	newInputProps := &ModelInputProperties{InsertSQLs: insertSQLs}
	newInputPropsMap := make(map[string]any)
	if err := mapstructure.Decode(newInputProps, &newInputPropsMap); err != nil {
		return nil, err
	}

	// Build the model executor options with updated input and output properties
	clone := *opts
	clone.InputProperties = newInputPropsMap
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *localFileToSelfExecutor) inferColumns(ctx context.Context, opts *drivers.ModelExecuteOptions, format string, localPaths []string) (string, error) {
	tempDir, err := os.MkdirTemp(opts.TempDir, "duckdb")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)
	connector, err := duckdb.NewConnector(filepath.Join(tempDir, "temp.db?threads=1&max_memory=256MB"), func(execer sqldriver.ExecerContext) error {
		_, err = execer.ExecContext(ctx, "SET autoinstall_known_extensions=1; SET autoload_known_extensions=1;", nil)
		return err
	})
	if err != nil {
		return "", fmt.Errorf("failed to create duckdb connector: %w", err)
	}
	defer connector.Close()

	db := sql.OpenDB(connector)
	defer db.Close()

	src, err := sourceReader(localPaths, format)
	if err != nil {
		return "", fmt.Errorf("failed to create source reader: %w", err)
	}

	// identify the columns and types
	rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT column_name, column_type FROM (DESCRIBE SELECT * FROM %s)", src))
	if err != nil {
		return "", fmt.Errorf("failed to describe table: %w", err)
	}
	defer rows.Close()

	var columns strings.Builder
	columns.WriteString("(")
	var name, typ string
	for rows.Next() {
		if err := rows.Scan(&name, &typ); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}
		if columns.Len() > 1 {
			columns.WriteString(", ")
		}
		columns.WriteString(fmt.Sprintf("%s %s", safeSQLName(name), typeFromDuckDBType(typ)))
	}
	if rows.Err() != nil {
		return "", fmt.Errorf("failed to iterate rows: %w", rows.Err())
	}
	columns.WriteString(")")
	return columns.String(), nil
}

func sourceReader(paths []string, format string) (string, error) {
	// Generate a "read" statement
	if containsAny(format, []string{"CSV", "TabSeparated"}) {
		// CSV reader
		return fmt.Sprintf("read_csv_auto(%s)", convertToStatementParamsStr(paths)), nil
	} else if strings.Contains(format, "Parquet") {
		// Parquet reader
		return fmt.Sprintf("read_parquet(%s)", convertToStatementParamsStr(paths)), nil
	} else if containsAny(format, []string{"JSON", "JSONEachRow"}) {
		// JSON reader
		return fmt.Sprintf("read_json_auto(%s)", convertToStatementParamsStr(paths)), nil
	}
	return "", fmt.Errorf("file type not supported : %s", format)
}

func containsAny(s string, targets []string) bool {
	for _, target := range targets {
		if strings.Contains(s, target) {
			return true
		}
	}
	return false
}

func convertToStatementParamsStr(paths []string) string {
	return fmt.Sprintf("['%s']", strings.Join(paths, "','"))
}

func typeFromDuckDBType(typ string) string {
	switch strings.ToLower(typ) {
	case "boolean":
		return "Bool"
	case "bigint":
		return "BIGINT"
	case "double":
		return "Float64"
	case "time":
		return "String"
	case "date":
		return "Date"
	case "varchar":
		return "String"
	case "timestamp":
		return "DateTime"
	default:
		return "String"
	}
}

func fileExtToFormat(ext string) (string, error) {
	switch ext {
	case ".csv":
		return "CSV", nil
	case ".tsv":
		return "TabSeparated", nil
	case ".txt":
		return "CSV", nil
	case ".parquet":
		return "Parquet", nil
	case ".json":
		return "JSON", nil
	case ".ndjson":
		return "JSONEachRow", nil
	default:
		return "", fmt.Errorf("unsupported file extension: %s, must be one of ['.csv', '.tsv', '.txt', '.parquet', '.json', '.ndjson'] for models that ingest from 'local_file' into 'clickhouse'", ext)
	}
}
