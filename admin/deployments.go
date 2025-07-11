package admin

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/client"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateDeploymentOptions struct {
	ProjectID   string
	OwnerUserID *string
	Environment string
	Annotations DeploymentAnnotations
	Branch      string
	Provisioner string
	Slots       int
	Version     string
	Variables   map[string]string
}

func (s *Service) CreateDeployment(ctx context.Context, opts *CreateDeploymentOptions) (*database.Deployment, error) {
	// Create the deployment
	depl, err := s.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         opts.ProjectID,
		OwnerUserID:       opts.OwnerUserID,
		Environment:       opts.Environment,
		Branch:            opts.Branch,
		RuntimeHost:       "", // Will be populated after provisioning in startDeploymentInner
		RuntimeInstanceID: "", // Will be populated after provisioning in startDeploymentInner
		RuntimeAudience:   "", // Will be populated after provisioning in startDeploymentInner
		Status:            database.DeploymentStatusPending,
		StatusMessage:     "Provisioning...",
	})
	if err != nil {
		return nil, err
	}

	// Initialize the deployment (by provisioning a runtime and creating an instance on it)
	depl, err = s.StartDeployment(ctx, depl, &StartDeploymentOptions{
		Annotations: opts.Annotations,
		Provisioner: opts.Provisioner,
		Slots:       opts.Slots,
		Version:     opts.Version,
		Variables:   opts.Variables,
	})
	if err != nil {
		return nil, err
	}

	return depl, nil
}

type StartDeploymentOptions struct {
	Annotations DeploymentAnnotations
	Provisioner string
	Slots       int
	Version     string
	Variables   map[string]string
}

func (s *Service) StartDeployment(ctx context.Context, depl *database.Deployment, opts *StartDeploymentOptions) (*database.Deployment, error) {
	// Update the deployment status to pending
	_, err := s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusPending, "Provisioning...")
	if err != nil {
		return nil, err
	}

	// Initialize the deployment (by provisioning a runtime and creating an instance on it)
	err = s.startDeploymentInner(ctx, depl, opts)
	if err != nil {
		// Mark deployment error
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		_, err2 := s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusError, fmt.Sprintf("Failed to provision runtime: %v", err))
		s.Logger.Error("start deployment: failed to provision runtime", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
		// TODO: The validate_deployments job will tear it down, but we should consider starting a background job to do so immediately.
		return nil, errors.Join(err, err2)
	}

	// Mark deployment ready
	depl1, err := s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusOK, "")
	if err != nil {
		// NOTE: Unlikely case – we'll leave it pending in this case, the user can reset.
		return nil, err
	}

	return depl1, nil
}

func (s *Service) StopDeployment(ctx context.Context, depl *database.Deployment) error {
	// Stop the deployment by tearing down its runtime instance and resources.
	err := s.stopDeploymentInner(ctx, depl)
	if err != nil {
		// Mark deployment error
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		_, err2 := s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusError, fmt.Sprintf("Failed to stop deployment: %v", err))
		s.Logger.Error("stop deployment: failed to stop deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
		return errors.Join(err, err2)
	}

	// Update the deployment status to stopped
	_, err = s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusStopped, "")
	if err != nil {
		return err
	}

	return nil
}

