package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
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
	var warnings []string
	unused, err := mapstructureutil.WeakDecodeWithWarnings(opts.InputProperties, inputProps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if len(unused) > 0 {
		if opts.Env.StrictModelProps {
			return nil, fmt.Errorf("undefined fields in input properties: %q", strings.Join(unused, ", "))
		}
		warnings = append(warnings, fmt.Sprintf("Undefined fields %q in input properties. Will be ignored.", strings.Join(unused, ", ")))
	}

	outputProps := &ModelOutputProperties{}
	unused, err = mapstructureutil.WeakDecodeWithWarnings(opts.OutputProperties, outputProps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if len(unused) > 0 {
		if opts.Env.StrictModelProps {
			return nil, fmt.Errorf("undefined fields in output properties: %q", strings.Join(unused, ", "))
		}
		warnings = append(warnings, fmt.Sprintf("Undefined fields %q in output properties. Will be ignored.", strings.Join(unused, ", ")))
	}

	// Validate the output properties
	err = e.c.validateAndApplyDefaults(opts, inputProps, outputProps)
	if err != nil {
		return nil, fmt.Errorf("invalid model properties: %w", err)
	}

	// Auto-detect references to Rill-managed named collections in the model SQL and verify they
	// exist on the server. This is the analog of DuckDB's `connectorsForSecrets` auto-detection
	// branch: users don't have to list named-collection connectors explicitly; we discover them
	// from `s3(rill_<conn>, ...)`-style references. Missing collections are reported as warnings
	// (not errors) so a user with multiple instance/cluster setups can still iterate locally;
	// the underlying CH error from the missing collection will surface during the actual query.
	if refs := DetectNamedCollectionRefs(inputProps.SQL); len(refs) > 0 {
		for _, connectorName := range refs {
			exists, checkErr := e.c.NamedCollectionExists(ctx, connectorName)
			if checkErr != nil {
				warnings = append(warnings, fmt.Sprintf("could not verify named collection %q on ClickHouse: %v", NamedCollectionName(connectorName), checkErr))
				continue
			}
			if !exists {
				warnings = append(warnings, fmt.Sprintf("model references named collection %q but it does not exist on the ClickHouse server; ensure the connector resource %q has reconciled successfully", NamedCollectionName(connectorName), connectorName))
			}
		}
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
		Warnings:     warnings,
	}, nil
}
