package duckdb

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type selfToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &selfToSelfExecutor{}

func (e *selfToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *selfToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	olap, ok := e.c.AsOLAP(e.c.instanceID)
	if !ok {
		return nil, fmt.Errorf("output connector is not OLAP")
	}

	inputProps := &ModelInputProperties{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if err := outputProps.Validate(opts); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	materialize := opts.Env.DefaultMaterialize
	if outputProps.Materialize != nil {
		materialize = *outputProps.Materialize
	}

	asView := !materialize
	tableName := outputProps.Table

	var (
		stats *drivers.TableStats
		err   error
	)
	if !opts.IncrementalRun {
		// Prepare for ingesting into the staging view/table.
		// NOTE: This intentionally drops the end table if not staging changes.
		stagingTableName := tableName
		if opts.Env.StageChanges {
			stagingTableName = stagingTableNameFor(tableName)
		}
		_ = olap.DropTable(ctx, stagingTableName)

		// Create the table
		createTableOpts := &drivers.CreateTableOptions{
			View:         asView,
			BeforeCreate: inputProps.PreExec,
			AfterCreate:  inputProps.PostExec,
		}
		stats, err = olap.CreateTableAsSelect(ctx, stagingTableName, inputProps.SQL, createTableOpts)
		if err != nil {
			_ = olap.DropTable(ctx, stagingTableName)
			return nil, fmt.Errorf("failed to create model: %w", err)
		}

		// Rename the staging table to the final table name
		if stagingTableName != tableName {
			err = olapForceRenameTable(ctx, olap, stagingTableName, asView, tableName)
			if err != nil {
				return nil, fmt.Errorf("failed to rename staged model: %w", err)
			}
		}
	} else {
		// Insert into the table
		insertTableOpts := &drivers.InsertTableOptions{
			BeforeInsert: inputProps.PreExec,
			AfterInsert:  inputProps.PostExec,
			ByName:       false,
			InPlace:      true,
			Strategy:     outputProps.IncrementalStrategy,
			UniqueKey:    outputProps.UniqueKey,
		}
		stats, err = olap.InsertTableAsSelect(ctx, tableName, inputProps.SQL, insertTableOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to incrementally insert into table: %w", err)
		}
		prevResult := &ModelResultProperties{}
		err := mapstructure.Decode(opts.PreviousResult.Properties, prevResult)
		if err != nil {
			return nil, fmt.Errorf("failed to parse previous result properties: %w", err)
		}
		stats.Size += prevResult.Size
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Table:         tableName,
		View:          asView,
		UsedModelName: usedModelName,
		Size:          stats.Size,
		ExecTime:      stats.ExecTime,
	}
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
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
