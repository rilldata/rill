package duckdb

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/file"
)

type selfToFileExecutor struct {
	c    *connection
	opts *drivers.ModelExecutorOptions
}

var _ drivers.ModelExecutor = &selfToFileExecutor{}

func (e *selfToFileExecutor) Execute(ctx context.Context) (*drivers.ModelResult, error) {
	olap, ok := e.c.AsOLAP(e.c.instanceID)
	if !ok {
		return nil, fmt.Errorf("output connector is not OLAP")
	}

	inputProps := &ModelInputProperties{}
	if err := mapstructure.WeakDecode(e.opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	outputProps := &file.ModelOutputProperties{}
	if err := mapstructure.WeakDecode(e.opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if err := outputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	if e.opts.IncrementalRun {
		return nil, fmt.Errorf("duckdb-to-file executor does not support incremental runs")
	}

	sql, err := exportSQL(inputProps.SQL, outputProps.Path, outputProps.Format)
	if err != nil {
		return nil, err
	}

	err = olap.Exec(ctx, &drivers.Statement{
		Query:    sql,
		Args:     inputProps.Args,
		Priority: e.opts.Priority,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Build result props
	resultProps := &file.ModelResultProperties{
		Path:   outputProps.Path,
		Format: outputProps.Format,
	}
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}
	return &drivers.ModelResult{
		Connector:  e.opts.OutputConnector,
		Properties: resultPropsMap,
	}, nil
}

func exportSQL(qry, path string, format drivers.FileFormat) (string, error) {
	switch format {
	case drivers.FileFormatParquet:
		return fmt.Sprintf("COPY (%s\n) TO '%s' (FORMAT PARQUET)", qry, path), nil
	case drivers.FileFormatCSV:
		return fmt.Sprintf("COPY (%s\n) TO '%s' (FORMAT CSV, HEADER true)", qry, path), nil
	case drivers.FileFormatJSON:
		return fmt.Sprintf("COPY (%s\n) TO '%s' (FORMAT JSON)", qry, path), nil
	default:
		return "", fmt.Errorf("duckdb: unsupported export format %q", format)
	}
}

func supportsExportFormat(format drivers.FileFormat) bool {
	switch format {
	case drivers.FileFormatParquet, drivers.FileFormatCSV, drivers.FileFormatJSON:
		return true
	default:
		return false
	}
}
