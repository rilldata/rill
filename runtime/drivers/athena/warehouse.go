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
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	types2 "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/tracing/smithyoteltracing"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/blob"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"gocloud.dev/blob/s3blob"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/drivers/athena")

var _ drivers.Warehouse = &Connection{}

func (c *Connection) QueryAsFiles(ctx context.Context, props map[string]any) (outIt drivers.FileIterator, outErr error) {
	ctx, span := tracer.Start(ctx, "Connection.QueryAsFiles")
	defer span.End()

	conf, err := parseSourceProperties(props)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	awsConfig, err := c.awsConfig(ctx, conf.AWSRegion)
	if err != nil {
		return nil, err
	}

	client := athena.NewFromConfig(awsConfig, func(o *athena.Options) {
		o.TracerProvider = smithyoteltracing.Adapt(otel.GetTracerProvider())
	})
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
		ctx, cancel := graceful.WithMinimumDuration(ctx, 10*time.Second)
		defer cancel()
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

	bucket, err := openBucket(ctx, awsConfig, bucketName, c.logger)
	if err != nil {
		return nil, fmt.Errorf("cannot open bucket %q: %w", bucketName, err)
	}

	it, err := bucket.Download(ctx, &blob.DownloadOptions{
		Glob:        unloadPath + "/**",
		Format:      "parquet",
		CloseBucket: true,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot download parquet output %q: %w", unloadPath, err)
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

	awsConfig, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return aws.Config{}, err
	}

	if c.config.RoleARN != "" {
		stsClient := sts.NewFromConfig(awsConfig)
		assumeRoleOptions := []func(*stscreds.AssumeRoleOptions){}
		if c.config.RoleSessionName != "" {
			assumeRoleOptions = append(assumeRoleOptions, func(o *stscreds.AssumeRoleOptions) {
				o.RoleSessionName = c.config.RoleSessionName
			})
		}
		if c.config.ExternalID != "" {
			assumeRoleOptions = append(assumeRoleOptions, func(o *stscreds.AssumeRoleOptions) {
				o.ExternalID = &c.config.ExternalID
			})
		}
		provider := stscreds.NewAssumeRoleProvider(stsClient, c.config.RoleARN, assumeRoleOptions...)
		awsConfig.Credentials = aws.NewCredentialsCache(provider)
	}

	return awsConfig, nil
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

func openBucket(ctx context.Context, cfg aws.Config, bucket string, logger *zap.Logger) (*blob.Bucket, error) {
	s3client := s3.NewFromConfig(cfg)
	s3bucket, err := s3blob.OpenBucketV2(ctx, s3client, bucket, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q: %w", bucket, err)
	}
	return blob.NewBucket(s3bucket, logger)
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
