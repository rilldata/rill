package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/client"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *Service) CreateProject(ctx context.Context, opts *database.InsertProjectOptions) (*database.Project, error) {
	// TODO: Make this actually fault tolerant.

	org, err := s.DB.FindOrganization(ctx, opts.OrganizationID)
	if err != nil {
		return nil, err
	}

	txCtx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Create the project
	proj, err := s.DB.InsertProject(txCtx, opts)
	if err != nil {
		return nil, err
	}

	adminRole, err := s.DB.FindProjectRole(txCtx, database.ProjectRoleNameAdmin)
	if err != nil {
		panic(errors.Wrap(err, "failed to find project admin role"))
	}

	// add project admin role to the user
	err = s.DB.InsertProjectMemberUser(txCtx, proj.ID, opts.UserID, adminRole.ID)
	if err != nil {
		return nil, err
	}

	// add project viewer role to the all_user_group of the org
	viewerRole, err := s.DB.FindProjectRole(txCtx, database.ProjectRoleNameViewer)
	if err != nil {
		panic(errors.Wrap(err, "failed to find project viewer role"))
	}
	err = s.DB.InsertProjectMemberUsergroup(txCtx, *org.AllUsergroupID, proj.ID, viewerRole.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if proj.GithubURL == nil || proj.GithubInstallationID == nil {
		return nil, fmt.Errorf("cannot create project without github info")
	}

	// Provision it
	provOpts := &provisioner.ProvisionOptions{
		OLAPDriver:           proj.ProdOLAPDriver,
		OLAPDSN:              proj.ProdOLAPDSN,
		Region:               proj.Region,
		Slots:                proj.ProdSlots,
		GithubURL:            *proj.GithubURL,
		GitBranch:            proj.ProdBranch,
		GithubInstallationID: *proj.GithubInstallationID,
		Variables:            proj.ProdVariables,
	}
	// start using original context again since transaction in txCtx is done
	inst, err := s.provisioner.Provision(ctx, provOpts)
	if err != nil {
		err = fmt.Errorf("provisioner: %w", err)
		err2 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2)
	}

	// Store deployment
	depl, err := s.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         proj.ID,
		Branch:            proj.ProdBranch,
		Slots:             proj.ProdSlots,
		RuntimeHost:       inst.Host,
		RuntimeInstanceID: inst.InstanceID,
		RuntimeAudience:   inst.Audience,
		Status:            database.DeploymentStatusPending,
		Logs:              "",
	})
	if err != nil {
		err2 := s.provisioner.Teardown(ctx, inst.Host, inst.InstanceID, proj.ProdOLAPDriver)
		err3 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2, err3)
	}

	// Update prod deployment on project
	res, err := s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		ProdBranch:           proj.ProdBranch,
		ProdVariables:        proj.ProdVariables,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		ProdDeploymentID:     &depl.ID,
	})
	if err != nil {
		err2 := s.DB.DeleteDeployment(ctx, depl.ID)
		err3 := s.provisioner.Teardown(ctx, inst.Host, inst.InstanceID, proj.ProdOLAPDriver)
		err4 := s.DB.DeleteProject(ctx, proj.ID)
		return nil, multierr.Combine(err, err2, err3, err4)
	}

	// Trigger reconcile
	err = s.TriggerReconcile(ctx, depl.ID)
	if err != nil {
		// This error is weird. But it's safe not to teardown the rest.
		return nil, err
	}

	return res, nil
}

