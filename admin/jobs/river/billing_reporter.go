package river

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	chdriver "github.com/rilldata/rill/runtime/drivers/clickhouse"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type BillingReporterArgs struct{}

func (BillingReporterArgs) Kind() string { return "billing_reporter" }

type BillingReporterWorker struct {
	river.WorkerDefaults[BillingReporterArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// NewBillingReporterWorker creates a new worker that reports billing information.
func (w *BillingReporterWorker) Work(ctx context.Context, job *river.Job[BillingReporterArgs]) error {
	// Always sync CHC info regardless of billing outcome
	defer w.syncCHCCloudInfo(ctx)

	// Get reporting granularity
	var granularity time.Duration
	var sqlGrainIdentifier string
	var gracePeriod time.Duration
	switch w.admin.Biller.GetReportingGranularity() {
	case billing.UsageReportingGranularityHour:
		granularity = time.Hour
		gracePeriod = time.Hour // keep 1 hour of delay as buffer, cron job runs at 55 minutes of each hour, so effectively we will report until the end of the last to last hour
		sqlGrainIdentifier = "hour"
	case billing.UsageReportingGranularityNone:
		w.logger.Debug("skipping usage reporting: no reporting granularity configured")
		return nil
	default:
		return fmt.Errorf("unsupported reporting granularity: %s", w.admin.Biller.GetReportingGranularity())
	}

	t, err := w.admin.DB.FindBillingUsageReportedOn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last usage reporting time: %w", err)
	}

	// after going back by grace period, report until the "end" of the last grain period
	endTime := time.Now().UTC().Add(-gracePeriod).Truncate(granularity)

	// start reporting from the last reported time or from the "start" of the last grain period for first time reporting
	var startTime time.Time
	if t.IsZero() {
		startTime = endTime.Add(-granularity)
	} else {
		startTime = t.UTC()
	}

	if !startTime.Before(endTime) {
		w.logger.Debug("skipping usage reporting: no new usage data available", zap.Time("start_time", startTime), zap.Time("end_time", endTime))
		return nil
	}

	// Get metrics client
	client, ok, err := w.admin.OpenMetricsProject(ctx)
	if err != nil {
		w.logger.Error("failed to report usage: unable to get metrics client", zap.Error(err))
		return err
	}
	if !ok {
		w.logger.Debug("skipping usage reporting: no metrics project configured")
		return nil
	}

	reportedOrgs := make(map[string]struct{})
	stop := false
	limit := 10000
	afterTime := time.Time{}
	afterOrgID := ""
	afterProjectID := ""
	afterEventName := ""

	checkPoint := startTime
	maxEndTime := time.Time{}
	// loop until all the usage data is reported
	for !stop {
		u, err := client.GetUsageMetrics(ctx, startTime, endTime, afterTime, afterOrgID, afterProjectID, afterEventName, sqlGrainIdentifier, limit)
		if err != nil {
			return fmt.Errorf("failed to get usage metrics: %w", err)
		}

		if len(u) == 0 {
			break
		}

		if len(u) < limit {
			stop = true
		} else {
			afterTime = u[len(u)-1].StartTime
			afterOrgID = u[len(u)-1].OrgID
			afterProjectID = u[len(u)-1].ProjectID
			afterEventName = u[len(u)-1].EventName
		}
		// since the usage data is ordered by start time first and end time is just (start time + grain), we can directly get the max end time
		maxEndTime = u[len(u)-1].EndTime

		var usage []*billing.Usage
		for _, m := range u {
			reportedOrgs[m.OrgID] = struct{}{}

			customerID := m.OrgID
			if m.BillingCustomerID != nil && *m.BillingCustomerID != "" {
				// org might have been deleted or recently created in both cases billing customer id will be null. If billing not initialized for the org, then it will be empty string
				// in all cases just use org ID to report in hope that org ID will be set as billing customer id in the future if not reported values will be ignored
				customerID = *m.BillingCustomerID
			}

			usage = append(usage, &billing.Usage{
				CustomerID:     customerID,
				MetricName:     m.EventName,
				Value:          m.MaxValue,
				ReportingGrain: w.admin.Biller.GetReportingGranularity(),
				StartTime:      m.StartTime,
				EndTime:        m.EndTime,
				Metadata:       map[string]interface{}{"org_id": m.OrgID, "project_id": m.ProjectID, "project_name": m.ProjectName, "billing_service": m.BillingService},
			})
		}

		err = w.admin.Biller.ReportUsage(ctx, usage)
		if err != nil {
			return fmt.Errorf("failed to report usage: %w", err)
		}

		if afterTime.After(checkPoint) {
			checkPoint = afterTime
			err = w.admin.DB.UpdateBillingUsageReportedOn(ctx, checkPoint)
			if err != nil {
				return fmt.Errorf("failed to update last usage reporting time: %w", err)
			}
		}
	}

	if len(reportedOrgs) == 0 {
		w.logger.Named("billing").Warn("skipping usage reporting: no usage data available", zap.Time("start_time", startTime), zap.Time("end_time", endTime))
		return nil
	}

	// should never happen, adding a check for safety
	if maxEndTime.IsZero() {
		return fmt.Errorf("failed to update last usage reporting time: max end time not updated after reporting usage data")
	}

	err = w.admin.DB.UpdateBillingUsageReportedOn(ctx, maxEndTime)
	if err != nil {
		return fmt.Errorf("failed to update last usage reporting time: %w", err)
	}

	// TODO move the validation to background job
	// get orgs which have billing customer id
	orgs, err := w.admin.DB.FindOrganizationIDsWithBilling(ctx)
	if err != nil {
		return fmt.Errorf("failed to report usage: unable to fetch orgs: %w", err)
	}

	// get orgs which have billing customer id and not reported in this run
	for _, org := range orgs {
		if _, ok := reportedOrgs[org]; !ok {
			// count the projects which are not hibernated and created before the given time
			count, err := w.admin.DB.CountBillingProjectsForOrganization(ctx, org, endTime)
			if err != nil {
				w.logger.Warn("failed to validate active projects for org", zap.String("org_id", org), zap.Error(err))
				continue
			}
			if count > 0 {
				w.logger.Warn("skipped usage reporting for org as no usage data was available", zap.String("org_id", org), zap.Time("start_time", startTime), zap.Time("end_time", endTime))
			}
		}
	}

	return nil
}

// syncCHCCloudInfo fetches the latest ClickHouse Cloud service info from each runtime
// and persists chc_cluster_size and rill_min_slots on the project in postgres.
func (w *BillingReporterWorker) syncCHCCloudInfo(ctx context.Context) {
	projects, err := w.admin.DB.FindProjectsWithCHC(ctx)
	if err != nil {
		w.logger.Warn("CHC sync: failed to find projects", zap.Error(err))
		return
	}

	for _, proj := range projects {
		if proj.PrimaryDeploymentID == nil {
			continue
		}
		depl, err := w.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
		if err != nil {
			continue
		}
		if depl.Status != database.DeploymentStatusRunning {
			continue
		}

		rt, err := w.admin.OpenRuntimeClient(depl)
		if err != nil {
			continue
		}

		resp, err := rt.GetInstance(ctx, &runtimev1.GetInstanceRequest{
			InstanceId: depl.RuntimeInstanceID,
			Sensitive:  true,
		})
		rt.Close()
		if err != nil {
			continue
		}

		// Extract host from the connector config (may be templated but resolved_host is injected)
		var connHost string
		for _, conn := range resp.Instance.ProjectConnectors {
			if conn.Name != resp.Instance.OlapConnector || conn.Config == nil {
				continue
			}
			if v, ok := conn.Config.Fields["resolved_host"]; ok {
				connHost = v.GetStringValue()
			}
			if connHost == "" {
				if v, ok := conn.Config.Fields["host"]; ok {
					connHost = v.GetStringValue()
				}
			}
			if connHost == "" {
				if v, ok := conn.Config.Fields["dsn"]; ok {
					connHost = v.GetStringValue()
				}
			}
			break
		}

		// Always call the CHC Cloud API directly for the authoritative status
		if connHost == "" || !strings.Contains(strings.ToLower(connHost), ".clickhouse.cloud") {
			continue
		}
		info := w.fetchCHCInfoDirectly(ctx, proj, connHost)
		if info == nil {
			continue
		}
		maxMemoryGB := info.MaxMemoryGB
		cloudStatus := info.Status

		if maxMemoryGB == 0 {
			continue
		}

		// Compute min slots from the cluster memory
		minSlots := admin.CHCMinSlotsForMemory(maxMemoryGB)
		minSlotsInt64 := int64(minSlots)

		// Update cluster size and min slots if changed (DB-only; no deployment reconciliation needed)
		if proj.ChcClusterSize == nil || *proj.ChcClusterSize != maxMemoryGB ||
			proj.RillMinSlots == nil || *proj.RillMinSlots != minSlotsInt64 {
			proj, err = w.admin.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
				Name:                 proj.Name,
				Description:          proj.Description,
				Public:               proj.Public,
				DirectoryName:        proj.DirectoryName,
				ArchiveAssetID:       proj.ArchiveAssetID,
				GitRemote:            proj.GitRemote,
				GithubInstallationID: proj.GithubInstallationID,
				GithubRepoID:         proj.GithubRepoID,
				ManagedGitRepoID:     proj.ManagedGitRepoID,
				Subpath:              proj.Subpath,
				ProdVersion:          proj.ProdVersion,
				PrimaryBranch:        proj.PrimaryBranch,
				PrimaryDeploymentID:  proj.PrimaryDeploymentID,
				ProdSlots:            proj.ProdSlots,
				ProdTTLSeconds:       proj.ProdTTLSeconds,
				DevSlots:             proj.DevSlots,
				DevTTLSeconds:        proj.DevTTLSeconds,
				Provisioner:          proj.Provisioner,
				Annotations:          proj.Annotations,
				ChcClusterSize:       &maxMemoryGB,
				RillMinSlots:         &minSlotsInt64,
			})
			if err != nil {
				w.logger.Warn("CHC sync: failed to update project cluster info", zap.String("project_id", proj.ID), zap.Error(err))
				continue
			}
			w.logger.Info("CHC sync: updated project cluster info",
				zap.String("project_id", proj.ID),
				zap.Float64("max_memory_gb", maxMemoryGB),
				zap.Int("rill_min_slots", minSlots),
			)
		}

		// Auto-scale slots based on CHC cloud status
		const autoScaleAnnotation = "rill.dev/chc-auto-scaled-slots"
		if cloudStatus == "idle" || cloudStatus == "stopped" {
			// CHC is hibernated: scale down to 1 slot to minimize cost
			if proj.ProdSlots > 1 {
				annotations := maps.Clone(proj.Annotations)
				if annotations == nil {
					annotations = make(map[string]string)
				}
				annotations[autoScaleAnnotation] = "true"

				_, err = w.admin.UpdateProject(ctx, proj, &database.UpdateProjectOptions{
					Name:                 proj.Name,
					Description:          proj.Description,
					Public:               proj.Public,
					DirectoryName:        proj.DirectoryName,
					ArchiveAssetID:       proj.ArchiveAssetID,
					GitRemote:            proj.GitRemote,
					GithubInstallationID: proj.GithubInstallationID,
					GithubRepoID:         proj.GithubRepoID,
					ManagedGitRepoID:     proj.ManagedGitRepoID,
					Subpath:              proj.Subpath,
					ProdVersion:          proj.ProdVersion,
					PrimaryBranch:        proj.PrimaryBranch,
					PrimaryDeploymentID:  proj.PrimaryDeploymentID,
					ProdSlots:            1,
					ProdTTLSeconds:       proj.ProdTTLSeconds,
					DevSlots:             proj.DevSlots,
					DevTTLSeconds:        proj.DevTTLSeconds,
					Provisioner:          proj.Provisioner,
					Annotations:          annotations,
					ChcClusterSize:       proj.ChcClusterSize,
					RillMinSlots:         proj.RillMinSlots,
				})
				if err != nil {
					w.logger.Warn("CHC sync: failed to auto-scale slots down", zap.String("project_id", proj.ID), zap.Error(err))
					continue
				}
				w.logger.Info("CHC sync: auto-scaled slots to 1 (CHC hibernated)",
					zap.String("project_id", proj.ID),
					zap.Int("previous_slots", proj.ProdSlots),
				)
			}
		} else if cloudStatus == "running" {
			// CHC is running: restore slots to rill_min_slots if needed
			if proj.RillMinSlots != nil && proj.ProdSlots < int(*proj.RillMinSlots) {
				annotations := maps.Clone(proj.Annotations)
				if annotations != nil {
					delete(annotations, autoScaleAnnotation)
				}

				_, err = w.admin.UpdateProject(ctx, proj, &database.UpdateProjectOptions{
					Name:                 proj.Name,
					Description:          proj.Description,
					Public:               proj.Public,
					DirectoryName:        proj.DirectoryName,
					ArchiveAssetID:       proj.ArchiveAssetID,
					GitRemote:            proj.GitRemote,
					GithubInstallationID: proj.GithubInstallationID,
					GithubRepoID:         proj.GithubRepoID,
					ManagedGitRepoID:     proj.ManagedGitRepoID,
					Subpath:              proj.Subpath,
					ProdVersion:          proj.ProdVersion,
					PrimaryBranch:        proj.PrimaryBranch,
					PrimaryDeploymentID:  proj.PrimaryDeploymentID,
					ProdSlots:            int(*proj.RillMinSlots),
					ProdTTLSeconds:       proj.ProdTTLSeconds,
					DevSlots:             proj.DevSlots,
					DevTTLSeconds:        proj.DevTTLSeconds,
					Provisioner:          proj.Provisioner,
					Annotations:          annotations,
					ChcClusterSize:       proj.ChcClusterSize,
					RillMinSlots:         proj.RillMinSlots,
				})
				if err != nil {
					w.logger.Warn("CHC sync: failed to restore slots on wake-up", zap.String("project_id", proj.ID), zap.Error(err))
					continue
				}
				w.logger.Info("CHC sync: restored slots on CHC wake-up",
					zap.String("project_id", proj.ID),
					zap.Int64("restored_slots", *proj.RillMinSlots),
				)
			}
		}
	}
}

