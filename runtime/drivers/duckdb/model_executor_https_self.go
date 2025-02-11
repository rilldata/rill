package duckdb

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/https"
)

type httpsToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &httpsToSelfExecutor{}

func (e *httpsToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *httpsToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Build the model executor options with updated input properties
	clone := *opts
	newInputProps, err := e.modelInputProperties(opts.ModelName, opts.InputConnector, opts.InputHandle, opts.InputProperties)
	if err != nil {
		return nil, err
	}
	clone.InputProperties = newInputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *httpsToSelfExecutor) modelInputProperties(modelName, inputConnector string, inputHandle drivers.Handle, props map[string]any) (map[string]any, error) {
	// somewhat unlikely but we also allow to define a http connector which models can refer
	cfg := inputHandle.Config()
	// config properties can also be set as input properties
	maps.Copy(cfg, props)
	inputProps := &https.ConfigProperties{}
	if err := mapstructure.WeakDecode(cfg, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}

	m := &ModelInputProperties{}
	safeSecret := safeSQLName(fmt.Sprintf("%s__%s__secret__", modelName, inputConnector))
	if len(inputProps.Headers) != 0 {
		m.PreExec = createSecretSQL(safeSecret, inputProps.Headers)
	}
	m.SQL = "SELECT * FROM " + safeSQLString(inputProps.ResolvePath())
	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}
	return propsMap, nil
}

func createSecretSQL(safeName string, headers map[string]string) string {
	var headerStrings []string
	for key, value := range headers {
		headerStrings = append(headerStrings, fmt.Sprintf("%s: %s", safeSQLString(key), safeSQLString(value)))
	}
	headersSQL := strings.Join(headerStrings, ", ")
	return fmt.Sprintf(`CREATE OR REPLACE TEMPORARY SECRET %s (TYPE HTTP, EXTRA_HTTP_HEADERS MAP { %s })`, safeName, headersSQL)
}