func (s *Service) TeardownProject(ctx context.Context, p *database.Project) error {
	// TODO: Make this actually fault tolerant.

	ds, err := s.DB.FindDeployments(ctx, p.ID)
	if err != nil {
		return err
	}

	for _, d := range ds {
		err := s.provisioner.Teardown(ctx, d.RuntimeHost, d.RuntimeInstanceID, p.ProdOLAPDriver)
		if err != nil {
			return err
		}

		err = s.DB.DeleteDeployment(ctx, d.ID)
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

// New method for trigger redeploy (tear down instance and deploy again)
func (s *Service) TriggerRedeploy(ctx context.Context, orgName, name string) error {
	// Run it all in the background
	go func() {
		// Use s.closeCtx to cancel if the service is stopped
		ctx := s.closeCtx

		proj, err := s.DB.FindProjectByName(ctx, orgName, name)
		if err != nil {
			s.logger.Error("trigger redeploy: could not find project", zap.String("project name", name), zap.Error(err))
			return
		}

		// Get current deployment_id
		deploymentID := *proj.ProdDeploymentID

		// Get current deployment
		depl, err := s.DB.FindDeployment(ctx, deploymentID)
		if err != nil {
			s.logger.Error("trigger redeploy: could not find deployment", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		// Check status
		if depl.Status == database.DeploymentStatusReconciling && time.Since(depl.UpdatedOn) < 30*time.Minute {
			s.logger.Error("trigger redeploy: skipping because it is already running", zap.String("deployment_id", deploymentID))
			return
		}

		// Trigger a new deployment
		err = s.triggerDeployment(ctx, proj)
		if err != nil {
			s.logger.Error("Error in new deployment", zap.Error(err))
			return
		}

		// Set old deployment status to reconciling
		depl, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusReconciling, "")
		if err != nil {
			s.logger.Error("trigger redeploy: could not update status", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		err = s.provisioner.Teardown(ctx, depl.RuntimeHost, depl.RuntimeInstanceID, proj.ProdOLAPDriver)
		if err != nil {
			s.logger.Error("trigger redeploy: could not tear down instance", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		// Deleting the deployment
		err = s.DB.DeleteDeployment(ctx, depl.ID)
		if err != nil {
			s.logger.Error("trigger redeploy: could not delete deployment", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}
	}()
	return nil
}

// New method for refresh source
func (s *Service) TriggerRefreshSource(ctx context.Context, orgName, name, sourceName string) error {
	// Run it all in the background
	go func() {
		// Use s.closeCtx to cancel if the service is stopped
		ctx := s.closeCtx

		proj, err := s.DB.FindProjectByName(ctx, orgName, name)
		if err != nil {
			s.logger.Error("trigger redeploy: could not find project", zap.String("project name", name), zap.Error(err))
			return
		}

		deploymentID := *proj.ProdDeploymentID

		s.logger.Info("refresh source: starting", zap.String("deployment_id", deploymentID))

		// Get deployment
		depl, err := s.DB.FindDeployment(ctx, deploymentID)
		if err != nil {
			s.logger.Error("refresh source: could not find deployment", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		// Check status
		if depl.Status == database.DeploymentStatusReconciling && time.Since(depl.UpdatedOn) < 30*time.Minute {
			s.logger.Error("refresh source: skipping because it is already running", zap.String("deployment_id", deploymentID))
			return
		}

		// Set deployment status to reconciling
		depl, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusReconciling, "")
		if err != nil {
			s.logger.Error("refresh source: could not update status", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		// Get superuser token for runtime host
		jwt, err := s.issuer.NewToken(auth.TokenOptions{
			AudienceURL:         depl.RuntimeAudience,
			TTL:                 time.Hour,
			InstancePermissions: map[string][]auth.Permission{depl.RuntimeInstanceID: {auth.EditInstance}},
		})
		if err != nil {
			s.logger.Error("refresh source: could not get token", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		// Make runtime client
		rt, err := client.New(depl.RuntimeHost, jwt)
		if err != nil {
			s.logger.Error("refresh source: could not create client", zap.String("deployment_id", deploymentID), zap.Error(err))
			return
		}

		// Call refresh source with source name
		res, err := rt.TriggerRefresh(ctx, &runtimev1.TriggerRefreshRequest{InstanceId: depl.RuntimeInstanceID, Name: sourceName})
		if err != nil {
			s.logger.Error("refresh source: rpc error", zap.String("deployment_id", deploymentID), zap.Error(err))
			_, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusError, err.Error())
			if err != nil {
				s.logger.Error("refresh source: could not update logs", zap.String("deployment_id", deploymentID), zap.Error(err))
			}
			return
		}

		// Set status
		if len(res.Errors) > 0 {
			json, err := protojson.Marshal(res)
			if err != nil {
				s.logger.Error("refresh source: json error", zap.String("deployment_id", deploymentID), zap.Error(err))
				return
			}

			_, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusError, string(json))
			if err != nil {
				s.logger.Error("refresh source: could not update logs", zap.String("deployment_id", deploymentID), zap.Error(err))
				return
			}
		} else {
			_, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusOK, "")
			if err != nil {
				s.logger.Error("refresh source: could not clear logs", zap.String("deployment_id", deploymentID), zap.Error(err))
				return
			}
		}

		s.logger.Info("refresh source: completed", zap.String("deployment_id", deploymentID))
	}()
	return nil
}

func (s *Service) TriggerReconcile(ctx context.Context, deploymentID string) error {
	// TODO: Make this actually fault tolerant

	// Run it all in the background
	go func() {
		// Use s.closeCtx to cancel if the service is stopped
		ctx := s.closeCtx

		s.logger.Info("reconcile: starting", zap.String("deployment_id", deploymentID), observability.ZapCtx(ctx))

		// Get deployment
		depl, err := s.DB.FindDeployment(ctx, deploymentID)
		if err != nil {
			s.logger.Error("reconcile: could not find deployment", zap.String("deployment_id", deploymentID), zap.Error(err), observability.ZapCtx(ctx))
			return
		}

		// Check status
		if depl.Status == database.DeploymentStatusReconciling && time.Since(depl.UpdatedOn) < 30*time.Minute {
			s.logger.Error("reconcile: skipping because it is already running", zap.String("deployment_id", deploymentID), observability.ZapCtx(ctx))
			return
		}

		// Set deployment status to reconciling
		depl, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusReconciling, "")
		if err != nil {
			s.logger.Error("reconcile: could not update status", zap.String("deployment_id", deploymentID), zap.Error(err), observability.ZapCtx(ctx))
			return
		}

		// Get superuser token for runtime host
		jwt, err := s.issuer.NewToken(auth.TokenOptions{
			AudienceURL:         depl.RuntimeAudience,
			TTL:                 time.Hour,
			InstancePermissions: map[string][]auth.Permission{depl.RuntimeInstanceID: {auth.EditInstance}},
		})
		if err != nil {
			s.logger.Error("reconcile: could not get token", zap.String("deployment_id", deploymentID), zap.Error(err), observability.ZapCtx(ctx))
			return
		}

		// Make runtime client
		rt, err := client.New(depl.RuntimeHost, jwt)
		if err != nil {
			s.logger.Error("reconcile: could not create client", zap.String("deployment_id", deploymentID), zap.Error(err), observability.ZapCtx(ctx))
			return
		}

		// Call reconcile
		res, err := rt.Reconcile(ctx, &runtimev1.ReconcileRequest{InstanceId: depl.RuntimeInstanceID})
		if err != nil {
			s.logger.Error("reconcile: rpc error", zap.String("deployment_id", deploymentID), zap.Error(err), observability.ZapCtx(ctx))
			_, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusError, err.Error())
			if err != nil {
				s.logger.Error("reconcile: could not update logs", zap.String("deployment_id", deploymentID), zap.Error(err), observability.ZapCtx(ctx))
			}
			return
		}

		// Set status
		if len(res.Errors) > 0 {
			json, err := protojson.Marshal(res)
			if err != nil {
				s.logger.Error("reconcile: json error", zap.String("deployment_id", deploymentID), zap.Error(err), observability.ZapCtx(ctx))
				return
			}

			_, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusError, string(json))
			if err != nil {
				s.logger.Error("reconcile: could not update logs", zap.String("deployment_id", deploymentID), zap.Error(err), observability.ZapCtx(ctx))
				return
			}
		} else {
			_, err = s.DB.UpdateDeploymentStatus(ctx, deploymentID, database.DeploymentStatusOK, "")
			if err != nil {
				s.logger.Error("reconcile: could not clear logs", zap.String("deployment_id", deploymentID), zap.Error(err), observability.ZapCtx(ctx))
				return
			}
		}

		s.logger.Info("reconcile: completed", zap.String("deployment_id", deploymentID), observability.ZapCtx(ctx))
	}()
	return nil
}

func (s *Service) UpdateProject(ctx context.Context, projID string, opts *database.UpdateProjectOptions) (*database.Project, error) {
	// TODO: Make this actually fault tolerant.

	// TODO: Handle if ProdBranch or GithubURL was changed.

	ds, err := s.DB.FindDeployments(ctx, projID)
	if err != nil {
		return nil, err
	}

	for _, d := range ds {
		if err := s.editInstance(ctx, d, opts.ProdVariables); err != nil {
			return nil, err
		}
	}

	// Update the project
	proj, err := s.DB.UpdateProject(ctx, projID, opts)
	if err != nil {
		return nil, err
	}
	return proj, nil
}

func (s *Service) editInstance(ctx context.Context, d *database.Deployment, variables map[string]string) error {
	jwt, err := s.issuer.NewToken(auth.TokenOptions{
		AudienceURL:       d.RuntimeAudience,
		TTL:               time.Hour,
		SystemPermissions: []auth.Permission{auth.ManageInstances, auth.ReadInstance},
	})
	if err != nil {
		return err
	}

	rt, err := client.New(d.RuntimeHost, jwt)
	if err != nil {
		return err
	}
	defer rt.Close()

	resp, err := rt.GetInstance(ctx, &runtimev1.GetInstanceRequest{
		InstanceId: d.RuntimeInstanceID,
	})
	if err != nil {
		return err
	}

	// Edit the instance
	inst := resp.Instance
	_, err = rt.EditInstance(ctx, &runtimev1.EditInstanceRequest{
		InstanceId:          d.RuntimeInstanceID,
		OlapDriver:          inst.OlapDriver,
		OlapDsn:             inst.OlapDsn,
		RepoDriver:          inst.RepoDriver,
		RepoDsn:             inst.RepoDsn,
		EmbedCatalog:        inst.EmbedCatalog,
		Variables:           variables,
		IngestionLimitBytes: inst.IngestionLimitBytes,
	})
	return err
}

func (s *Service) triggerDeployment(ctx context.Context, proj *database.Project) error {
	// Provision new deployment
	provOpts := &provisioner.ProvisionOptions{
		OLAPDriver:           proj.ProdOLAPDriver,
		OLAPDSN:              proj.ProdOLAPDSN,
		Region:               proj.Region,
		Slots:                proj.ProdSlots,
		GithubURL:            *proj.GithubURL,
		GitBranch:            proj.ProdBranch,
		GithubInstallationID: *proj.GithubInstallationID,
		Variables:            proj.ProdVariables,
	}
	inst, err := s.provisioner.Provision(ctx, provOpts)
	if err != nil {
		err = fmt.Errorf("provisioner: %w", err)
		err2 := s.DB.DeleteProject(ctx, proj.ID)
		return fmt.Errorf("trigger redeploy: could not provision instance, instance_id %s, Error:%w", inst.InstanceID, multierr.Combine(err, err2))
	}

	// Store new deployment details
	depl, err := s.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         proj.ID,
		Branch:            proj.ProdBranch,
		Slots:             proj.ProdSlots,
		RuntimeHost:       inst.Host,
		RuntimeInstanceID: inst.InstanceID,
		RuntimeAudience:   inst.Audience,
		Status:            database.DeploymentStatusPending,
		Logs:              "",
	})
	if err != nil {
		err2 := s.provisioner.Teardown(ctx, inst.Host, inst.InstanceID, proj.ProdOLAPDriver)
		err3 := s.DB.DeleteProject(ctx, proj.ID)
		return fmt.Errorf("trigger redeploy: could not insert deployment, instance_id %s, Error:%w", inst.InstanceID, multierr.Combine(err2, err3))
	}

	// Update prod deployment on project
	_, err = s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		ProdBranch:           proj.ProdBranch,
		ProdVariables:        proj.ProdVariables,
		GithubURL:            proj.GithubURL,
		GithubInstallationID: proj.GithubInstallationID,
		ProdDeploymentID:     &depl.ID,
	})
	if err != nil {
		err2 := s.DB.DeleteDeployment(ctx, depl.ID)
		err3 := s.provisioner.Teardown(ctx, inst.Host, inst.InstanceID, proj.ProdOLAPDriver)
		err4 := s.DB.DeleteProject(ctx, proj.ID)
		return fmt.Errorf("trigger redeploy: could not update project, instance_id %s, Error:%w", inst.InstanceID, multierr.Combine(err, err2, err3, err4))
	}

	// Trigger reconcile
	err = s.TriggerReconcile(ctx, depl.ID)
	if err != nil {
		// This error is weird. But it's safe not to teardown the rest.
		return fmt.Errorf("trigger redeploy: could not reconcile project, deployment_id %s, Error:%w", depl.ID, err)
	}

	return nil
}
