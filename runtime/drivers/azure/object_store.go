package azure

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"gocloud.dev/blob/azureblob"
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

	client, err := c.newStorageClient()
	if err != nil {
		return nil, "", err
	}
	opts := &azblob.ListContainersOptions{}
	if validPageSize > 0 {
		v := int32(validPageSize)
		opts.MaxResults = &v
	}
	if unmarshalPageToken != "" {
		opts.Marker = &unmarshalPageToken
	}
	pager := client.NewListContainersPager(opts)

	var buckets []string
	var next string

	if pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("azure list containers failed: %w", err)
		}

		for _, c := range page.ContainerItems {
			buckets = append(buckets, *c.Name)
		}
		if page.NextMarker != nil {
			next = *page.NextMarker
		}
	}
	if next != "" {
		next = pagination.MarshalPageToken(next)
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

// DownloadFiles returns a file iterator over objects stored in azure blob storage.
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
	url, err := globutil.ParseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q: %w", path, err)
	}
	if url.Scheme != "az" && url.Scheme != "azure" {
		return nil, fmt.Errorf("invalid Azure path %q: should start with az://", path)
	}
	return url, nil
}

func (c *Connection) openBucket(ctx context.Context, bucket string, anonymous bool) (*blob.Bucket, error) {
	var client *container.Client
	var err error
	if anonymous {
		client, err = c.newAnonymousClient(bucket)
	} else {
		client, err = c.newClient(bucket)
	}
	if err != nil {
		return nil, err
	}

	azureBucket, err := azureblob.OpenBucket(ctx, client, nil)
	if err != nil {
		return nil, err
	}

	return blob.NewBucket(azureBucket, c.logger)
}

// newClient returns a new azure blob client.
func (c *Connection) newClient(bucket string) (*container.Client, error) {
	client, err := c.newStorageClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Azure storage client: %w", err)
	}
	return client.NewContainerClient(bucket), nil
}

// newStorageClient returns a service client.
func (c *Connection) newStorageClient() (*service.Client, error) {
	connectionString := c.config.GetConnectionString()
	if connectionString != "" {
		client, err := service.NewClientFromConnectionString(connectionString, nil)
		if err != nil {
			return nil, fmt.Errorf("failed service.NewClientFromConnectionString: %w", err)
		}
		return client, nil
	}

	if c.config.GetAccount() != "" {
		svcURL := fmt.Sprintf("https://%s.blob.core.windows.net/", c.config.GetAccount())
		cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
			DisableInstanceDiscovery: true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed azidentity.NewDefaultAzureCredential: %w", err)
		}
		client, err := service.NewClient(svcURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed service.NewClient: %w", err)
		}
		return client, nil
	}

	return nil, errors.New("can't access remote host without credentials: no credentials provided")
}

func (c *Connection) newAnonymousClient(bucket string) (*container.Client, error) {
	accountName := c.config.GetAccount()
	if accountName == "" {
		return nil, fmt.Errorf("AccountName can't be empty")
	}

	svcURL := fmt.Sprintf("https://%s.blob.core.windows.net", accountName)
	containerURL, err := url.JoinPath(svcURL, bucket)
	if err != nil {
		return nil, err
	}
	client, err := container.NewClientWithNoCredential(containerURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed container.NewClientWithNoCredential: %w", err)
	}

	return client, nil
}
