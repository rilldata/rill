package clickhouse

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/s3"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

type s3ToSelfExecutor struct {
	s3   drivers.Handle
	c    *connection
	opts *drivers.ModelExecutorOptions
}

type inputProps struct {
	Path   string             `mapstructure:"path"`
	Format drivers.FileFormat `mapstructure:"format"`
}

func (p *inputProps) Validate() error {
	if p.Path == "" {
		return fmt.Errorf("path is mandatory for s3 input connector")
	}
	return nil
}

func (e *s3ToSelfExecutor) Execute(ctx context.Context) (*drivers.ModelResult, error) {
	inputProps := &inputProps{}
	if err := mapstructure.WeakDecode(e.opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	var glob string
	if isGlob(inputProps.Path) {
		glob = inputProps.Path
	} else if filepath.Ext(inputProps.Path) != "" {
		glob = inputProps.Path
	} else {
		if inputProps.Format == "" {
			return nil, fmt.Errorf("clickhouse executor requires a format to be specified for non-glob paths")
		}
		var err error
		glob, err = url.JoinPath(inputProps.Path, "**")
		if err != nil {
			return nil, err
		}
	}

	sql, err := e.genSQL(glob, format(inputProps.Format))
	if err != nil {
		return nil, err
	}
	props := &ModelInputProperties{SQL: sql}
	propsMap := make(map[string]any)
	if err := mapstructure.Decode(props, &propsMap); err != nil {
		return nil, err
	}
	// Build the model executor options with updated input properties
	opts := &drivers.ModelExecutorOptions{
		Env:              e.opts.Env,
		ModelName:        e.opts.ModelName,
		InputConnector:   e.opts.OutputConnector,
		InputProperties:  propsMap,
		OutputConnector:  e.opts.OutputConnector,
		OutputProperties: e.opts.OutputProperties,
		Priority:         e.opts.Priority,
		Incremental:      e.opts.Incremental,
		IncrementalRun:   e.opts.IncrementalRun,
	}
	executor := &selfToSelfExecutor{c: e.c, opts: opts}
	return executor.Execute(ctx)
}

func (e *s3ToSelfExecutor) genSQL(glob, format string) (string, error) {
	props := &s3.ConfigProperties{}
	if err := mapstructure.Decode(e.s3.Config(), props); err != nil {
		return "", err
	}

	// SELECT * FROM S3(path, [id, secret], format)
	var sb strings.Builder
	sb.WriteString("SELECT * FROM s3(")
	sb.WriteString(fmt.Sprintf("'%s'", glob))
	if props.AccessKeyID != "" {
		sb.WriteString(", ")
		sb.WriteString(fmt.Sprintf("'%s'", props.AccessKeyID))
		sb.WriteString(", ")
		sb.WriteString(fmt.Sprintf("'%s'", props.SecretAccessKey))
	}
	if format != "" {
		sb.WriteString(", ")
		sb.WriteString(format)
	}
	sb.WriteString(")")
	return sb.String(), nil
}

func isGlob(path string) bool {
	_, glob := doublestar.SplitPattern(path)
	return fileutil.IsGlob(glob)
}

func format(f drivers.FileFormat) string {
	switch f {
	case drivers.FileFormatCSV:
		return "CSV"
	case drivers.FileFormatJSON:
		return "JSONEachRow"
	case drivers.FileFormatParquet:
		return "Parquet"
	default:
		return ""
	}
}
