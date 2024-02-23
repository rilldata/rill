package admin

import (
	"context"
	"fmt"
	"path"
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
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type createDeploymentOptions struct {
	ProjectID      string
	Provisioner    string
	Annotations    DeploymentAnnotations
	VersionNumber  string
	ProdBranch     string
	ProdVariables  map[string]string
	ProdOLAPDriver string
	ProdOLAPDSN    string
	ProdSlots      int
	ProdVersion    string
}

func (s *Service) createDeployment(ctx context.Context, opts *createDeploymentOptions) (*database.Deployment, error) {
	// We require a branch to be specified to create a deployment
	if opts.ProdBranch == "" {
		return nil, fmt.Errorf("cannot create project without a branch")
	}

	// Use default if no provisioner is specified
	if opts.Provisioner == "" {
		opts.Provisioner = s.opts.DefaultProvisioner
	}

	// Get provisioner from the set
	p, ok := s.ProvisionerSet[opts.Provisioner]
	if !ok {
		return nil, fmt.Errorf("provisioner: %q is not in the provisioner set", opts.Provisioner)
	}

	// Resolve runtime version
	runtimeVersion := opts.ProdVersion
	if runtimeVersion == "latest" && opts.VersionNumber != "" {
		// Resolve latest version from config
		runtimeVersion = opts.VersionNumber
	}

	// Verify version is a valid SemVer or 'latest'
	if runtimeVersion != "latest" {
		_, err := version.NewVersion(runtimeVersion)
		if err != nil {
			return nil, err
		}
	}

	// Create instance ID and use the same ID for the provision ID
	instanceID := strings.ReplaceAll(uuid.New().String(), "-", "")
	provisionID := instanceID

	// Get a runtime with capacity for the deployment
	alloc, err := p.Provision(ctx, &provisioner.ProvisionOptions{
		ProvisionID:    provisionID,
		RuntimeVersion: runtimeVersion,
		Slots:          opts.ProdSlots,
		Annotations:    opts.Annotations.toMap(),
	})
	if err != nil {
		s.Logger.Error("provisioner: failed provisioning", zap.String("project_id", opts.ProjectID), zap.String("provisioner", opts.Provisioner), zap.String("provision_id", provisionID), zap.Error(err), observability.ZapCtx(ctx))
		return nil, err
	}

	// Prepare instance config
	var connectors []*runtimev1.Connector
	modelDefaultMaterialize, err := defaultModelMaterialize(opts.ProdVariables)
	if err != nil {
		return nil, err
	}

	// Always configure a DuckDB connector, even if it's not set as the default OLAP connector
	connectors = append(connectors, &runtimev1.Connector{
		Name: "duckdb",
		Type: "duckdb",
		Config: map[string]string{
			"dsn":                    fmt.Sprintf("%s.db", path.Join(alloc.DataDir, instanceID, "main")),
			"cpu":                    strconv.Itoa(alloc.CPU),
			"memory_limit_gb":        strconv.Itoa(alloc.MemoryGB),
			"storage_limit_bytes":    strconv.FormatInt(alloc.StorageBytes, 10),
			"external_table_storage": strconv.FormatBool(true),
		},
	})

	// Determine the default OLAP connector
	var olapConnector string
	switch opts.ProdOLAPDriver {
	case "duckdb", "duckdb-ext-storage":
		if opts.ProdSlots == 0 {
			return nil, fmt.Errorf("slot count can't be 0 for OLAP driver 'duckdb'")
		}
		olapConnector = "duckdb"
		// Already configured DuckDB above
	default:
		olapConnector = opts.ProdOLAPDriver
		connectors = append(connectors, &runtimev1.Connector{
			Name: opts.ProdOLAPDriver,
			Type: opts.ProdOLAPDriver,
			Config: map[string]string{
				"dsn": opts.ProdOLAPDSN,
			},
		})
	}

	// Create the deployment
	depl, err := s.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         opts.ProjectID,
		Provisioner:       opts.Provisioner,
		ProvisionID:       provisionID,
		Branch:            opts.ProdBranch,
		Slots:             opts.ProdSlots,
		RuntimeHost:       alloc.Host,
		RuntimeInstanceID: instanceID,
		RuntimeAudience:   alloc.Audience,
		RuntimeVersion:    runtimeVersion,
		Status:            database.DeploymentStatusPending,
	})
	if err != nil {
		return nil, err
	}

	// Wait for the runtime to be ready
	err = p.AwaitReady(ctx, provisionID)
	if err != nil {
		s.Logger.Error("provisioner: failed awaiting runtime to be ready", zap.String("project_id", opts.ProjectID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.Error(err), observability.ZapCtx(ctx))
		// Mark deployment error
		_, err2 := s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusError, err.Error())
		return nil, multierr.Combine(err, err2)
	}

	// Open a runtime client
	rt, err := s.openRuntimeClient(alloc.Host, alloc.Audience)
	if err != nil {
		err2 := p.Deprovision(ctx, provisionID)
		err3 := s.DB.DeleteDeployment(ctx, depl.ID)
		return nil, multierr.Combine(err, err2, err3)
	}
	defer rt.Close()

	// Create an access token the deployment can use to authenticate with the admin server.
	dat, err := s.IssueDeploymentAuthToken(ctx, depl.ID, nil)
	if err != nil {
		err2 := p.Deprovision(ctx, provisionID)
		err3 := s.DB.DeleteDeployment(ctx, depl.ID)
		return nil, multierr.Combine(err, err2, err3)
	}
	adminAuthToken := dat.Token().String()

	// Add the admin connector
	connectors = append(connectors, &runtimev1.Connector{
		Name: "admin",
		Type: "admin",
		Config: map[string]string{
			"admin_url":    s.opts.ExternalURL,
			"access_token": adminAuthToken,
			"project_id":   opts.ProjectID,
			"branch":       opts.ProdBranch,
			"nonce":        time.Now().Format(time.RFC3339Nano), // Only set for consistency with updateDeployment
		},
	})

	// Create the instance
	_, err = rt.CreateInstance(ctx, &runtimev1.CreateInstanceRequest{
		InstanceId:              instanceID,
		Environment:             "prod",
		OlapConnector:           olapConnector,
		RepoConnector:           "admin",
		AdminConnector:          "admin",
		AiConnector:             "admin",
		Connectors:              connectors,
		Variables:               opts.ProdVariables,
		Annotations:             opts.Annotations.toMap(),
		EmbedCatalog:            false,
		StageChanges:            true,
		ModelDefaultMaterialize: modelDefaultMaterialize,
	})
	if err != nil {
		err2 := p.Deprovision(ctx, provisionID)
		err3 := s.DB.DeleteDeployment(ctx, depl.ID)
		return nil, multierr.Combine(err, err2, err3)
	}

	// Mark deployment ready
	depl, err = s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusOK, "")
	if err != nil {
		// NOTE: Unlikely case – we'll leave it pending in this case, the user can reset.
		return nil, err
	}

	return depl, nil
}

