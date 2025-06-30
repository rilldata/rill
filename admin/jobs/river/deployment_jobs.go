package river

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/riverqueue/river"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

const validateDeploymentsForProjectTimeout = 5 * time.Minute

type ResetAllDeploymentsArgs struct{}

func (ResetAllDeploymentsArgs) Kind() string { return "reset_all_deployments" }

type ResetAllDeploymentsWorker struct {
	river.WorkerDefaults[ResetAllDeploymentsArgs]
	admin *admin.Service
}

func (w *ResetAllDeploymentsWorker) Work(ctx context.Context, job *river.Job[ResetAllDeploymentsArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.resetAllDeployments)
}

func (w *ResetAllDeploymentsWorker) resetAllDeployments(ctx context.Context) error {
	// Iterate over batches of projects to redeploy
	limit := 100
	afterName := ""
	stop := false
	for !stop {
		// Get batch and update iterator variables
		projs, err := w.admin.DB.FindProjects(ctx, afterName, limit)
		if err != nil {
			return err
		}
		if len(projs) < limit {
			stop = true
		}
		if len(projs) != 0 {
			afterName = projs[len(projs)-1].Name
		}

		// Process batch
		for _, proj := range projs {
			err := w.resetAllDeploymentsForProject(ctx, proj)
			if err != nil {
				// We log the error, but continues to the next project
				w.admin.Logger.Error("reset all deployments: failed to reset project deployments", zap.String("project_id", proj.ID), observability.ZapCtx(ctx), zap.Error(err))
			}
		}
	}

	return nil
}

func (w *ResetAllDeploymentsWorker) resetAllDeploymentsForProject(ctx context.Context, proj *database.Project) error {
	depls, err := w.admin.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return err
	}

	for _, depl := range depls {
		w.admin.Logger.Info("reset all deployments: redeploying deployment", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
		_, err = w.admin.RedeployProject(ctx, proj, depl)
		if err != nil {
			return err
		}
		w.admin.Logger.Info("reset all deployments: redeployed deployment", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
	}

	return nil
}

type ValidateDeploymentsArgs struct{}

func (ValidateDeploymentsArgs) Kind() string { return "validate_deployments" }

type ValidateDeploymentsWorker struct {
	river.WorkerDefaults[ValidateDeploymentsArgs]
	admin *admin.Service
}

func (w *ValidateDeploymentsWorker) Work(ctx context.Context, job *river.Job[ValidateDeploymentsArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.validateDeployments)
}

func (w *ValidateDeploymentsWorker) validateDeployments(ctx context.Context) error {
	var wg sync.WaitGroup
	ch := make(chan *database.Project)

	concurrency := 30
	if w.admin.ProvisionerMaxConcurrency > 0 {
		concurrency = w.admin.ProvisionerMaxConcurrency
	} else {
		w.admin.Logger.Warn("validate deployments: provisioner max concurrency invalid, using default concurrency of 30", zap.Int("provisioner_max_concurrency", w.admin.ProvisionerMaxConcurrency), observability.ZapCtx(ctx))
	}

	// Setup concurrent workers
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Read projects from shared channel
			for proj := range ch {
				err := w.validateDeploymentsForProject(ctx, proj)
				if err != nil {
					// We log the error, but continue to the next project
					w.admin.Logger.Error("validate deployments: failed to validate project deployments", zap.String("project_id", proj.ID), zap.Error(err), observability.ZapCtx(ctx))
				}
			}
		}()
	}

	// Iterate over batches of projects and add them to the shared channel
	limit := 100
	afterName := ""
	stop := false
	for !stop {
		// Get batch and update iterator variables
		projs, err := w.admin.DB.FindProjects(ctx, afterName, limit)
		if err != nil {
			return err
		}
		if len(projs) < limit {
			stop = true
		}
		if len(projs) != 0 {
			afterName = projs[len(projs)-1].Name
		}

		for _, proj := range projs {
			ch <- proj
		}
	}

	close(ch)
	wg.Wait()

	return nil
}

