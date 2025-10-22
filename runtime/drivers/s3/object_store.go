package s3

import (
	"context"
	"fmt"
	"strings"

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

	bucket, err := c.openBucket(ctx, url.Host, false)
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
