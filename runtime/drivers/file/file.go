package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("file", driver{name: "file"})
	drivers.Register("local_file", driver{name: "local_file"})
	drivers.RegisterAsConnector("local_file", driver{name: "local_file"})
}

var spec = drivers.Spec{
	DisplayName: "Local file",
	Description: "Import Locally Stored File.",
	SourceProperties: []drivers.PropertySchema{
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
}

type driver struct {
	name string
}

func (d driver) Open(config map[string]any, logger *zap.Logger) (drivers.Connection, error) {
	dsn, ok := config["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("require dsn to open file connection")
	}

	path, err := fileutil.ExpandHome(dsn)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	c := &connection{
		root:         absPath,
		driverConfig: config,
		driverName:   d.name,
	}
	if err := c.checkRoot(); err != nil {
		return nil, err
	}
	return c, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src drivers.Source, logger *zap.Logger) (bool, error) {
	return true, nil
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
	// root should be absolute path
	root         string
	driverConfig map[string]any
	driverName   string

	watcherMu    sync.Mutex
	watcherCount int
	watcher      *watcher
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	return c.driverConfig
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return nil
}

// Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *connection) AsRepoStore() (drivers.RepoStore, bool) {
	return c, true
}

// OLAP implements drivers.Connection.
func (c *connection) AsOLAP() (drivers.OLAPStore, bool) {
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

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Connection) (drivers.Transporter, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return c, true
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