type UpdateDeploymentOptions struct {
	Version         string
	Branch          string
	Variables       map[string]string
	Annotations     DeploymentAnnotations
	EvictCachedRepo bool // Set to true if config returned by GetRepoMeta has changed such that the runtime should do a fresh clone instead of a pull.
}

func (s *Service) UpdateDeployment(ctx context.Context, depl *database.Deployment, opts *UpdateDeploymentOptions) error {
	if opts.Branch == "" {
		return fmt.Errorf("cannot update deployment without specifying a valid branch")
	}

	var modelDefaultMaterialize *bool
	if opts.Variables != nil { // if variables are nil, it means they were not changed
		val, err := defaultModelMaterialize(opts.Variables)
		if err != nil {
			return err
		}
		modelDefaultMaterialize = &val
	}

	// Update the provisioned runtime if the version has changed
	if opts.Version != depl.RuntimeVersion {
		// Get provisioner from the set
		p, ok := s.ProvisionerSet[depl.Provisioner]
		if !ok {
			return fmt.Errorf("provisioner: %q is not in the provisioner set", depl.Provisioner)
		}

		// Update the runtime
		err := p.Update(ctx, depl.ProvisionID, opts.Version)
		if err != nil {
			s.Logger.Error("provisioner: failed to update", zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.Error(err), observability.ZapCtx(ctx))
			return err
		}

		// Wait for the runtime to be ready after update
		err = p.AwaitReady(ctx, depl.ProvisionID)
		if err != nil {
			s.Logger.Error("provisioner: failed awaiting runtime to be ready after update", zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.Error(err), observability.ZapCtx(ctx))
			// Mark deployment error
			_, err2 := s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusError, err.Error())
			return multierr.Combine(err, err2)
		}

		// Update the deployment runtime version
		_, err = s.DB.UpdateDeploymentRuntimeVersion(ctx, depl.ID, opts.Version)
		if err != nil {
			// NOTE: If the update was triggered by a scheduled job like 'upgrade_latest_version_projects',
			// then this error will cause the update to be retried on the next job invocation and it should eventually become consistent.

			// TODO: Handle inconsistent state when a manually triggered update failed, where we can't rely on job retries.
			return err
		}
	}

	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	res, err := rt.GetInstance(ctx, &runtimev1.GetInstanceRequest{InstanceId: depl.RuntimeInstanceID})
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
		InstanceId:              depl.RuntimeInstanceID,
		Connectors:              connectors,
		Annotations:             opts.Annotations.toMap(),
		Variables:               opts.Variables,
		ModelDefaultMaterialize: modelDefaultMaterialize,
	})
	if err != nil {
		return err
	}

	// Branch is the only property that's persisted on the Deployment
	if opts.Branch != depl.Branch {
		newDepl, err := s.DB.UpdateDeploymentBranch(ctx, depl.ID, opts.Branch)
		if err != nil {
			// TODO: Handle inconsistent state (instance updated successfully, but deployment did not update)
			return err
		}
		depl.Branch = opts.Branch
		depl.UpdatedOn = newDepl.UpdatedOn
	}

	return nil
}

