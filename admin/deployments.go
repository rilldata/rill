package admin

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/client"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateDeploymentOptions struct {
	ProjectID   string
	Annotations DeploymentAnnotations
	Branch      string
	Provisioner string
	Slots       int
	Version     string
	Variables   map[string]string
	OLAPDriver  string
	OLAPDSN     string
}

func (s *Service) CreateDeployment(ctx context.Context, opts *CreateDeploymentOptions) (*database.Deployment, error) {
	// Create the deployment
	depl, err := s.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         opts.ProjectID,
		Branch:            opts.Branch,
		RuntimeHost:       "", // Will be populated after provisioning in createDeploymentInner
		RuntimeInstanceID: "", // Will be populated after provisioning in createDeploymentInner
		RuntimeAudience:   "", // Will be populated after provisioning in createDeploymentInner
		Status:            database.DeploymentStatusPending,
		StatusMessage:     "Provisioning...",
	})
	if err != nil {
		return nil, err
	}

	// Initialize the deployment (by provisioning a runtime and creating an instance on it)
	depl, err = s.createDeploymentInner(ctx, depl, opts)
	if err != nil {
		// Mark deployment error
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		_, err2 := s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusError, fmt.Sprintf("Failed provisioning runtime: %v", err))
		return nil, errors.Join(err, err2)
	}

	// Mark deployment ready
	depl, err = s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusOK, "")
	if err != nil {
		// NOTE: Unlikely case – we'll leave it pending in this case, the user can reset.
		return nil, err
	}

	return depl, nil
}

// createDeploymentInner idempotently provisions a runtime and initializes an instance on it.
// The implementation is idempotent, enabling it to be moved to a retryable background job in the future.
func (s *Service) createDeploymentInner(ctx context.Context, d *database.Deployment, opts *CreateDeploymentOptions) (*database.Deployment, error) {
	// Validate the desired runtime version.
	// This is usually "latest", which the provisioner internally may resolve to an actual version.
	runtimeVersion := opts.Version
	err := validateRuntimeVersion(runtimeVersion)
	if err != nil {
		return nil, err
	}

	// Provision the runtime
	r, err := s.provisionRuntime(ctx, &provisionRuntimeOptions{
		DeploymentID: d.ID,
		Provisioner:  opts.Provisioner,
		Slots:        opts.Slots,
		Version:      runtimeVersion,
		Annotations:  opts.Annotations.ToMap(),
	})
	if err != nil {
		return nil, err
	}
	cfg, err := provisioner.NewRuntimeConfig(r.Config)
	if err != nil {
		return nil, err
	}

	// Update the deployment with the runtime details
	instanceID := strings.ReplaceAll(r.ID, "-", "") // Use the provisioned resource ID without dashes as the instance ID
	d, err = s.DB.UpdateDeployment(ctx, d.ID, &database.UpdateDeploymentOptions{
		Branch:            d.Branch,
		RuntimeHost:       cfg.Host,
		RuntimeInstanceID: instanceID,
		RuntimeAudience:   cfg.Audience,
		Status:            database.DeploymentStatusPending,
		StatusMessage:     "Creating instance...",
	})
	if err != nil {
		return nil, err
	}

	// Connect to the runtime
	rt, err := s.OpenRuntimeClient(d)
	if err != nil {
		return nil, err
	}
	defer rt.Close()

	// If the instance already exists, we can return now. (This can happen since this operation is idempotent and may be retried.)
	_, err = rt.GetInstance(ctx, &runtimev1.GetInstanceRequest{InstanceId: instanceID})
	if err != nil && status.Code(err) != codes.NotFound {
		return nil, err
	}
	if err == nil {
		// Instance already exists. We can return.
		return d, nil
	}

	// Prepare instance config
	var connectors []*runtimev1.Connector

	// Add the admin connector. It gets an access token that it can use to authenticate with the admin server.
	dat, err := s.IssueDeploymentAuthToken(ctx, d.ID, nil)
	if err != nil {
		return nil, err
	}
	connectors = append(connectors, &runtimev1.Connector{
		Name: "admin",
		Type: "admin",
		Config: map[string]string{
			"admin_url":    s.opts.ExternalURL,
			"access_token": dat.Token().String(),
			"project_id":   opts.ProjectID,
			"branch":       opts.Branch,
			"nonce":        time.Now().Format(time.RFC3339Nano), // Only set for consistency with updateDeployment
		},
	})

	// Always configure a DuckDB connector, even if it's not set as the default OLAP connector
	connectors = append(connectors, &runtimev1.Connector{
		Name: "duckdb",
		Type: "duckdb",
		Config: map[string]string{
			"cpu":                 strconv.Itoa(cfg.CPU),
			"memory_limit_gb":     strconv.Itoa(cfg.MemoryGB),
			"storage_limit_bytes": strconv.FormatInt(cfg.StorageBytes, 10),
		},
	})

	// Determine the default OLAP connector.
	// TODO: Remove this because it is deprecated and can now be configured directly using `rill.yaml` and `rill env`.
	var olapConnector string
	switch opts.OLAPDriver {
	case "duckdb", "duckdb-ext-storage":
		if opts.Slots == 0 {
			return nil, fmt.Errorf("slot count can't be 0 for OLAP driver 'duckdb'")
		}
		olapConnector = "duckdb"
		// Already configured DuckDB above
	default:
		olapConnector = opts.OLAPDriver
		connectors = append(connectors, &runtimev1.Connector{
			Name: opts.OLAPDriver,
			Type: opts.OLAPDriver,
			Config: map[string]string{
				"dsn": opts.OLAPDSN,
			},
		})
	}

	// Create the instance
	_, err = rt.CreateInstance(ctx, &runtimev1.CreateInstanceRequest{
		InstanceId:     instanceID,
		Environment:    "prod",
		OlapConnector:  olapConnector,
		RepoConnector:  "admin",
		AdminConnector: "admin",
		AiConnector:    "admin",
		Connectors:     connectors,
		Variables:      opts.Variables,
		Annotations:    opts.Annotations.ToMap(),
		EmbedCatalog:   false,
	})
	if err != nil {
		return nil, err
	}

	// Deployment is ready to use
	return d, nil
}

