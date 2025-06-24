package slack

import (
	"context"
	"embed"
	"fmt"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

//go:embed templates/slack/*
var templatesFS embed.FS

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

type configProperties struct {
	BotToken string `mapstructure:"bot_token"`
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

	var api *slack.Client
	if conf.BotToken != "" {
		api = slack.New(conf.BotToken)
	}

	conn := &handle{
		api:       api,
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

type handle struct {
	api       *slack.Client
	config    *configProperties
	logger    *zap.Logger
	templates *template.Template
}

var _ drivers.Handle = &handle{}

// Ping implements drivers.Handle.
func (h *handle) Ping(ctx context.Context) error {
	return drivers.ErrNotImplemented
}

// Driver implements drivers.Handle.
func (h *handle) Driver() string {
	return "slack"
}

// Config implements drivers.Handle.
func (h *handle) Config() map[string]any {
	return map[string]any{}
}

// Migrate implements drivers.Handle.
func (h *handle) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (h *handle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Close implements drivers.Handle.
func (h *handle) Close() error {
	return nil
}

// AsConnection implements drivers.Handle.
func (h *handle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (h *handle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (h *handle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (h *handle) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (h *handle) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Handle.
func (h *handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (h *handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (h *handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (h *handle) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (h *handle) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (h *handle) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsNotifier implements drivers.Handle.
func (h *handle) AsNotifier(properties map[string]any) (drivers.Notifier, bool) {
	return &notifier{h: h, props: properties}, true
}
