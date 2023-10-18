package s3

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
)

var spec = drivers.Spec{
	DisplayName: "Rill Admin",
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:    "access_token",
			Secret: true,
		},
	},
}

func init() {
	drivers.Register("admin", driver{})
}

type driver struct{}

var _ drivers.Driver = driver{}

type configProperties struct {
	AccessToken string `mapstructure:"access_token"`
}

func (d driver) Open(cfgMap map[string]any, shared bool, client activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("admin driver can't be shared")
	}

	cfg := &configProperties{}
	err := mapstructure.Decode(cfgMap, cfg)
	if err != nil {
		return nil, err
	}

	h := &Handle{
		config: cfg,
		logger: logger,
	}
	return h, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, props map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type Handle struct {
	config *configProperties
	logger *zap.Logger
}

var _ drivers.Handle = &Handle{}

// Driver implements drivers.Handle.
func (c *Handle) Driver() string {
	return "admin"
}

// Config implements drivers.Handle.
func (c *Handle) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, &m)
	return m
}

// Migrate implements drivers.Handle.
func (c *Handle) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *Handle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Close implements drivers.Handle.
func (c *Handle) Close() error {
	return nil
}

// Registry implements drivers.Handle.
func (c *Handle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Handle.
func (c *Handle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Handle.
func (c *Handle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Handle.
func (c *Handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (c *Handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (c *Handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsTransporter implements drivers.Handle.
func (c *Handle) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Handle.
func (c *Handle) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}