type UpdateDeploymentOptions struct {
	Annotations     DeploymentAnnotations
	Branch          string
	Version         string
	Variables       map[string]string // If empty, the existing variables are left unchanged.
	EvictCachedRepo bool              // Set to true to force the runtime to do a fresh repo clone instead of a pull.
}

func (s *Service) UpdateDeployment(ctx context.Context, d *database.Deployment, opts *UpdateDeploymentOptions) error {
	// Validate the desired runtime version.
	// This is usually "latest", which the provisioner internally may resolve to an actual version.
	runtimeVersion := opts.Version
	err := validateRuntimeVersion(runtimeVersion)
	if err != nil {
		return err
	}

	// Find the runtime provisioned for this deployment
	pr, err := s.findProvisionedRuntimeResource(ctx, d.ID)
	if err != nil {
		return err
	}
	if pr == nil {
		return fmt.Errorf("can't update deployment %q because its runtime has not been initialized yet", d.ID)
	}
	args, err := provisioner.NewRuntimeArgs(pr.Args)
	if err != nil {
		return err
	}

	// Provision the runtime. This is idempotent and will (partially) update the existing provisioned runtime if the config has changed.
	_, err = s.provisionRuntime(ctx, &provisionRuntimeOptions{
		DeploymentID: d.ID,
		Provisioner:  pr.Provisioner,
		Slots:        args.Slots,
		Version:      runtimeVersion,
		Annotations:  opts.Annotations.ToMap(),
	})
	if err != nil {
		return err
	}

	// Connect to the runtime, and update the instance's connectors/variables/annotations.
	rt, err := s.OpenRuntimeClient(d)
	if err != nil {
		return err
	}
	defer rt.Close()
	res, err := rt.GetInstance(ctx, &runtimev1.GetInstanceRequest{
		InstanceId: d.RuntimeInstanceID,
		Sensitive:  true,
	})
	if err != nil {
		return err
	}
	connectors := res.Instance.Connectors
	for _, c := range connectors {
		if c.Name == "admin" {
			if c.Config == nil {
				c.Config = make(map[string]string)
			}
			c.Config["branch"] = opts.Branch

			// Adding a nonce will cause the runtime to evict any currently open handle and open a new one.
			if opts.EvictCachedRepo {
				c.Config["nonce"] = time.Now().Format(time.RFC3339Nano)
			}
		}
	}
	_, err = rt.EditInstance(ctx, &runtimev1.EditInstanceRequest{
		InstanceId:  d.RuntimeInstanceID,
		Connectors:  connectors,
		Variables:   opts.Variables,
		Annotations: opts.Annotations.ToMap(),
	})
	if err != nil {
		return err
	}

	// Write the changed branch and status to the persisted deployment.
	d, err = s.DB.UpdateDeployment(ctx, d.ID, &database.UpdateDeploymentOptions{
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
					Config: pr.Config,
					State:  pr.State,
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
		return err
	}

	// The returned resource's state may have been updated, so we update the database accordingly.
	pr, err = s.DB.UpdateProvisionerResource(ctx, pr.ID, &database.UpdateProvisionerResourceOptions{
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
	return DeploymentAnnotations{
		orgID:           org.ID,
		orgName:         org.Name,
		projID:          proj.ID,
		projName:        proj.Name,
		projProdSlots:   fmt.Sprint(proj.ProdSlots),
		projProvisioner: proj.Provisioner,
		projAnnotations: proj.Annotations,
	}
}

type DeploymentAnnotations struct {
	orgID           string
	orgName         string
	projID          string
	projName        string
	projProdSlots   string
	projProvisioner string
	projAnnotations map[string]string
}

func (da *DeploymentAnnotations) ToMap() map[string]string {
	res := make(map[string]string, len(da.projAnnotations)+4)
	for k, v := range da.projAnnotations {
		res[k] = v
	}
	res["organization_id"] = da.orgID
	res["organization_name"] = da.orgName
	res["project_id"] = da.projID
	res["project_name"] = da.projName
	res["project_prod_slots"] = da.projProdSlots
	res["project_provisioner"] = da.projProvisioner
	return res
}

type provisionRuntimeOptions struct {
	DeploymentID string
	Provisioner  string
	Slots        int
	Version      string
	Annotations  map[string]string
}

func (s *Service) provisionRuntime(ctx context.Context, opts *provisionRuntimeOptions) (*database.ProvisionerResource, error) {
	// Get provisioner from the set.
	// Use default if no provisioner is specified.
	if opts.Provisioner == "" {
		opts.Provisioner = s.opts.DefaultProvisioner
	}
	p, ok := s.ProvisionerSet[opts.Provisioner]
	if !ok {
		return nil, fmt.Errorf("provisioner: %q is not in the provisioner set", opts.Provisioner)
	}

	// Create provisioner args
	args := &provisioner.RuntimeArgs{
		Slots:   opts.Slots,
		Version: opts.Version,
	}

	// Attempt to find an existing provisioned runtime for the deployment
	pr, err := s.findProvisionedRuntimeResource(ctx, opts.DeploymentID)
	if err != nil {
		return nil, err
	}
	if pr != nil && pr.Provisioner != opts.Provisioner {
		return nil, fmt.Errorf("provisioner: cannot change provisioner from %q to %q for deployment %q", pr.Provisioner, opts.Provisioner, opts.DeploymentID)
	}

	// If we didn't find an existing DB entry, create one
	if pr == nil {
		pr, err = s.DB.InsertProvisionerResource(ctx, &database.InsertProvisionerResourceOptions{
			ID:            uuid.New().String(),
			DeploymentID:  opts.DeploymentID,
			Type:          string(provisioner.ResourceTypeRuntime),
			Name:          "", // Not giving runtime resources a name since there should only be one per deployment.
			Status:        database.ProvisionerResourceStatusPending,
			StatusMessage: "Provisioning...",
			Provisioner:   opts.Provisioner,
			Args:          args.AsMap(),
		})
		if err != nil {
			return nil, err
		}
	}

	// Provision the runtime
	r := &provisioner.Resource{
		ID:     pr.ID,
		Type:   provisioner.ResourceTypeRuntime,
		Config: pr.Config, // Empty if inserting
		State:  pr.State,  // Empty if inserting
	}
	r, err = p.Provision(ctx, r, &provisioner.ResourceOptions{
		Args:        args.AsMap(),
		Annotations: opts.Annotations,
		RillVersion: s.resolveRillVersion(),
	})
	if err != nil {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		_, _ = s.DB.UpdateProvisionerResource(ctx, pr.ID, &database.UpdateProvisionerResourceOptions{
			Status:        database.ProvisionerResourceStatusError,
			StatusMessage: fmt.Sprintf("Failed provisioning runtime: %v", err),
			Args:          pr.Args,
			State:         pr.State,
			Config:        pr.Config,
		})
		return nil, err
	}

	// Update the provisioner resource
	pr, err = s.DB.UpdateProvisionerResource(ctx, pr.ID, &database.UpdateProvisionerResourceOptions{
		Status:        database.ProvisionerResourceStatusOK,
		StatusMessage: "",
		Args:          args.AsMap(),
		State:         r.State,
		Config:        r.Config,
	})
	if err != nil {
		return nil, err
	}

	// Await the runtime to be ready
	err = p.AwaitReady(ctx, r)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *Service) findProvisionedRuntimeResource(ctx context.Context, deploymentID string) (*database.ProvisionerResource, error) {
	prs, err := s.DB.FindProvisionerResourcesForDeployment(ctx, deploymentID)
	if err != nil {
		return nil, err
	}
	for _, val := range prs {
		if provisioner.ResourceType(val.Type) == provisioner.ResourceTypeRuntime {
			return val, nil
		}
	}
	return nil, nil
}

func (s *Service) resolveRillVersion() string {
	if s.VersionNumber != "" {
		return s.VersionNumber
	}
	if s.VersionCommit != "" {
		return s.VersionCommit
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
