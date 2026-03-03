package clickhouse

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type selfToSelfExecutor struct {
	c *Connection
}

var _ drivers.ModelExecutor = &selfToSelfExecutor{}

func (e *selfToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return desired, true
	}
	return _defaultConcurrentInserts, true
}

func (e *selfToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Parse the input and output properties
	inputProps := &ModelInputProperties{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}

	// Validate the output properties
	err := e.c.validateAndApplyDefaults(opts, inputProps, outputProps)
	if err != nil {
		return nil, fmt.Errorf("invalid model properties: %w", err)
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	asView := outputProps.Typ == "VIEW"
	tableName := outputProps.Table

	var metrics *tableWriteMetrics
	if !opts.IncrementalRun {
		stagingTableName := tableName
		if opts.Env.StageChanges {
			stagingTableName = stagingTableNameFor(tableName)
		}

		// Drop the staging view/table if it exists.
		// NOTE: This intentionally drops the end table if not staging changes.
		_ = e.c.dropTable(ctx, stagingTableName)

		// Create the table
		var err error
		metrics, err = e.c.createTableAsSelect(ctx, stagingTableName, inputProps.SQL, outputProps, inputProps.PreExec, inputProps.PostExec)
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
	} else {
		// Insert into the table
		var err error
		metrics, err = e.c.insertTableAsSelect(ctx, tableName, inputProps.SQL, &InsertTableOptions{
			Strategy:     outputProps.IncrementalStrategy,
			BeforeInsert: inputProps.PreExec,
			AfterInsert:  inputProps.PostExec,
		}, outputProps)
		if err != nil {
			return nil, fmt.Errorf("failed to incrementally insert into table: %w", err)
		}
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Table:         tableName,
		View:          asView,
		Typ:           outputProps.Typ,
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
