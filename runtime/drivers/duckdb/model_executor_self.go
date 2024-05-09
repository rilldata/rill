package duckdb

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type selfToSelfExecutor struct {
	c    *connection
	opts *drivers.ModelExecutorOptions
}

func (e *selfToSelfExecutor) Execute(ctx context.Context) (*drivers.ModelResult, error) {
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

	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(e.opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if err := outputProps.Validate(e.opts); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = e.opts.ModelName
		usedModelName = true
	}

	materialize := e.opts.Env.DefaultMaterialize
	if outputProps.Materialize != nil {
		materialize = *outputProps.Materialize
	}

	asView := !materialize
	tableName := outputProps.Table
	stagingTableName := tableName
	if e.opts.Env.StageChanges {
		stagingTableName = stagingTableNameFor(tableName)
	}

	// Drop the staging view/table if it exists.
	// NOTE: This intentionally drops the end table if not staging changes.
	if t, err := olap.InformationSchema().Lookup(ctx, "", "", stagingTableName); err == nil {
		_ = olap.DropTable(ctx, stagingTableName, t.View)
	}

	if !e.opts.IncrementalRun {
		// Create the table
		err := olap.CreateTableAsSelect(ctx, stagingTableName, asView, inputProps.SQL)
		if err != nil {
			_ = olap.DropTable(ctx, stagingTableName, asView)
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
		err := olap.InsertTableAsSelect(ctx, tableName, inputProps.SQL, false, false, outputProps.IncrementalStrategy, outputProps.UniqueKey)
		if err != nil {
			return nil, fmt.Errorf("failed to incrementally insert into table: %w", err)
		}
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Table:         tableName,
		View:          asView,
		UsedModelName: usedModelName,
	}
	resultPropsMap := map[string]interface{}{}
	err := mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	// Done
	return &drivers.ModelResult{
		Connector:  e.opts.OutputConnector,
		Properties: resultPropsMap,
		Table:      tableName,
	}, nil
}
