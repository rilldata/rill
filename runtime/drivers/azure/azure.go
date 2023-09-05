package azure

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"go.uber.org/zap"
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
			Placeholder: "az://container-name/path/to/file.csv",
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
			Key:  "azure.storage.account",
			Hint: "Enter path of file to load from.",
			ValidateFunc: func(any interface{}) error {
				val := any.(string)
				if val == "" {
					// user can chhose to leave empty for public sources
					return nil
				}

				path, err := fileutil.ExpandHome(strings.TrimSpace(val))
				if err != nil {
					return err
				}

				_, err = os.Stat(path)
				return err
			},
		},
	},
}

type driver struct{}

type configProperties struct {
	Account string `mapstructure:"azure.storage.account"`
}

func (d driver) Open(config map[string]any, shared bool, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("gcs driver can't be shared")
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

func (d driver) HasAnonymousSourceAccess(context.Context, drivers.Source, *zap.Logger) (bool, error) {
	return false, nil
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

// // DownloadFiles returns a file iterator over objects stored in azure blob storage.
// func (c *Connection) DownloadFiles(ctx context.Context, source *drivers.BucketSource) (drivers.FileIterator, error) {
// 	return nil, fmt.Errorf("not implemented")
// }

// DownloadFiles returns a file iterator over objects stored in azure blob storage.
func (c *Connection) DownloadFiles(ctx context.Context, source *drivers.BucketSource) (drivers.FileIterator, error) {
	conf, err := parseSourceProperties(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	opts := azureblob.NewDefaultServiceURLOptions()
	serviceURL, err := azureblob.NewServiceURL(opts)
	if err != nil {
		return nil, err
	}

	c.logger.Named("console").Info("Connecting to Azure Blob Storage", zap.String("account", conf.url.Host))

	client, err := azureblob.NewDefaultClient(serviceURL, azureblob.ContainerName(conf.url.Host))
	if err != nil {
		return nil, err
	}

	// bucketObj, err := blob.OpenBucket(ctx, "azblob://temp")
	// if err != nil {
	// 	return nil, err
	// }
	// defer bucketObj.Close()

	// Create a *blob.Bucket.
	bucketObj, err := azureblob.OpenBucket(ctx, client, nil)
	if err != nil {
		return nil, err
	}
	defer bucketObj.Close()

	// // Now we can use b to read or write files to the container.
	// data, err := bucketObj.ReadAll(ctx, "my-key")
	// if err != nil {
	// 	return nil, err
	// }

	// fmt.Println("Datat is ", data)

	// prepare fetch configs
	opts1 := rillblob.Options{
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
		GlobPattern:           conf.url.Path,
		ExtractPolicy:         source.ExtractPolicy,
	}

	iter, err := rillblob.NewIterator(ctx, bucketObj, opts1, c.logger)
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
		// ideally this should be validated at much earlier stage
		// keeping it here to have gcs specific validations
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
