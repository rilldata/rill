package salesforce

import (
	"context"
	"errors"
	"fmt"
	"maps"

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
	DocsURL:     "https://docs.rilldata.com/connect/data-source/salesforce",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "username",
			Type:        drivers.StringPropertyType,
			DisplayName: "Salesforce Username",
			Placeholder: "user@example.com",
			Hint:        "Your Salesforce username (usually an email address). Required for username-password authentication.",
			Required:    true,
		},
		{
			Key:         "password",
			Type:        drivers.StringPropertyType,
			DisplayName: "Salesforce Password",
			Hint:        "Your Salesforce password, optionally followed by a security token if required.",
			Secret:      true,
		},
		{
			Key:         "key",
			Type:        drivers.StringPropertyType,
			DisplayName: "JWT Key for Authentication",
			Hint:        "Paste your JWT private key for token-based authentication. Used with Connected App and Client ID.",
			Secret:      true,
		},
		{
			Key:         "endpoint",
			Type:        drivers.StringPropertyType,
			DisplayName: "Login Endpoint",
			Placeholder: "login.salesforce.com",
			Hint:        "The Salesforce login URL (e.g., login.salesforce.com or test.salesforce.com for sandboxes).",
			Required:    true,
		},
		{
			Key:         "client_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "Connected App Client Id",
			Default:     defaultClientID,
			Hint:        "The client ID (consumer key) from your Salesforce Connected App. Required for JWT authentication.",
			NoPrompt:    true,
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
	var username, password, endpoint, key, clientID string

	if u, ok := c.config["username"].(string); ok && u != "" {
		username = u
	} else {
		// backwards compatibility: return early because this can be defined in sourceProp
		return nil
	}

	if e, ok := c.config["endpoint"].(string); ok && e != "" {
		endpoint = e
	} else {
		// backwards compatibility: return early because this can be defined in sourceProp
		return nil
	}

	if c, ok := c.config["client_id"].(string); ok && c != "" {
		clientID = c
	} else {
		clientID = defaultClientID
	}

	if p, ok := c.config["password"].(string); ok && p != "" {
		password = p
	}

	if k, ok := c.config["key"].(string); ok && k != "" {
		key = k
	}

	if password == "" && key == "" {
		// backwards compatibility: return early because this can be defined in sourceProp
		return nil
	}

	authOptions := authenticationOptions{
		Username:     username,
		Password:     password,
		JWT:          key,
		Endpoint:     endpoint,
		ConnectedApp: clientID,
	}

	_, err := authenticate(authOptions)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return nil
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
	return maps.Clone(c.config)
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

// AsInformationSchema implements drivers.Connection.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
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
