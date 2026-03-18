package river

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/client"
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
	// Sync OLAP connector types for running deployments (best-effort; runs even if billing fails).
	w.syncOlapConnectors(ctx)

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

	const slotRatePerHr = 0.15 // $/slot/hr for credit accounting

	reportedOrgs := make(map[string]struct{})
	orgCreditCost := make(map[string]float64) // org_id -> accumulated dollar cost for this reporting window
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

			// Accumulate slot cost per org for free-plan credit accounting
			hours := m.EndTime.Sub(m.StartTime).Hours()
			orgCreditCost[m.OrgID] += m.MaxValue * slotRatePerHr * hours

			customerID := m.OrgID
			if m.BillingCustomerID != nil && *m.BillingCustomerID != "" {
				// org might have been deleted or recently created in both cases billing customer id will be null. If billing not initialized for the org, then it will be empty string
				// in all cases just use org ID to report in hope that org ID will be set as billing customer id in the future if not reported values will be ignored
				customerID = *m.BillingCustomerID
			}

			meta := map[string]interface{}{
				"org_id":          m.OrgID,
				"project_id":     m.ProjectID,
				"project_name":   m.ProjectName,
				"billing_service": m.BillingService,
			}

			usage = append(usage, &billing.Usage{
				CustomerID:     customerID,
				MetricName:     m.EventName,
				Value:          m.MaxValue,
				ReportingGrain: w.admin.Biller.GetReportingGranularity(),
				StartTime:      m.StartTime,
				EndTime:        m.EndTime,
				Metadata:       meta,
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

	// Increment credit_used for free-plan orgs based on slot usage cost
	for orgID, cost := range orgCreditCost {
		if cost <= 0 {
			continue
		}
		if err := w.admin.DB.IncrementOrganizationCreditUsed(ctx, orgID, cost); err != nil {
			w.logger.Warn("failed to increment credit usage for org", zap.String("org_id", orgID), zap.Float64("cost", cost), zap.Error(err))
		}
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

// syncOlapConnectors detects and persists the OLAP connector driver type for running deployments.
// This allows the frontend to show the correct engine label even when a project is hibernated.
func (w *BillingReporterWorker) syncOlapConnectors(ctx context.Context) {
	afterID := ""
	limit := 100
	for {
		depls, err := w.admin.DB.FindDeployments(ctx, afterID, limit)
		if err != nil {
			w.logger.Warn("olap sync: failed to list deployments", zap.Error(err))
			return
		}
		for _, depl := range depls {
			if depl.Status != database.DeploymentStatusRunning {
				continue
			}
			w.syncOlapConnectorForDeployment(ctx, depl)
		}
		if len(depls) < limit {
			break
		}
		afterID = depls[len(depls)-1].ID
	}
}

func (w *BillingReporterWorker) syncOlapConnectorForDeployment(ctx context.Context, depl *database.Deployment) {
	proj, err := w.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		w.logger.Debug("olap sync: failed to find project", zap.String("project_id", depl.ProjectID), zap.Error(err))
		return
	}

	rt, err := w.admin.OpenRuntimeClient(depl)
	if err != nil {
		w.logger.Debug("olap sync: failed to open runtime client", zap.String("project_id", depl.ProjectID), zap.Error(err))
		return
	}
	defer rt.Close()

	resp, err := rt.GetInstance(ctx, &runtimev1.GetInstanceRequest{
		InstanceId: depl.RuntimeInstanceID,
		Sensitive:  true,
	})
	if err != nil {
		w.logger.Debug("olap sync: failed to get instance", zap.String("project_id", depl.ProjectID), zap.Error(err))
		return
	}

	olapConnectorName := resp.Instance.OlapConnector
	if olapConnectorName == "" {
		return
	}

	// Resolve the connector driver type from the connector name
	var connectorType string
	var connector *runtimev1.Connector
	for _, c := range resp.Instance.ProjectConnectors {
		if c.Name == olapConnectorName {
			connectorType = c.Type
			connector = c
			break
		}
	}
	if connectorType == "" {
		return
	}

	// Update OLAP connector type if changed
	if proj.OlapConnector == nil || *proj.OlapConnector != connectorType {
		if err := w.admin.DB.UpdateProjectOlapConnector(ctx, proj.ID, connectorType); err != nil {
			w.logger.Warn("olap sync: failed to update olap connector", zap.String("project_id", proj.ID), zap.Error(err))
		}
	}

	// Detect and sync cluster slots for non-DuckDB connectors (Live Connect)
	w.syncClusterSlots(ctx, proj, depl, rt, connector)
}

// clusterDetectionSQL returns the SQL query for detecting cluster vCPUs, or empty if unsupported.
func clusterDetectionSQL(connector *runtimev1.Connector) string {
	if connector == nil {
		return ""
	}
	switch connector.Type {
	case "clickhouse":
		return `SELECT
    if(cgroup_cpu > 0, cgroup_cpu, os_cpu) AS vcpus
FROM (
    SELECT
        (SELECT value FROM system.asynchronous_metrics WHERE metric = 'CGroupMaxCPU') AS cgroup_cpu,
        (SELECT value FROM system.asynchronous_metrics WHERE metric = 'OSProcessorCount') AS os_cpu
) AS hw`
	case "duckdb":
		// MotherDuck: detect from DuckDB settings
		cfg := connector.Config
		if cfg == nil || cfg.Fields == nil {
			return ""
		}
		var pathVal, tokenVal string
		if v := cfg.Fields["path"]; v != nil {
			pathVal = v.GetStringValue()
		}
		if v := cfg.Fields["token"]; v != nil {
			tokenVal = v.GetStringValue()
		}
		isMotherDuck := strings.HasPrefix(pathVal, "md:") || tokenVal != ""
		if isMotherDuck {
			return `SELECT MAX(CASE WHEN name = 'threads' THEN CAST(value AS INT) END) AS vcpus FROM duckdb_settings() WHERE name = 'threads'`
		}
		return ""
	default:
		return ""
	}
}

// syncClusterSlots runs a SQL query against the OLAP connector to detect the cluster's vCPU count
// and persists it as cluster_slots in the projects table.
func (w *BillingReporterWorker) syncClusterSlots(ctx context.Context, proj *database.Project, depl *database.Deployment, rt *client.Client, connector *runtimev1.Connector) {
	sql := clusterDetectionSQL(connector)
	if sql == "" {
		return
	}

	qc := rt.QueryServiceClient()
	qResp, err := qc.Query(ctx, &runtimev1.QueryRequest{
		InstanceId: depl.RuntimeInstanceID,
		Connector:  connector.Name,
		Sql:        sql,
		Priority:   -1,
	})
	if err != nil {
		w.logger.Debug("cluster slots sync: query failed", zap.String("project_id", proj.ID), zap.Error(err))
		return
	}

	if len(qResp.Data) == 0 {
		return
	}

	row := qResp.Data[0]
	vcpusField := row.Fields["vcpus"]
	if vcpusField == nil {
		return
	}
	vcpus := int64(vcpusField.GetNumberValue())
	if vcpus <= 0 {
		return
	}

	// Only update if the value has changed
	if proj.ClusterSlots != nil && *proj.ClusterSlots == vcpus {
		return
	}

	if err := w.admin.DB.UpdateProjectClusterSlots(ctx, proj.ID, vcpus); err != nil {
		w.logger.Warn("cluster slots sync: failed to update", zap.String("project_id", proj.ID), zap.Int64("vcpus", vcpus), zap.Error(err))
	} else {
		w.logger.Info("cluster slots sync: updated", zap.String("project_id", proj.ID), zap.Int64("vcpus", vcpus))
	}
}
