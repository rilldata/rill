package azure

import (
	"context"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("azure", driver{})
	drivers.RegisterAsConnector("azure", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Azure Blob Storage",
	Description: "Connect to Azure Blob Storage.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/azure",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "azure_storage_account",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
		{
			Key:    "azure_storage_key",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
		{
			Key:    "azure_storage_sas_token",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
		{
			Key:    "azure_storage_connection_string",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			DisplayName: "Blob URI",
			Description: "Path to file on the disk.",
			Placeholder: "azure://container-name/path/to/file.csv",
			Required:    true,
			Hint:        "Glob patterns are supported",
		},
		{
			Key:         "account",
			Type:        drivers.StringPropertyType,
			DisplayName: "Account name",
			Description: "Azure storage account name.",
			Required:    false,
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
	ImplementsObjectStore: true,
}

type driver struct{}

type configProperties struct {
	Account          string `mapstructure:"azure_storage_account"`
	Key              string `mapstructure:"azure_storage_key"`
	SASToken         string `mapstructure:"azure_storage_sas_token"`
	ConnectionString string `mapstructure:"azure_storage_connection_string"`
	AllowHostAccess  bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(instanceID string, config map[string]any, storage *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("azure driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config:  conf,
		storage: storage,
		logger:  logger,
	}
	return conn, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, props map[string]any, logger *zap.Logger) (bool, error) {
	conf, err := parseSourceProperties(props)
	if err != nil {
		return false, fmt.Errorf("failed to parse config: %w", err)
	}

	conn := &Connection{
		config: &configProperties{},
		logger: logger,
	}

	bucketObj, err := conn.openBucketWithNoCredentials(ctx, conf)
	if err != nil {
		return false, fmt.Errorf("failed to open container %q, %w", conf.url.Host, err)
	}
	defer bucketObj.Close()

	return bucketObj.IsAccessible(ctx)
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type Connection struct {
	config  *configProperties
	storage *storage.Client
	logger  *zap.Logger
}

var _ drivers.Handle = &Connection{}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	return drivers.ErrNotImplemented
}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "azure"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, &m)
	return m
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	return nil
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

// AsTransporter implements drivers.Connection.
func (c *Connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *Connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Connection.
func (c *Connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *Connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}
