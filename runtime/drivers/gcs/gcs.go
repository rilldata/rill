package gcs

import (
	"context"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/gcputil"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"gocloud.dev/gcp"
)

func init() {
	drivers.Register("gcs", driver{})
	drivers.RegisterAsConnector("gcs", driver{})

	// Alternate name
	drivers.Register("gs", driver{})
	drivers.RegisterAsConnector("gs", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Google Cloud Storage",
	Description: "Connect to Google Cloud Storage.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/gcs",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:  "google_application_credentials",
			Type: drivers.FilePropertyType,
			Hint: "Enter path of file to load from.",
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			DisplayName: "GS URI",
			Description: "Path to file on the disk.",
			Placeholder: "gs://bucket-name/path/to/file.csv",
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
	SecretJSON      string `mapstructure:"google_application_credentials"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
	// When working in s3 compatible mode
	KeyID  string `mapstructure:"key_id"`
	Secret string `mapstructure:"secret"`
}

func NewConfigProperties(in map[string]any) (*ConfigProperties, error) {
	gcsConfig := &ConfigProperties{}
	err := mapstructure.WeakDecode(in, gcsConfig)
	if err != nil {
		return nil, err
	}
	return gcsConfig, nil
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("gcs driver can't be shared")
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

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
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
	if c.config.SecretJSON != "" {
		creds, err := gcputil.Credentials(ctx, c.config.SecretJSON, c.config.AllowHostAccess)
		if err != nil {
			return fmt.Errorf("failed to load credentials: %w", err)
		}

		ts := gcp.CredentialsTokenSource(creds)
		_, err = ts.Token()
		if err != nil {
			return fmt.Errorf("failed to retrieve access token: %w", err)
		}
	}

	if c.config.KeyID != "" && c.config.Secret != "" {
		// TODO: handle in case of HMAC key secret
		return nil
	}

	return nil
}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "gcs"
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
	return c, true
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

func (c *Connection) newClient(ctx context.Context) (*gcp.HTTPClient, error) {
	creds, err := gcputil.Credentials(ctx, c.config.SecretJSON, c.config.AllowHostAccess)
	if err != nil {
		if !errors.Is(err, gcputil.ErrNoCredentials) {
			return nil, err
		}

		// no credentials set, we try with a anonymous client in case user is trying to access public buckets
		return gcp.NewAnonymousHTTPClient(gcp.DefaultTransport()), nil
	}
	// the token source returned from credentials works for all kind of credentials like serviceAccountKey, credentialsKey etc.
	return gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
}
