package duckdb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/exportutil"
)

type selfToObjectStoreExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &selfToObjectStoreExecutor{}

func (e *selfToObjectStoreExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *selfToObjectStoreExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
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

	outputProps := &drivers.ObjectStoreModelOutputProperties{}
	if err := mapstructure.Decode(opts.OutputProperties, outputProps); err != nil {
		return nil, err
	}

	if outputProps.Format != "" && outputProps.Format != drivers.FileFormatParquet {
		return nil, fmt.Errorf("duckdb-to-objectstore executor only support 'parquet' format")
	}

	if opts.IncrementalRun {
		return nil, fmt.Errorf("duckdb-to-objectstore executor does not support incremental runs")
	}

	// Parse the output path
	bucket, _, fullPath, err := exportutil.ParsePath(outputProps.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output path: %w", err)
	}

	connectorsForSecrets, autoDetected := connectorsForSecrets(inputProps.CreateSecretsFromConnectors, e.c.config.CreateSecretsFromConnectors, opts.Env.Connectors)
	var createSecretSQLs, dropSecretSQLs []string
	for _, connector := range connectorsForSecrets {
		// We need to pass the bucket we are using because of S3 region detection
		createSecretSQL, dropSecretSQL, _, err := generateSecretSQL(ctx, opts, connector, bucket, nil)
		if err != nil {
			// Silently ignore when auto detected connector or when using native GCS credentials (since it's not supported by DuckDB)
			if autoDetected || errors.Is(err, errGCSUsesNativeCreds) {
				continue
			}
			return nil, fmt.Errorf("failed to create secret for connector %q: %w", connector, err)
		}
		createSecretSQLs = append(createSecretSQLs, createSecretSQL)
		dropSecretSQLs = append(dropSecretSQLs, dropSecretSQL)
	}

	sql := fmt.Sprintf(
		"%s; COPY (%s\n) TO '%s'; %s",
		strings.Join(createSecretSQLs, "; "),
		inputProps.SQL,
		fullPath,
		strings.Join(dropSecretSQLs, "; "),
	)
	err = olap.Exec(ctx, &drivers.Statement{
		Query:    sql,
		Args:     inputProps.Args,
		Priority: opts.Priority,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	resProps := &drivers.ObjectStoreModelResultProperties{
		Path:   fullPath,
		Format: string(drivers.FileFormatParquet),
	}
	resPropsMap := make(map[string]any)
	err = mapstructure.Decode(resProps, &resPropsMap)
	if err != nil {
		return nil, err
	}

	return &drivers.ModelResult{
		Connector:  opts.OutputConnector,
		Properties: resPropsMap,
	}, nil
}
