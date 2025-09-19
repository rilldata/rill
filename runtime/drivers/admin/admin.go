package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/admin/client"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/drivers/admin")

var spec = drivers.Spec{
	DisplayName: "Rill Admin",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:      "admin_url",
			Type:     drivers.StringPropertyType,
			Required: true,
		},
		{
			Key:      "access_token",
			Type:     drivers.StringPropertyType,
			Required: true,
			Secret:   true,
		},
	},
}

func init() {
	drivers.Register("admin", driver{})
}

type driver struct{}

var _ drivers.Driver = driver{}

type configProperties struct {
	AdminURL    string `mapstructure:"admin_url"`
	AccessToken string `mapstructure:"access_token"`
	ProjectID   string `mapstructure:"project_id"`
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("admin driver can't be shared")
	}

	cfg := &configProperties{}
	err := mapstructure.WeakDecode(config, cfg)
	if err != nil {
		return nil, err
	}

	admin, err := client.New(cfg.AdminURL, cfg.AccessToken, "rill-runtime")
	if err != nil {
		return nil, fmt.Errorf("failed to open admin client: %w", err)
	}

	h := &Handle{
		config:  cfg,
		logger:  logger,
		storage: st,
		admin:   admin,
	}
	h.repo = newRepo(h)

	return h, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, props map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type Handle struct {
	config  *configProperties
	logger  *zap.Logger
	storage *storage.Client
	admin   *client.Client
	repo    *repo
}

var _ drivers.Handle = &Handle{}

// Ping implements drivers.Handle.
func (h *Handle) Ping(ctx context.Context) error {
	// Check connectivity with admin service.
	_, err := h.admin.Ping(ctx, &adminv1.PingRequest{})
	if err != nil {
		return err
	}

	// Check for a repo error
	_, err = h.repo.checkReady(ctx)
	return err
}

// Driver implements drivers.Handle.
func (h *Handle) Driver() string {
	return "admin"
}

// Config implements drivers.Handle.
func (h *Handle) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(h.config, &m)
	return m
}

// Migrate implements drivers.Handle.
func (h *Handle) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (h *Handle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Close implements drivers.Handle.
func (h *Handle) Close() error {
	return errors.Join(
		h.repo.close(),
		h.admin.Close(),
	)
}

// AsRegistry implements drivers.Handle.
func (h *Handle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (h *Handle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (h *Handle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return h.repo, true
}

// AsAdmin implements drivers.Handle.
func (h *Handle) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return h, true
}

// AsAI implements drivers.Handle.
func (h *Handle) AsAI(instanceID string) (drivers.AIService, bool) {
	return h, true
}

// AsOLAP implements drivers.Handle.
func (h *Handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// InformationSchema implements drivers.Handle.
func (h *Handle) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (h *Handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (h *Handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (h *Handle) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (h *Handle) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (h *Handle) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsNotifier implements drivers.Handle.
func (h *Handle) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}
