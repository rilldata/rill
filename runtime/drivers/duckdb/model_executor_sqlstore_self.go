package duckdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type sqlStoreToSelfInputProps struct {
	SQL string `mapstructure:"sql"`
	DSN string `mapstructure:"dsn"`
}

func (p *sqlStoreToSelfInputProps) Validate() error {
	if p.SQL == "" {
		return fmt.Errorf("missing property 'sql'")
	}
	if p.DSN == "" {
		return fmt.Errorf("missing property `dsn`")
	}
	return nil
}

type sqlStoreToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &sqlStoreToSelfExecutor{}

func (e *sqlStoreToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *sqlStoreToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	inputProps := &sqlStoreToSelfInputProps{}
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

	// Build the model executor options with updated input and output properties
	clone := *opts

	newInputProps, err := e.modelInputProperties(opts.ModelName, opts.InputHandle.Driver(), inputProps)
	if err != nil {
		return nil, err
	}
	clone.InputProperties = newInputProps

	newOutputProps := make(map[string]any)
	err = mapstructure.WeakDecode(outputProps, &newOutputProps)
	if err != nil {
		return nil, err
	}
	clone.OutputProperties = newOutputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *sqlStoreToSelfExecutor) modelInputProperties(modelName, inputDriver string, inputProps *sqlStoreToSelfInputProps) (map[string]any, error) {
	m := &ModelInputProperties{}
	dbName := modelName + "_external_db_"
	safeDBName := safeName(dbName)
	userQuery, _ := strings.CutSuffix(inputProps.SQL, ";") // trim trailing semi colon
	switch inputDriver {
	case "mysql":
		dsn := rewriteMySQLDSN(inputProps.DSN)
		m.PreExec = fmt.Sprintf("INSTALL 'MYSQL'; LOAD 'MYSQL'; ATTACH %s AS %s (TYPE mysql, READ_ONLY)", safeSQLString(dsn), safeDBName)
		m.SQL = fmt.Sprintf("SELECT * FROM mysql_query(%s, %s)", safeSQLString(dbName), safeSQLString(userQuery))
	case "postgres":
		m.PreExec = fmt.Sprintf("INSTALL 'POSTGRES'; LOAD 'POSTGRES'; ATTACH %s AS %s (TYPE postgres, READ_ONLY)", safeSQLString(inputProps.DSN), safeDBName)
		m.SQL = fmt.Sprintf("SELECT * FROM postgres_query(%s, %s)", safeSQLString(dbName), safeSQLString(userQuery))
	default:
		return nil, fmt.Errorf("internal error: unsupported external database: %s", inputDriver)
	}
	m.PostExec = fmt.Sprintf("DETACH %s", safeDBName)
	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}
	return propsMap, nil
}
