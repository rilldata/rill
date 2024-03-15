package athena

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	types2 "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

func (c *Connection) Query(_ context.Context, _ map[string]any) (drivers.RowIterator, error) {
	return nil, drivers.ErrNotImplemented
}

func (c *Connection) QueryAsFiles(ctx context.Context, props map[string]any, _ *drivers.QueryOption, _ drivers.Progress) (outIt drivers.FileIterator, outErr error) {
	conf, err := parseSourceProperties(props)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	awsConfig, err := c.awsConfig(ctx, conf.AWSRegion)
	if err != nil {
		return nil, err
	}

	client := athena.NewFromConfig(awsConfig)
	outputLocation, err := resolveOutputLocation(ctx, client, conf)
	if err != nil {
		return nil, err
	}

	outputURL, err := url.Parse(outputLocation)
	if err != nil {
		return nil, err
	}

	// outputLocation s3://bucket/path
	// unloadLocation s3://bucket/path/rill-tmp-<uuid>
	// unloadPath path/rill-tmp-<uuid>
	unloadFolderName := "rill-tmp-" + uuid.New().String()
	bucketName := outputURL.Hostname()
	unloadURL := outputURL.JoinPath(unloadFolderName)
	unloadLocation := unloadURL.String()
	unloadPath := strings.TrimPrefix(unloadURL.Path, "/")

	cleanupFn := func() error {
		return deleteObjectsInPrefix(ctx, awsConfig, bucketName, unloadPath)
	}

	err = c.unload(ctx, client, conf, unloadLocation)
	if err != nil {
		unloadErr := fmt.Errorf("failed to unload: %w", err)
		cleanupErr := cleanupFn()
		if cleanupErr != nil {
			cleanupErr = fmt.Errorf("cleanup error: %w", cleanupErr)
		}
		return nil, errors.Join(unloadErr, cleanupErr)
	}

	defer func() {
		if outErr != nil {
			cleanupErr := cleanupFn()
			if cleanupErr != nil {
				outErr = errors.Join(outErr, fmt.Errorf("cleanup error: %w", cleanupErr))
			}
		}
	}()

	bucketObj, err := openBucket(ctx, awsConfig, bucketName)
	if err != nil {
		return nil, fmt.Errorf("cannot open bucket %q: %w", bucketName, err)
	}

	opts := rillblob.Options{
		GlobPattern: unloadPath + "/**",
		Format:      "parquet",
	}

	it, err := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
	if err != nil {
		return nil, fmt.Errorf("cannot download parquet output %q %w", opts.GlobPattern, err)
	}

	return autoDeleteFileIterator{
		FileIterator: it,
		cleanupFn:    cleanupFn,
	}, nil
}

func (c *Connection) awsConfig(ctx context.Context, awsRegion string) (aws.Config, error) {
	loadOptions := []func(*config.LoadOptions) error{
		// Setting the default region to an empty string, will result in the default region value being ignored
		config.WithDefaultRegion("us-east-1"),
		// Setting the region to an empty string, will result in the region value being ignored
		config.WithRegion(awsRegion),
	}

	// If one of the static properties is specified: access key, secret key, or session token, use static credentials,
	// Else fallback to the SDK's default credential chain (environment, instance, etc) unless AllowHostAccess is false
	if c.config.AccessKeyID != "" || c.config.SecretAccessKey != "" || c.config.SessionToken != "" {
		p := credentials.NewStaticCredentialsProvider(c.config.AccessKeyID, c.config.SecretAccessKey, c.config.SessionToken)
		loadOptions = append(loadOptions, config.WithCredentialsProvider(p))
	} else if !c.config.AllowHostAccess {
		return aws.Config{}, fmt.Errorf("static creds are not provided, and host access is not allowed")
	}

	return config.LoadDefaultConfig(ctx, loadOptions...)
}

