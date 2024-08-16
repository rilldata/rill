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

type s3ToSelfExecutor struct {
	s3 drivers.Handle
	c  *connection
}

var _ drivers.ModelExecutor = &s3ToSelfExecutor{}

func (e *s3ToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return desired, true
	}
	return _defaultConcurrentInserts, true
}

func (e *s3ToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	inputProps := &inputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
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

	// Build the model executor options with updated input and output properties
	clone := *opts
	clone.InputProperties = propsMap
	newOpts := &clone

	// Ensure materialize is true because the selfToSelfExecutor is not able to infer it independently.
	outputProps := &ModelOutputProperties{}
	err = mapstructure.WeakDecode(opts.OutputProperties, &outputProps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}

	if outputProps.Materialize != nil && !*outputProps.Materialize {
		return nil, fmt.Errorf("models must be materialized when fetching data from s3")
	}
	outputProps.Materialize = boolPtr(true)
	err = mapstructure.WeakDecode(outputProps, &newOpts.OutputProperties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}

	// execute
	executor := &selfToSelfExecutor{c: e.c}
	return executor.Execute(ctx, newOpts)
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