// startDeploymentInner provisions a runtime and initializes an instance on it.
// The implementation is idempotent, enabling it to be moved to a retryable background job in the future.
func (s *Service) startDeploymentInner(ctx context.Context, depl *database.Deployment, opts *StartDeploymentOptions) error {
	// Validate the desired runtime version.
	// This is usually "latest", which the provisioner internally may resolve to an actual version.
	runtimeVersion := opts.Version
	err := validateRuntimeVersion(runtimeVersion)
	if err != nil {
		return err
	}

	// Provision the runtime
	r, err := s.provisionRuntime(ctx, &provisionRuntimeOptions{
		DeploymentID: depl.ID,
		Environment:  depl.Environment,
		Provisioner:  opts.Provisioner,
		Slots:        opts.Slots,
		Version:      runtimeVersion,
		Annotations:  opts.Annotations.ToMap(),
	})
	if err != nil {
		return err
	}
	cfg, err := provisioner.NewRuntimeConfig(r.Config)
	if err != nil {
		return err
	}

	// Update the deployment with the runtime details
	instanceID := strings.ReplaceAll(r.ID, "-", "") // Use the provisioned resource ID without dashes as the instance ID
	depl, err = s.DB.UpdateDeployment(ctx, depl.ID, &database.UpdateDeploymentOptions{
		Branch:            depl.Branch,
		RuntimeHost:       cfg.Host,
		RuntimeInstanceID: instanceID,
		RuntimeAudience:   cfg.Audience,
		Status:            database.DeploymentStatusPending,
		StatusMessage:     "Creating instance...",
	})
	if err != nil {
		return err
	}

	// Connect to the runtime
	rt, err := s.OpenRuntimeClient(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	// If the instance already exists, we can return now. (This can happen since this operation is idempotent and may be retried.)
	_, err = rt.GetInstance(ctx, &runtimev1.GetInstanceRequest{InstanceId: instanceID})
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}
	if err == nil {
		// Instance already exists. We can return.
		return nil
	}

	// Create an access token that it can use to authenticate with the admin server.
	dat, err := s.IssueDeploymentAuthToken(ctx, depl.ID, nil)
	if err != nil {
		return err
	}

	// Prepare connectors
	connectors := []*runtimev1.Connector{
		// The admin connector
		{
			Name: "admin",
			Type: "admin",
			Config: map[string]string{
				"admin_url":    s.opts.ExternalURL,
				"access_token": dat.Token().String(),
				"project_id":   depl.ProjectID,
			},
		},
		// Always configure a DuckDB connector, even if it's not set as the default OLAP connector
		{
			Name: "duckdb",
			Type: "duckdb",
			Config: map[string]string{
				"cpu":                 strconv.Itoa(cfg.CPU),
				"memory_limit_gb":     strconv.Itoa(cfg.MemoryGB),
				"storage_limit_bytes": strconv.FormatInt(cfg.StorageBytes, 10),
			},
		},
	}

	// Create the instance
	_, err = rt.CreateInstance(ctx, &runtimev1.CreateInstanceRequest{
		InstanceId:     instanceID,
		Environment:    depl.Environment,
		OlapConnector:  "duckdb", // Default OLAP connector for backwards compatibility with projects that don't specify olap_connector in rill.yaml
		RepoConnector:  "admin",
		AdminConnector: "admin",
		AiConnector:    "admin",
		Connectors:     connectors,
		Variables:      opts.Variables,
		Annotations:    opts.Annotations.ToMap(),
	})
	if err != nil {
		return err
	}

	// Deployment is ready to use
	return nil
}

