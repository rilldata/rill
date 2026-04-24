package local

import (
	"context"
	"errors"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

// localAdminService implements drivers.AdminService by using user's admin token stored locally and calling Rill's admin API.
type localAdminService struct {
	ch          *cmdutil.Helper
	root        string
	environment string
	frontendURL string
}

var (
	_ drivers.AdminService = &localAdminService{}
	_ drivers.Handle       = &localAdminService{}
)

var s = &localAdminService{}

func init() {
	drivers.Register("local_admin", drivers.NewEmbeddedDriver(s, drivers.Spec{
		ImplementsAdmin: true,
	}))
}

func initLocalAdminService(ch *cmdutil.Helper, root, environment, frontendURL string) {
	s.ch = ch
	s.root = root
	s.environment = environment
	s.frontendURL = frontendURL
}

// GetAlertMetadata implements drivers.AdminService.
func (l *localAdminService) GetAlertMetadata(ctx context.Context, alertName, ownerID string, emailRecipients []string, anonRecipients bool, annotations map[string]string, queryForUserID, queryForUserEmail string) (*drivers.AlertMetadata, error) {
	return nil, drivers.ErrAlertsNotSupported
}

// GetDeploymentConfig implements drivers.AdminService.
func (l *localAdminService) GetDeploymentConfig(ctx context.Context) (*drivers.DeploymentConfig, error) {
	return nil, drivers.ErrNotImplemented
}

// GetReportMetadata implements drivers.AdminService.
func (l *localAdminService) GetReportMetadata(ctx context.Context, reportName, ownerID, webOpenMode string, emailRecipients []string, anonRecipients bool, executionTime time.Time) (*drivers.ReportMetadata, error) {
	return nil, drivers.ErrReportsNotSupported
}

// ProvisionConnector implements drivers.AdminService.
func (l *localAdminService) ProvisionConnector(ctx context.Context, name, driver string, args map[string]any) (map[string]any, error) {
	return nil, drivers.ErrProvisioningNotSupported
}

// ListDeployments implements drivers.AdminService.
func (l *localAdminService) ListDeployments(ctx context.Context) ([]*drivers.Deployment, error) {
	if l.ch.AdminToken() == "" {
		return nil, drivers.ErrNotAuthenticated
	}

	client, err := l.ch.Client()
	if err != nil {
		return nil, err
	}

	projects, err := l.ch.InferProjects(ctx, l.ch.Org, l.root)
	if err != nil {
		if errors.Is(err, cmdutil.ErrInferProjectFailed) {
			// Succeed with an empty list
			return nil, nil
		}
		return nil, err
	}
	project := projects[0] // InferProjects always returns at least one project in case of no error

	resp, err := client.ListDeployments(ctx, &adminv1.ListDeploymentsRequest{
		Org:     project.OrgName,
		Project: project.Name,
	})
	if err != nil {
		return nil, err
	}

	res := make([]*drivers.Deployment, 0, len(resp.Deployments))
	for _, d := range resp.Deployments {
		res = append(res, &drivers.Deployment{
			Branch: d.Branch,
		})
	}

	return res, nil
}

// AsAI implements [drivers.Handle].
func (l *localAdminService) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsAdmin implements [drivers.Handle].
func (l *localAdminService) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return l, true
}

// AsCatalogStore implements [drivers.Handle].
func (l *localAdminService) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsFileStore implements [drivers.Handle].
func (l *localAdminService) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsInformationSchema implements [drivers.Handle].
func (l *localAdminService) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsModelExecutor implements [drivers.Handle].
func (l *localAdminService) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements [drivers.Handle].
func (l *localAdminService) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// AsNotifier implements [drivers.Handle].
func (l *localAdminService) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotImplemented
}

// AsOLAP implements [drivers.Handle].
func (l *localAdminService) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements [drivers.Handle].
func (l *localAdminService) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsRegistry implements [drivers.Handle].
func (l *localAdminService) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsRepoStore implements [drivers.Handle].
func (l *localAdminService) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsWarehouse implements [drivers.Handle].
func (l *localAdminService) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// Close implements [drivers.Handle].
func (l *localAdminService) Close() error {
	return nil
}

// Config implements [drivers.Handle].
func (l *localAdminService) Config() map[string]any {
	return map[string]any{}
}

// Driver implements [drivers.Handle].
func (l *localAdminService) Driver() string {
	return "local_admin"
}

// Migrate implements [drivers.Handle].
func (l *localAdminService) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements [drivers.Handle].
func (l *localAdminService) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Ping implements [drivers.Handle].
func (l *localAdminService) Ping(ctx context.Context) error {
	return nil
}
