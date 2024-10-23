package admin

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: The functions in this file are not truly fault tolerant. They should be refactored to run as idempotent, retryable background tasks.

// CreateProject creates a new project and provisions and reconciles a prod deployment for it.
func (s *Service) CreateProject(ctx context.Context, org *database.Organization, opts *database.InsertProjectOptions) (*database.Project, error) {
	isGitInfoEmpty := opts.GithubURL == nil || opts.GithubInstallationID == nil || opts.ProdBranch == ""
	if (opts.ArchiveAssetID == nil) == isGitInfoEmpty {
		return nil, fmt.Errorf("either github info or archive_asset_id must be set")
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
	if opts.CreatedByUserID != nil {
		err = s.DB.InsertProjectMemberUser(txCtx, proj.ID, *opts.CreatedByUserID, adminRole.ID)
		if err != nil {
			return nil, err
		}
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
		Provisioner:    proj.Provisioner,
		Annotations:    s.NewDeploymentAnnotations(org, proj),
		ProdBranch:     proj.ProdBranch,
		ProdVariables:  nil,
		ProdOLAPDriver: proj.ProdOLAPDriver,
		ProdOLAPDSN:    proj.ProdOLAPDSN,
		ProdSlots:      proj.ProdSlots,
		ProdVersion:    proj.ProdVersion,
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
		ArchiveAssetID:       proj.ArchiveAssetID,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		Provisioner:          proj.Provisioner,
		ProdVersion:          proj.ProdVersion,
		ProdBranch:           proj.ProdBranch,
		Subpath:              proj.Subpath,
		ProdSlots:            proj.ProdSlots,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		ProdDeploymentID:     &depl.ID,
		Annotations:          proj.Annotations,
	})
	if err != nil {
		err2 := s.TeardownDeployment(ctx, depl)
		err3 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2, err3)
	}

	// Log project creation
	s.Logger.Info("created project", zap.String("id", proj.ID), zap.String("name", proj.Name), zap.String("org", org.Name), zap.Any("user_id", opts.CreatedByUserID))

	return res, nil
}