// stopDeploymentInner stops a deployment by tearing down its runtime instance and resources.
// The implementation is idempotent, enabling it to be moved to a retryable background job in the future.
func (s *Service) stopDeploymentInner(ctx context.Context, depl *database.Deployment) error {
	// Connect to the deployment's runtime and delete the instance
	rt, err := s.OpenRuntimeClient(depl)
	if err != nil {
		s.Logger.Error("failed to open runtime client", zap.String("deployment_id", depl.ID), zap.String("runtime_instance_id", depl.RuntimeInstanceID), zap.Error(err), observability.ZapCtx(ctx))
	} else {
		defer rt.Close()
		_, err = rt.DeleteInstance(ctx, &runtimev1.DeleteInstanceRequest{
			InstanceId: depl.RuntimeInstanceID,
		})
		if err != nil {
			s.Logger.Error("failed to delete instance", zap.String("deployment_id", depl.ID), zap.String("runtime_instance_id", depl.RuntimeInstanceID), zap.Error(err), observability.ZapCtx(ctx))
		}
	}

	// Delete all provisioned resources for the deployment
	prs, err := s.DB.FindProvisionerResourcesForDeployment(ctx, depl.ID)
	if err != nil {
		s.Logger.Error("failed to find provisioner resources for deployment", zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
	} else {
		for _, pr := range prs {
			p, ok := s.ProvisionerSet[pr.Provisioner]
			if !ok {
				s.Logger.Warn("provisioner: deprovisioning skipped, provisioner not found", zap.String("deployment_id", depl.ID), zap.String("provisioner", pr.Provisioner), zap.String("provision_id", pr.ID), observability.ZapCtx(ctx))
			} else {
				err = p.Deprovision(ctx, &provisioner.Resource{
					ID:     pr.ID,
					Type:   provisioner.ResourceType(pr.Type),
					State:  pr.State,
					Config: pr.Config,
				})
				if err != nil {
					s.Logger.Error("provisioner: failed to deprovision", zap.String("deployment_id", depl.ID), zap.String("provisioner", pr.Provisioner), zap.String("provision_id", pr.ID), zap.Error(err), observability.ZapCtx(ctx))
				}
			}

			err = s.DB.DeleteProvisionerResource(ctx, pr.ID)
			if err != nil {
				s.Logger.Error("failed to delete provisioner resource", zap.String("deployment_id", depl.ID), zap.String("provisioner_resource_id", pr.ID), zap.Error(err), observability.ZapCtx(ctx))
			}
		}
	}

	return nil
}

// UpdateDeploymentsForProject updates the deployments of a project.
// In normal operation, projects only have one deployment. But during (re)deployment and in various error scenarios, there may be multiple deployments.
// Care must be taken to avoid one broken deployment from blocking updates to other healthy deployments.
func (s *Service) UpdateDeploymentsForProject(ctx context.Context, p *database.Project, opts *UpdateDeploymentOptions) error {
	ds, err := s.DB.FindDeploymentsForProject(ctx, p.ID)
	if err != nil {
		return err
	}

	grp, ctx := errgroup.WithContext(ctx)
	grp.SetLimit(5)
	var prodErr error
	for _, d := range ds {
		d := d
		grp.Go(func() error {
			err := s.UpdateDeployment(ctx, d, opts)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if p.ProdDeploymentID != nil && *p.ProdDeploymentID == d.ID {
					prodErr = err
				}
				s.Logger.Warn("failed to update deployment", zap.String("deployment_id", d.ID), zap.Error(err), observability.ZapCtx(ctx))
			}
			return nil
		})
	}

	err = grp.Wait()
	if err != nil {
		return err
	}

	return prodErr
}

type UpdateDeploymentOptions struct {
	Annotations     DeploymentAnnotations
	Branch          string
	Version         string
	EvictCachedRepo bool // Set to true to force the runtime to do a fresh repo clone instead of a pull.
}