func (w *ValidateDeploymentsWorker) validateDeploymentsForProject(ctx context.Context, proj *database.Project) error {
	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, validateDeploymentsForProjectTimeout)
	defer cancel()

	// Get all project deployments
	depls, err := w.admin.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return err
	}
	if len(depls) == 0 {
		return nil
	}

	// Get project organization, we need this to create the deployment annotations
	org, err := w.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return err
	}

	// Determine the current production deployment, if any
	var prodDeplID string
	if proj.ProdDeploymentID != nil {
		prodDeplID = *proj.ProdDeploymentID
	}

	for _, depl := range depls {
		// If it appears to be an orphaned deployment, we tear it down.
		// This might for example happen if a redeploy failed after switching to the new deployment.
		// We consider a deployment orphaned if it is not the prod deployment and has not been updated in 3 hours.
		// The 3 hour delay is to ensure we don't tear down a deployment that is in the process of being created and is to become the new prod deployment.
		if depl.ID != prodDeplID && depl.UpdatedOn.Add(3*time.Hour).Before(time.Now()) {
			w.admin.Logger.Info("validate deployments: removing deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx))
			err = w.admin.TeardownDeployment(ctx, depl)
			if err != nil {
				w.admin.Logger.Error("validate deployments: failed to remove deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx), zap.Error(err))
				continue
			}
			w.admin.Logger.Info("validate deployments: removed deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx))
			continue
		}

		// Retrieve the deployment's provisioned resources
		prs, err := w.admin.DB.FindProvisionerResourcesForDeployment(ctx, depl.ID)
		if err != nil {
			return err
		}
		if len(prs) == 0 {
			continue
		}

		// Build annotations for the deployment
		annotations := w.admin.NewDeploymentAnnotations(org, proj)

		// Validate each provisioned resource
		for _, pr := range prs {
			w.admin.Logger.Info("validate deployments: checking resource", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("resource_id", pr.ID), zap.String("provisioner", pr.Provisioner), observability.ZapCtx(ctx))
			err := w.admin.CheckProvisionerResource(ctx, pr, annotations)
			if err != nil {
				w.admin.Logger.Error("validate deployments: failed to check resource", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("resource_id", pr.ID), zap.String("provisioner", pr.Provisioner), zap.Error(err), observability.ZapCtx(ctx))
				continue
			}
			w.admin.Logger.Info("validate deployments: checked resource", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("resource_id", pr.ID), zap.String("provisioner", pr.Provisioner), observability.ZapCtx(ctx))
		}
	}

	return nil
}

type DeploymentHealthCheckArgs struct{}

func (DeploymentHealthCheckArgs) Kind() string { return "deployments_health_check" }

type DeploymentHealthCheckWorker struct {
	river.WorkerDefaults[DeploymentHealthCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *DeploymentHealthCheckWorker) Work(ctx context.Context, job *river.Job[DeploymentHealthCheckArgs]) error {
	afterID := ""
	limit := 100
	seenHosts := map[string]bool{}
	expectedInstances := map[string][]string{}
	actualInstances := map[string][]string{}
	var mu sync.Mutex
	for {
		deployments, err := w.admin.DB.FindDeployments(ctx, afterID, limit)
		if err != nil {
			return fmt.Errorf("deployment health check: failed to get deployments: %w", err)
		}
		if len(deployments) == 0 {
			break
		}

		group, cctx := errgroup.WithContext(ctx)
		group.SetLimit(32)
		for _, d := range deployments {
			d := d
			if d.Status != database.DeploymentStatusOK {
				if time.Since(d.UpdatedOn) > time.Hour {
					w.logger.Error("deployment health check: deployment not ok", zap.String("project_id", d.ProjectID), zap.String("deployment_id", d.ID), zap.String("status", d.Status.String()), zap.Time("since", d.UpdatedOn), observability.ZapCtx(ctx))
				}
				continue
			}
			addExpectedInstance(expectedInstances, d)
			if seenHosts[d.RuntimeHost] {
				continue
			}
			seenHosts[d.RuntimeHost] = true
			group.Go(func() error {
				instances, ok := w.deploymentHealthCheck(cctx, d)
				if ok {
					mu.Lock()
					actualInstances[d.RuntimeHost] = instances
					mu.Unlock()
				}
				return nil
			})
		}
		if err := group.Wait(); err != nil {
			return err
		}
		if len(deployments) < limit {
			break
		}
		afterID = deployments[len(deployments)-1].ID
	}

	for host, expected := range expectedInstances {
		actual, ok := actualInstances[host]
		if !ok {
			continue
		}
		for _, instance := range expected {
			if slices.Contains(actual, instance) {
				continue
			}
			d, err := w.admin.DB.FindDeploymentByInstanceID(ctx, instance)
			if err != nil {
				if errors.Is(err, database.ErrNotFound) {
					continue
				}
				w.logger.Error("deployment health check: failed to find deployment", zap.String("instance_id", instance), zap.Error(err), observability.ZapCtx(ctx))
				continue
			}
			annotations, err := w.annotationsForDeployment(ctx, d)
			if err != nil {
				w.logger.Error("deployment health check: failed to find deployment_annotations", zap.String("project_id", d.ProjectID), zap.String("deployment_id", d.ID), zap.Error(err), observability.ZapCtx(ctx))
				continue
			}
			f := []zap.Field{zap.String("project_id", d.ProjectID), zap.String("deployment_id", d.ID), zap.String("instance_id", instance), zap.String("host", d.RuntimeHost), observability.ZapCtx(ctx)}
			for k, v := range annotations.ToMap() {
				f = append(f, zap.String(k, v))
			}
			w.logger.Error("deployment health check: missing instance on runtime", f...)
		}
	}
	return nil
}

func (w *DeploymentHealthCheckWorker) deploymentHealthCheck(ctx context.Context, d *database.Deployment) (instances []string, runtimeOK bool) {
	ctx, span := tracer.Start(ctx, "deploymentHealthCheck", trace.WithAttributes(attribute.String("project_id", d.ProjectID), attribute.String("deployment_id", d.ID)))
	defer span.End()

	client, err := w.admin.OpenRuntimeClient(d)
	if err != nil {
		w.logger.Error("deployment health check: failed to open runtime client", zap.String("project_id", d.ProjectID), zap.String("deployment_id", d.ID), zap.String("host", d.RuntimeHost), zap.Error(err), observability.ZapCtx(ctx))
		return nil, false
	}
	defer client.Close()

	resp, err := client.Health(ctx, &runtimev1.HealthRequest{})
	if err != nil {
		if status.Code(err) != codes.Unavailable {
			w.logger.Error("deployment health check: health check call failed", zap.String("project_id", d.ProjectID), zap.String("deployment_id", d.ID), zap.String("host", d.RuntimeHost), zap.Error(err), observability.ZapCtx(ctx))
			return nil, false
		}
		// an unavailable error could also be because the deployment got deleted
		d, dbErr := w.admin.DB.FindDeployment(ctx, d.ID)
		if dbErr != nil {
			if errors.Is(dbErr, database.ErrNotFound) {
				return nil, false
			}
			w.logger.Error("deployment health check: failed to find deployment", zap.String("project_id", d.ProjectID), zap.String("deployment_id", d.ID), zap.Error(dbErr), observability.ZapCtx(ctx))
			return nil, false
		}
		if d.Status == database.DeploymentStatusOK {
			w.logger.Error("deployment health check: health check call failed", zap.String("project_id", d.ProjectID), zap.String("deployment_id", d.ID), zap.String("host", d.RuntimeHost), zap.Error(err), observability.ZapCtx(ctx))
		}
		return nil, false
	}

	if runtimeUnhealthy(resp) {
		f := []zap.Field{zap.String("project_id", d.ProjectID), zap.String("deployment_id", d.ID), zap.String("host", d.RuntimeHost), observability.ZapCtx(ctx)}
		if resp.LimiterError != "" {
			f = append(f, zap.String("limiter_error", resp.LimiterError))
		}
		if resp.ConnCacheError != "" {
			f = append(f, zap.String("conn_cache_error", resp.ConnCacheError))
		}
		if resp.MetastoreError != "" {
			f = append(f, zap.String("metastore_error", resp.MetastoreError))
		}
		if resp.NetworkError != "" {
			f = append(f, zap.String("network_error", resp.NetworkError))
		}
		w.logger.Error("deployment health check: runtime is unhealthy", f...)
		return nil, false
	}

	for instanceID, health := range resp.InstancesHealth {
		instances = append(instances, instanceID)
		if !instanceUnhealthy(health) {
			continue
		}

		d := d
		if d.RuntimeInstanceID != instanceID {
			d, err = w.admin.DB.FindDeploymentByInstanceID(ctx, instanceID)
			if err != nil {
				w.logger.Error("deployment health check: failed to find deployment", zap.String("instance_id", instanceID), zap.Error(err), observability.ZapCtx(ctx))
				continue
			}
		}

		annotations, err := w.annotationsForDeployment(ctx, d)
		if err != nil {
			w.logger.Error("deployment health check: failed to find deployment_annotations", zap.String("project_id", d.ProjectID), zap.String("deployment_id", d.ID), zap.Error(err), observability.ZapCtx(ctx))
			continue
		}
		f := []zap.Field{zap.String("deployment_id", d.ID), zap.String("host", d.RuntimeHost), zap.String("instance_id", instanceID), observability.ZapCtx(ctx)}
		for k, v := range annotations.ToMap() {
			f = append(f, zap.String(k, v))
		}

		for d, err := range health.MetricsViewErrors {
			f := slices.Clone(f)
			f = append(f, zap.String("metrics_view", d), zap.String("error", err))
			w.logger.Warn("deployment health check: metrics view error", f...)
		}

		logAtError := false
		if health.ControllerError != "" {
			logAtError = true
			f = append(f, zap.String("controller_error", health.ControllerError))
		}
		if health.OlapError != "" {
			logAtError = true
			f = append(f, zap.String("olap_error", health.OlapError))
		}
		if health.RepoError != "" {
			logAtError = true
			f = append(f, zap.String("repo_error", health.RepoError))
		}
		if len(health.MetricsViewErrors) > 0 {
			f = append(f, zap.Int("metrics_view_errors", len(health.MetricsViewErrors)))
		}
		if health.ParseErrorCount > 0 {
			f = append(f, zap.Int32("parse_errors", health.ParseErrorCount))
		}
		if health.ReconcileErrorCount > 0 {
			f = append(f, zap.Int32("reconcile_errors", health.ReconcileErrorCount))
		}
		if logAtError {
			w.logger.Error("deployment health check: instance is unhealthy", f...)
		} else {
			w.logger.Warn("deployment health check: instance is unhealthy", f...)
		}
	}
	return instances, true
}

func (w *DeploymentHealthCheckWorker) annotationsForDeployment(ctx context.Context, d *database.Deployment) (*admin.DeploymentAnnotations, error) {
	proj, err := w.admin.DB.FindProject(ctx, d.ProjectID)
	if err != nil {
		return nil, err
	}
	org, err := w.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}
	annotations := w.admin.NewDeploymentAnnotations(org, proj)
	return &annotations, nil
}

func runtimeUnhealthy(r *runtimev1.HealthResponse) bool {
	return r.LimiterError != "" || r.ConnCacheError != "" || r.MetastoreError != "" || r.NetworkError != ""
}

func instanceUnhealthy(i *runtimev1.InstanceHealth) bool {
	return i.OlapError != "" || i.ControllerError != "" || i.RepoError != "" || len(i.MetricsViewErrors) != 0 || i.ParseErrorCount > 0 || i.ReconcileErrorCount > 0
}

func addExpectedInstance(expectedInstances map[string][]string, d *database.Deployment) {
	if expectedInstances[d.RuntimeHost] == nil {
		expectedInstances[d.RuntimeHost] = []string{}
	}
	expectedInstances[d.RuntimeHost] = append(expectedInstances[d.RuntimeHost], d.RuntimeInstanceID)
}

type HibernateExpiredDeploymentsArgs struct{}

func (HibernateExpiredDeploymentsArgs) Kind() string { return "hibernate_expired_deployments" }

type HibernateExpiredDeploymentsWorker struct {
	river.WorkerDefaults[HibernateExpiredDeploymentsArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *HibernateExpiredDeploymentsWorker) Work(ctx context.Context, job *river.Job[HibernateExpiredDeploymentsArgs]) error {
	depls, err := w.admin.DB.FindExpiredDeployments(ctx)
	if err != nil {
		return err
	}
	if len(depls) == 0 {
		return nil
	}

	w.logger.Info("hibernate: starting", zap.Int("deployments", len(depls)))

	for _, depl := range depls {
		w.logger.Info("hibernate: deleting deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID))
		err := w.hibernateExpiredDeployment(ctx, depl)
		if err != nil {
			w.logger.Error("hibernate: failed to delete deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
			continue
		}
		w.logger.Info("hibernate: deleted deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID))
	}

	w.logger.Info("hibernate: completed", zap.Int("deployments", len(depls)))

	return nil
}

func (w *HibernateExpiredDeploymentsWorker) hibernateExpiredDeployment(ctx context.Context, depl *database.Deployment) error {
	proj, err := w.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return err
	}

	switch depl.Environment {
	case "prod":
		// Tear down prod deployments on hibernation
		// TODO: update this to stop deployment instead of tearing it down when the frontend supports it
		if proj.ProdDeploymentID != nil && *proj.ProdDeploymentID == depl.ID {
			_, err = w.admin.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
				Name:                 proj.Name,
				Description:          proj.Description,
				Public:               proj.Public,
				Provisioner:          proj.Provisioner,
				ArchiveAssetID:       proj.ArchiveAssetID,
				GitRemote:            proj.GitRemote,
				GithubInstallationID: proj.GithubInstallationID,
				GithubRepoID:         proj.GithubRepoID,
				ManagedGitRepoID:     proj.ManagedGitRepoID,
				ProdVersion:          proj.ProdVersion,
				ProdBranch:           proj.ProdBranch,
				Subpath:              proj.Subpath,
				ProdSlots:            proj.ProdSlots,
				ProdTTLSeconds:       proj.ProdTTLSeconds,
				ProdDeploymentID:     nil,
				DevSlots:             proj.DevSlots,
				DevTTLSeconds:        proj.DevTTLSeconds,
				Annotations:          proj.Annotations,
			})
			if err != nil {
				return err
			}
		}

		err = w.admin.TeardownDeployment(ctx, depl)
		if err != nil {
			return err
		}
	case "dev":
		// For dev deployments we stop the deployment
		err = w.admin.StopDeployment(ctx, depl)
		if err != nil {
			return err
		}
	}

	return nil
}
