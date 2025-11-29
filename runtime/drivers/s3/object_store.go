package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"gocloud.dev/blob/s3blob"
)

func (c *Connection) ListBuckets(ctx context.Context, pageSize uint32, pageToken string) ([]string, string, error) {
	// If PathPrefixes is configured, return buckets derived from those prefixes.
	// This is used when ListBuckets permissions may not be available, or when
	// the user explicitly wants to restrict access to specific buckets.
	if len(c.config.PathPrefixes) > 0 {
		return drivers.ListBucketsFromPathPrefixes(c.config.PathPrefixes, pageSize, pageToken)
	}

	validPageSize := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	unmarshalPageToken := ""
	if pageToken != "" {
		if err := pagination.UnmarshalPageToken(pageToken, &unmarshalPageToken); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}
	client, err := getS3Client(ctx, c.config, "")
	if err != nil {
		return nil, "", err
	}

	input := &s3.ListBucketsInput{
		MaxBuckets: aws.Int32(int32(validPageSize)),
	}
	if unmarshalPageToken != "" {
		input.ContinuationToken = aws.String(unmarshalPageToken)
	}
	output, err := client.ListBuckets(ctx, input)
	if err != nil {
		return nil, "", err
	}
	buckets := make([]string, 0, len(output.Buckets))
	for _, bucket := range output.Buckets {
		if bucket.Name != nil {
			buckets = append(buckets, *bucket.Name)
		}
	}
	next := ""
	if output.ContinuationToken != nil {
		next = pagination.MarshalPageToken(*output.ContinuationToken)
	}
	return buckets, next, nil
}

// ListObjects implements drivers.ObjectStore.
func (c *Connection) ListObjects(ctx context.Context, bucket, path, delimiter string, pageSize uint32, pageToken string) ([]drivers.ObjectStoreEntry, string, error) {
	blobBucket, err := c.openBucket(ctx, bucket, false)
	if err != nil {
		return nil, "", err
	}
	defer blobBucket.Close()
	blobListfn := func(ctx context.Context, p string, d string, s uint32, t string) ([]drivers.ObjectStoreEntry, string, error) {
		return blobBucket.ListObjects(ctx, p, d, s, t)
	}
	return drivers.ListObjects(ctx, c.config.PathPrefixes, blobListfn, bucket, path, delimiter, pageSize, pageToken)
}

// ListObjectsForGlob implements drivers.ObjectStore.
func (c *Connection) ListObjectsForGlob(ctx context.Context, bucket, glob string) ([]drivers.ObjectStoreEntry, error) {
	blobBucket, err := c.openBucket(ctx, bucket, false)
	if err != nil {
		return nil, err
	}
	defer blobBucket.Close()

	return blobBucket.ListObjectsForGlob(ctx, glob)
}

// DownloadFiles implements drivers.ObjectStore.
func (c *Connection) DownloadFiles(ctx context.Context, path string) (drivers.FileIterator, error) {
	url, err := c.parseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q: %w", path, err)
	}

	bucket, err := c.openBucket(ctx, url.Host, false)
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

func (c *Connection) parseBucketURL(path string) (*globutil.URL, error) {
	path = c.rewriteToS3Path(path)
	url, err := globutil.ParseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q: %w", path, err)
	}
	if url.Scheme != "s3" {
		return nil, fmt.Errorf("invalid S3 path %q: should start with s3://", path)
	}
	return url, nil
}

func (c *Connection) openBucket(ctx context.Context, bucket string, anonymous bool) (*blob.Bucket, error) {
	var s3client *s3.Client
	var err error
	if anonymous {
		s3client, err = getAnonymousS3Client(ctx, c.config, bucket)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		s3client, err = getS3Client(ctx, c.config, bucket)
		if err != nil {
			return nil, err
		}
	}

	s3Bucket, err := s3blob.OpenBucketV2(ctx, s3client, bucket, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q: %w", bucket, err)
	}

	return blob.NewBucket(s3Bucket, c.logger)
}

func (c *Connection) rewriteToS3Path(s string) string {
	switch c.config.Endpoint {
	case "storage.googleapis.com":
		if after, ok := strings.CutPrefix(s, "gs://"); ok {
			return "s3://" + after
		}
		if after, ok := strings.CutPrefix(s, "gcs://"); ok {
			return "s3://" + after
		}
		return s
	default:
		return s
	}
}
