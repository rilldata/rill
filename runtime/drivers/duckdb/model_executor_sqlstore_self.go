package duckdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/mysql"
	"github.com/rilldata/rill/runtime/drivers/postgres"
)

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
	inputProps := &ModelInputProperties{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	// Build the model executor options with updated input properties
	clone := *opts
	newInputProps, err := e.modelInputProperties(opts.ModelName, opts.InputConnector, opts.InputHandle, inputProps)
	if err != nil {
		return nil, err
	}
	clone.InputProperties = newInputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *sqlStoreToSelfExecutor) modelInputProperties(modelName, inputConnector string, inputHandle drivers.Handle, inputProps *ModelInputProperties) (map[string]any, error) {
	m := &ModelInputProperties{}
	dbName := fmt.Sprintf("%s__%s", modelName, inputConnector)
	safeDBName := safeName(dbName)
	userQuery, _ := strings.CutSuffix(inputProps.SQL, ";") // trim trailing semi colon
	switch inputHandle.Driver() {
	case "mysql":
		var config *mysql.ConfigProperties
		if err := mapstructure.Decode(inputHandle.Config(), &config); err != nil {
			return nil, err
		}
		dsn := rewriteMySQLDSN(config.DSN)
		m.PreExec = fmt.Sprintf("INSTALL 'MYSQL'; LOAD 'MYSQL'; ATTACH %s AS %s (TYPE mysql, READ_ONLY)", safeSQLString(dsn), safeDBName)
		m.SQL = fmt.Sprintf("SELECT * FROM mysql_query(%s, %s)", safeSQLString(dbName), safeSQLString(userQuery))
	case "postgres":
		var config *postgres.ConfigProperties
		if err := mapstructure.Decode(inputHandle.Config(), &config); err != nil {
			return nil, err
		}
		m.PreExec = fmt.Sprintf("INSTALL 'POSTGRES'; LOAD 'POSTGRES'; ATTACH %s AS %s (TYPE postgres, READ_ONLY)", safeSQLString(config.DatabaseURL), safeDBName)
		m.SQL = fmt.Sprintf("SELECT * FROM postgres_query(%s, %s)", safeSQLString(dbName), safeSQLString(userQuery))
	default:
		return nil, fmt.Errorf("internal error: unsupported external database: %s", inputHandle.Driver())
	}
	m.PostExec = fmt.Sprintf("DETACH %s", safeDBName)
	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}
	return propsMap, nil
}
