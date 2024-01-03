package admin

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// TODO: The functions in this file are not truly fault tolerant. They should be refactored to run as idempotent, retryable background tasks.

// CreateProject creates a new project and provisions and reconciles a prod deployment for it.
func (s *Service) CreateProject(ctx context.Context, org *database.Organization, userID string, opts *database.InsertProjectOptions) (*database.Project, error) {
	// Check Github info is set (presently required for deployments)
	if opts.GithubURL == nil || opts.GithubInstallationID == nil || opts.ProdBranch == "" {
		return nil, fmt.Errorf("cannot create project without github info")
	}

	// Get roles for initial setup
	adminRole, err := s.DB.FindProjectRole(ctx, database.ProjectRoleNameAdmin)
	if err != nil {
		panic(err)
	}
	viewerRole, err := s.DB.FindProjectRole(ctx, database.ProjectRoleNameViewer)
	if err != nil {
		panic(err)
	}

	// Create the project and add initial members using a transaction.
	// The transaction is not used for provisioning and deployments, since they involve external services.
	txCtx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	proj, err := s.DB.InsertProject(txCtx, opts)
	if err != nil {
		return nil, err
	}

	// The creating user becomes project admin
	err = s.DB.InsertProjectMemberUser(txCtx, proj.ID, userID, adminRole.ID)
	if err != nil {
		return nil, err
	}

	// All org members as a group get viewer role
	err = s.DB.InsertProjectMemberUsergroup(txCtx, *org.AllUsergroupID, proj.ID, viewerRole.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// Provision prod deployment.
	// Start using original context again since transaction in txCtx is done.
	depl, err := s.createDeployment(ctx, &createDeploymentOptions{
		ProjectID:      proj.ID,
		Region:         proj.Region,
		ProdBranch:     proj.ProdBranch,
		ProdVariables:  proj.ProdVariables,
		ProdOLAPDriver: proj.ProdOLAPDriver,
		ProdOLAPDSN:    proj.ProdOLAPDSN,
		ProdSlots:      proj.ProdSlots,
		Annotations:    newDeploymentAnnotations(org, proj),
	})
	if err != nil {
		err2 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2)
	}

	// Update prod deployment on project
	res, err := s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		ProdBranch:           proj.ProdBranch,
		ProdVariables:        proj.ProdVariables,
		ProdSlots:            proj.ProdSlots,
		Region:               proj.Region,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		Tags:                 proj.Tags,
		ProdDeploymentID:     &depl.ID,
	})
	if err != nil {
		err2 := s.teardownDeployment(ctx, proj, depl)
		err3 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2, err3)
	}

	// Log project creation
	s.Logger.Info("created project", zap.String("id", proj.ID), zap.String("name", proj.Name), zap.String("org", org.Name), zap.String("user_id", userID))

	return res, nil
}

// TeardownProject tears down a project and all its deployments.
func (s *Service) TeardownProject(ctx context.Context, p *database.Project) error {
	ds, err := s.DB.FindDeploymentsForProject(ctx, p.ID)
	if err != nil {
		return err
	}

	for _, d := range ds {
		err := s.teardownDeployment(ctx, p, d)
		if err != nil {
			return err
		}
	}

	err = s.DB.DeleteProject(ctx, p.ID)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProject updates a project and any impacted deployments.
// It runs a reconcile if deployment parameters (like branch or variables) have been changed and reconcileDeployment is set.
func (s *Service) UpdateProject(ctx context.Context, proj *database.Project, opts *database.UpdateProjectOptions) (*database.Project, error) {
	requiresReset := (proj.Region != opts.Region) || (proj.ProdSlots != opts.ProdSlots)

	impactsDeployments := (requiresReset ||
		(proj.Name != opts.Name) ||
		(proj.ProdBranch != opts.ProdBranch) ||
		!reflect.DeepEqual(proj.ProdVariables, opts.ProdVariables) ||
		!reflect.DeepEqual(proj.GithubURL, opts.GithubURL) ||
		!reflect.DeepEqual(proj.GithubInstallationID, opts.GithubInstallationID))

	proj, err := s.DB.UpdateProject(ctx, proj.ID, opts)
	if err != nil {
		return nil, err
	}

	if !impactsDeployments {
		return proj, nil
	}

	if requiresReset {
		s.Logger.Info("update project: resetting deployment", observability.ZapCtx(ctx))

		var oldDepl *database.Deployment
		if proj.ProdDeploymentID != nil {
			oldDepl, err = s.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
			if err != nil && !errors.Is(err, database.ErrNotFound) {
				return nil, err
			}
		}

		return s.TriggerRedeploy(ctx, proj, oldDepl)
	}

	s.Logger.Info("update project: updating deployments", observability.ZapCtx(ctx))

	org, err := s.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}
	annotations := newDeploymentAnnotations(org, proj)

	ds, err := s.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return nil, err
	}

	// NOTE: This assumes every deployment (almost always, there's just one) deploys the prod branch.
	// It needs to be refactored when implementing preview deploys.
	for _, d := range ds {
		err := s.updateDeployment(ctx, d, &updateDeploymentOptions{
			Branch:          opts.ProdBranch,
			Variables:       opts.ProdVariables,
			Annotations:     annotations,
			EvictCachedRepo: true,
		})
		if err != nil {
			// TODO: This may leave things in an inconsistent state. (Although presently, there's almost never multiple deployments.)
			return nil, err
		}
	}

	return proj, nil
}

