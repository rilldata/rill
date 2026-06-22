package river

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// counterMetrics are billable metrics whose period total is sum(value) rather than max(value).
// Generic usage events ("query", "request_time_ms", "tool_call") carry a "source" attribute (and request_time_ms also
// carries embed/user_id); the metrics project applies the billing-specific filtering and distinct counting downstream.
var counterMetrics = map[string]bool{
	"slot_seconds_spend":  true,
	"query":               true,
	"tool_call":           true,
	"input_tokens":        true,
	"cached_input_tokens": true,
	"output_tokens":       true,
}

// orgUsageMetric computes a billable usage value for an organization from the admin database.
// Add entries here to report additional admin-derived billable metrics; they are reported by the billing
// reporter alongside the runtime-derived metrics for every org that reported usage (see Work).
type orgUsageMetric struct {
	name    string
	collect func(ctx context.Context, adm *admin.Service, orgID string) (float64, error)
}

var orgUsageMetrics = []orgUsageMetric{
	{
		name: "seats",
		collect: func(ctx context.Context, adm *admin.Service, orgID string) (float64, error) {
			// Exclude internal Rill users from billable seat counts.
			n, err := adm.DB.CountOrganizationMemberUsers(ctx, orgID, "", "%@"+billing.InternalEmailDomain, true)
			if err != nil {
				return 0, err
			}
			return float64(n), nil
		},
	},
}

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

	reportedOrgs := make(map[string]string) // org ID -> billing customer ID
	stop := false
	limit := 10000
	afterTime := time.Time{}
	afterOrgID := ""
	afterProjectID := ""
	afterInstanceID := ""
	afterBillingService := ""
	afterEventName := ""

	checkPoint := startTime
	maxEndTime := time.Time{}
	// loop until all the usage data is reported
	for !stop {
		u, err := client.GetUsageMetrics(ctx, startTime, endTime, afterTime, afterOrgID, afterProjectID, afterInstanceID, afterBillingService, afterEventName, sqlGrainIdentifier, limit)
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
			afterInstanceID = u[len(u)-1].InstanceID
			afterBillingService = u[len(u)-1].BillingService
			afterEventName = u[len(u)-1].EventName
		}
		// since the usage data is ordered by start time first and end time is just (start time + grain), we can directly get the max end time
		maxEndTime = u[len(u)-1].EndTime

		var usage []*billing.Usage
		for _, m := range u {
			customerID := m.OrgID
			if m.BillingCustomerID != nil && *m.BillingCustomerID != "" {
				// org might have been deleted or recently created in both cases billing customer id will be null. If billing not initialized for the org, then it will be empty string
				// in all cases just use org ID to report in hope that org ID will be set as billing customer id in the future if not reported values will be ignored
				customerID = *m.BillingCustomerID
			}
			reportedOrgs[m.OrgID] = customerID

			value := m.MaxValue
			if counterMetrics[m.EventName] {
				value = m.SumValue
			}
			usage = append(usage, &billing.Usage{
				CustomerID:     customerID,
				MetricName:     m.EventName,
				Value:          value,
				ReportingGrain: w.admin.Biller.GetReportingGranularity(),
				StartTime:      m.StartTime,
				EndTime:        m.EndTime,
				Metadata:       map[string]interface{}{"org_id": m.OrgID, "project_id": m.ProjectID, "project_name": m.ProjectName, "instance_id": m.InstanceID, "billing_service": m.BillingService},
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

	// Report admin-database-derived billable metrics (e.g. seats) for every org that reported usage in this run.
	// These are current gauges, so we report them for the last completed grain period; the biller aggregates over the billing period.
	w.reportOrgUsageMetrics(ctx, reportedOrgs, endTime.Add(-granularity), endTime)

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

// reportOrgUsageMetrics reports the admin-database-derived billable metrics (orgUsageMetrics) for the given orgs.
// It is best-effort: a failure for one org or metric is logged and skipped so the rest still get reported.
func (w *BillingReporterWorker) reportOrgUsageMetrics(ctx context.Context, reportedOrgs map[string]string, startTime, endTime time.Time) {
	grain := w.admin.Biller.GetReportingGranularity()
	for orgID, customerID := range reportedOrgs {
		var usage []*billing.Usage
		for _, m := range orgUsageMetrics {
			val, err := m.collect(ctx, w.admin, orgID)
			if err != nil {
				w.logger.Error("failed to collect admin usage metric", zap.String("metric", m.name), zap.String("org_id", orgID), zap.Error(err))
				continue
			}
			usage = append(usage, &billing.Usage{
				CustomerID:     customerID,
				MetricName:     m.name,
				Value:          val,
				ReportingGrain: grain,
				StartTime:      startTime,
				EndTime:        endTime,
				Metadata:       map[string]interface{}{"org_id": orgID},
			})
		}
		if len(usage) == 0 {
			continue
		}
		if err := w.admin.Biller.ReportUsage(ctx, usage); err != nil {
			w.logger.Error("failed to report admin usage metrics", zap.String("org_id", orgID), zap.Error(err))
		}
	}
}