// HibernateDeployments tears down unused deployments
func (s *Service) HibernateDeployments(ctx context.Context) error {
	depls, err := s.DB.FindExpiredDeployments(ctx)
	if err != nil {
		return err
	}

	if len(depls) == 0 {
		return nil
	}

	s.Logger.Info("hibernate: starting", zap.Int("deployments", len(depls)))

	for _, depl := range depls {
		proj, err := s.DB.FindProject(ctx, depl.ProjectID)
		if err != nil {
			s.Logger.Error("hibernate: find project error", zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
			continue
		}

		s.Logger.Info("hibernate: deleting deployment", zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID))

		err = s.teardownDeployment(ctx, proj, depl)
		if err != nil {
			s.Logger.Error("hibernate: teardown deployment error", zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
			continue
		}

		// Update prod deployment on project
		_, err = s.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
			Name:                 proj.Name,
			Description:          proj.Description,
			Public:               proj.Public,
			Provisioner:          proj.Provisioner,
			GithubURL:            proj.GithubURL,
			GithubInstallationID: proj.GithubInstallationID,
			ProdVersion:          proj.ProdVersion,
			ProdBranch:           proj.ProdBranch,
			ProdVariables:        proj.ProdVariables,
			ProdSlots:            proj.ProdSlots,
			ProdTTLSeconds:       proj.ProdTTLSeconds,
			ProdDeploymentID:     nil,
			Annotations:          proj.Annotations,
		})
		if err != nil {
			return err
		}
	}

	s.Logger.Info("hibernate: completed", zap.Int("deployments", len(depls)))

	return nil
}

func (s *Service) teardownDeployment(ctx context.Context, proj *database.Project, depl *database.Deployment) error {
	// Connect to the deployment's runtime
	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	// Delete the instance
	_, err = rt.DeleteInstance(ctx, &runtimev1.DeleteInstanceRequest{
		InstanceId: depl.RuntimeInstanceID,
	})
	if err != nil {
		return err
	}

	// Get provisioner and deprovision, skip if the provisioner is no longer defined in the provisioner set
	p, ok := s.ProvisionerSet[depl.Provisioner]
	if ok {
		err = p.Deprovision(ctx, depl.ProvisionID)
		if err != nil {
			return err
		}
	} else {
		s.Logger.Warn("provisioner: deprovisioning skipped, provisioner not found", zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.Error(err), observability.ZapCtx(ctx))
	}

	// Delete the deployment
	err = s.DB.DeleteDeployment(ctx, depl.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) openRuntimeClientForDeployment(d *database.Deployment) (*client.Client, error) {
	return s.openRuntimeClient(d.RuntimeHost, d.RuntimeAudience)
}

func (s *Service) openRuntimeClient(host, audience string) (*client.Client, error) {
	jwt, err := s.issuer.NewToken(auth.TokenOptions{
		AudienceURL:       audience,
		TTL:               time.Hour,
		SystemPermissions: []auth.Permission{auth.ManageInstances, auth.ReadInstance, auth.EditInstance, auth.ReadObjects},
	})
	if err != nil {
		return nil, err
	}

	rt, err := client.New(host, jwt)
	if err != nil {
		return nil, err
	}

	return rt, nil
}

type DeploymentAnnotations struct {
	orgID           string
	orgName         string
	projID          string
	projName        string
	projAnnotations map[string]string
}

func (s *Service) NewDeploymentAnnotations(org *database.Organization, proj *database.Project) DeploymentAnnotations {
	return DeploymentAnnotations{
		orgID:           org.ID,
		orgName:         org.Name,
		projID:          proj.ID,
		projName:        proj.Name,
		projAnnotations: proj.Annotations,
	}
}

func (da *DeploymentAnnotations) toMap() map[string]string {
	res := make(map[string]string, len(da.projAnnotations)+4)
	for k, v := range da.projAnnotations {
		res[k] = v
	}
	res["organization_id"] = da.orgID
	res["organization_name"] = da.orgName
	res["project_id"] = da.projID
	res["project_name"] = da.projName
	return res
}

// defaultModelMaterialize determines whether to materialize models by default for deployed projects.
// It defaults to true, but can be overridden with the __materialize_default variable.
func defaultModelMaterialize(vars map[string]string) (bool, error) {
	// Temporary hack to enable configuring ModelDefaultMaterialize using a variable.
	// Remove when we have a way to conditionally configure it using code files.

	if vars == nil {
		return true, nil
	}

	s, ok := vars["__materialize_default"]
	if !ok {
		return true, nil
	}

	val, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("invalid __materialize_default value %q: %w", s, err)
	}

	return val, nil
}
