package webhook

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
)

var spec = drivers.Spec{
	DisplayName: "Webhook",
	Description: "Webhook Notifier",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "signing_secret",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
	},
	ImplementsNotifier: true,
}

func init() {
	drivers.Register("webhook", driver{})
	drivers.RegisterAsConnector("webhook", driver{})
}

type driver struct{}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) Open(_, instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, fmt.Errorf("webhook driver can't be shared")
	}
	conf := &configProperties{}
	err := mapstructure.Decode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &handle{
		config: conf,
		logger: logger,
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
	config *configProperties
	logger *zap.Logger
}

var _ drivers.Handle = &handle{}

// Ping implements drivers.Handle.
func (h *handle) Ping(ctx context.Context) error {
	// Delivery URLs are provided per alert/report, so there is no receiver to contact at
	// the connector level. Validate that the signing secret (if any) is well-formed so a
	// misconfiguration surfaces here instead of on the first delivery.
	_, err := signingKey(h.config.SigningSecret)
	return err
}

func (h *handle) Driver() string {
	return "webhook"
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

func (h *handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsInformationSchema implements drivers.Handle.
func (h *handle) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

func (h *handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

func (h *handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (h *handle) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

func (h *handle) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (h *handle) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

func (h *handle) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return newNotifier(h.config, properties)
}

type configProperties struct {
	// SigningSecret enables signing payloads following the Standard Webhooks specification
	// (https://www.standardwebhooks.com/). Secrets prefixed with "whsec_" are treated as
	// base64-encoded (the format the spec defines); other values are used as raw key bytes.
	SigningSecret string `mapstructure:"signing_secret"`
	// Headers are static headers added to every delivery, e.g. an Authorization header for
	// receivers behind an API gateway.
	Headers map[string]string `mapstructure:"headers"`
}
