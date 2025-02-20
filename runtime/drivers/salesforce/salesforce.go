package salesforce

import (
	"context"
	"errors"

	force "github.com/ForceCLI/force/lib"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("salesforce", driver{})
	drivers.RegisterAsConnector("salesforce", driver{})
	force.Log = silentLogger{}
}

type silentLogger struct{}

func (silentLogger) Info(args ...any) {
}

var spec = drivers.Spec{
	DisplayName: "Salesforce",
	Description: "Connect to Salesforce.",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "username",
			Type:   drivers.StringPropertyType,
			Secret: false,
		},
		{
			Key:    "password",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
		{
			Key:    "key",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
		{
			Key:    "endpoint",
			Type:   drivers.StringPropertyType,
			Secret: false,
		},
		{
			Key:    "client_id",
			Type:   drivers.StringPropertyType,
			Secret: false,
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "soql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SOQL",
			Description: "SOQL Query to extract data from Salesforce.",
			Placeholder: "SELECT Id, CreatedDate, Name FROM Opportunity",
		},
		{
			Key:         "sobject",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SObject",
			Description: "SObject to query in Salesforce.",
			Placeholder: "Opportunity",
		},
		{
			Key:         "queryAll",
			Type:        drivers.BooleanPropertyType,
			Required:    false,
			DisplayName: "Query All",
			Description: "Include deleted and archived records",
		},
		{
			Key:         "username",
			Type:        drivers.StringPropertyType,
			DisplayName: "Salesforce Username",
			Required:    false,
			Placeholder: "user@example.com",
			Hint:        "Either set this or pass --env connector.salesforce.username=... to rill start",
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			DisplayName: "Salesforce Password",
			Required:    false,
			Hint:        "Either set this or pass --env connector.salesforce.password=... to rill start",
		},
		{
			Key:         "key",
			Type:        drivers.StringPropertyType,
			DisplayName: "JWT Key for Authentication",
			Required:    false,
			Hint:        "Either set this or pass --env connector.salesforce.key=... to rill start",
		},
		{
			Key:         "endpoint",
			Type:        drivers.StringPropertyType,
			DisplayName: "Login Endpoint",
			Required:    false,
			Default:     "login.salesforce.com",
			Placeholder: "login.salesforce.com",
			Hint:        "Either set this or pass --env connector.salesforce.endpoint=... to rill start",
		},
		{
			Key:         "client_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "Connected App Client Id",
			Required:    false,
			Default:     defaultClientID,
			Hint:        "Either set this or pass --env connector.salesforce.client_id=... to rill start",
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
	ImplementsWarehouse: true,
}

type driver struct{}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("salesforce driver can't be shared")
	}
	// actual db connection is opened during query
	return &connection{
		config: config,
		logger: logger,
	}, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type connection struct {
	config map[string]any
	logger *zap.Logger
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	return drivers.ErrNotImplemented
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "salesforce"
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	return c.config
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

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return c, true
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}
