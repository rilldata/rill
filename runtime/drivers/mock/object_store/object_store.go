package mock

import (
	"context"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	rillblob "github.com/rilldata/rill/runtime/pkg/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"gocloud.dev/blob"

	// Use the file-backed bucket driver for mocked buckets
	_ "gocloud.dev/blob/fileblob"
)

func init() {
	drivers.Register("mock_object_store", driver{})
	drivers.RegisterAsConnector("mock_object_store", driver{})
}

type configProperties struct {
	// Path to a directory on the local file system containing files to serve as objects.
	Path string `mapstructure:"path"`
}

type driver struct{}

var _ drivers.Driver = driver{}

// Spec implements drivers.Driver.
func (driver) Spec() drivers.Spec {
	return drivers.Spec{}
}

// Open implements drivers.Driver.
func (driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	cfg := &configProperties{}
	err := mapstructure.WeakDecode(config, cfg)
	if err != nil {
		return nil, err
	}

	bucket, err := blob.OpenBucket(context.Background(), "file://"+cfg.Path)
	if err != nil {
		return nil, err
	}

	return &Connection{
		logger: logger,
		cfg:    cfg,
		bucket: bucket,
	}, nil
}

// HasAnonymousSourceAccess implements drivers.Driver.
func (driver) HasAnonymousSourceAccess(ctx context.Context, srcProps map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

// TertiarySourceConnectors implements drivers.Driver.
func (driver) TertiarySourceConnectors(ctx context.Context, srcProps map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type Connection struct {
	logger *zap.Logger
	cfg    *configProperties
	bucket *blob.Bucket
}

var _ drivers.Handle = &Connection{}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	return nil
}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "s3"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	return nil
}

// InformationSchema implements drivers.Handle.
func (c *Connection) InformationSchema() drivers.InformationSchema {
	return &drivers.NotImplementedInformationSchema{}
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	return c.bucket.Close()
}

// AsRegistry implements drivers.Connection.
func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Connection.
func (c *Connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Connection.
func (c *Connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *Connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *Connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Connection.
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

// AsModelExecutor implements drivers.Handle.
func (c *Connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (c *Connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *Connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *Connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// ListObjects implements drivers.ObjectStore.
func (c *Connection) ListObjects(ctx context.Context, path string) ([]drivers.ObjectStoreEntry, error) {
	url, err := globutil.ParseBucketURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", path, err)
	}

	bucket, err := rillblob.NewBucket(c.bucket, c.logger)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	return bucket.ListObjects(ctx, url.Path)
}

// DownloadFiles implements drivers.ObjectStore.
func (c *Connection) DownloadFiles(ctx context.Context, path string) (drivers.FileIterator, error) {
	return nil, errors.New("not implemented")
}