// TeardownProject tears down a project and all its deployments.
func (s *Service) TeardownProject(ctx context.Context, p *database.Project) error {
	ds, err := s.DB.FindDeploymentsForProject(ctx, p.ID)
	if err != nil {
		return err
	}

	for _, d := range ds {
		err := s.TeardownDeployment(ctx, d)
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
	requiresReset := (proj.Provisioner != opts.Provisioner) || (proj.ProdSlots != opts.ProdSlots) || (proj.ProdVersion != opts.ProdVersion)

	impactsDeployments := requiresReset ||
		(proj.Name != opts.Name) ||
		(proj.Subpath != opts.Subpath) ||
		(proj.ProdBranch != opts.ProdBranch) ||
		!reflect.DeepEqual(proj.Annotations, opts.Annotations) ||
		!reflect.DeepEqual(proj.GithubURL, opts.GithubURL) ||
		!reflect.DeepEqual(proj.GithubInstallationID, opts.GithubInstallationID) ||
		!reflect.DeepEqual(proj.ArchiveAssetID, opts.ArchiveAssetID)

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

		return s.RedeployProject(ctx, proj, oldDepl)
	}

	s.Logger.Info("update project: updating deployments", observability.ZapCtx(ctx))

	org, err := s.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}
	annotations := s.NewDeploymentAnnotations(org, proj)

	ds, err := s.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return nil, err
	}

	// NOTE: This assumes every deployment (almost always, there's just one) deploys the prod branch.
	// It needs to be refactored when implementing preview deploys.
	for _, d := range ds {
		err := s.UpdateDeployment(ctx, d, &UpdateDeploymentOptions{
			Version:         d.RuntimeVersion,
			Branch:          opts.ProdBranch,
			Variables:       nil,
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

// UpdateProjectVariables updates a project's variables and runs reconcile on the deployments.
func (s *Service) UpdateProjectVariables(ctx context.Context, project *database.Project, environment string, vars map[string]string, unsetVars []string, userID string) ([]*database.ProjectVariable, error) {
	if len(vars) == 0 && len(unsetVars) == 0 {
		return nil, nil
	}
	txCtx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Upsert variables
	var updatedVars []*database.ProjectVariable
	if len(vars) > 0 {
		updatedVars, err = s.DB.UpsertProjectVariable(txCtx, project.ID, environment, vars, userID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Delete unset variables
	if len(unsetVars) > 0 {
		err = s.DB.DeleteProjectVariables(txCtx, project.ID, environment, unsetVars)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Update deployments
	s.Logger.Info("update project variables: updating deployments", observability.ZapCtx(ctx))

	org, err := s.DB.FindOrganization(ctx, project.OrganizationID)
	if err != nil {
		return nil, err
	}

	annotations := s.NewDeploymentAnnotations(org, project)

	ds, err := s.DB.FindDeploymentsForProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	vars, err = s.ResolveVariables(ctx, project.ID, "prod")
	if err != nil {
		return nil, err
	}

	if len(vars) == 0 {
		// edge case : no prod variables to set (variable was deleted)
		// but the runtime does not update variables if the new map is empty
		// so we need to set a dummy variable to trigger the update
		vars = map[string]string{"rill.internal.nonce": time.Now().Format(time.RFC3339Nano)}
	}

	// NOTE: This assumes every deployment (almost always, there's just one) deploys the prod branch.
	// It needs to be refactored when implementing preview deploys.
	for _, d := range ds {
		err := s.UpdateDeployment(ctx, d, &UpdateDeploymentOptions{
			Version:         d.RuntimeVersion,
			Branch:          project.ProdBranch,
			Variables:       vars,
			Annotations:     annotations,
			EvictCachedRepo: true,
		})
		if err != nil {
			// TODO: This may leave things in an inconsistent state. (Although presently, there's almost never multiple deployments.)
			return nil, err
		}
	}

	return updatedVars, nil
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
				err := s.UpdateDeployment(ctx, d, &UpdateDeploymentOptions{
					Version:         d.RuntimeVersion,
					Branch:          proj.ProdBranch,
					Variables:       nil,
					Annotations:     s.NewDeploymentAnnotations(org, proj),
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

// RedeployProject de-provisions and re-provisions a project's prod deployment.
func (s *Service) RedeployProject(ctx context.Context, proj *database.Project, prevDepl *database.Deployment) (*database.Project, error) {
	org, err := s.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	vars, err := s.ResolveVariables(ctx, proj.ID, "prod")
	if err != nil {
		return nil, err
	}

	// Provision new deployment
	newDepl, err := s.createDeployment(ctx, &createDeploymentOptions{
		ProjectID:      proj.ID,
		Provisioner:    proj.Provisioner,
		Annotations:    s.NewDeploymentAnnotations(org, proj),
		ProdVersion:    proj.ProdVersion,
		ProdBranch:     proj.ProdBranch,
		ProdVariables:  vars,
		ProdOLAPDriver: proj.ProdOLAPDriver,
		ProdOLAPDSN:    proj.ProdOLAPDSN,
		ProdSlots:      proj.ProdSlots,
	})
	if err != nil {
		return nil, err
	}

	// Update prod deployment on project
	proj, err = s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		Provisioner:          proj.Provisioner,
		ArchiveAssetID:       proj.ArchiveAssetID,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		ProdVersion:          proj.ProdVersion,
		ProdBranch:           proj.ProdBranch,
		Subpath:              proj.Subpath,
		ProdDeploymentID:     &newDepl.ID,
		ProdSlots:            proj.ProdSlots,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		Annotations:          proj.Annotations,
	})
	if err != nil {
		err2 := s.TeardownDeployment(ctx, newDepl)
		return nil, multierr.Combine(err, err2)
	}

	// Delete old prod deployment if exists
	if prevDepl != nil {
		err = s.TeardownDeployment(ctx, prevDepl)
		if err != nil {
			s.Logger.Error("trigger redeploy: could not teardown old deployment", zap.String("deployment_id", prevDepl.ID), zap.Error(err), observability.ZapCtx(ctx))
		}
	}

	return proj, nil
}

// HibernateProject hibernates a project by tearing down its prod deployment.
func (s *Service) HibernateProject(ctx context.Context, proj *database.Project) (*database.Project, error) {
	depls, err := s.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return nil, err
	}

	for _, depl := range depls {
		err = s.TeardownDeployment(ctx, depl)
		if err != nil {
			return nil, err
		}
	}

	proj, err = s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		Provisioner:          proj.Provisioner,
		ArchiveAssetID:       proj.ArchiveAssetID,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		ProdVersion:          proj.ProdVersion,
		ProdBranch:           proj.ProdBranch,
		Subpath:              proj.Subpath,
		ProdDeploymentID:     nil,
		ProdSlots:            proj.ProdSlots,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		Annotations:          proj.Annotations,
	})
	if err != nil {
		return nil, err
	}

	return proj, nil
}

// TriggerParser triggers the deployment's project parser to do a new pull and parse.
func (s *Service) TriggerParser(ctx context.Context, depl *database.Deployment) (err error) {
	s.Logger.Info("reconcile: triggering pull", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
	defer func() {
		if err != nil {
			s.Logger.Error("reconcile: trigger pull failed", zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
		} else {
			s.Logger.Info("reconcile: trigger pull completed", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
		}
	}()

	rt, err := s.OpenRuntimeClient(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	_, err = rt.CreateTrigger(ctx, &runtimev1.CreateTriggerRequest{
		InstanceId: depl.RuntimeInstanceID,
		Parser:     true,
	})
	return err
}

// TriggerParserAndAwaitResource triggers the parser and polls the runtime until the given resource's spec version has been updated (or ctx is canceled).
func (s *Service) TriggerParserAndAwaitResource(ctx context.Context, depl *database.Deployment, name, kind string) error {
	rt, err := s.OpenRuntimeClient(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	resourceReq := &runtimev1.GetResourceRequest{
		InstanceId: depl.RuntimeInstanceID,
		Name: &runtimev1.ResourceName{
			Kind: kind,
			Name: name,
		},
	}

	// Get old spec version
	var oldSpecVersion *int64
	r, err := rt.GetResource(ctx, resourceReq)
	if err == nil {
		oldSpecVersion = &r.Resource.Meta.SpecVersion
	}

	// Trigger parser
	_, err = rt.CreateTrigger(ctx, &runtimev1.CreateTriggerRequest{
		InstanceId: depl.RuntimeInstanceID,
		Parser:     true,
	})
	if err != nil {
		return err
	}

	// Poll every 1 seconds until the resource is found or the ctx is cancelled or times out
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		r, err := rt.GetResource(ctx, resourceReq)
		if err != nil {
			if s, ok := status.FromError(err); !ok || s.Code() != codes.NotFound {
				return fmt.Errorf("failed to poll for resource: %w", err)
			}
			if oldSpecVersion != nil {
				// Success - previously the resource was found, now we cannot find it anymore
				return nil
			}
			// Continue polling
			continue
		}
		if oldSpecVersion == nil {
			// Success - previously the resource was not found, now we found one
			return nil
		}
		if *oldSpecVersion != r.Resource.Meta.SpecVersion {
			// Success - the spec version has changed
			return nil
		}
	}
}

// ResolveVariables resolves the project's variables for the given environment.
// It fetches the variable specific to the environment plus the default variables not set exclusively for the environment.
func (s *Service) ResolveVariables(ctx context.Context, projectID, environment string) (map[string]string, error) {
	vars, err := s.DB.FindProjectVariables(ctx, projectID, &environment)
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for _, v := range vars {
		res[v.Name] = string(v.Value)
	}
	return res, nil
}
