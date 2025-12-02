package admin

import (
	"context"
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
func (s *Service) CreateProject(ctx context.Context, org *database.Organization, opts *database.InsertProjectOptions, deploy bool) (*database.Project, error) {
	// Get roles for initial setup
	adminRole, err := s.DB.FindProjectRole(ctx, database.ProjectRoleNameAdmin)
	if err != nil {
		return nil, err
	}

	// Get the autogroup:members group
	allMembers, err := s.DB.FindUsergroupByName(ctx, org.Name, database.UsergroupNameAutogroupMembers)
	if err != nil {
		return nil, err
	}

	// Create the project and add initial members using a transaction.
	// The transaction is not used for provisioning and deployments, since they involve external services.
	txCtx, tx, err := s.DB.NewTx(ctx, false)
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
		err = s.InsertProjectMemberUser(txCtx, org.ID, proj.ID, *opts.CreatedByUserID, adminRole.ID, nil)
		if err != nil {
			return nil, err
		}
	}

	// Add the system-managed autogroup:members group to the project with the org.DefaultProjectRoleID role (if configured)
	if org.DefaultProjectRoleID != nil {
		err = s.DB.InsertProjectMemberUsergroup(txCtx, allMembers.ID, proj.ID, *org.DefaultProjectRoleID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var createdByID, createdByEmail string
	if opts.CreatedByUserID != nil {
		user, err := s.DB.FindUser(ctx, *proj.CreatedByUserID)
		if err == nil {
			createdByID = user.ID
			createdByEmail = user.Email
		}
	}

	s.Logger.Info("created project", zap.String("id", proj.ID), zap.String("name", proj.Name), zap.String("org", org.Name), zap.String("user_id", createdByID), zap.String("user_email", createdByEmail))

	// Exit early if not deploying
	if !deploy {
		return proj, nil
	}

	// Check if the project has an archive or git info
	hasArchive := opts.ArchiveAssetID != nil
	hasGitInfo := opts.GitRemote != nil && opts.GithubInstallationID != nil && opts.ProdBranch != ""
	if !hasArchive && !hasGitInfo {
		return nil, fmt.Errorf("failed to deploy project: either an archive or git info must be provided")
	}

	// Provision prod deployment.
	// Start using original context again since transaction in txCtx is done.
	depl, err := s.CreateDeployment(ctx, &CreateDeploymentOptions{
		ProjectID:   proj.ID,
		OwnerUserID: nil,
		Environment: "prod",
		Branch:      proj.ProdBranch,
	})
	if err != nil {
		return nil, err
	}

	// Update prod deployment on project
	res, err := s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		DirectoryName:        proj.DirectoryName,
		ArchiveAssetID:       proj.ArchiveAssetID,
		GitRemote:            proj.GitRemote,
		GithubInstallationID: proj.GithubInstallationID,
		GithubRepoID:         proj.GithubRepoID,
		ManagedGitRepoID:     proj.ManagedGitRepoID,
		Provisioner:          proj.Provisioner,
		ProdVersion:          proj.ProdVersion,
		ProdBranch:           proj.ProdBranch,
		Subpath:              proj.Subpath,
		ProdSlots:            proj.ProdSlots,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		ProdDeploymentID:     &depl.ID,
		DevSlots:             proj.DevSlots,
		DevTTLSeconds:        proj.DevTTLSeconds,
		Annotations:          proj.Annotations,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// TeardownProject tears down a project and all its deployments.
func (s *Service) TeardownProject(ctx context.Context, p *database.Project) error {
	ds, err := s.DB.FindDeploymentsForProject(ctx, p.ID)
	if err != nil {
		return err
	}

	// Teardown all deployments in background jobs.
	for _, d := range ds {
		err := s.TeardownDeployment(ctx, d)
		if err != nil {
			return err
		}
	}

	// Poll until all deployments are deleted with a timeout.
	pollCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	for {
		select {
		case <-pollCtx.Done():
			return pollCtx.Err()
		case <-time.After(2 * time.Second):
			// Ready to check again.
		}
		depls, err := s.DB.FindDeploymentsForProject(ctx, p.ID)
		if err != nil {
			return err
		}
		if len(depls) == 0 {
			break
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
	impactsDeployments := (proj.ProdVersion != opts.ProdVersion) ||
		(proj.ProdSlots != opts.ProdSlots) ||
		(proj.Name != opts.Name) ||
		(proj.Subpath != opts.Subpath) ||
		(proj.ProdBranch != opts.ProdBranch) ||
		!reflect.DeepEqual(proj.Annotations, opts.Annotations) ||
		!reflect.DeepEqual(proj.GitRemote, opts.GitRemote) ||
		!reflect.DeepEqual(proj.GithubInstallationID, opts.GithubInstallationID) ||
		!reflect.DeepEqual(proj.ArchiveAssetID, opts.ArchiveAssetID)

	proj, err := s.DB.UpdateProject(ctx, proj.ID, opts)
	if err != nil {
		return nil, err
	}

	if !impactsDeployments {
		return proj, nil
	}

	s.Logger.Info("update project: updating deployments", observability.ZapCtx(ctx))

	err = s.UpdateDeploymentsForProject(ctx, proj)
	if err != nil {
		return nil, err
	}

	return proj, nil
}

// UpdateProjectVariables updates a project's variables and runs reconcile on the deployments.
func (s *Service) UpdateProjectVariables(ctx context.Context, project *database.Project, environment string, vars map[string]string, unsetVars []string, userID string) error {
	if len(vars) == 0 && len(unsetVars) == 0 {
		return nil
	}
	txCtx, tx, err := s.DB.NewTx(ctx, false)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Upsert variables
	if len(vars) > 0 {
		_, err = s.DB.UpsertProjectVariable(txCtx, project.ID, environment, vars, userID)
		if err != nil {
			return err
		}
	}

	// Delete unset variables
	if len(unsetVars) > 0 {
		err = s.DB.DeleteProjectVariables(txCtx, project.ID, environment, unsetVars)
		if err != nil {
			return err
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	// Update deployments
	s.Logger.Info("update project variables: updating deployments", observability.ZapCtx(ctx))

	err = s.UpdateDeploymentsForProject(ctx, project)
	if err != nil {
		return err
	}

	return nil
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
			err := s.UpdateDeploymentsForProject(ctx, proj)
			if err != nil {
				return err
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
	// Provision new deployment
	newDepl, err := s.CreateDeployment(ctx, &CreateDeploymentOptions{
		ProjectID:   proj.ID,
		OwnerUserID: nil,
		Environment: "prod",
		Branch:      proj.ProdBranch,
	})
	if err != nil {
		return nil, err
	}

	// Update prod deployment on project
	proj, err = s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		DirectoryName:        proj.DirectoryName,
		Provisioner:          proj.Provisioner,
		ArchiveAssetID:       proj.ArchiveAssetID,
		GitRemote:            proj.GitRemote,
		GithubInstallationID: proj.GithubInstallationID,
		GithubRepoID:         proj.GithubRepoID,
		ManagedGitRepoID:     proj.ManagedGitRepoID,
		ProdVersion:          proj.ProdVersion,
		ProdBranch:           proj.ProdBranch,
		Subpath:              proj.Subpath,
		ProdDeploymentID:     &newDepl.ID,
		ProdSlots:            proj.ProdSlots,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		DevSlots:             proj.DevSlots,
		DevTTLSeconds:        proj.DevTTLSeconds,
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

// HibernateProject hibernates a project by tearing down its deployment.
func (s *Service) HibernateProject(ctx context.Context, proj *database.Project) (*database.Project, error) {
	depls, err := s.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return nil, err
	}

	for _, depl := range depls {
		err = s.StopDeployment(ctx, depl)
		if err != nil {
			return nil, err
		}
	}

	proj, err = s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		DirectoryName:        proj.DirectoryName,
		Provisioner:          proj.Provisioner,
		ArchiveAssetID:       proj.ArchiveAssetID,
		GitRemote:            proj.GitRemote,
		GithubInstallationID: proj.GithubInstallationID,
		GithubRepoID:         proj.GithubRepoID,
		ManagedGitRepoID:     proj.ManagedGitRepoID,
		ProdVersion:          proj.ProdVersion,
		ProdBranch:           proj.ProdBranch,
		Subpath:              proj.Subpath,
		ProdDeploymentID:     nil,
		ProdSlots:            proj.ProdSlots,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		DevSlots:             proj.DevSlots,
		DevTTLSeconds:        proj.DevTTLSeconds,
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
func (s *Service) ResolveVariables(ctx context.Context, projectID, environment string, forWriting bool) (map[string]string, error) {
	vars, err := s.DB.FindProjectVariables(ctx, projectID, &environment)
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for _, v := range vars {
		res[v.Name] = v.Value
	}
	if forWriting && len(res) == 0 {
		// edge case : no prod variables to set (variable was deleted)
		// but the runtime does not update variables if the new map is empty
		// so we need to set a dummy variable to trigger the update
		res["rill.internal.nonce"] = time.Now().Format(time.RFC3339Nano)
	}
	return res, nil
}
