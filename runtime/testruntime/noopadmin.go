package testruntime

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
)

type noopAdminService struct{}

var (
	_ drivers.AdminService = &noopAdminService{}
	_ drivers.Handle       = &noopAdminService{}
)

func init() {
	drivers.Register("noop_admin", drivers.NewEmbeddedDriver(&noopAdminService{}, drivers.Spec{
		ImplementsAdmin: true,
	}))
}

func (n *noopAdminService) GetAlertMetadata(ctx context.Context, alertName, ownerID string, emailRecipients []string, anonRecipients bool, annotations map[string]string, queryForUserID, queryForUserEmail string) (*drivers.AlertMetadata, error) {
	return nil, drivers.ErrNotImplemented
}

func (n *noopAdminService) GetDeploymentConfig(ctx context.Context) (*drivers.DeploymentConfig, error) {
	return nil, drivers.ErrNotImplemented
}

func (n *noopAdminService) GetReportMetadata(ctx context.Context, reportName, ownerID, webOpenMode string, emailRecipients []string, anonRecipients bool, executionTime time.Time) (*drivers.ReportMetadata, error) {
	return nil, drivers.ErrNotImplemented
}

func (n *noopAdminService) ProvisionConnector(ctx context.Context, name, driver string, args map[string]any) (map[string]any, error) {
	return nil, drivers.ErrNotImplemented
}

func (n *noopAdminService) ListDeployments(ctx context.Context) ([]*drivers.Deployment, error) {
	return nil, drivers.ErrNotImplemented
}

// AsAI implements [drivers.Handle].
func (n *noopAdminService) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsAdmin implements [drivers.Handle].
func (n *noopAdminService) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return n, true
}

// AsCatalogStore implements [drivers.Handle].
func (n *noopAdminService) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsFileStore implements [drivers.Handle].
func (n *noopAdminService) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsInformationSchema implements [drivers.Handle].
func (n *noopAdminService) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsModelExecutor implements [drivers.Handle].
func (n *noopAdminService) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements [drivers.Handle].
func (n *noopAdminService) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// AsNotifier implements [drivers.Handle].
func (n *noopAdminService) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotImplemented
}

// AsOLAP implements [drivers.Handle].
func (n *noopAdminService) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements [drivers.Handle].
func (n *noopAdminService) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsRegistry implements [drivers.Handle].
func (n *noopAdminService) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsRepoStore implements [drivers.Handle].
func (n *noopAdminService) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsWarehouse implements [drivers.Handle].
func (n *noopAdminService) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// Close implements [drivers.Handle].
func (n *noopAdminService) Close() error {
	return nil
}

// Config implements [drivers.Handle].
func (n *noopAdminService) Config() map[string]any {
	return map[string]any{}
}

// Driver implements [drivers.Handle].
func (n *noopAdminService) Driver() string {
	return "noop_admin"
}

// Migrate implements [drivers.Handle].
func (n *noopAdminService) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements [drivers.Handle].
func (n *noopAdminService) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Ping implements [drivers.Handle].
func (n *noopAdminService) Ping(ctx context.Context) error {
	return nil
}
