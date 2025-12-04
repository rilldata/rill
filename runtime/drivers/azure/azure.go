package azure

import (
	"context"
	"errors"
	"fmt"
	"os"

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
	DocsURL:     "https://docs.rilldata.com/build/connectors/data-source/azure",
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
	// Important: Any edits to the below properties must be accompanied by changes to the client-side form validation schemas.
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

type ConfigProperties struct {
	Account          string `mapstructure:"azure_storage_account"`
	Key              string `mapstructure:"azure_storage_key"`
	SASToken         string `mapstructure:"azure_storage_sas_token"`
	ConnectionString string `mapstructure:"azure_storage_connection_string"`
	// A list of container or virtual directory prefixes that this connector is allowed to access.
	// Useful when different containers or paths use different credentials, allowing the system
	// to route access through the appropriate connector based on the blob path.
	// Example formats: `azure://my-container/` `azure://my-container/path/` `azure://my-container/path/prefix`
	PathPrefixes    []string `mapstructure:"path_prefixes"`
	AllowHostAccess bool     `mapstructure:"allow_host_access"`
}

func (c *ConfigProperties) GetAccount() string {
	if c.Account == "" && c.AllowHostAccess {
		return os.Getenv("AZURE_STORAGE_ACCOUNT")
	}
	return c.Account
}

func (c *ConfigProperties) GetKey() string {
	if c.Key == "" && c.AllowHostAccess {
		return os.Getenv("AZURE_STORAGE_KEY")
	}
	return c.Key
}

func (c *ConfigProperties) GetSASToken() string {
	if c.SASToken == "" && c.AllowHostAccess {
		return os.Getenv("AZURE_STORAGE_SAS_TOKEN")
	}
	return c.SASToken
}

func (c *ConfigProperties) GetConnectionString() string {
	if c.AllowHostAccess {
		if c.ConnectionString == "" {
			c.ConnectionString = os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
		}
	}

	if c.ConnectionString != "" {
		return c.ConnectionString
	}

	// If no auth provided, return empty so caller can use AAD
	if c.GetKey() == "" && c.GetSASToken() == "" {
		return ""
	}

	dsn := fmt.Sprintf("AccountName=%s;", c.GetAccount())

	var authPart string
	switch {
	case c.GetKey() != "":
		authPart = fmt.Sprintf("AccountKey=%s;", c.GetKey())
	case c.GetSASToken() != "":
		authPart = fmt.Sprintf("SharedAccessSignature=%s;", c.GetSASToken())
	}
	dsn += authPart

	return dsn
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("azure driver can't be shared")
	}

	conf := &ConfigProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config:  conf,
		storage: st,
		logger:  logger,
	}
	return conn, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, props map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type Connection struct {
	config  *ConfigProperties
	storage *storage.Client
	logger  *zap.Logger
}

var _ drivers.Handle = &Connection{}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	client, err := c.newStorageClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Azure storage client: %w", err)
	}

	_, err = client.GetAccountInfo(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get Azure account info: %w", err)
	}

	return nil
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

// ParsedConfig returns the parsed configuration of the connection.
func (c *Connection) ParsedConfig() *ConfigProperties {
	cpy := *c.config
	return &cpy
}

// InformationSchema implements drivers.Handle.
func (c *Connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
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
func (c *Connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *Connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

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
