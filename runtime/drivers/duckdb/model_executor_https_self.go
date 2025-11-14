package duckdb

import (
	"context"
	"fmt"
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

func (e *httpsToSelfExecutor) modelInputProperties(modelName, inputConnector string, inputHandle drivers.Handle, inputProperties map[string]any) (map[string]any, error) {
	connectorProps := &https.ConfigProperties{}
	if err := mapstructure.WeakDecode(inputHandle.Config(), connectorProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}

	srcProp := &https.SourceProperties{}
	if err := mapstructure.WeakDecode(inputProperties, srcProp); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	// Merge headers
	headers := make(map[string]string)
	for k, v := range connectorProps.Headers {
		headers[k] = v
	}
	for k, v := range srcProp.Headers {
		headers[k] = v // srcProp overrides connectorProps
	}

	path := srcProp.ResolvePath()
	if path == "" {
		return nil, fmt.Errorf("missing required property: `path`")
	}

	m := &ModelInputProperties{}
	safeSecret := safeSQLName(fmt.Sprintf("%s__%s__secret__", modelName, inputConnector))
	if len(headers) != 0 {
		m.PreExec = createSecretSQL(safeSecret, headers)
	}
	m.SQL = "SELECT * FROM " + safeSQLString(path)
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
