package bigquery

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type selfToGCSExecutor struct {
	c     *Connection
	store drivers.ObjectStore
}

var _ drivers.ModelExecutor = &selfToGCSExecutor{}

func (e *selfToGCSExecutor) Concurrency(desired int) (int, bool) {
	if desired > 0 {
		return desired, true
	}
	return 10, true // Default
}

func (e *selfToGCSExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	props := &drivers.ObjectStoreModelOutputProperties{}
	if err := mapstructure.Decode(opts.OutputProperties, props); err != nil {
		return nil, err
	}
	var format drivers.FileFormat
	if props.Format != "" {
		format = props.Format
	} else {
		format = drivers.FileFormatParquet
	}
	outputLocation, err := e.export(ctx, opts.InputProperties, props.Path, format)
	if err != nil {
		return nil, err
	}
	resProps := &drivers.ObjectStoreModelResultProperties{Path: outputLocation, Format: string(format)}
	res := make(map[string]any)
	err = mapstructure.Decode(resProps, &res)
	if err != nil {
		return nil, err
	}

	return &drivers.ModelResult{
		Connector:  opts.OutputConnector,
		Properties: res,
	}, nil
}

func (e *selfToGCSExecutor) export(ctx context.Context, props map[string]any, outputLocation string, format drivers.FileFormat) (string, error) {
	conf, err := e.c.parseSourceProperties(props)
	if err != nil {
		return "", err
	}

	client, err := e.c.getClient(ctx)
	if err != nil {
		return "", err
	}

	outputLocation, err = url.JoinPath(outputLocation, "rill-tmp-"+uuid.New().String(), "/")
	if err != nil {
		return "", err
	}

	exportOptionsStr, err := exportOptions(outputLocation, format)
	if err != nil {
		return "", err
	}

	// Construct EXPORT DATA SQL
	query := fmt.Sprintf(`
		EXPORT DATA OPTIONS (
			%s
		) AS (%s);`, exportOptionsStr, conf.SQL)

	// Run the EXPORT DATA SQL query
	q := client.Query(query)

	job, err := q.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start query job: %w", err)
	}

	// Wait for completion
	_, err = job.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("query job failed: %w", err)
	}

	return outputLocation + "*", nil
}

func exportOptions(outputLocation string, format drivers.FileFormat) (string, error) {
	switch format {
	case drivers.FileFormatCSV:
		return fmt.Sprintf(`
			uri = '%s*',
			format = 'CSV',
			overwrite = true,
			header = true,
			compression = 'GZIP',
			field_delimiter = ','
			`, outputLocation), nil
	case drivers.FileFormatJSON:
		return fmt.Sprintf(`
			uri = '%s*',
			format = 'JSON',
			compression = 'GZIP',
			overwrite = true`,
			outputLocation), nil
	case drivers.FileFormatParquet:
		return fmt.Sprintf(`
			uri = '%s*',
			format = 'PARQUET',
			compression = 'SNAPPY',
			overwrite = true`,
			outputLocation), nil
	default:
		return "", errors.New("invalid format: must be 'CSV', 'JSON', or 'PARQUET'")
	}
}
