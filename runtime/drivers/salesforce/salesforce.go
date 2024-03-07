package salesforce

import (
	"context"

	force "github.com/ForceCLI/force/lib"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
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
	SourceProperties: []drivers.PropertySchema{
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
			DisplayName: "Salesforce Username",
			Type:        drivers.StringPropertyType,
			Required:    false,
			Placeholder: "user@example.com",
			Hint:        "Either set this or pass --var connector.salesforce.username=... to rill start",
		},
		{
			Key:         "password",
			DisplayName: "Salesforce Password",
			Type:        drivers.StringPropertyType,
			Required:    false,
			Hint:        "Either set this or pass --var connector.salesforce.password=... to rill start",
		},
		{
			Key:         "key",
			DisplayName: "JWT Key for Authentication",
			Type:        drivers.StringPropertyType,
			Required:    false,
			Hint:        "Either set this or pass --var connector.salesforce.key=... to rill start",
		},
		{
			Key:         "endpoint",
			DisplayName: "Login Endpoint",
			Type:        drivers.StringPropertyType,
			Required:    false,
			Default:     "login.salesforce.com",
			Placeholder: "login.salesforce.com",
			Hint:        "Either set this or pass --var connector.salesforce.endpoint=... to rill start",
		},
		{
			Key:         "client_id",
			DisplayName: "Connected App Client Id",
			Type:        drivers.StringPropertyType,
			Required:    false,
			Default:     defaultClientID,
			Hint:        "Either set this or pass --var connector.salesforce.client_id=... to rill start",
		},
	},
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:    "username",
			Secret: false,
		},
		{
			Key:    "password",
			Secret: true,
		},
		{
			Key:    "key",
			Secret: true,
		},
		{
			Key:    "endpoint",
			Secret: false,
		},
		{
			Key:    "client_id",
			Secret: false,
		},
	},
}

type driver struct{}

func (d driver) Open(config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	// actual db connection is opened during query
	return &connection{
		config: config,
		logger: logger,
	}, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
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

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Connection.
func (c *connection) AsSQLStore() (drivers.SQLStore, bool) {
	return c, true
}
