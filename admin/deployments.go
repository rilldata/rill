package admin

import (
	"context"
	"encoding/json"
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
	"github.com/rilldata/rill/runtime/drivers/github"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type createDeploymentOptions struct {
	ProjectID            string
	Region               string
	GithubURL            *string
	GithubInstallationID *int64
	Subpath              string
	ProdBranch           string
	ProdVariables        database.Variables
	ProdOLAPDriver       string
	ProdOLAPDSN          string
	ProdSlots            int
	Annotations          deploymentAnnotations
}

func (s *Service) createDeployment(ctx context.Context, opts *createDeploymentOptions) (*database.Deployment, error) {
	// We require Github info on project to create a deployment
	if opts.GithubURL == nil || opts.GithubInstallationID == nil || opts.ProdBranch == "" {
		return nil, fmt.Errorf("cannot create project without github info")
	}
	repoDriver, repoDSN, err := githubRepoInfoForRuntime(*opts.GithubURL, *opts.GithubInstallationID, opts.Subpath, opts.ProdBranch)
	if err != nil {
		return nil, err
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
	var ingestionLimit int64
	switch olapDriver {
	case "duckdb":
		if opts.ProdOLAPDSN != "" {
			return nil, fmt.Errorf("passing a DSN is not allowed for driver 'duckdb'")
		}
		if opts.ProdSlots == 0 {
			return nil, fmt.Errorf("slot count can't be 0 for driver 'duckdb'")
		}

		olapConfig["dsn"] = fmt.Sprintf("%s.db?max_memory=%dGB", path.Join(alloc.DataDir, instanceID), alloc.MemoryGB)
		olapConfig["pool_size"] = strconv.Itoa(alloc.CPU)
		embedCatalog = true
		ingestionLimit = alloc.StorageBytes
	case "duckdb-vip":
		if opts.ProdOLAPDSN != "" {
			return nil, fmt.Errorf("passing a DSN is not allowed for driver 'duckdb-vip'")
		}
		if opts.ProdSlots == 0 {
			return nil, fmt.Errorf("slot count can't be 0 for driver 'duckdb-vip'")
		}

		// NOTE: Rewriting to a "duckdb" driver without CPU, memory, or storage limits
		olapDriver = "duckdb"
		olapConfig["dsn"] = fmt.Sprintf("%s.db", path.Join(alloc.DataDir, instanceID))
		olapConfig["pool_size"] = "8"
		embedCatalog = true
		ingestionLimit = 0
	default:
		olapConfig["dsn"] = opts.ProdOLAPDSN
		embedCatalog = false
		ingestionLimit = 0
	}

	// Open a runtime client
	rt, err := s.openRuntimeClient(alloc.Host, alloc.Audience)
	if err != nil {
		return nil, err
	}
	defer rt.Close()

	// Create the instance
	_, err = rt.CreateInstance(ctx, &runtimev1.CreateInstanceRequest{
		InstanceId:          instanceID,
		OlapConnector:       olapDriver,
		RepoConnector:       "repo",
		EmbedCatalog:        embedCatalog,
		Variables:           opts.ProdVariables,
		IngestionLimitBytes: ingestionLimit,
		Annotations:         opts.Annotations.toMap(),
		Connectors: []*runtimev1.Connector{
			{
				Name:   olapDriver,
				Type:   olapDriver,
				Config: olapConfig,
			},
			{
				Name:   "repo",
				Type:   repoDriver,
				Config: map[string]string{"dsn": repoDSN},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Create the deployment
	depl, err := s.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         opts.ProjectID,
		Branch:            opts.ProdBranch,
		Slots:             opts.ProdSlots,
		RuntimeHost:       alloc.Host,
		RuntimeInstanceID: instanceID,
		RuntimeAudience:   alloc.Audience,
		Status:            database.DeploymentStatusPending,
		Logs:              "",
	})
	if err != nil {
		_, err2 := rt.DeleteInstance(ctx, &runtimev1.DeleteInstanceRequest{
			InstanceId: instanceID,
			DropDb:     olapDriver == "duckdb", // Only drop DB if it's DuckDB
		})
		return nil, multierr.Combine(err, err2)
	}

	return depl, nil
}

type updateDeploymentOptions struct {
	GithubURL            *string
	GithubInstallationID *int64
	Subpath              string
	Branch               string
	Variables            map[string]string
	Annotations          *deploymentAnnotations
}

func (s *Service) updateDeployment(ctx context.Context, depl *database.Deployment, opts *updateDeploymentOptions) error {
	if opts.GithubURL == nil || opts.GithubInstallationID == nil || opts.Branch == "" {
		return fmt.Errorf("cannot update deployment without github info")
	}

	repoDriver, repoDSN, err := githubRepoInfoForRuntime(*opts.GithubURL, *opts.GithubInstallationID, opts.Subpath, opts.Branch)
	if err != nil {
		return err
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
		if c.Name == "repo" {
			if c.Config == nil {
				c.Config = make(map[string]string)
			}
			c.Config["dsn"] = repoDSN
			c.Type = repoDriver
		}
	}

	var annotations map[string]string
	if opts.Annotations != nil { // annotations changed
		annotations = opts.Annotations.toMap()
	}
	_, err = rt.EditInstance(ctx, &runtimev1.EditInstanceRequest{
		InstanceId:  depl.RuntimeInstanceID,
		Connectors:  connectors,
		Annotations: annotations,
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

	if err := s.triggerReconcile(ctx, depl); err != nil {
		s.logger.Error("failed to trigger reconcile", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
		return err
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

	s.logger.Info("hibernate: starting", zap.Int("deployments", len(depls)))

	for _, depl := range depls {
		if depl.Status == database.DeploymentStatusReconciling && time.Since(depl.UpdatedOn) < 30*time.Minute {
			s.logger.Info("hibernate: skipping deployment because it is reconciling", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
			continue
		}

		proj, err := s.DB.FindProject(ctx, depl.ProjectID)
		if err != nil {
			s.logger.Error("hibernate: find project error", zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
			continue
		}

		s.logger.Info("hibernate: deleting deployment", zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID))

		err = s.teardownDeployment(ctx, proj, depl)
		if err != nil {
			s.logger.Error("hibernate: teardown deployment error", zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
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
		})
		if err != nil {
			return err
		}
	}

	s.logger.Info("hibernate: completed", zap.Int("deployments", len(depls)))

	return nil
}

func (s *Service) updateDeplVariables(ctx context.Context, depl *database.Deployment, variables map[string]string) error {
	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	_, err = rt.EditInstanceVariables(ctx, &runtimev1.EditInstanceVariablesRequest{
		InstanceId: depl.RuntimeInstanceID,
		Variables:  variables,
	})
	return err
}

func (s *Service) updateDeplAnnotations(ctx context.Context, depl *database.Deployment, annotations deploymentAnnotations) error {
	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	_, err = rt.EditInstanceAnnotations(ctx, &runtimev1.EditInstanceAnnotationsRequest{
		InstanceId:  depl.RuntimeInstanceID,
		Annotations: annotations.toMap(),
	})
	return err
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
		DropDb:     proj.ProdOLAPDriver == "duckdb", // Only drop DB if it's DuckDB
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

func githubRepoInfoForRuntime(githubURL string, installationID int64, subPath, branch string) (string, string, error) {
	dsn, err := json.Marshal(github.DSN{
		GithubURL:      githubURL,
		InstallationID: installationID,
		Subpath:        subPath,
		Branch:         branch,
	})
	if err != nil {
		return "", "", err
	}

	return "github", string(dsn), nil
}

type deploymentAnnotations struct {
	orgID    string
	orgName  string
	projID   string
	projName string
}

func newDeploymentAnnotations(org *database.Organization, proj *database.Project) deploymentAnnotations {
	return deploymentAnnotations{
		orgID:    org.ID,
		orgName:  org.Name,
		projID:   proj.ID,
		projName: proj.Name,
	}
}

func (da *deploymentAnnotations) toMap() map[string]string {
	return map[string]string{
		"organization_id":   da.orgID,
		"organization_name": da.orgName,
		"project_id":        da.projID,
		"project_name":      da.projName,
	}
}
