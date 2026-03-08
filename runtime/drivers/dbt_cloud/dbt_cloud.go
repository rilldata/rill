package dbt_cloud

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("dbt_cloud", driver{})
	drivers.RegisterAsConnector("dbt_cloud", driver{})
}

var spec = drivers.Spec{
	DisplayName: "dbt Cloud",
	Description: "Connect to dbt Cloud to import semantic layer metrics.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connectors/dbt-cloud",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "api_token",
			Type:        drivers.StringPropertyType,
			DisplayName: "API Token",
			Description: "dbt Cloud API token for authentication",
			Required:    true,
			Secret:      true,
		},
		{
			Key:         "account_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "Account ID",
			Description: "Your dbt Cloud account ID",
			Required:    true,
		},
		{
			Key:         "environment_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "Environment ID",
			Description: "The dbt Cloud environment to fetch manifests from",
			Required:    true,
		},
		{
			Key:         "webhook_secret",
			Type:        drivers.StringPropertyType,
			DisplayName: "Webhook Secret",
			Description: "HMAC secret for validating dbt Cloud webhook payloads",
			Required:    false,
			Secret:      true,
		},
		{
			Key:         "required_connectors",
			Type:        drivers.StringPropertyType,
			DisplayName: "Required Connectors",
			Description: "Comma-separated list of Rill connector names that must be healthy before importing metrics",
			Required:    false,
		},
		{
			Key:         "base_url",
			Type:        drivers.StringPropertyType,
			DisplayName: "Base URL",
			Description: "dbt Cloud host URL (e.g. https://bh369.us1.dbt.com); defaults to https://cloud.getdbt.com",
			Required:    false,
		},
	},
}

// manifestCacheTTL is how long a cached manifest stays valid before refetching.
const manifestCacheTTL = 5 * time.Minute

type driver struct{}

var _ drivers.Driver = driver{}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) Open(connectorName, instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, fmt.Errorf("dbt_cloud driver requires an instance ID")
	}

	cfg := &configProperties{}
	if err := mapstructure.WeakDecode(config, cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("api_token is required")
	}
	if cfg.AccountID == "" {
		return nil, fmt.Errorf("account_id is required")
	}
	if cfg.EnvironmentID == "" {
		return nil, fmt.Errorf("environment_id is required")
	}

	client := NewClient(cfg.APIToken, cfg.AccountID, cfg.BaseURL)
	return &connection{
		config: cfg,
		client: client,
		logger: logger,
	}, nil
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, srcProps map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, srcProps map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

// configProperties holds the decoded connector config.
type configProperties struct {
	APIToken           string `mapstructure:"api_token"`
	AccountID          string `mapstructure:"account_id"`
	EnvironmentID      string `mapstructure:"environment_id"`
	BaseURL            string `mapstructure:"base_url"`
	WebhookSecret      string `mapstructure:"webhook_secret"`
	RequiredConnectors string `mapstructure:"required_connectors"`
	WarehouseConnector string `mapstructure:"warehouse_connector"`
}

// connection is the handle for a dbt_cloud connector.
type connection struct {
	config *configProperties
	client *Client
	logger *zap.Logger

	// Cached manifest
	mu        sync.RWMutex
	manifest  *Manifest
	fetchedAt time.Time
}

var _ drivers.Handle = &connection{}

func (c *connection) Driver() string {
	return "dbt_cloud"
}

func (c *connection) Config() map[string]any {
	m := make(map[string]any)
	_ = mapstructure.Decode(c.config, &m)
	return m
}

func (c *connection) Migrate(ctx context.Context) error {
	return nil
}

func (c *connection) MigrationStatus(ctx context.Context) (current int, desired int, err error) {
	return 0, 0, nil
}

func (c *connection) Close() error {
	return nil
}

func (c *connection) Ping(ctx context.Context) error {
	_, err := c.client.FetchLatestRunWithArtifacts(ctx, c.config.EnvironmentID)
	if err != nil {
		return fmt.Errorf("failed to connect to dbt Cloud: %w", err)
	}
	return nil
}

// GetManifest returns the cached manifest or fetches a fresh one if the cache is expired.
func (c *connection) GetManifest(ctx context.Context) (*Manifest, error) {
	c.mu.RLock()
	if c.manifest != nil && time.Since(c.fetchedAt) < manifestCacheTTL {
		m := c.manifest
		c.mu.RUnlock()
		return m, nil
	}
	c.mu.RUnlock()

	return c.fetchAndCacheManifest(ctx)
}

// InvalidateManifest clears the cached manifest, forcing a refetch on next access.
func (c *connection) InvalidateManifest() {
	c.mu.Lock()
	c.manifest = nil
	c.fetchedAt = time.Time{}
	c.mu.Unlock()
}

// ParseRequiredConnectors returns the list of required connector names.
func (c *connection) ParseRequiredConnectors() []string {
	if c.config.RequiredConnectors == "" {
		return nil
	}
	var res []string
	for _, name := range strings.Split(c.config.RequiredConnectors, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			res = append(res, name)
		}
	}
	return res
}

func (c *connection) fetchAndCacheManifest(ctx context.Context) (*Manifest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check; another goroutine may have fetched while we waited for the lock
	if c.manifest != nil && time.Since(c.fetchedAt) < manifestCacheTTL {
		return c.manifest, nil
	}

	run, err := c.client.FetchLatestRunWithArtifacts(ctx, c.config.EnvironmentID)
	if err != nil {
		return nil, err
	}

	manifest, err := c.client.FetchManifest(ctx, run.ID)
	if err != nil {
		return nil, err
	}

	c.manifest = manifest
	c.fetchedAt = time.Now()
	return manifest, nil
}

// Handle interface: none of these capabilities are supported by dbt_cloud

func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	if opts.InputHandle != c {
		return nil, drivers.ErrNotImplemented
	}
	if _, ok := opts.OutputHandle.AsOLAP(instanceID); !ok {
		return nil, drivers.ErrNotImplemented
	}
	return &dbtCloudToOLAPExecutor{conn: c, instanceID: instanceID}, nil
}

func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// ManifestInvalidator is an interface for invalidating the cached manifest.
// Used by the webhook handler to avoid importing the concrete connection type.
type ManifestInvalidator interface {
	InvalidateManifest()
}

// ManifestProvider is an interface for fetching the dbt manifest.
// Used by the import handler and model executor.
type ManifestProvider interface {
	GetManifest(ctx context.Context) (*Manifest, error)
}
