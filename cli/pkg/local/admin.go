package local

import (
	"context"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

// localAdminService implements drivers.AdminService by using user's admin token stored locally and calling Rill's admin API.
// TODO: revisit this implementation.
type localAdminService struct {
	ch   *cmdutil.Helper
	root string
}

var _ drivers.AdminService = &localAdminService{}

func newLocalAdminService(ch *cmdutil.Helper, root string) drivers.AdminService {
	return &localAdminService{
		ch:   ch,
		root: root,
	}
}

// GetAlertMetadata implements drivers.AdminService.
func (l *localAdminService) GetAlertMetadata(ctx context.Context, alertName, ownerID string, emailRecipients []string, anonRecipients bool, annotations map[string]string, queryForUserID, queryForUserEmail string) (*drivers.AlertMetadata, error) {
	return nil, drivers.ErrNotImplemented
}

// GetDeploymentConfig implements drivers.AdminService.
func (l *localAdminService) GetDeploymentConfig(ctx context.Context) (*drivers.DeploymentConfig, error) {
	return nil, drivers.ErrNotImplemented
}

// GetReportMetadata implements drivers.AdminService.
func (l *localAdminService) GetReportMetadata(ctx context.Context, reportName, resolver, ownerID, webOpenMode string, emailRecipients []string, anonRecipients bool, executionTime time.Time) (*drivers.ReportMetadata, error) {
	return nil, drivers.ErrNotImplemented
}

// ProvisionConnector implements drivers.AdminService.
func (l *localAdminService) ProvisionConnector(ctx context.Context, name, driver string, args map[string]any) (map[string]any, error) {
	return nil, drivers.ErrNotImplemented
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
