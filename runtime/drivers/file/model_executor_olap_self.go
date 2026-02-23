package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/driverutil"
)

type olapToSelfExecutor struct {
	c    *connection
	olap drivers.OLAPStore
}

var _ drivers.ModelExecutor = &olapToSelfExecutor{}

func (e *olapToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *olapToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Parse SQL from input properties
	inputProps := &struct {
		SQL  string `mapstructure:"sql"`
		Args []any  `mapstructure:"args"`
	}{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if inputProps.SQL == "" {
		return nil, errors.New("missing SQL in input properties")
	}

	// Parse output properties
	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if err := outputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	// Execute the SQL
	res, err := e.olap.Query(ctx, &drivers.Statement{
		Query:    inputProps.SQL,
		Args:     inputProps.Args,
		Priority: opts.Priority,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	f, err := os.Create(outputProps.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var fw io.Writer = f
	if outputProps.FileSizeLimitBytes > 0 {
		fw = &limitedWriter{W: fw, N: outputProps.FileSizeLimitBytes}
	}

	err = driverutil.ResultToFile(res, fw, outputProps.Format, outputProps.Headers)
	if err != nil {
		if errors.Is(err, io.ErrShortWrite) {
			return nil, fmt.Errorf("file exceeds size limit %q", datasize.ByteSize(outputProps.FileSizeLimitBytes).HumanReadable())
		}
		return nil, fmt.Errorf("failed to write format %q: %w", outputProps.Format, err)
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Path:   outputProps.Path,
		Format: outputProps.Format,
	}
	resultPropsMap := map[string]any{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}
	return &drivers.ModelResult{
		Connector:  opts.OutputConnector,
		Properties: resultPropsMap,
	}, nil
}

// A limitedWriter writes to W but limits the amount of
// data written to just N bytes.
//
// Modified from github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/util/ioutils/ioutils.go
type limitedWriter struct {
	W io.Writer // underlying writer
	N int64     // max bytes remaining
}

func (l *limitedWriter) Write(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.ErrShortWrite
	}
	if int64(len(p)) > l.N {
		return 0, io.ErrShortWrite
	}
	n, err = l.W.Write(p)
	l.N -= int64(n)
	return
}
