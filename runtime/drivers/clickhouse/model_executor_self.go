package clickhouse

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
		return desired, true
	}
	return _defaultConcurrentInserts, true
}

func (e *selfToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
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
	if outputProps.Typ == "" && outputProps.Materialize == nil {
		outputProps.Materialize = &opts.Env.DefaultMaterialize
	}
	if err := outputProps.Validate(opts); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}
	if outputProps.Typ != "DICTIONARY" && inputProps.SQL == "" {
		return nil, fmt.Errorf("input SQL is required")
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	asView := outputProps.Typ == "VIEW"
	tableName := outputProps.Table
	if outputProps.QuerySettings != "" {
		// Note: This will lead to failures if user sets settings both in query and output properties
		inputProps.SQL = inputProps.SQL + " SETTINGS " + outputProps.QuerySettings
	}

	if !opts.IncrementalRun {
		stagingTableName := tableName
		if opts.Env.StageChanges {
			stagingTableName = stagingTableNameFor(tableName)
		}

		// Drop the staging view/table if it exists.
		// NOTE: This intentionally drops the end table if not staging changes.
		if t, err := e.c.InformationSchema().Lookup(ctx, e.c.config.Database, "", stagingTableName); err == nil {
			_ = e.c.DropTable(ctx, stagingTableName, t.View)
		}

		// Create the table
		err := e.c.CreateTableAsSelect(ctx, stagingTableName, asView, inputProps.SQL, mustToMap(outputProps))
		if err != nil {
			_ = e.c.DropTable(ctx, stagingTableName, asView)
			return nil, fmt.Errorf("failed to create model: %w", err)
		}

		// Rename the staging table to the final table name
		if stagingTableName != tableName {
			err = olapForceRenameTable(ctx, e.c, stagingTableName, asView, tableName)
			if err != nil {
				return nil, fmt.Errorf("failed to rename staged model: %w", err)
			}
		}
	} else {
		// Insert into the table
		err := e.c.InsertTableAsSelect(ctx, tableName, inputProps.SQL, false, true, outputProps.IncrementalStrategy, outputProps.UniqueKey)
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
		Connector:  opts.OutputConnector,
		Properties: resultPropsMap,
		Table:      tableName,
	}, nil
}

func mustToMap(o *ModelOutputProperties) map[string]any {
	m := make(map[string]any)
	err := mapstructure.WeakDecode(o, &m)
	if err != nil {
		panic(fmt.Errorf("failed to encode output properties: %w", err))
	}
	return m
}