// fetchCHCInfoDirectly calls the ClickHouse Cloud API using the project's stored API keys.
// Used as a fallback when the runtime connector can't open (e.g. CHC is stopped).
func (w *BillingReporterWorker) fetchCHCInfoDirectly(ctx context.Context, proj *database.Project, host string) *chdriver.CloudServiceInfo {
	env := "prod"
	vars, err := w.admin.DB.FindProjectVariables(ctx, proj.ID, &env)
	if err != nil {
		w.logger.Debug("CHC sync: failed to find project variables", zap.String("project_id", proj.ID), zap.Error(err))
		return nil
	}

	var keyID, keySecret string
	for _, v := range vars {
		switch v.Name {
		case "CLICKHOUSE_CLOUD_API_KEY_ID":
			keyID = v.Value
		case "CLICKHOUSE_CLOUD_API_KEY_SECRET":
			keySecret = v.Value
		}
	}
	if keyID == "" || keySecret == "" {
		return nil
	}

	client := chdriver.NewCloudAPIClient(keyID, keySecret)
	if client == nil {
		return nil
	}

	info, err := client.FindServiceByHost(ctx, host)
	if err != nil {
		w.logger.Debug("CHC sync: direct API lookup failed", zap.String("project_id", proj.ID), zap.Error(err))
		return nil
	}
	return info
}
