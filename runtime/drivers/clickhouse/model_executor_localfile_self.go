package clickhouse

import (
	"context"
	"database/sql"
	sqldriver "database/sql/driver"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

type localInputProps struct {
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
	from, ok := e.fileStore.AsFileStore()
	if !ok {
		return nil, fmt.Errorf("input handle %q does not implement filestore", opts.InputHandle.Driver())
	}

	if opts.IncrementalRun {
		return nil, fmt.Errorf("clickhouse: incremental models are not supported for local_file connector")
	}

	// Parse the input and output properties
	inputProps := &localInputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}

	// Require materialization for local_file
	if outputProps.Materialize != nil && !*outputProps.Materialize {
		return nil, fmt.Errorf("models with input connector `local_file` must be materialized")
	}
	outputProps.Materialize = boolptr(true)

	// Validate the output properties
	err := e.c.validateAndApplyDefaults(opts, nil, outputProps)
	if err != nil {
		return nil, fmt.Errorf("invalid model properties: %w", err)
	}

	// Extra validation: the model should be a table or dictionary
	if outputProps.Typ != "TABLE" && outputProps.Typ != "DICTIONARY" {
		return nil, fmt.Errorf("models with input connector `local_file` must be materialized as `TABLE` or `DICTIONARY`")
	}

	// get the local file path
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

	// Infer the schema if not provided
	if outputProps.Columns == "" {
		outputProps.Columns, err = e.inferColumns(ctx, opts, inputProps.Format, localPaths)
		if err != nil {
			return nil, fmt.Errorf("failed to infer columns: %w", err)
		}
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}
	tableName := outputProps.Table

	// Prepare for ingesting into the staging view/table.
	// NOTE: This intentionally drops the end table if not staging changes.
	stagingTableName := tableName
	if opts.Env.StageChanges || outputProps.Typ == "DICTIONARY" {
		stagingTableName = stagingTableNameFor(tableName)
	}
	_ = e.c.dropTable(ctx, stagingTableName)

	// create the table
	err = e.c.createTable(ctx, stagingTableName, "", outputProps)
	if err != nil {
		_ = e.c.dropTable(ctx, stagingTableName)
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// ingest the data
	for _, path := range localPaths {
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", path, err)
		}

		query := fmt.Sprintf("INSERT INTO %s FORMAT %s\n", safeSQLName(stagingTableName), inputProps.Format) + string(contents)
		_, err = e.c.writeDB.DB.ExecContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to insert data: %w", err)
		}
	}

	if outputProps.Typ == "DICTIONARY" {
		err = e.c.createDictionary(ctx, tableName, fmt.Sprintf("SELECT * FROM %s", safeSQLName(stagingTableName)), outputProps)
		// drop the temp table
		_ = e.c.dropTable(ctx, stagingTableName)
		if err != nil {
			return nil, fmt.Errorf("failed to create dictionary: %w", err)
		}
	} else if stagingTableName != tableName {
		// Rename the staging table to the final table name
		err = e.c.forceRenameTable(ctx, stagingTableName, false, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to rename staged model: %w", err)
		}
	}

	// Build result props
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(&ModelResultProperties{
		Table:         tableName,
		View:          false,
		Typ:           outputProps.Typ,
		UsedModelName: usedModelName,
	}, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	// Done
	return &drivers.ModelResult{
		Connector:  opts.OutputConnector,
		Properties: resultPropsMap,
		Table:      tableName,
	}, nil
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

func boolptr(b bool) *bool {
	return &b
}
