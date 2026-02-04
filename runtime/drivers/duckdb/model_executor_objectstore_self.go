package duckdb

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

var errGCSUsesNativeCreds = errors.New("GCS uses native credentials")

type objectStoreToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &objectStoreToSelfExecutor{}

func (e *objectStoreToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *objectStoreToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Build the model executor options with updated input properties
	clone := *opts
	newInputProps, err := e.modelInputProperties(ctx, opts)
	if err != nil {
		if errors.Is(err, errGCSUsesNativeCreds) {
			e := &objectStoreToSelfExecutorNonNative{c: e.c}
			return e.Execute(ctx, opts)
		}
		return nil, err
	}
	clone.InputProperties = newInputProps
	newOpts := &clone

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
}

func (e *objectStoreToSelfExecutor) modelInputProperties(ctx context.Context, opts *drivers.ModelExecuteOptions) (map[string]any, error) {
	parsed := &drivers.ObjectStoreModelInputProperties{}
	if err := parsed.Decode(opts.InputProperties); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}

	m := &ModelInputProperties{}
	var format string
	if parsed.Format != "" {
		format = fmt.Sprintf(".%s", parsed.Format)
	} else {
		format = fileutil.FullExt(parsed.Path)
	}

	// Generate secret SQL to access the to access object store using duckdb
	var err error
	m.InternalCreateSecretSQL, m.InternalDropSecretSQL, _, err = generateSecretSQL(ctx, opts, opts.InputConnector, parsed.Path, opts.InputProperties)
	if err != nil {
		return nil, err
	}

	// Set SQL to read from the external source
	from, err := sourceReader([]string{parsed.Path}, format, parsed.DuckDB)
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

// objectStoreToSelfExecutorNonNative is a non-native implementation of objectStoreToSelfExecutor.
// It uses Rill's own connectors instead of duckdb's native connectors.
type objectStoreToSelfExecutorNonNative struct {
	c *connection
}

func (e *objectStoreToSelfExecutorNonNative) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	parsed := &drivers.ObjectStoreModelInputProperties{}
	if err := parsed.Decode(opts.InputProperties); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}

	store, ok := opts.InputHandle.AsObjectStore()
	if !ok {
		return nil, fmt.Errorf("input handle is not an object store")
	}

	iter, err := store.DownloadFiles(ctx, parsed.Path)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	// We want to batch all the files to avoid issues with schema compatibility and partition_overwrite inserts.
	// If a user encounters performance issues, we should encourage them to use `partitions:` without `incremental:` to break ingestion into smaller batches.
	iter.SetKeepFilesUntilClose()
	var files []string
	for {
		batch, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		files = append(files, batch...)
	}
	if len(files) == 0 {
		return nil, drivers.ErrNoRows
	}

	var format string
	if parsed.Format != "" {
		format = fmt.Sprintf(".%s", parsed.Format)
	} else {
		format = fileutil.FullExt(parsed.Path)
	}

	fromClause, err := sourceReader(files, format, parsed.DuckDB)
	if err != nil {
		return nil, err
	}

	m := &ModelInputProperties{SQL: "SELECT * FROM " + fromClause}
	propsMap := make(map[string]any)
	if err := mapstructure.Decode(m, &propsMap); err != nil {
		return nil, err
	}
	opts.InputProperties = propsMap

	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, opts)
}
