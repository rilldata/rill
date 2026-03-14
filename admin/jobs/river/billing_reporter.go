package river

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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

	// Sync ClickHouse Cloud cluster info from runtime to postgres
	w.syncCHCCloudInfo(ctx)

	return nil
}

// syncCHCCloudInfo fetches the latest ClickHouse Cloud service info from each runtime
// and persists chc_cluster_size and rill_min_slots on the project in postgres.
func (w *BillingReporterWorker) syncCHCCloudInfo(ctx context.Context) {
	projects, err := w.admin.DB.FindProjectsWithCHC(ctx)
	if err != nil {
		w.logger.Warn("CHC sync: failed to find projects with CHC", zap.Error(err))
		return
	}

	for _, proj := range projects {
		if proj.PrimaryDeploymentID == nil {
			continue
		}
		depl, err := w.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
		if err != nil {
			w.logger.Debug("CHC sync: failed to find deployment", zap.String("project_id", proj.ID), zap.Error(err))
			continue
		}
		if depl.Status != database.DeploymentStatusRunning {
			continue
		}

		rt, err := w.admin.OpenRuntimeClient(depl)
		if err != nil {
			w.logger.Debug("CHC sync: failed to open runtime client", zap.String("project_id", proj.ID), zap.Error(err))
			continue
		}

		resp, err := rt.GetInstance(ctx, &runtimev1.GetInstanceRequest{
			InstanceId: depl.RuntimeInstanceID,
			Sensitive:  true,
		})
		rt.Close()
		if err != nil {
			w.logger.Debug("CHC sync: failed to get instance", zap.String("project_id", proj.ID), zap.Error(err))
			continue
		}

		// Find the OLAP connector config with cloud_max_memory_gb
		var maxMemoryGB float64
		found := false
		for _, conn := range resp.Instance.ProjectConnectors {
			if conn.Name != resp.Instance.OlapConnector || conn.Config == nil {
				continue
			}
			if v, ok := conn.Config.Fields["cloud_max_memory_gb"]; ok {
				maxMemoryGB = v.GetNumberValue()
				found = true
			}
			break
		}
		if !found || maxMemoryGB == 0 {
			continue
		}

		// Compute min slots from the cluster memory
		minSlots := admin.CHCMinSlotsForMemory(maxMemoryGB)
		minSlotsInt64 := int64(minSlots)

		// Skip update if nothing changed
		if proj.ChcClusterSize != nil && *proj.ChcClusterSize == maxMemoryGB &&
			proj.RillMinSlots != nil && *proj.RillMinSlots == minSlotsInt64 {
			continue
		}

		_, err = w.admin.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
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
			w.logger.Warn("CHC sync: failed to update project", zap.String("project_id", proj.ID), zap.Error(err))
			continue
		}

		w.logger.Info("CHC sync: updated project cluster info",
			zap.String("project_id", proj.ID),
			zap.Float64("max_memory_gb", maxMemoryGB),
			zap.Int("rill_min_slots", minSlots),
		)
	}
}