func (s *Service) UpdateDeployment(ctx context.Context, d *database.Deployment, opts *UpdateDeploymentOptions) error {
	// Validate the desired runtime version.
	// This is usually "latest", which the provisioner internally may resolve to an actual version.
	runtimeVersion := opts.Version
	err := validateRuntimeVersion(runtimeVersion)
	if err != nil {
		return err
	}

	// Resolve the deployment's variables
	vars, err := s.ResolveVariables(ctx, d.ProjectID, d.Environment, true)
	if err != nil {
		return err
	}

	// Find the runtime provisioned for this deployment
	pr, ok, err := s.findProvisionedRuntimeResource(ctx, d.ID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("can't update deployment %q because its runtime has not been initialized yet", d.ID)
	}
	args, err := provisioner.NewRuntimeArgs(pr.Args)
	if err != nil {
		return err
	}

	// Provision the runtime. This is idempotent and will (partially) update the existing provisioned runtime if the config has changed.
	_, err = s.provisionRuntime(ctx, &provisionRuntimeOptions{
		DeploymentID: d.ID,
		Environment:  d.Environment,
		Provisioner:  pr.Provisioner,
		Slots:        args.Slots,
		Version:      runtimeVersion,
		Annotations:  opts.Annotations.ToMap(),
	})
	if err != nil {
		return err
	}

	// Connect to the runtime, and update the instance's variables/annotations.
	// Any call to EditInstance will also force it to check for any repo config changes (e.g. branch or archive ID).
	rt, err := s.OpenRuntimeClient(d)
	if err != nil {
		return err
	}
	defer rt.Close()
	_, err = rt.EditInstance(ctx, &runtimev1.EditInstanceRequest{
		InstanceId:  d.RuntimeInstanceID,
		Variables:   vars,
		Annotations: opts.Annotations.ToMap(),
	})
	if err != nil {
		return err
	}

	// Write the changed branch and status to the persisted deployment.
	_, err = s.DB.UpdateDeployment(ctx, d.ID, &database.UpdateDeploymentOptions{
		Branch:            opts.Branch,
		RuntimeHost:       d.RuntimeHost,
		RuntimeInstanceID: d.RuntimeInstanceID,
		RuntimeAudience:   d.RuntimeAudience,
		Status:            database.DeploymentStatusOK,
		StatusMessage:     "",
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) TeardownDeployment(ctx context.Context, depl *database.Deployment) error {
	// Stop runtime instance and resources.
	err := s.stopDeploymentInner(ctx, depl)
	if err != nil {
		return err
	}

	// Delete the deployment
	err = s.DB.DeleteDeployment(ctx, depl.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CheckProvisionerResource(ctx context.Context, pr *database.ProvisionerResource, annotations DeploymentAnnotations) error {
	// Find the provisioner
	p, ok := s.ProvisionerSet[pr.Provisioner]
	if !ok {
		return fmt.Errorf("provisioner: %q is not in the provisioner set", pr.Provisioner)
	}

	// Run a check
	r := &provisioner.Resource{
		ID:     pr.ID,
		Type:   provisioner.ResourceType(pr.Type),
		State:  pr.State,
		Config: pr.Config,
	}
	r, err := p.CheckResource(ctx, r, &provisioner.ResourceOptions{
		Args:        pr.Args,
		Annotations: annotations.ToMap(),
		RillVersion: s.resolveRillVersion(),
	})
	if err != nil {
		// For cancellations, we exit early without updating the status in the DB
		if errors.Is(err, ctx.Err()) {
			return err
		}

		// Set the status as errored
		_, err2 := s.DB.UpdateProvisionerResource(ctx, pr.ID, &database.UpdateProvisionerResourceOptions{
			Status:        database.ProvisionerResourceStatusError,
			StatusMessage: fmt.Sprintf("check failed: %s", err.Error()),
			Args:          pr.Args,
			State:         pr.State,
			Config:        pr.Config,
		})
		if err2 != nil {
			return errors.Join(err, err2)
		}

		return err
	}

	// The returned resource's state may have been updated, so we update the database accordingly.
	_, err = s.DB.UpdateProvisionerResource(ctx, pr.ID, &database.UpdateProvisionerResourceOptions{
		Status:        database.ProvisionerResourceStatusOK,
		StatusMessage: "",
		Args:          pr.Args,
		State:         r.State,
		Config:        r.Config,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) OpenRuntimeClient(depl *database.Deployment) (*client.Client, error) {
	if depl.RuntimeHost == "" {
		if depl.Status == database.DeploymentStatusError {
			return nil, fmt.Errorf("deployment %q has no runtime host: %s", depl.ID, depl.StatusMessage)
		}
		return nil, fmt.Errorf("deployment %q has no runtime host", depl.ID)
	}

	jwt, err := s.IssueRuntimeManagementToken(depl.RuntimeAudience)
	if err != nil {
		return nil, err
	}

	rt, err := client.New(depl.RuntimeHost, jwt)
	if err != nil {
		return nil, err
	}

	return rt, nil
}

func (s *Service) IssueRuntimeManagementToken(aud string) (string, error) {
	jwt, err := s.issuer.NewToken(auth.TokenOptions{
		AudienceURL:       aud,
		Subject:           "admin-service",
		TTL:               time.Hour,
		SystemPermissions: []auth.Permission{auth.ManageInstances, auth.ReadInstance, auth.EditInstance, auth.EditTrigger, auth.ReadObjects},
	})
	if err != nil {
		return "", err
	}

	return jwt, nil
}

func (s *Service) NewDeploymentAnnotations(org *database.Organization, proj *database.Project) DeploymentAnnotations {
	var orgBillingPlanName string
	if org.BillingPlanName != nil {
		orgBillingPlanName = *org.BillingPlanName
	}
	return DeploymentAnnotations{
		orgID:              org.ID,
		orgName:            org.Name,
		orgBillingPlanName: orgBillingPlanName,
		projID:             proj.ID,
		projName:           proj.Name,
		projProdSlots:      fmt.Sprint(proj.ProdSlots),
		projProvisioner:    proj.Provisioner,
		projAnnotations:    proj.Annotations,
	}
}

type DeploymentAnnotations struct {
	orgID              string
	orgName            string
	orgBillingPlanName string
	projID             string
	projName           string
	projProdSlots      string
	projProvisioner    string
	projAnnotations    map[string]string
}

func (da *DeploymentAnnotations) ToMap() map[string]string {
	res := make(map[string]string, len(da.projAnnotations)+7)
	for k, v := range da.projAnnotations {
		res[k] = v
	}
	res["organization_id"] = da.orgID
	res["organization_name"] = da.orgName
	res["organization_plan"] = da.orgBillingPlanName
	res["project_id"] = da.projID
	res["project_name"] = da.projName
	res["project_prod_slots"] = da.projProdSlots
	res["project_provisioner"] = da.projProvisioner
	return res
}

type provisionRuntimeOptions struct {
	DeploymentID string
	Environment  string
	Provisioner  string
	Slots        int
	Version      string
	Annotations  map[string]string
}

func (s *Service) provisionRuntime(ctx context.Context, opts *provisionRuntimeOptions) (*database.ProvisionerResource, error) {
	// Use default if no provisioner is specified.
	if opts.Provisioner == "" {
		opts.Provisioner = s.opts.DefaultProvisioner
	}

	// Create provisioner args
	args := &provisioner.RuntimeArgs{
		Slots:       opts.Slots,
		Version:     opts.Version,
		Environment: opts.Environment,
	}

	// Call into the generic provision function
	pr, err := s.Provision(ctx, &ProvisionOptions{
		DeploymentID: opts.DeploymentID,
		Type:         provisioner.ResourceTypeRuntime,
		Name:         "", // Not giving runtime resources a name since there should only be one per deployment.
		Provisioner:  opts.Provisioner,
		Args:         args.AsMap(),
		Annotations:  opts.Annotations,
	})
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *Service) findProvisionedRuntimeResource(ctx context.Context, deploymentID string) (*database.ProvisionerResource, bool, error) {
	pr, err := s.DB.FindProvisionerResourceByTypeAndName(ctx, deploymentID, string(provisioner.ResourceTypeRuntime), "")
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return pr, true, nil
}

func (s *Service) resolveRillVersion() string {
	if s.Version.Number != "" {
		return s.Version.Number
	}
	if s.Version.Commit != "" {
		return s.Version.Commit
	}
	return "latest"
}

func validateRuntimeVersion(ver string) error {
	// Verify version is a valid SemVer, a full Git commit hash or 'latest'
	if ver != "latest" {
		_, err := version.NewVersion(ver)
		if err != nil {
			// Not a valid SemVer, try as a full commit hash
			matched, err := regexp.MatchString(`\b([a-f0-9]{40})\b`, ver)
			if err != nil {
				return err
			}
			if !matched {
				return fmt.Errorf("not a valid version %q", ver)
			}
		}
	}

	return nil
}
