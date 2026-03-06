package duckdb

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

type localFileToSelfExecutor struct {
	c    *connection
	from drivers.FileStore
}

var _ drivers.ModelExecutor = &localFileToSelfExecutor{}

type inputProps struct {
	InvalidateOnChange bool           `mapstructure:"invalidate_on_change"`
	Format             string         `mapstructure:"format"`
	DuckDB             map[string]any `mapstructure:"duckdb"`
}

func (p *inputProps) Validate() error {
	return nil
}

func (e *localFileToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *localFileToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	if opts.IncrementalRun {
		return nil, fmt.Errorf("duckdb: incremental models are not supported for the local_file connector")
	}

	inputProps := &inputProps{}
	unused, err := mapstructureutil.WeakDecodeWithWarnings(opts.InputProperties, inputProps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if len(unused) > 0 {
		e.c.logger.Warn("Undefined fields in input properties. Will be ignored", zap.String("model", opts.ModelName), zap.Strings("fields", unused), observability.ZapCtx(ctx))
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	outputProps := &ModelOutputProperties{}
	unused, err = mapstructureutil.WeakDecodeWithWarnings(opts.OutputProperties, outputProps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if len(unused) > 0 {
		e.c.logger.Warn("Undefined fields in output properties. Will be ignored", zap.String("model", opts.ModelName), zap.Strings("fields", unused), observability.ZapCtx(ctx))
	}
	if err := outputProps.validateAndApplyDefaults(opts, &ModelInputProperties{}, outputProps); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	materialize := true
	if outputProps.Materialize != nil {
		materialize = *outputProps.Materialize
	}

	asView := !materialize
	tableName := outputProps.Table

	// Prepare for ingesting into the staging view/table.
	// NOTE: This intentionally drops the end table if not staging changes.
	stagingTableName := tableName
	if opts.Env.StageChanges {
		stagingTableName = stagingTableNameFor(tableName)
	}
	_ = e.c.dropTable(ctx, stagingTableName)

	// get the local file path
	localPaths, err := e.from.FilePaths(ctx, opts.InputProperties)
	if err != nil {
		return nil, err
	}
	if len(localPaths) == 0 {
		return nil, fmt.Errorf("no files to ingest")
	}

	if inputProps.Format == "" {
		inputProps.Format = fileutil.FullExt(localPaths[0])
	} else {
		inputProps.Format = "." + inputProps.Format
	}

	from, err := sourceReader(localPaths, inputProps.Format, inputProps.DuckDB)
	if err != nil {
		return nil, err
	}

	// create the table
	metrics, err := e.c.createTableAsSelect(ctx, stagingTableName, "SELECT * FROM "+from, &createTableOptions{view: asView})
	if err != nil {
		_ = e.c.dropTable(ctx, stagingTableName)
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// Rename the staging table to the final table name
	if stagingTableName != tableName {
		err = e.c.forceRenameTable(ctx, stagingTableName, asView, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to rename staged model: %w", err)
		}
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Table:         tableName,
		View:          asView,
		UsedModelName: usedModelName,
	}
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	// Done
	return &drivers.ModelResult{
		Connector:    opts.OutputConnector,
		Properties:   resultPropsMap,
		Table:        tableName,
		ExecDuration: metrics.duration,
	}, nil
}
