package duckdb

import (
	"context"
	"errors"
	"io"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

type warehouseToSelfExecutor struct {
	c *connection
	w drivers.Warehouse
}

var _ drivers.ModelExecutor = &warehouseToSelfExecutor{}

func (e *warehouseToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *warehouseToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	iter, err := e.w.QueryAsFiles(ctx, opts.InputProperties)
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

	format := fileutil.FullExt(files[0])
	if iter.Format() != "" {
		format += "." + iter.Format()
	}

	fromClause, err := sourceReader(files, format, make(map[string]any))
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
