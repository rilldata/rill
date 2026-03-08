package dbt_cloud

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

// DbtModelInputProperties are the input properties for a model with dbt_cloud as the input connector.
type DbtModelInputProperties struct {
	DbtMetricRef       string `mapstructure:"dbt_metric_ref"`
	WarehouseConnector string `mapstructure:"warehouse_connector"`
}

// dbtCloudToOLAPExecutor executes a dbt_cloud model by reading data from a warehouse connector
// and writing it to the output OLAP connector.
type dbtCloudToOLAPExecutor struct {
	conn       *connection
	instanceID string
}

var _ drivers.ModelExecutor = &dbtCloudToOLAPExecutor{}

func (e *dbtCloudToOLAPExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *dbtCloudToOLAPExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Parse input properties
	inputProps := &DbtModelInputProperties{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to decode dbt model input properties: %w", err)
	}
	if inputProps.DbtMetricRef == "" {
		return nil, fmt.Errorf("dbt_metric_ref is required")
	}

	// Resolve warehouse connector name; fall back to connector-level config
	warehouseConnector := inputProps.WarehouseConnector
	if warehouseConnector == "" {
		warehouseConnector = e.conn.config.WarehouseConnector
	}
	if warehouseConnector == "" {
		return nil, fmt.Errorf("warehouse_connector is required: set it on the model or the dbt_cloud connector")
	}

	// Fetch manifest and resolve the metric to an output table
	manifest, err := e.conn.GetManifest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dbt manifest: %w", err)
	}

	database, schema, table, err := GetOutputTable(manifest, inputProps.DbtMetricRef)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dbt metric %q: %w", inputProps.DbtMetricRef, err)
	}

	// Build a fully qualified table reference for the SQL query
	qualifiedTable := qualifyTableName(database, schema, table)
	query := fmt.Sprintf("SELECT * FROM %s", qualifiedTable)

	// Acquire the warehouse connector
	warehouseHandle, release, err := opts.Env.AcquireConnector(ctx, warehouseConnector)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire warehouse connector %q: %w", warehouseConnector, err)
	}
	defer release()

	// Delegate execution to the output connector with the warehouse as input.
	// This reuses existing execution patterns (e.g. DuckDB's warehouseToSelfExecutor).
	delegateOpts := &drivers.ModelExecutorOptions{
		Env:                         opts.Env,
		ModelName:                   opts.ModelName,
		InputHandle:                 warehouseHandle,
		InputConnector:              warehouseConnector,
		PreliminaryInputProperties:  map[string]any{"sql": query},
		OutputHandle:                opts.OutputHandle,
		OutputConnector:             opts.OutputConnector,
		PreliminaryOutputProperties: opts.PreliminaryOutputProperties,
	}

	executor, err := opts.OutputHandle.AsModelExecutor(e.instanceID, delegateOpts)
	if err != nil {
		return nil, fmt.Errorf("output connector %q cannot execute from warehouse %q: %w", opts.OutputConnector, warehouseConnector, err)
	}

	// Execute with the warehouse SQL as input
	delegateExecOpts := &drivers.ModelExecuteOptions{
		ModelExecutorOptions: delegateOpts,
		InputProperties:      map[string]any{"sql": query},
		OutputProperties:     opts.OutputProperties,
		Priority:             opts.Priority,
		Incremental:          opts.Incremental,
		IncrementalRun:       opts.IncrementalRun,
		PartitionRun:         opts.PartitionRun,
		PartitionKey:         opts.PartitionKey,
		TempDir:              opts.TempDir,
	}

	return executor.Execute(ctx, delegateExecOpts)
}

// qualifyTableName builds a SQL table reference from database, schema, and table components.
func qualifyTableName(database, schema, table string) string {
	if database != "" && schema != "" {
		return fmt.Sprintf("%s.%s.%s", database, schema, table)
	}
	if schema != "" {
		return fmt.Sprintf("%s.%s", schema, table)
	}
	return table
}
