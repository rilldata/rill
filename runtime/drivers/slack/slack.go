package slack

import (
	"context"
	"fmt"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
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

func (d driver) Open(instanceID string, config map[string]any, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, fmt.Errorf("slack driver can't be shared")
	}
	conf := &configProperties{}
	err := mapstructure.Decode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &handle{
		config:    conf,
		logger:    logger,
		templates: template.Must(template.New("").ParseFS(templatesFS, "templates/slack/*.slack")),
	}
	return conn, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return nil
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, props map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type handle struct {
	config    *configProperties
	logger    *zap.Logger
	templates *template.Template
}

var _ drivers.Handle = &handle{}

func (h *handle) Driver() string {
	return "slack"
}

func (h *handle) Config() map[string]any {
	return map[string]any{}
}

func (h *handle) Migrate(ctx context.Context) error {
	return nil
}

func (h *handle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

func (h *handle) Close() error {
	return nil
}

func (h *handle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

func (h *handle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

func (h *handle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

func (h *handle) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

func (h *handle) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

func (h *handle) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

func (h *handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

func (h *handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

func (h *handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

func (h *handle) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

func (h *handle) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return newNotifier(h.config.BotToken, properties)
}

type configProperties struct {
	BotToken string `mapstructure:"bot_token"`
}
