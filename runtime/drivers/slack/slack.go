package slack

import (
	"context"
	"fmt"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
)

var spec = drivers.Spec{
	DisplayName: "Slack",
	Description: "Slack Notifier",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "bot_token",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
	},
	ImplementsNotifier: true,
}

func init() {
	drivers.Register("slack", driver{})
	drivers.RegisterAsConnector("slack", driver{})
}

type driver struct{}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, fmt.Errorf("slack driver can't be shared")
	}
	conf := &configProperties{}
	err := mapstructure.Decode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config:    conf,
		logger:    logger,
		templates: template.Must(template.New("").ParseFS(templatesFS, "templates/slack/*.slack")),
	}
	return conn, nil
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, props map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type Connection struct {
	config    *configProperties
	logger    *zap.Logger
	templates *template.Template
}

var _ drivers.Handle = &Connection{}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	if c.config.BotToken == "" {
		return fmt.Errorf("bot token not configured")
	}

	// Create a test notifier to verify the token
	notifier, err := newNotifier(c.config.BotToken, nil)
	if err != nil {
		return fmt.Errorf("failed to create notifier: %w", err)
	}

	_, err = notifier.api.AuthTest()
	if err != nil {
		return fmt.Errorf("failed to verify bot token: %w", err)
	}

	return nil
}

func (c *Connection) Driver() string {
	return "slack"
}

func (c *Connection) Config() map[string]any {
	return map[string]any{}
}

func (c *Connection) Migrate(ctx context.Context) error {
	return nil
}

func (c *Connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// InformationSchema implements drivers.Handle.
func (c *Connection) InformationSchema() drivers.InformationSchema {
	return &drivers.NotImplementedInformationSchema{}
}

func (c *Connection) Close() error {
	return nil
}

func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

func (c *Connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

func (c *Connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

func (c *Connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

func (c *Connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

func (c *Connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

func (c *Connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *Connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

func (c *Connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (c *Connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

func (c *Connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return newNotifier(c.config.BotToken, properties)
}

type configProperties struct {
	BotToken string `mapstructure:"bot_token"`
}
