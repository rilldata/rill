package redshift

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
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

	sourceProperties, err := parseSourceProperties(props)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	awsConfig, err := c.awsConfig(ctx, sourceProperties.ResolveRegion(c.config))
	if err != nil {
		return nil, err
	}

	client := redshiftdata.NewFromConfig(awsConfig, func(o *redshiftdata.Options) {
		o.TracerProvider = smithyoteltracing.Adapt(otel.GetTracerProvider())
	})

	outputURL, err := url.Parse(sourceProperties.OutputLocation)
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

	err = c.unload(ctx, client, sourceProperties, unloadLocation)
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

func (c *Connection) unload(ctx context.Context, client *redshiftdata.Client, sourceProperties *sourceProperties, unloadLocation string) error {
	finalSQL := fmt.Sprintf("UNLOAD ('%s') TO '%s/' IAM_ROLE '%s' FORMAT AS PARQUET", sourceProperties.SQL, unloadLocation, sourceProperties.RoleARN)

	_, err := c.executeQuery(ctx, client, finalSQL,
		sourceProperties.ResolveDatabase(c.config),
		sourceProperties.ResolveWorkgroup(c.config),
		sourceProperties.ResolveClusterIdentifier(c.config), nil)
	return err
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

		if out.IsTruncated != nil && *out.IsTruncated && out.NextContinuationToken != nil {
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

func (s *sourceProperties) ResolveRegion(config *configProperties) string {
	if s.AWSRegion != "" {
		return s.AWSRegion
	}
	return config.AWSRegion
}

func (s *sourceProperties) ResolveWorkgroup(config *configProperties) string {
	if s.Workgroup != "" {
		return s.Workgroup
	}
	return config.Workgroup
}

func (s *sourceProperties) ResolveDatabase(config *configProperties) string {
	if s.Database != "" {
		return s.Database
	}
	return config.Database
}

func (s *sourceProperties) ResolveClusterIdentifier(config *configProperties) string {
	if s.ClusterIdentifier != "" {
		return s.ClusterIdentifier
	}
	return config.ClusterIdentifier
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
