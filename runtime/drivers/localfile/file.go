package localfile

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
)

type config struct {
	Path   string `mapstructure:"path"`
	Format string `mapstructure:"format"`
}

func parseConfig(props map[string]any) (*config, error) {
	conf := &config{}
	err := mapstructure.Decode(props, &conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func init() {
	drivers.Register("local_file", driver{})
}

type driver struct{}

func (d driver) Open(config map[string]any, logger *zap.Logger) (drivers.Connection, error) {
	conn := &connection{
		config: config,
	}
	return conn, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

type connection struct {
	config map[string]any
}

var _ drivers.Connection = &connection{}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "local_file"
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return nil
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	return c.config
}

// Registry implements drivers.Connection.
func (c *connection) RegistryStore() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) CatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *connection) RepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *connection) OLAPStore() (drivers.OLAPStore, bool) {
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

// AsFileStore implements drivers.Connection
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return c, true
}

// FilePaths implements drivers.FileStore
func (c *connection) FilePaths(ctx context.Context, src *drivers.FilesSource) ([]string, error) {
	conf, err := parseConfig(src.Properties)
	if err != nil {
		return nil, err
	}

	path, err := c.resolveLocalPath(conf.Path)
	if err != nil {
		return nil, err
	}

	// get all files in case glob passed
	localPaths, err := doublestar.FilepathGlob(path)
	if err != nil {
		return nil, err
	}
	if len(localPaths) == 0 {
		return nil, fmt.Errorf("file does not exist at %s", conf.Path)
	}

	return localPaths, nil
}

func (c *connection) resolveLocalPath(path string) (string, error) {
	path, err := fileutil.ExpandHome(path)
	if err != nil {
		return "", err
	}

	var repoRoot string
	val, ok := c.config["repo_root"]
	if ok {
		repoRoot = val.(string)
	}
	finalPath := path
	if !filepath.IsAbs(path) {
		finalPath = filepath.Join(repoRoot, path)
	}

	allowHostAccess := false
	if val, ok := c.config["allow_host_access"]; ok {
		allowHostAccess = val.(bool)
	}
	if !allowHostAccess && !strings.HasPrefix(finalPath, repoRoot) {
		// path is outside the repo root
		return "", fmt.Errorf("file connector cannot ingest source: path is outside repo root")
	}
	return finalPath, nil
}

// AsConnector implements drivers.Connection.
func (c *connection) AsConnector() (drivers.Connector, bool) {
	return c, true
}
