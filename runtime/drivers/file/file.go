package file

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/filewatcher"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func init() {
	drivers.Register("file", driver{name: "file"})
	drivers.Register("local_file", driver{name: "local_file"})
	drivers.RegisterAsConnector("local_file", driver{name: "local_file"})
}

var spec = drivers.Spec{
	DisplayName: "Local file",
	Description: "Import Locally Stored File.",
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Path",
			Description: "Path or URL to file",
			Placeholder: "/path/to/file",
		},
		{
			Key:         "format",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Format",
			Description: "Either CSV or Parquet. Inferred if not set.",
			Placeholder: "csv",
		},
	},
	ImplementsFileStore: true,
}

type driver struct {
	name string
}

type configProperties struct {
	DSN             string `mapstructure:"dsn"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
	AdminURL        string `mapstructure:"admin_url"`
	Org             string `mapstructure:"org"`
	AccessToken     string `mapstructure:"access_token"`
}

// a smaller subset of relevant parts of rill.yaml
type rillYAML struct {
	IgnorePaths []string `yaml:"ignore_paths"`
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("file driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	path, err := fileutil.ExpandHome(conf.DSN)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	admin, err := client.New(conf.AdminURL, conf.AccessToken, "rill-runtime")
	if err != nil {
		return nil, fmt.Errorf("failed to open admin client: %w", err)
	}

	c := &connection{
		logger:       logger,
		root:         absPath,
		driverConfig: conf,
		driverName:   d.name,
		admin:        admin,
	}
	if err := c.checkRoot(); err != nil {
		return nil, err
	}

	// Read rill.yaml and fill in `ignore_paths`
	rawYaml, err := c.Get(context.Background(), "/rill.yaml")
	if err == nil {
		yml := &rillYAML{}
		err = yaml.Unmarshal([]byte(rawYaml), yml)
		if err == nil {
			c.ignorePaths = yml.IgnorePaths
		}
	}

	// Setup the file watcher
	c.watcher = filewatcher.NewLazyWatcher(c.root, c.ignorePaths, c.logger)

	return c, nil
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

type sourceProperties struct {
	Path   string `mapstructure:"path"`
	Format string `mapstructure:"format"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

type connection struct {
	logger       *zap.Logger
	root         string // root should be absolute path
	driverConfig *configProperties
	driverName   string

	admin     *client.Client  // admin client for admin service
	gitConfig *gitutil.Config // git config for repo
	gitMu     sync.Mutex      // mutex to protect git operations

	watcher     *filewatcher.LazyWatcher
	ignorePaths []string
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	return c.checkRoot()
}

// Driver implements drivers.Handle.
func (c *connection) Driver() string {
	return c.driverName
}

// Config implements drivers.Handle.
func (c *connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.driverConfig, &m)
	return m
}

// InformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// Close implements drivers.Handle.
func (c *connection) Close() error {
	return nil
}

// AsRegistry implements drivers.Handle.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return c, true
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Handle.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// Migrate implements drivers.Handle.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Handle.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	if opts.OutputHandle == c {
		if olap, ok := opts.InputHandle.AsOLAP(instanceID); ok {
			return &olapToSelfExecutor{c, olap}, nil
		}
	}
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return c, true
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Handle.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// checkPath checks that the connection's root is a valid directory.
func (c *connection) checkRoot() error {
	info, err := os.Stat(c.root)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("repo: directory does not exist at '%s'", c.root)
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("repo: file is not a directory '%s'", c.root)
	}

	return nil
}
