package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/runtime/drivers"
)

// executeExport enables exporting data from a connector to a temporary local file in the given format.
// The inputConnector and inputProps must be valid for use in a ModelExecutor.
//
// executeExport works by simulating a model that outputs to a file.
// This means it creates a ModelExecutor with the provided input connector and props as input,
// and with the "file" driver as the output connector targeting a temporary output path.
func (e *Executor) executeExport(ctx context.Context, format drivers.FileFormat, inputConnector string, inputProps map[string]any, headers []string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultExportTimeout)
	defer cancel()

	name, err := randomString("export-", 16)
	if err != nil {
		return "", err
	}
	name = format.Filename(name)

	tempDir, err := e.rt.TempDir(e.instanceID)
	if err != nil {
		return "", err
	}
	tempPath := filepath.Join(tempDir, name)

	ic, ir, err := e.rt.AcquireHandle(ctx, e.instanceID, inputConnector)
	if err != nil {
		return "", err
	}
	defer ir()

	oc, or, err := e.rt.AcquireHandle(ctx, e.instanceID, "file")
	if err != nil {
		ir()
		return "", err
	}
	defer or()

	outputProps := map[string]any{
		"path":                  tempPath,
		"format":                format,
		"headers":               headers,
		"file_size_limit_bytes": e.instanceCfg.DownloadLimitBytes,
	}

	opts := &drivers.ModelExecutorOptions{
		Env: &drivers.ModelEnv{
			AllowHostAccess: e.rt.AllowHostAccess(),
			AcquireConnector: func(ctx context.Context, name string) (drivers.Handle, func(), error) {
				return e.rt.AcquireHandle(ctx, e.instanceID, name)
			},
		},
		ModelName:                   "metrics_export", // This isn't a real model; just setting for nicer log messages
		InputHandle:                 ic,
		InputConnector:              inputConnector,
		PreliminaryInputProperties:  inputProps,
		OutputHandle:                oc,
		OutputConnector:             "file",
		PreliminaryOutputProperties: outputProps,
	}

	me, err := ic.AsModelExecutor(e.instanceID, opts)
	if err != nil {
		if !errors.Is(err, drivers.ErrNotImplemented) {
			return "", err
		}
		me, err = oc.AsModelExecutor(e.instanceID, opts)
		if err != nil {
			if !errors.Is(err, drivers.ErrNotImplemented) {
				return "", err
			}
			return "", fmt.Errorf("cannot execute export: input connector %q and output connector %q are not compatible", opts.InputConnector, opts.OutputConnector)
		}
	}

	_, err = me.Execute(ctx, &drivers.ModelExecuteOptions{
		ModelExecutorOptions: opts,
		InputProperties:      inputProps,
		OutputProperties:     outputProps,
		Priority:             e.priority,
		TempDir:              tempPath,
	})
	if err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to execute export: %w", err)
	}

	return tempPath, nil
}
