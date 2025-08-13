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
	"github.com/rilldata/rill/runtime/drivers/gcs"
	"github.com/rilldata/rill/runtime/drivers/s3"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

type objectStoreToSelfExecutor struct {
	objectStore drivers.Handle
	c           *Connection
}

var _ drivers.ModelExecutor = &objectStoreToSelfExecutor{}

func (e *objectStoreToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return desired, true
	}
	return _defaultConcurrentInserts, true
}

func (e *objectStoreToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	inputProps := &drivers.ObjectStoreModelInputProperties{}
	err := inputProps.Decode(opts.InputProperties)
	if err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	sql := inputProps.SQL
	if sql == "" {
		sql, err = e.genPathSQL(inputProps.Path, inputProps.Format)
		if err != nil {
			return nil, err
		}
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
		return nil, fmt.Errorf("models must be materialized when fetching data from object store")
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

func (e *objectStoreToSelfExecutor) genPathSQL(path string, format drivers.FileFormat) (string, error) {
	var glob string
	if isGlob(path) {
		glob = path
	} else if filepath.Ext(path) != "" {
		glob = path
	} else {
		if format == "" {
			return "", fmt.Errorf("clickhouse: format is required for non-glob paths")
		}
		var err error
		glob, err = url.JoinPath(path, "**")
		if err != nil {
			return "", err
		}
	}

	switch e.objectStore.Driver() {
	case "s3":
		props := &s3.ConfigProperties{}
		if err := mapstructure.Decode(e.objectStore.Config(), props); err != nil {
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
			sb.WriteString(formatToSQL(format))
		}
		sb.WriteString(")")
		return sb.String(), nil
	case "gcs":
		props := &gcs.ConfigProperties{}
		if err := mapstructure.Decode(e.objectStore.Config(), props); err != nil {
			return "", err
		}

		// SELECT * FROM gcs(path, [id, secret], format)
		var sb strings.Builder
		sb.WriteString("SELECT * FROM gcs(")
		sb.WriteString(fmt.Sprintf("'%s'", glob))
		if props.KeyID != "" {
			sb.WriteString(", ")
			sb.WriteString(fmt.Sprintf("'%s'", props.KeyID))
			sb.WriteString(", ")
			sb.WriteString(fmt.Sprintf("'%s'", props.Secret))
		}
		if format != "" {
			sb.WriteString(", ")
			sb.WriteString(formatToSQL(format))
		}
		sb.WriteString(")")
		return sb.String(), nil
	default:
		return "", fmt.Errorf("internal error: unsupported object store: %s", e.objectStore.Driver())
	}
}

func isGlob(path string) bool {
	_, glob := doublestar.SplitPattern(path)
	return fileutil.IsGlob(glob)
}

func formatToSQL(f drivers.FileFormat) string {
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
