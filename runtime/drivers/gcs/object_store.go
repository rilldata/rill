package gcs

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/blob"
	"github.com/rilldata/rill/runtime/pkg/gcputil"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"gocloud.dev/blob/gcsblob"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
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

	credentials, err := gcputil.Credentials(ctx, c.config.SecretJSON, c.config.AllowHostAccess)
	if err != nil {
		return nil, "", err
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, "", err
	}
	defer client.Close()

	projectID, err := gcputil.ProjectID(credentials)
	if err != nil {
		return nil, "", err
	}

	pager := iterator.NewPager(client.Buckets(ctx, projectID), validPageSize, unmarshalPageToken)
	buckets := make([]*storage.BucketAttrs, 0)
	next, err := pager.NextPage(&buckets)
	if err != nil {
		return nil, "", err
	}
	names := make([]string, len(buckets))
	for i := 0; i < len(buckets); i++ {
		names[i] = buckets[i].Name
	}

	if next != "" {
		next = pagination.MarshalPageToken(next)
	}
	return names, next, nil
}

// ListObjects implements drivers.ObjectStore.
func (c *Connection) ListObjects(ctx context.Context, bucket, path, delimiter string, pageSize uint32, pageToken string) ([]drivers.ObjectStoreEntry, string, error) {
	blobBucket, err := c.openBucket(ctx, bucket)
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
func (c *Connection) ListObjectsForGlob(ctx context.Context, bucket, glob string, pageSize uint32, pageToken string) ([]drivers.ObjectStoreEntry, string, error) {
	blobBucket, err := c.openBucket(ctx, bucket)
	if err != nil {
		return nil, "", err
	}
	defer blobBucket.Close()

	return blobBucket.ListObjectsForGlob(ctx, glob, pageSize, pageToken)
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
