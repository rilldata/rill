package duckdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type httpsToSelfInputProps struct {
	Path    string            `mapstructure:"path"`
	URI     string            `mapstructure:"uri"`
	Headers map[string]string `mapstructure:"headers"`
	Format  string
}

func (p *httpsToSelfInputProps) resolvePath() string {
	// Backwards compatibility for "uri" renamed to "path"
	if p.URI != "" {
		return p.URI
	}
	return p.Path
}

func (p *httpsToSelfInputProps) Validate() error {
	if p.URI == "" && p.Path == "" {
		return fmt.Errorf("missing property 'path'")
	}
	if p.URI != "" && p.Path != "" {
		return fmt.Errorf("cannot set both 'uri' and 'path'")
	}
	return nil
}

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
	inputProps := &httpsToSelfInputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	// Build the model executor options with updated input properties
	clone := *opts
	newInputProps, err := e.modelInputProperties(opts.ModelName, opts.InputConnector, inputProps)
	if err != nil {
		return nil, err
	}
	clone.InputProperties = newInputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *httpsToSelfExecutor) modelInputProperties(modelName, inputConnector string, inputProps *httpsToSelfInputProps) (map[string]any, error) {
	m := &ModelInputProperties{}
	safeSecret := safeSQLName(fmt.Sprintf("%s__%s__secret__", modelName, inputConnector))
	if len(inputProps.Headers) != 0 {
		m.PreExec = createSecretSQL(safeSecret, inputProps.Headers)
	}
	m.SQL = "SELECT * FROM " + safeSQLString(inputProps.resolvePath())
	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}
	fmt.Printf("propsMap: %v\n", propsMap)
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
