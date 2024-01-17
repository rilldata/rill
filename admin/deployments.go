package admin

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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
	Region         string
	ProdBranch     string
	ProdVariables  map[string]string
	ProdOLAPDriver string
	ProdOLAPDSN    string
	ProdSlots      int
	Annotations    deploymentAnnotations
}

func (s *Service) createDeployment(ctx context.Context, opts *createDeploymentOptions) (*database.Deployment, error) {
	// We require a branch to be specified to create a deployment
	if opts.ProdBranch == "" {
		return nil, fmt.Errorf("cannot create project without a branch")
	}

	// Get a runtime with capacity for the deployment
	alloc, err := s.Provisioner.Provision(ctx, &provisioner.ProvisionOptions{
		OLAPDriver: opts.ProdOLAPDriver,
		Slots:      opts.ProdSlots,
		Region:     opts.Region,
	})
	if err != nil {
		return nil, err
	}

	// Build instance config
	instanceID := strings.ReplaceAll(uuid.New().String(), "-", "")
	olapDriver := opts.ProdOLAPDriver
	olapConfig := map[string]string{}
	var embedCatalog bool
	switch olapDriver {
	case "duckdb":
		if opts.ProdOLAPDSN != "" {
			return nil, fmt.Errorf("passing a DSN is not allowed for driver 'duckdb'")
		}
		if opts.ProdSlots == 0 {
			return nil, fmt.Errorf("slot count can't be 0 for driver 'duckdb'")
		}

		olapConfig["dsn"] = fmt.Sprintf("%s.db", path.Join(alloc.DataDir, instanceID))
		olapConfig["cpu"] = strconv.Itoa(alloc.CPU)
		olapConfig["memory_limit_gb"] = strconv.Itoa(alloc.MemoryGB)
		olapConfig["storage_limit_bytes"] = strconv.FormatInt(alloc.StorageBytes, 10)
		embedCatalog = false
	case "duckdb-ext-storage": // duckdb driver having capability to store table as view
		if opts.ProdOLAPDSN != "" {
			return nil, fmt.Errorf("passing a DSN is not allowed for driver 'duckdb-ext-storage'")
		}
		if opts.ProdSlots == 0 {
			return nil, fmt.Errorf("slot count can't be 0 for driver 'duckdb-ext-storage'")
		}

		olapDriver = "duckdb"
		olapConfig["dsn"] = fmt.Sprintf("%s.db", path.Join(alloc.DataDir, instanceID, "main"))
		olapConfig["cpu"] = strconv.Itoa(alloc.CPU)
		olapConfig["memory_limit_gb"] = strconv.Itoa(alloc.MemoryGB)
		olapConfig["storage_limit_bytes"] = strconv.FormatInt(alloc.StorageBytes, 10)
		olapConfig["external_table_storage"] = strconv.FormatBool(true)
		embedCatalog = false
	default:
		olapConfig["dsn"] = opts.ProdOLAPDSN
		embedCatalog = false
		olapConfig["storage_limit_bytes"] = "0"
	}

	modelDefaultMaterialize, err := defaultModelMaterialize(opts.ProdVariables)
	if err != nil {
		return nil, err
	}

	// Open a runtime client
	rt, err := s.openRuntimeClient(alloc.Host, alloc.Audience)
	if err != nil {
		return nil, err
	}
	defer rt.Close()

	// Create the deployment
	depl, err := s.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         opts.ProjectID,
		Branch:            opts.ProdBranch,
		Slots:             opts.ProdSlots,
		RuntimeHost:       alloc.Host,
		RuntimeInstanceID: instanceID,
		RuntimeAudience:   alloc.Audience,
		Status:            database.DeploymentStatusPending,
	})
	if err != nil {
		return nil, err
	}

	// Create an access token the deployment can use to authenticate with the admin server.
	dat, err := s.IssueDeploymentAuthToken(ctx, depl.ID, nil)
	if err != nil {
		err2 := s.DB.DeleteDeployment(ctx, depl.ID)
		return nil, multierr.Combine(err, err2)
	}
	adminAuthToken := dat.Token().String()

	// Create the instance
	_, err = rt.CreateInstance(ctx, &runtimev1.CreateInstanceRequest{
		InstanceId:     instanceID,
		OlapConnector:  olapDriver,
		RepoConnector:  "admin",
		AdminConnector: "admin",
		Connectors: []*runtimev1.Connector{
			{
				Name:   olapDriver,
				Type:   olapDriver,
				Config: olapConfig,
			},
			{
				Name: "admin",
				Type: "admin",
				Config: map[string]string{
					"admin_url":    s.opts.ExternalURL,
					"access_token": adminAuthToken,
					"project_id":   opts.ProjectID,
					"branch":       opts.ProdBranch,
					"nonce":        time.Now().Format(time.RFC3339Nano), // Only set for consistency with updateDeployment
				},
			},
		},
		Variables:               opts.ProdVariables,
		Annotations:             opts.Annotations.toMap(),
		EmbedCatalog:            embedCatalog,
		StageChanges:            true,
		ModelDefaultMaterialize: modelDefaultMaterialize,
	})
	if err != nil {
		err2 := s.DB.DeleteDeployment(ctx, depl.ID)
		return nil, multierr.Combine(err, err2)
	}

	// Mark deployment ready
	depl, err = s.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusOK, "")
	if err != nil {
		// NOTE: Unlikely case – we'll leave it pending in this case, the user can reset.
		return nil, err
	}

	return depl, nil
}

type updateDeploymentOptions struct {
	Branch          string
	Variables       map[string]string
	Annotations     deploymentAnnotations
	EvictCachedRepo bool // Set to true if config returned by GetRepoMeta has changed such that the runtime should do a fresh clone instead of a pull.
}

func (s *Service) updateDeployment(ctx context.Context, depl *database.Deployment, opts *updateDeploymentOptions) error {
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
			GithubURL:            proj.GithubURL,
			GithubInstallationID: proj.GithubInstallationID,
			ProdBranch:           proj.ProdBranch,
			ProdVariables:        proj.ProdVariables,
			ProdSlots:            proj.ProdSlots,
			Region:               proj.Region,
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

	// Delete the deployment
	err = s.DB.DeleteDeployment(ctx, depl.ID)
	if err != nil {
		return err
	}

	// Delete the instance
	_, err = rt.DeleteInstance(ctx, &runtimev1.DeleteInstanceRequest{
		InstanceId: depl.RuntimeInstanceID,
		DropDb:     strings.Contains(proj.ProdOLAPDriver, "duckdb"), // Only drop DB if it's DuckDB
	})
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

type deploymentAnnotations struct {
	orgID           string
	orgName         string
	projID          string
	projName        string
	projAnnotations map[string]string
}

func newDeploymentAnnotations(org *database.Organization, proj *database.Project) deploymentAnnotations {
	return deploymentAnnotations{
		orgID:           org.ID,
		orgName:         org.Name,
		projID:          proj.ID,
		projName:        proj.Name,
		projAnnotations: proj.Annotations,
	}
}

func (da *deploymentAnnotations) toMap() map[string]string {
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

func defaultModelMaterialize(vars map[string]string) (bool, error) {
	// Temporary hack to enable configuring ModelDefaultMaterialize using a variable.
	// Remove when we have a way to conditionally configure it using code files.

	if vars == nil {
		return false, nil
	}

	s, ok := vars["__materialize_default"]
	if !ok {
		return false, nil
	}

	val, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("invalid __materialize_default value %q: %w", s, err)
	}

	return val, nil
}
