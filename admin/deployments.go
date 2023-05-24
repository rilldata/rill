package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
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

func (s *Service) createDeployment(ctx context.Context, proj *database.Project) (*database.Deployment, error) {
	// We require Github info on project to create a deployment
	if proj.GithubURL == nil || proj.GithubInstallationID == nil || proj.ProdBranch == "" {
		return nil, fmt.Errorf("cannot create project without github info")
	}
	repoDriver, repoDSN, err := githubRepoInfoForRuntime(*proj.GithubURL, *proj.GithubInstallationID, proj.Subpath, proj.ProdBranch)
	if err != nil {
		return nil, err
	}

	// Get a runtime with capacity for the deployment
	alloc, err := s.Provisioner.Provision(ctx, &provisioner.ProvisionOptions{
		OLAPDriver: proj.ProdOLAPDriver,
		Slots:      proj.ProdSlots,
		Region:     proj.Region,
	})
	if err != nil {
		return nil, err
	}

	// Build instance config
	instanceID := strings.ReplaceAll(uuid.New().String(), "-", "")
	olapDriver := proj.ProdOLAPDriver
	olapDSN := proj.ProdOLAPDSN
	var embedCatalog bool
	var ingestionLimit int64
	if olapDriver == "duckdb" {
		if olapDSN != "" {
			return nil, fmt.Errorf("passing a DSN is not allowed for driver 'duckdb'")
		}
		if proj.ProdSlots == 0 {
			return nil, fmt.Errorf("slot count can't be 0 for driver 'duckdb'")
		}

		embedCatalog = true
		ingestionLimit = alloc.StorageBytes

		olapDSN = fmt.Sprintf("%s.db?rill_pool_size=%d&threads=%d&max_memory=%dGB", path.Join(alloc.DataDir, instanceID), alloc.CPU, alloc.CPU, alloc.MemoryGB)
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
		OlapDriver:          olapDriver,
		OlapDsn:             olapDSN,
		RepoDriver:          repoDriver,
		RepoDsn:             repoDSN,
		EmbedCatalog:        embedCatalog,
		Variables:           proj.ProdVariables,
		IngestionLimitBytes: ingestionLimit,
	})
	if err != nil {
		return nil, err
	}

	// Create the deployment
	depl, err := s.DB.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         proj.ID,
		Branch:            proj.ProdBranch,
		Slots:             proj.ProdSlots,
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
	Reconcile            bool
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
	inst := res.Instance

	_, err = rt.EditInstance(ctx, &runtimev1.EditInstanceRequest{
		InstanceId:          inst.InstanceId,
		OlapDriver:          inst.OlapDriver,
		OlapDsn:             inst.OlapDsn,
		RepoDriver:          repoDriver,
		RepoDsn:             repoDSN,
		EmbedCatalog:        inst.EmbedCatalog,
		Variables:           opts.Variables,
		IngestionLimitBytes: inst.IngestionLimitBytes,
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

	if opts.Reconcile {
		if err := s.triggerReconcile(ctx, depl); err != nil {
			s.logger.Error("failed to trigger reconcile", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
			return err
		}
	}

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
