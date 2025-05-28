package redshift

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
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	redshift_types "github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/tracing/smithyoteltracing"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/blob"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"gocloud.dev/blob/s3blob"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/drivers/redshift")

var _ drivers.Warehouse = &Connection{}

func (c *Connection) QueryAsFiles(ctx context.Context, props map[string]any) (outIt drivers.FileIterator, outErr error) {
	ctx, span := tracer.Start(ctx, "Connection.QueryAsFiles")
	defer func() {
		if outErr != nil {
			span.SetStatus(codes.Error, outErr.Error())
		}
		span.End()
	}()

	conf, err := parseSourceProperties(props)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	awsConfig, err := c.awsConfig(ctx, conf.AWSRegion)
	if err != nil {
		return nil, err
	}

	client := redshiftdata.NewFromConfig(awsConfig, func(o *redshiftdata.Options) {
		o.TracerProvider = smithyoteltracing.Adapt(otel.GetTracerProvider())
	})

	outputURL, err := url.Parse(conf.OutputLocation)
	if err != nil {
		return nil, err
	}

	// outputLocation s3://bucket/path
	// unloadLocation s3://bucket/path/rill-tmp-<uuid>
	// unloadPath path/rill-tmp-redshift-<uuid>
	unloadFolderName := "rill-tmp-redshift-" + uuid.New().String()
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

	tempDir, err := c.storage.TempDir()
	if err != nil {
		return nil, err
	}

	it, err := bucket.Download(ctx, &blob.DownloadOptions{
		Glob:        unloadPath + "/**",
		Format:      "parquet",
		TempDir:     tempDir,
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
	if c.config.AccessKeyID != "" || c.config.SecretAccessKey != "" {
		p := credentials.NewStaticCredentialsProvider(c.config.AccessKeyID, c.config.SecretAccessKey, c.config.SessionToken)
		loadOptions = append(loadOptions, config.WithCredentialsProvider(p))
	} else if !c.config.AllowHostAccess {
		return aws.Config{}, fmt.Errorf("static creds are not provided, and host access is not allowed")
	}

	return config.LoadDefaultConfig(ctx, loadOptions...)
}

func (c *Connection) unload(ctx context.Context, client *redshiftdata.Client, conf *sourceProperties, unloadLocation string) error {
	finalSQL := fmt.Sprintf("UNLOAD ('%s') TO '%s/' IAM_ROLE '%s' FORMAT AS PARQUET", conf.SQL, unloadLocation, conf.RoleARN)

	executeParams := &redshiftdata.ExecuteStatementInput{
		Sql:      &finalSQL,
		Database: &conf.Database,
	}

	if conf.ClusterIdentifier != "" { // ClusterIdentifier and Workgroup are interchangeable
		executeParams.ClusterIdentifier = aws.String(conf.ClusterIdentifier)
	}

	if conf.Workgroup != "" {
		executeParams.WorkgroupName = &conf.Workgroup
	}

	queryExecutionOutput, err := client.ExecuteStatement(ctx, executeParams)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			cancelCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			_, err = client.CancelStatement(cancelCtx, &redshiftdata.CancelStatementInput{
				Id: queryExecutionOutput.Id,
			})
			cancel()
			return errors.Join(ctx.Err(), err)
		case <-ticker.C:
			status, err := client.DescribeStatement(ctx, &redshiftdata.DescribeStatementInput{
				Id: queryExecutionOutput.Id,
			})
			if err != nil {
				return err
			}

			state := status.Status

			if status.Error != nil {
				return fmt.Errorf("Redshift query execution failed %s", *status.Error)
			}

			if state != redshift_types.StatusStringSubmitted && state != redshift_types.StatusStringStarted && state != redshift_types.StatusStringPicked {
				return nil
			}
		}
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
	SQL               string `mapstructure:"sql"`
	OutputLocation    string `mapstructure:"output_location"`
	Workgroup         string `mapstructure:"workgroup"`
	Database          string `mapstructure:"database"`
	ClusterIdentifier string `mapstructure:"cluster_identifier"`
	RoleARN           string `mapstructure:"role_arn"`
	AWSRegion         string `mapstructure:"region"`
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