func (c *Connection) unload(ctx context.Context, client *athena.Client, conf *sourceProperties, unloadLocation string) error {
	finalSQL := fmt.Sprintf("UNLOAD (%s\n) TO '%s' WITH (format = 'PARQUET')", conf.SQL, unloadLocation)

	executeParams := &athena.StartQueryExecutionInput{
		QueryString: aws.String(finalSQL),
	}

	if conf.OutputLocation != "" {
		executeParams.ResultConfiguration = &types2.ResultConfiguration{
			OutputLocation: aws.String(conf.OutputLocation),
		}
	}

	if conf.Workgroup != "" { // primary is used if nothing is set
		executeParams.WorkGroup = aws.String(conf.Workgroup)
	}

	queryExecutionOutput, err := client.StartQueryExecution(ctx, executeParams)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			_, err = client.StopQueryExecution(ctx, &athena.StopQueryExecutionInput{
				QueryExecutionId: queryExecutionOutput.QueryExecutionId,
			})
			return errors.Join(ctx.Err(), err)
		default:
			status, err := client.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
				QueryExecutionId: queryExecutionOutput.QueryExecutionId,
			})
			if err != nil {
				return err
			}

			switch status.QueryExecution.Status.State {
			case types2.QueryExecutionStateSucceeded:
				return nil
			case types2.QueryExecutionStateCancelled:
				return fmt.Errorf("Athena query execution cancelled")
			case types2.QueryExecutionStateFailed:
				return fmt.Errorf("Athena query execution failed %s", *status.QueryExecution.Status.AthenaError.ErrorMessage)
			}
		}
		time.Sleep(time.Second)
	}
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func resolveOutputLocation(ctx context.Context, client *athena.Client, conf *sourceProperties) (string, error) {
	if conf.OutputLocation != "" {
		return conf.OutputLocation, nil
	}

	workgroup := conf.Workgroup
	// fallback to "primary" (default) workgroup if no workgroup is specified
	if workgroup == "" {
		workgroup = "primary"
	}

	wo, err := client.GetWorkGroup(ctx, &athena.GetWorkGroupInput{
		WorkGroup: aws.String(workgroup),
	})
	if err != nil {
		return "", err
	}

	resultConfiguration := wo.WorkGroup.Configuration.ResultConfiguration
	if resultConfiguration != nil && resultConfiguration.OutputLocation != nil && *resultConfiguration.OutputLocation != "" {
		return *resultConfiguration.OutputLocation, nil
	}

	return "", fmt.Errorf("either output_location or workgroup with an output location must be set")
}

func openBucket(ctx context.Context, cfg aws.Config, bucket string) (*blob.Bucket, error) {
	s3client := s3.NewFromConfig(cfg)
	return s3blob.OpenBucketV2(ctx, s3client, bucket, nil)
}

func deleteObjectsInPrefix(ctx context.Context, cfg aws.Config, bucketName, prefix string) error {
	s3client := s3.NewFromConfig(cfg)

	deleteBatch := func(objects []types.ObjectIdentifier) error {
		_, err := s3client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &bucketName,
			Delete: &types.Delete{
				Objects: objects,
			},
		})
		return err
	}

	var continuationToken *string
	for {
		out, err := s3client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            &bucketName,
			Prefix:            &prefix,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return err
		}

		ids := make([]types.ObjectIdentifier, 0, len(out.Contents))
		for _, o := range out.Contents {
			ids = append(ids, types.ObjectIdentifier{
				Key: o.Key,
			})
		}

		if len(ids) > 0 {
			if err := deleteBatch(ids); err != nil {
				return err
			}
		}

		if *out.IsTruncated && out.NextContinuationToken != nil {
			continuationToken = out.NextContinuationToken
		} else {
			break
		}
	}

	return nil
}

type sourceProperties struct {
	SQL            string `mapstructure:"sql"`
	OutputLocation string `mapstructure:"output_location"`
	Workgroup      string `mapstructure:"workgroup"`
	AWSRegion      string `mapstructure:"region"`
}

type autoDeleteFileIterator struct {
	drivers.FileIterator
	cleanupFn func() error
}

func (i autoDeleteFileIterator) Close() error {
	err := i.FileIterator.Close()
	if err != nil {
		return err
	}

	return i.cleanupFn()
}
