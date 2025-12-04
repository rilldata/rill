package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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

	if !e.c.config.isClickhouseCloud() {
		connectorsForNamedCollections, autoDetected := connectorsForNameCollection(inputProps.CreateNamedCollectionsFromConnectors, e.c.config.CreateNamedCollectionsFromConnectors, opts.Env.Connectors)
		err = e.c.createOrReplaceNamedCollections(ctx, connectorsForNamedCollections, autoDetected, opts.Env)
		if err != nil {
			return nil, err
		}
	}

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
		metrics, err = e.c.createTableAsSelect(ctx, stagingTableName, inputProps.SQL, outputProps)
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
			Strategy: outputProps.IncrementalStrategy,
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

// connectorsForNamedCollections returns the list of connectors to be used for NamedCollection creation.
// Priority:
// 1. If the model configuration specifies connector names, use those.
// 2. if duckdb connector configuration specifies connector names, use those
// 3. If neither is configured, automatically detect all connectors of type s3, azure, gcs, or https.
// The boolean return value is true if the list of connectors was automatically detected.
func connectorsForNameCollection(modelNamedCollections, duckdbNamedCollections []string, allConnectors []*runtimev1.Connector) ([]string, bool) {
	var configuredConnectorsForNamedCollections []string
	if len(modelNamedCollections) > 0 {
		configuredConnectorsForNamedCollections = append(configuredConnectorsForNamedCollections, modelNamedCollections...)
	} else if len(duckdbNamedCollections) > 0 {
		configuredConnectorsForNamedCollections = append(configuredConnectorsForNamedCollections, duckdbNamedCollections...)
	}

	// If no connectors are configured, automatically detect all connectors of type s3, azure, gcs, or https from the project.
	// If a single configured value contains a comma-separated list of connector names, split it into individual entries.
	// Otherwise, return the explicitly configured list of connectors.
	if len(configuredConnectorsForNamedCollections) == 0 {
		var res []string
		for _, c := range allConnectors {
			if c.Type == "s3" || c.Type == "gcs" {
				res = append(res, c.Name)
			}
		}
		return res, true
	} else if len(configuredConnectorsForNamedCollections) == 1 && strings.Contains(configuredConnectorsForNamedCollections[0], ",") {
		res := strings.Split(configuredConnectorsForNamedCollections[0], ",")
		for i, s := range res {
			res[i] = strings.TrimSpace(s)
		}
		return res, false
	}
	return configuredConnectorsForNamedCollections, false
}
