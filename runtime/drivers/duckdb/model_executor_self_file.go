package duckdb

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/file"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

type selfToFileExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &selfToFileExecutor{}

func (e *selfToFileExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *selfToFileExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	olap, ok := e.c.AsOLAP(e.c.instanceID)
	if !ok {
		return nil, fmt.Errorf("output connector is not OLAP")
	}

	inputProps := &ModelInputProperties{}
	unused, err := mapstructureutil.WeakDecodeWithWarnings(opts.InputProperties, inputProps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if len(unused) > 0 {
		e.c.logger.Warn("Undefined fields in input properties. Will be ignored", zap.String("model", opts.ModelName), zap.Strings("fields", unused), observability.ZapCtx(ctx))
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	outputProps := &file.ModelOutputProperties{}
	unused, err = mapstructureutil.WeakDecodeWithWarnings(opts.OutputProperties, outputProps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if len(unused) > 0 {
		e.c.logger.Warn("Undefined fields in output properties. Will be ignored", zap.String("model", opts.ModelName), zap.Strings("fields", unused), observability.ZapCtx(ctx))
	}
	if err := outputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	if opts.IncrementalRun {
		return nil, fmt.Errorf("duckdb-to-file executor does not support incremental runs")
	}

	sql, err := exportSQL(inputProps.SQL, outputProps.Path, outputProps.Format)
	if err != nil {
		return nil, err
	}

	// Check the output file size does not exceed the configured limit.
	overLimit := atomic.Bool{}
	if outputProps.FileSizeLimitBytes > 0 {
		var cancel context.CancelFunc
		// override the parent context
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		go func() {
			ticker := time.NewTicker(time.Millisecond * 500)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					f, err := os.Stat(outputProps.Path)
					if err != nil { // ignore error since file may not be created yet
						continue
					}
					if f.Size() > outputProps.FileSizeLimitBytes {
						overLimit.Store(true)
						cancel()
					}
				}
			}
		}()
	}

	err = olap.Exec(ctx, &drivers.Statement{
		Query:    sql,
		Args:     inputProps.Args,
		Priority: opts.Priority,
	})
	if err != nil {
		if overLimit.Load() {
			return nil, fmt.Errorf("file exceeds size limit %q", datasize.ByteSize(outputProps.FileSizeLimitBytes).HumanReadable())
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// check the size again since duckdb writes data with high throughput
	// and it is possible that the entire file is written
	// before we check size in background goroutine
	if outputProps.FileSizeLimitBytes > 0 {
		f, err := os.Stat(outputProps.Path)
		if err != nil {
			return nil, err
		}
		if f.Size() > outputProps.FileSizeLimitBytes {
			return nil, fmt.Errorf("file exceeds size limit %q", datasize.ByteSize(outputProps.FileSizeLimitBytes).HumanReadable())
		}
	}

	// Build result props
	resultProps := &file.ModelResultProperties{
		Path:   outputProps.Path,
		Format: outputProps.Format,
	}
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}
	return &drivers.ModelResult{
		Connector:  opts.OutputConnector,
		Properties: resultPropsMap,
	}, nil
}

func exportSQL(qry, path string, format drivers.FileFormat) (string, error) {
	switch format {
	case drivers.FileFormatParquet:
		return fmt.Sprintf("COPY (%s\n) TO '%s' (FORMAT PARQUET)", qry, path), nil
	case drivers.FileFormatCSV:
		return fmt.Sprintf("COPY (%s\n) TO '%s' (FORMAT CSV, HEADER true, DATEFORMAT '%%x', TIMESTAMPFORMAT '%%c')", qry, path), nil
	case drivers.FileFormatJSON:
		return fmt.Sprintf("COPY (%s\n) TO '%s' (FORMAT JSON)", qry, path), nil
	default:
		return "", fmt.Errorf("duckdb: unsupported export format %q", format)
	}
}

func supportsExportFormat(format drivers.FileFormat, headers []string) bool {
	switch format {
	case drivers.FileFormatParquet, drivers.FileFormatJSON:
		return true
	case drivers.FileFormatCSV:
		// Avoid using model_executor_self_file when headers are present,because DuckDB's Prefix option requires header=false and suffix.
		// Also, DuckDB XLSX (currently we are not using it) writer doesn't support headers.
		if len(headers) == 0 {
			return true
		}
	}
	return false
}
