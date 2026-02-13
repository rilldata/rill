package s3

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/driverutil"
	"github.com/rilldata/rill/runtime/pkg/exportutil"
)

type olapToSelfExecutor struct {
	c    *Connection
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

	outputProps := &drivers.ObjectStoreModelOutputProperties{}
	if err := mapstructure.Decode(opts.OutputProperties, outputProps); err != nil {
		return nil, err
	}

	if outputProps.Format != "" && outputProps.Format != drivers.FileFormatParquet {
		return nil, fmt.Errorf("olap-to-objectstore executor only support 'parquet' format")
	}

	if opts.IncrementalRun {
		return nil, fmt.Errorf("olap-to-objectstore executor does not support incremental runs")
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

	// Create a temporary file to write result to
	f, err := os.CreateTemp("", "olap-to-s3-*.parquet")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// Make sure to clean up the file after we're done
	defer os.Remove(f.Name())

	err = driverutil.ResultToFile(res, f, drivers.FileFormatParquet, nil)
	if err != nil {
		return nil, err
	}

	// We need to re-open the file since it was closed after writing the result to it
	f, err = os.Open(f.Name())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Parse the output path
	bucket, key, fullPath, err := exportutil.ParsePath(outputProps.Path)
	if err != nil {
		return nil, err
	}

	client, err := getS3Client(ctx, e.c.config, bucket)
	if err != nil {
		return nil, err
	}

	// Upload the file to the object store
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	})
	if err != nil {
		return nil, err
	}

	resProps := &drivers.ObjectStoreModelResultProperties{
		Path:   fullPath,
		Format: string(drivers.FileFormatParquet),
	}
	resPropsMap := make(map[string]any)
	err = mapstructure.Decode(resProps, &resPropsMap)
	if err != nil {
		return nil, err
	}

	return &drivers.ModelResult{
		Connector:  opts.OutputConnector,
		Properties: resPropsMap,
	}, nil
}
