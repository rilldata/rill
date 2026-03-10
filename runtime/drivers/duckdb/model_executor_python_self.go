package duckdb

import (
	"context"
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/python"
	"go.uber.org/zap"
)

type pythonToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &pythonToSelfExecutor{}

func (e *pythonToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *pythonToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	if opts.IncrementalRun {
		return nil, fmt.Errorf("duckdb: incremental models are not supported for the python connector")
	}

	// Parse Python-specific input properties
	inputProps := &python.ModelInputProperties{}
	if err := inputProps.Decode(opts.InputProperties); err != nil {
		return nil, err
	}

	e.c.logger.Info("python executor: parsed input props",
		zap.String("code_path", inputProps.CodePath),
		zap.Strings("create_secrets_from_connectors", inputProps.CreateSecretsFromConnectors),
	)

	// Parse the connector config to get python_path
	connConfig := &python.ConfigProperties{}
	if err := mapstructure.WeakDecode(opts.InputHandle.Config(), connConfig); err != nil {
		return nil, fmt.Errorf("failed to parse python connector config: %w", err)
	}

	// Resolve env vars from referenced connectors (create_secrets_from_connectors)
	connectorEnvVars, err := python.ResolveConnectorEnvVars(ctx, inputProps.CreateSecretsFromConnectors, opts.Env.AcquireConnector)
	if err != nil {
		return nil, err
	}

	// Debug: log resolved env var names (not values, for security)
	envVarNames := make([]string, 0, len(connectorEnvVars))
	for k, v := range connectorEnvVars {
		envVarNames = append(envVarNames, fmt.Sprintf("%s (len=%d)", k, len(v)))
	}
	e.c.logger.Info("python executor: resolved connector env vars", zap.Strings("env_vars", envVarNames))

	// Run the Python script
	outputPath, err := python.ExecuteScript(ctx, &python.ExecuteOptions{
		CodePath:         inputProps.CodePath,
		PythonPath:       connConfig.PythonPath,
		RepoRoot:         opts.Env.RepoRoot,
		AllowHostAccess:  opts.Env.AllowHostAccess,
		TempDir:          opts.TempDir,
		Args:             inputProps.Args,
		ExtraEnv:         inputProps.Env,
		ConnectorEnvVars: connectorEnvVars,
	})
	if err != nil {
		return nil, err
	}
	defer os.Remove(outputPath)

	// Build DuckDB input properties: read the Parquet file produced by the script
	m := &ModelInputProperties{
		SQL: "SELECT * FROM read_parquet(" + safeSQLString(outputPath) + ")",
	}
	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}

	// Delegate to selfToSelfExecutor
	clone := *opts
	clone.InputProperties = propsMap
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, &clone)
}
