package localfile

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type Config struct {
	Path   string `mapstructure:"path"`
	Format string `mapstructure:"format"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
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

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "local_file"
}

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
func (c *connection) AsTransporter(from drivers.Connection, to drivers.Connection) (drivers.Transporter, bool) {
	return nil, false
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return c, true
}

func (c *connection) FilePaths(ctx context.Context, src *drivers.FilesSource) ([]string, error) {
	return []string{src.Properties["paths"].(string)}, nil
}

// AsConnector implements drivers.Connection.
func (c *connection) AsConnector() (drivers.Connector, bool) {
	return c, true
}
