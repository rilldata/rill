package azure

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/azureblob"
)

func init() {
	drivers.Register("azure", driver{})
	drivers.RegisterAsConnector("azure", driver{})
}

var spec = drivers.Spec{
	DisplayName:        "Azure Blob Storage",
	Description:        "Connect to Azure Blob Storage.",
	ServiceAccountDocs: "https://docs.rilldata.com/deploy/credentials/azure",
	SourceProperties: []drivers.PropertySchema{
		{
			Key:         "path",
			DisplayName: "Blob URI",
			Description: "Path to file on the disk.",
			Placeholder: "azblob://container-name/path/to/file.csv",
			Type:        drivers.StringPropertyType,
			Required:    true,
			Hint:        "Glob patterns are supported",
		},
		{
			Key:         "azure.storage.account",
			DisplayName: "Azure Storage Account",
			Description: "Azure Storage Account inferred from your local environment.",
			Type:        drivers.InformationalPropertyType,
			Hint:        "Set your local credentials: <code>az login</code> Click to learn more.",
			Href:        "https://docs.rilldata.com/develop/import-data#configure-credentials-for-azure",
		},
	},
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:    "azure_storage_account",
			Secret: true,
		},
		{
			Key:    "azure_storage_key",
			Secret: true,
		},
		{
			Key:    "azure_storage_sas_token",
			Secret: true,
		},
	},
}

type driver struct{}

type configProperties struct {
	Account         string `mapstructure:"azure_storage_account"`
	Key             string `mapstructure:"azure_storage_key"`
	SASToken        string `mapstructure:"azure_storage_sas_token"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(config map[string]any, shared bool, client activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("azure driver does not support shared connections")
	}
	conf := &configProperties{}
	err := mapstructure.Decode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config: conf,
		logger: logger,
	}
	return conn, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src drivers.Source, logger *zap.Logger) (bool, error) {
	b, ok := src.BucketSource()
	if !ok {
		return false, fmt.Errorf("require bucket source")
	}
	conf, err := parseSourceProperties(b.Properties)
	if err != nil {
		return false, fmt.Errorf("failed to parse config: %w", err)
	}

	c, err := d.Open(map[string]any{}, false, activity.NewNoopClient(), logger)
	if err != nil {
		return false, err
	}

	conn := c.(*Connection)
	bucketObj, err := conn.openBucket(ctx, conf)
	if err != nil {
		return false, fmt.Errorf("failed to open container %q, %w", conf.url.Host, err)
	}
	defer bucketObj.Close()

	return bucketObj.IsAccessible(ctx)
}

type Connection struct {
	config *configProperties
	logger *zap.Logger
}

var _ drivers.Handle = &Connection{}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "azure"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, m)
	return m
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	return nil
}

// Registry implements drivers.Connection.
func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *Connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *Connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *Connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// Migrate implements drivers.Connection.
func (c *Connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *Connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *Connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return c, true
}

// AsTransporter implements drivers.Connection.
func (c *Connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Connection.
func (c *Connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// DownloadFiles returns a file iterator over objects stored in azure blob storage.
func (c *Connection) DownloadFiles(ctx context.Context, source *drivers.BucketSource) (drivers.FileIterator, error) {
	conf, err := parseSourceProperties(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	client, err := c.getClient(ctx, conf)
	if err != nil {
		return nil, err
	}

	// Create a *blob.Bucket.
	bucketObj, err := azureblob.OpenBucket(ctx, client, nil)
	if err != nil {
		return nil, err
	}
	defer bucketObj.Close()

	// prepare fetch configs
	opts := rillblob.Options{
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
		GlobPattern:           conf.url.Path,
		ExtractPolicy:         source.ExtractPolicy,
	}

	iter, err := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
	if err != nil {
		return nil, err
	}

	return iter, nil
}

type sourceProperties struct {
	Path                  string `key:"path"`
	GlobMaxTotalSize      int64  `mapstructure:"glob.max_total_size"`
	GlobMaxObjectsMatched int    `mapstructure:"glob.max_objects_matched"`
	GlobMaxObjectsListed  int64  `mapstructure:"glob.max_objects_listed"`
	GlobPageSize          int    `mapstructure:"glob.page_size"`
	url                   *globutil.URL
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if !doublestar.ValidatePattern(conf.Path) {
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}
	url, err := globutil.ParseBucketURL(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", conf.Path, err)
	}
	if url.Scheme != "azblob" {
		return nil, fmt.Errorf("invalid scheme %q in path %q", url.Scheme, conf.Path)
	}

	conf.url = url
	return conf, nil
}

// getClient returns a new azure blob client.
func (c *Connection) getClient(ctx context.Context, conf *sourceProperties) (*container.Client, error) {
	name := c.config.Account
	key := c.config.Key

	if c.config.AllowHostAccess {
		name = os.Getenv("AZURE_STORAGE_ACCOUNT")
		key = os.Getenv("AZURE_STORAGE_KEY")
	}

	credential, err := azblob.NewSharedKeyCredential(name, key)
	if err != nil {
		return nil, err
	}

	containerURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", name, conf.url.Host)
	client, err := container.NewClientWithNoCredential(containerURL, nil)
	if err != nil {
		return nil, err
	}

	if key != "" {
		client, err = container.NewClientWithSharedKeyCredential(containerURL, credential, nil)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *Connection) openBucket(ctx context.Context, conf *sourceProperties) (*blob.Bucket, error) {
	// Create containerURL object.
	containerURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", c.config.Account, conf.url.Host)
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
