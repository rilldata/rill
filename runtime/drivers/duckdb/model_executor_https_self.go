package duckdb

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/https"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
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
	newInputProps, err := e.modelInputProperties(ctx, opts)
	if err != nil {
		return nil, err
	}
	clone.InputProperties = newInputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *httpsToSelfExecutor) modelInputProperties(ctx context.Context, opts *drivers.ModelExecuteOptions) (map[string]any, error) {
	parsed := &https.ModelInputProperties{}
	unused, err := parsed.DecodeWithWarnings(opts.InputProperties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if len(unused) > 0 {
		e.c.logger.Warn("Undefined fields in input properties. Will be ignored", zap.String("model", opts.ModelName), zap.Strings("fields", unused), observability.ZapCtx(ctx))
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
		return nil, err
	}

	// Set SQL to read from the external source
	from, err := sourceReader([]string{parsed.Path}, format, map[string]any{})
	if err != nil {
		return nil, err
	}

	m.SQL = "SELECT * FROM " + from

	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}
	return propsMap, nil
}