// UpdateOrgDeploymentAnnotations iterates over projects of the given org and
// updates annotations of corresponding deployments with the new organization name
// NOTE : this does not trigger reconcile.
func (s *Service) UpdateOrgDeploymentAnnotations(ctx context.Context, org *database.Organization) error {
	limit := 10
	afterProjectName := ""
	for {
		projs, err := s.DB.FindProjectsForOrganization(ctx, org.ID, afterProjectName, limit)
		if err != nil {
			return err
		}

		for _, proj := range projs {
			ds, err := s.DB.FindDeploymentsForProject(ctx, proj.ID)
			if err != nil {
				return err
			}

			for _, d := range ds {
				err := s.updateDeployment(ctx, d, &updateDeploymentOptions{
					Branch:          proj.ProdBranch,
					Variables:       proj.ProdVariables,
					Annotations:     newDeploymentAnnotations(org, proj),
					EvictCachedRepo: false,
				})
				if err != nil {
					return err
				}
			}

			afterProjectName = proj.Name
		}

		if len(projs) < limit {
			break
		}
	}

	return nil
}

// TriggerRedeploy de-provisions and re-provisions a project's prod deployment.
func (s *Service) TriggerRedeploy(ctx context.Context, proj *database.Project, prevDepl *database.Deployment) (*database.Project, error) {
	org, err := s.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	// Provision new deployment
	newDepl, err := s.createDeployment(ctx, &createDeploymentOptions{
		ProjectID:      proj.ID,
		Region:         proj.Region,
		ProdBranch:     proj.ProdBranch,
		ProdVariables:  proj.ProdVariables,
		ProdOLAPDriver: proj.ProdOLAPDriver,
		ProdOLAPDSN:    proj.ProdOLAPDSN,
		ProdSlots:      proj.ProdSlots,
		Annotations:    newDeploymentAnnotations(org, proj),
	})
	if err != nil {
		return nil, err
	}

	// Update prod deployment on project
	proj, err = s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		ProdBranch:           proj.ProdBranch,
		ProdVariables:        proj.ProdVariables,
		ProdDeploymentID:     &newDepl.ID,
		ProdSlots:            proj.ProdSlots,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		Region:               proj.Region,
		Tags:                 proj.Tags,
	})
	if err != nil {
		err2 := s.teardownDeployment(ctx, proj, newDepl)
		return nil, multierr.Combine(err, err2)
	}

	// Delete old prod deployment if exists
	if prevDepl != nil {
		err = s.teardownDeployment(ctx, proj, prevDepl)
		if err != nil {
			s.Logger.Error("trigger redeploy: could not teardown old deployment", zap.String("deployment_id", prevDepl.ID), zap.Error(err), observability.ZapCtx(ctx))
		}
	}

	return proj, nil
}

// TriggerReconcile triggers a reconcile for a deployment.
func (s *Service) TriggerReconcile(ctx context.Context, depl *database.Deployment) (err error) {
	s.Logger.Info("reconcile: triggering pull", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
	defer func() {
		if err != nil {
			s.Logger.Error("reconcile: trigger pull failed", zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
		} else {
			s.Logger.Info("reconcile: trigger pull completed", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
		}
	}()

	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	_, err = rt.CreateTrigger(ctx, &runtimev1.CreateTriggerRequest{
		InstanceId: depl.RuntimeInstanceID,
		Trigger: &runtimev1.CreateTriggerRequest_PullTriggerSpec{
			PullTriggerSpec: &runtimev1.PullTriggerSpec{},
		},
	})
	return err
}

// TriggerRefreshSource triggers refresh of a deployment's sources. If the sources slice is nil, it will refresh all sources.
func (s *Service) TriggerRefreshSources(ctx context.Context, depl *database.Deployment, sources []string) (err error) {
	s.Logger.Info("reconcile: triggering refresh", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
	defer func() {
		if err != nil {
			s.Logger.Error("reconcile: trigger refresh failed", zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
		} else {
			s.Logger.Info("reconcile: trigger refresh completed", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
		}
	}()

	names := make([]*runtimev1.ResourceName, 0, len(sources))
	for _, source := range sources {
		// NOTE: When keeping Kind empty, the RefreshTrigger will match both sources and models
		names = append(names, &runtimev1.ResourceName{Name: source})
	}

	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	_, err = rt.CreateTrigger(ctx, &runtimev1.CreateTriggerRequest{
		InstanceId: depl.RuntimeInstanceID,
		Trigger: &runtimev1.CreateTriggerRequest_RefreshTriggerSpec{
			RefreshTriggerSpec: &runtimev1.RefreshTriggerSpec{OnlyNames: names},
		},
	})
	return err
}
