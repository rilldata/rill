package azure

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/azureblob"
	"gocloud.dev/gcerrors"
)

type sourceProperties struct {
	Path                  string         `mapstructure:"path"`
	Account               string         `mapstructure:"account"`
	URI                   string         `mapstructure:"uri"`
	Extract               map[string]any `mapstructure:"extract"`
	GlobMaxTotalSize      int64          `mapstructure:"glob.max_total_size"`
	GlobMaxObjectsMatched int            `mapstructure:"glob.max_objects_matched"`
	GlobMaxObjectsListed  int64          `mapstructure:"glob.max_objects_listed"`
	GlobPageSize          int            `mapstructure:"glob.page_size"`
	BatchSize             string         `mapstructure:"batch_size"`
	url                   *globutil.URL
	extractPolicy         *rillblob.ExtractPolicy
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.WeakDecode(props, conf)
	if err != nil {
		return nil, err
	}
	if !doublestar.ValidatePattern(conf.Path) {
		return nil, fmt.Errorf("glob pattern %q is invalid", conf.Path)
	}
	bucketURL, err := globutil.ParseBucketURL(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q: %w", conf.Path, err)
	}
	if bucketURL.Scheme != "azure" {
		return nil, fmt.Errorf("invalid scheme %q in path %q", bucketURL.Scheme, conf.Path)
	}

	conf.url = bucketURL
	return conf, nil
}

// ListObjects implements drivers.ObjectStore.
func (c *Connection) ListObjects(ctx context.Context, propsMap map[string]any) ([]drivers.ObjectStoreEntry, error) {
	props, err := parseSourceProperties(propsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	client, err := c.getClient(props)
	if err != nil {
		return nil, err
	}

	azureBucket, err := azureblob.OpenBucket(ctx, client, nil)
	if err != nil {
		return nil, err
	}

	bucket, err := rillblob.NewBucket(azureBucket, c.logger)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	return bucket.ListObjects(ctx, props.url.Path)
}

// DownloadFiles returns a file iterator over objects stored in azure blob storage.
func (c *Connection) DownloadFiles(ctx context.Context, props map[string]any) (drivers.FileIterator, error) {
	conf, err := parseSourceProperties(props)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	client, err := c.getClient(conf)
	if err != nil {
		return nil, err
	}

	// Create a *blob.Bucket.
	bucketObj, err := azureblob.OpenBucket(ctx, client, nil)
	if err != nil {
		return nil, err
	}
	defer bucketObj.Close()

	var batchSize datasize.ByteSize
	if conf.BatchSize == "-1" {
		batchSize = math.MaxInt64 // download everything in one batch
	} else {
		batchSize, err = datasize.ParseString(conf.BatchSize)
		if err != nil {
			return nil, err
		}
	}

	// prepare fetch configs
	opts := rillblob.Options{
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
		GlobPattern:           conf.url.Path,
		ExtractPolicy:         conf.extractPolicy,
		BatchSizeBytes:        int64(batchSize.Bytes()),
		KeepFilesUntilClose:   conf.BatchSize == "-1",
		TempDir:               c.storage.TempDir(),
	}

	iter, err := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
	if err != nil {
		// If the err is due to not using the anonymous client for a public container, we want to retry.
		var respErr *azcore.ResponseError
		if gcerrors.Code(err) == gcerrors.Unknown ||
			(errors.As(err, &respErr) && respErr.RawResponse.StatusCode == http.StatusForbidden && (respErr.ErrorCode == "AuthorizationPermissionMismatch" || respErr.ErrorCode == "AuthenticationFailed")) {
			c.logger.Debug("Azure Blob Storage account does not have permission to list blobs. Falling back to anonymous access.", zap.Error(err))

			client, err := c.createAnonymousClient(conf)
			if err != nil {
				return nil, err
			}

			bucketObj, err := azureblob.OpenBucket(ctx, client, nil)
			if err != nil {
				return nil, err
			}

			anonIt, anonErr := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
			if anonErr == nil {
				return anonIt, nil
			}
		}

		// If there's still an err, return it
		respErr = nil
		if errors.As(err, &respErr) && respErr.StatusCode == http.StatusForbidden {
			return nil, drivers.NewPermissionDeniedError(fmt.Sprintf("failed to create iterator: %v", respErr))
		}
		return nil, err
	}

	return iter, nil
}

// getClient returns a new azure blob client.
func (c *Connection) getClient(conf *sourceProperties) (*container.Client, error) {
	var accountKey, sasToken, connectionString string

	accountName, err := c.getAccountName(conf)
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
		client, err := container.NewClientFromConnectionString(connectionString, conf.url.Host, nil)
		if err != nil {
			return nil, fmt.Errorf("failed container.NewClientFromConnectionString: %w", err)
		}
		return client, nil
	}

	if accountName != "" {
		svcURL := fmt.Sprintf("https://%s.blob.core.windows.net", accountName)
		containerURL, err := url.JoinPath(svcURL, conf.url.Host)
		if err != nil {
			return nil, err
		}

		var sharedKeyCred *azblob.SharedKeyCredential

		if accountKey != "" {
			sharedKeyCred, err = azblob.NewSharedKeyCredential(accountName, accountKey)
			if err != nil {
				return nil, fmt.Errorf("failed azblob.NewSharedKeyCredential: %w", err)
			}

			client, err := container.NewClientWithSharedKeyCredential(containerURL, sharedKeyCred, nil)
			if err != nil {
				return nil, fmt.Errorf("failed container.NewClientWithSharedKeyCredential: %w", err)
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

			containerURL, err := url.JoinPath(string(serviceURL), conf.url.Host)
			if err != nil {
				return nil, err
			}

			client, err := container.NewClientWithNoCredential(containerURL, nil)
			if err != nil {
				return nil, fmt.Errorf("failed container.NewClientWithNoCredential: %w", err)
			}
			return client, nil
		}

		cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
			DisableInstanceDiscovery: true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed azidentity.NewDefaultAzureCredential: %w", err)
		}
		client, err := container.NewClient(containerURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed container.NewClient: %w", err)
		}
		return client, nil
	}

	return nil, drivers.NewPermissionDeniedError("can't access remote host without credentials: no credentials provided")
}

