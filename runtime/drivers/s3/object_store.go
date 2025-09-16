package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"gocloud.dev/blob/s3blob"
)

// ListObjects implements drivers.ObjectStore.
func (c *Connection) ListObjects(ctx context.Context, path string) ([]drivers.ObjectStoreEntry, error) {
	url, err := c.parseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q: %w", path, err)
	}

	bucket, err := c.openBucket(ctx, url.Host)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	return bucket.ListObjects(ctx, url.Path)
}

// DownloadFiles implements drivers.ObjectStore.
func (c *Connection) DownloadFiles(ctx context.Context, path string) (drivers.FileIterator, error) {
	url, err := c.parseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q: %w", path, err)
	}

	bucket, err := c.openBucket(ctx, url.Host)
	if err != nil {
		return nil, err
	}

	tempDir, err := c.storage.TempDir()
	if err != nil {
		return nil, err
	}

	return bucket.Download(ctx, &blob.DownloadOptions{
		Glob:        url.Path,
		TempDir:     tempDir,
		CloseBucket: true,
	})
}

// BucketRegion returns the region to use for the given bucket.
func (c *Connection) BucketRegion(ctx context.Context, bucket string) (string, error) {
	// If custom endpoint is set, use the configured region
	if c.config.Endpoint != "" {
		if c.config.Region != "" {
			return c.config.Region, nil
		}
		return "us-east-1", nil // default for S3-compatible endpoints
	}

	// For AWS endpoints, try to get the bucket region
	cfg, err := c.GetAWSConfig(ctx)
	client := c.GetS3Client(cfg, c.config.Region)

	// Try to get bucket location
	result, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get bucket location: %w", err)
	}

	// AWS returns "us-east-1" as null for the default region
	if result.LocationConstraint == "" {
		return "us-east-1", nil
	}

	return string(result.LocationConstraint), nil
}

func (c *Connection) parseBucketURL(path string) (*globutil.URL, error) {
	url, err := globutil.ParseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q: %w", path, err)
	}
	if url.Scheme != "s3" {
		return nil, fmt.Errorf("invalid S3 path %q: should start with s3://", path)
	}
	return url, nil
}

func (c *Connection) openBucket(ctx context.Context, bucket string) (*blob.Bucket, error) {
	// Determine region if needed (AWS endpoints only)
	region := c.config.Region
	if c.config.Endpoint == "" && region == "" {
		if r, err := c.BucketRegion(ctx, bucket); err == nil && r != "" {
			region = r
		}
	}
	cfg, err := c.GetAWSConfig(ctx)
	if err != nil {
		return nil, err
	}
	s3client := c.GetS3Client(cfg, region)

	s3Bucket, err := s3blob.OpenBucketV2(ctx, s3client, bucket, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q: %w", bucket, err)
	}

	return blob.NewBucket(s3Bucket, c.logger)
}
