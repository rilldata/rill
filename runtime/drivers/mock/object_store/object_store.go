package mock

import (
	"context"
	"errors"
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"github.com/rilldata/rill/runtime/pkg/activity"
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

	return &handle{
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

type handle struct {
	logger *zap.Logger
	cfg    *configProperties
	bucket *blob.Bucket
}

var _ drivers.Handle = &handle{}

// Ping implements drivers.Handle.
func (h *handle) Ping(ctx context.Context) error {
	return drivers.ErrNotImplemented
}

// Driver implements drivers.Connection.
func (h *handle) Driver() string {
	return "s3"
}

// Config implements drivers.Connection.
func (h *handle) Config() map[string]any {
	return nil
}

// Close implements drivers.Connection.
func (h *handle) Close() error {
	return h.bucket.Close()
}

// AsRegistry implements drivers.Connection.
func (h *handle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Connection.
func (h *handle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Connection.
func (h *handle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (h *handle) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (h *handle) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Connection.
func (h *handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// Migrate implements drivers.Connection.
func (h *handle) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (h *handle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (h *handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return h, true
}

// AsModelExecutor implements drivers.Handle.
func (h *handle) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (h *handle) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsTransporter implements drivers.Connection.
func (h *handle) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (h *handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (h *handle) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (h *handle) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

type sourceProperties struct {
	Path string `mapstructure:"path"`
	url  *globutil.URL
}

func parseSourceProperties(propsMap map[string]any) (*sourceProperties, error) {
	props := &sourceProperties{}
	err := mapstructure.WeakDecode(propsMap, props)
	if err != nil {
		return nil, err
	}

	if !doublestar.ValidatePattern(props.Path) {
		return nil, fmt.Errorf("glob pattern %s is invalid", props.Path)
	}

	url, err := globutil.ParseBucketURL(props.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", props.Path, err)
	}
	props.url = url

	return props, nil
}

// ListObjects implements drivers.ObjectStore.
func (h *handle) ListObjects(ctx context.Context, propsMap map[string]any) ([]drivers.ObjectStoreEntry, error) {
	props, err := parseSourceProperties(propsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse propsig: %w", err)
	}

	bucket, err := rillblob.NewBucket(h.bucket, h.logger)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	return bucket.ListObjects(ctx, props.url.Path)
}

// DownloadFiles implements drivers.ObjectStore.
func (h *handle) DownloadFiles(ctx context.Context, src map[string]any) (drivers.FileIterator, error) {
	return nil, errors.New("not implemented")
}