// Create anonymous azure blob client.
func (c *Connection) createAnonymousClient(conf *sourceProperties) (*container.Client, error) {
	accountName, err := c.getAccountName(conf)
	if err != nil {
		return nil, err
	}

	svcURL := fmt.Sprintf("https://%s.blob.core.windows.net", accountName)
	containerURL, err := url.JoinPath(svcURL, conf.url.Host)
	if err != nil {
		return nil, err
	}
	client, err := container.NewClientWithNoCredential(containerURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed container.NewClientWithNoCredential: %w", err)
	}

	return client, nil
}

func (c *Connection) openBucketWithNoCredentials(ctx context.Context, conf *sourceProperties) (*blob.Bucket, error) {
	// Create containerURL object.
	accountName, err := c.getAccountName(conf)
	if err != nil {
		return nil, err
	}
	containerURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, conf.url.Host)
	client, err := container.NewClientWithNoCredential(containerURL, nil)
	if err != nil {
		return nil, err
	}

	// Create a *blob.Bucket.
	bucketObj, err := azureblob.OpenBucket(ctx, client, nil)
	if err != nil {
		return nil, err
	}

	return bucketObj, nil
}

func (c *Connection) getAccountName(conf *sourceProperties) (string, error) {
	if conf.Account != "" {
		return conf.Account, nil
	}

	if c.config.Account != "" {
		return c.config.Account, nil
	}

	if c.config.AllowHostAccess {
		return os.Getenv("AZURE_STORAGE_ACCOUNT"), nil
	}

	return "", errors.New("account name not found")
}
