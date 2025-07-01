package azure

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"gocloud.dev/blob/azureblob"
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
	var accountKey, sasToken, connectionString string

	accountName, err := c.accountName()
	if err != nil {
		return nil, err
	}

	if c.config.AllowHostAccess {
		accountKey = os.Getenv("AZURE_STORAGE_KEY")
		sasToken = os.Getenv("AZURE_STORAGE_SAS_TOKEN")
		connectionString = os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	}

	if c.config.Key != "" {
		accountKey = c.config.Key
	}
	if c.config.SASToken != "" {
		sasToken = c.config.SASToken
	}
	if c.config.ConnectionString != "" {
		connectionString = c.config.ConnectionString
	}

	if connectionString != "" {
		client, err := service.NewClientFromConnectionString(connectionString, nil)
		if err != nil {
			return nil, fmt.Errorf("failed service.NewClientFromConnectionString: %w", err)
		}
		return client, nil
	}

	if accountName != "" {
		svcURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)

		var sharedKeyCred *azblob.SharedKeyCredential

		if accountKey != "" {
			sharedKeyCred, err = azblob.NewSharedKeyCredential(accountName, accountKey)
			if err != nil {
				return nil, fmt.Errorf("failed azblob.NewSharedKeyCredential: %w", err)
			}

			client, err := service.NewClientWithSharedKeyCredential(svcURL, sharedKeyCred, nil)
			if err != nil {
				return nil, fmt.Errorf("failed service.NewClientWithSharedKeyCredential: %w", err)
			}
			return client, nil
		}

		if sasToken != "" {
			serviceURL, err := azureblob.NewServiceURL(&azureblob.ServiceURLOptions{
				AccountName: accountName,
				SASToken:    sasToken,
			})
			if err != nil {
				return nil, err
			}

			client, err := service.NewClientWithNoCredential(string(serviceURL), nil)
			if err != nil {
				return nil, fmt.Errorf("failed service.NewClientWithNoCredential: %w", err)
			}
			return client, nil
		}

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
	accountName, err := c.accountName()
	if err != nil {
		return nil, err
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

func (c *Connection) accountName() (string, error) {
	if c.config.Account != "" {
		return c.config.Account, nil
	}

	if c.config.AllowHostAccess {
		return os.Getenv("AZURE_STORAGE_ACCOUNT"), nil
	}

	return "", errors.New("account name not found")
}
