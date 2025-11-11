package https

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("https", driver{})
	drivers.RegisterAsConnector("https", driver{})
}

var spec = drivers.Spec{
	DisplayName: "https",
	Description: "Connect to a remote file.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connect/#adding-a-remote-source",
	// Important: Any edits to the below properties must be accompanied by changes to the client-side form validation schemas.
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			DisplayName: "Path",
			Description: "Path to the remote file.",
			Placeholder: "https://example.com/file.csv",
			Required:    true,
		},
		{
			Key:         "name",
			Type:        drivers.StringPropertyType,
			DisplayName: "Source name",
			Description: "The name of the source",
			Placeholder: "my_new_source",
			Required:    true,
		},
	},
	ImplementsFileStore: true,
}

type driver struct{}

type ConfigProperties struct {
	Headers map[string]string `mapstructure:"headers"`
}

type SourceProperties struct {
	Path    string            `mapstructure:"path"`
	URI     string            `mapstructure:"uri"`
	Headers map[string]string `mapstructure:"headers"`
}

func (s *SourceProperties) ResolvePath() string {
	// Backwards compatibility for "uri" renamed to "path"
	if s.URI != "" {
		return s.URI
	}
	return s.Path
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("https driver can't be shared")
	}

	cfg := &ConfigProperties{}
	err := mapstructure.WeakDecode(config, cfg)
	if err != nil {
		return nil, err
	}

	conn := &connection{
		config: config,
		logger: logger,
	}
	return conn, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return true, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type connection struct {
	config map[string]any
	logger *zap.Logger
}

var _ drivers.Handle = &connection{}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	// no properties to define in connector so ping always return true.
	return nil
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "https"
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	m := make(map[string]any)
	err := mapstructure.Decode(c.config, &m)
	if err != nil {
		c.logger.Warn("error in generating https config", zap.Error(err))
	}
	return m
}

// InformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return nil
}

// AsRegistry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Connection.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return c, true
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// FilePaths implements drivers.FileStore
func (c *connection) FilePaths(ctx context.Context, src map[string]any) ([]string, error) {
	srcProp := &SourceProperties{}
	if err := mapstructure.WeakDecode(src, srcProp); err != nil {
		return nil, fmt.Errorf("failed to parse properties: %w", err)
	}

	path := srcProp.ResolvePath()
	if path == "" {
		return nil, fmt.Errorf("missing required property: `path`")
	}
	extension, err := urlExtension(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension from path %s, %w", path, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for path %s:  %w", path, err)
	}

	for k, v := range srcProp.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url %s:  %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to fetch url %s: %s", path, resp.Status)
	}

	file, _, err := fileutil.CopyToTempFile(resp.Body, "", extension)
	if err != nil {
		return nil, err
	}

	return []string{file}, nil
}

func urlExtension(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	return fileutil.FullExt(u.Path), nil
}
