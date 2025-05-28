package gcs

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"gocloud.dev/blob/gcsblob"
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

// DownloadFiles returns a file iterator over objects stored in gcs.
// The credential json is read from config google_application_credentials.
// Additionally in case `allow_host_credentials` is true it looks for "Application Default Credentials" as well
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

func (c *Connection) parseBucketURL(path string) (*globutil.URL, error) {
	url, err := globutil.ParseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", path, err)
	}
	if url.Scheme != "gs" {
		return nil, fmt.Errorf("invalid GCS path %q: should start with gs://", path)
	}
	return url, nil
}

func (c *Connection) openBucket(ctx context.Context, bucket string) (*blob.Bucket, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	gcsBucket, err := gcsblob.OpenBucket(ctx, client, bucket, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q: %w", bucket, err)
	}

	return blob.NewBucket(gcsBucket, c.logger)
}
