package metricsview

import (
	"context"
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
func (e *Executor) executeExport(ctx context.Context, format, inputConnector string, inputProps map[string]any) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultExportTimeout)
	defer cancel()

	path := e.rt.TempDir(e.instanceID, "metrics_export")
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}

	name, err := randomString("export-", 16)
	if err != nil {
		return "", err
	}
	name = fmt.Sprintf("%s.%s", name, format)
	path = filepath.Join(path, name)

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

	opts := &drivers.ModelExecutorOptions{
		Env: &drivers.ModelEnv{
			AllowHostAccess: e.rt.AllowHostAccess(),
			AcquireConnector: func(ctx context.Context, name string) (drivers.Handle, func(), error) {
				return e.rt.AcquireHandle(ctx, e.instanceID, name)
			},
		},
		ModelName:       "metrics_export", // This isn't a real model; just setting for nicer log messages
		InputHandle:     ic,
		InputConnector:  inputConnector,
		InputProperties: inputProps,
		OutputHandle:    oc,
		OutputConnector: "file",
		OutputProperties: map[string]any{
			"path":   path,
			"format": format,
		},
		Priority: e.priority,
	}

	me, ok := ic.AsModelExecutor(e.instanceID, opts)
	if !ok {
		me, ok = oc.AsModelExecutor(e.instanceID, opts)
		if !ok {
			return "", fmt.Errorf("cannot execute export: input connector %q and output connector %q are not compatible", opts.InputConnector, opts.OutputConnector)
		}
	}

	_, err = me.Execute(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to execute export: %w", err)
	}

	return path, nil
}
