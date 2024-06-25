package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"go.uber.org/zap"
)

func (w *Worker) reportUsage(ctx context.Context) error {
	// Get reporting granularity
	var granularity time.Duration
	var sqlGrainIdentifier string
	switch w.admin.Biller.GetReportingGranularity() {
	case billing.UsageReportingGranularityHour:
		granularity = time.Hour
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

	// start reporting from the last reported time
	var startTime time.Time
	if t.IsZero() {
		startTime = time.Now().UTC().Truncate(granularity).Add(-granularity)
	} else {
		startTime = t.UTC()
	}
	// report until end of the last hour
	endTime := time.Now().UTC().Truncate(granularity)

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

	u, err := client.GetUsageMetrics(ctx, startTime, endTime, sqlGrainIdentifier)
	if err != nil {
		return fmt.Errorf("failed to get usage metrics: %w", err)
	}

	if len(u) == 0 {
		w.logger.Warn("skipping usage reporting: no usage data available", zap.Time("start_time", startTime), zap.Time("end_time", endTime))
		return nil
	}

	reportedOrgs := make(map[string]struct{})
	var maxEndTime time.Time
	var usage []*billing.Usage
	for _, m := range u {
		reportedOrgs[m.OrgID] = struct{}{}
		if m.EndTime.After(maxEndTime) {
			maxEndTime = m.EndTime
		}
		usage = append(usage, &billing.Usage{
			CustomerID:     m.OrgID,
			MetricName:     m.MetricName,
			Value:          m.Value,
			ReportingGrain: w.admin.Biller.GetReportingGranularity(),
			StartTime:      m.StartTime,
			EndTime:        m.EndTime,
			Metadata:       map[string]interface{}{"org_id": m.OrgID, "project_id": m.ProjectID},
		})
	}

	err = w.admin.Biller.ReportUsage(ctx, usage)
	if err != nil {
		return fmt.Errorf("failed to report usage: %w", err)
	}

	// update last reporting time to maxEndTime as it is the max processed event time
	err = w.admin.DB.UpdateBillingUsageReportedOn(ctx, maxEndTime)
	if err != nil {
		return fmt.Errorf("failed to update last usage reporting time: %w", err)
	}

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
				w.logger.Warn("failed to validate active projects for org", zap.String("org", org), zap.Error(err))
				continue
			}
			if count > 0 {
				w.logger.Warn("skipping usage reporting for org: no usage data available", zap.String("org", org), zap.Time("start_time", startTime), zap.Time("end_time", endTime))
			}
		}
	}
	return nil
}
