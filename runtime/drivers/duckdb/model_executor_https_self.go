package duckdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/https"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
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
	newInputProps, warnings, err := e.modelInputProperties(ctx, opts)
	if err != nil {
		return nil, err
	}
	clone.InputProperties = newInputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	res, err := executor.Execute(ctx, newOpts)
	if err != nil {
		return nil, err
	}
	res.Warnings = append(res.Warnings, warnings...)
	return res, nil
}

func (e *httpsToSelfExecutor) modelInputProperties(ctx context.Context, opts *drivers.ModelExecuteOptions) (map[string]any, []string, error) {
	parsed := &https.ModelInputProperties{}
	var warnings []string
	unused, err := parsed.DecodeWithWarnings(opts.InputProperties)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if len(unused) > 0 {
		if opts.Env.StrictModelProps {
			return nil, nil, fmt.Errorf("undefined fields in input properties: %s", strings.Join(unused, ", "))
		}
		warnings = append(warnings, fmt.Sprintf("Undefined fields %q in input properties. Will be ignored.", strings.Join(unused, ", ")))
	}

	var format string
	if parsed.Format != "" {
		format = fmt.Sprintf(".%s", parsed.Format)
	} else {
		format = fileutil.FullExt(parsed.Path)
	}

	m := &ModelInputProperties{}
	// Generate secret SQL to access the http url using duckdb
	m.InternalCreateSecretSQL, m.InternalDropSecretSQL, _, err = generateSecretSQL(ctx, opts, opts.InputConnector, parsed.Path, opts.InputProperties)
	if err != nil {
		return nil, nil, err
	}

	// Set SQL to read from the external source
	from, err := sourceReader([]string{parsed.Path}, format, map[string]any{})
	if err != nil {
		return nil, nil, err
	}

	m.SQL = "SELECT * FROM " + from

	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, nil, err
	}
	return propsMap, warnings, nil
}
