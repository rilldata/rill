package river

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// orgUsageMetric computes a billable usage value for an organization from the admin database.
// Add entries here to report additional admin-derived billable metrics (the runtime-derived metrics
// are reported separately by the billing reporter).
type orgUsageMetric struct {
	name    string
	collect func(ctx context.Context, adm *admin.Service, orgID string) (float64, error)
}

var orgUsageMetrics = []orgUsageMetric{
	{
		name: "seats",
		collect: func(ctx context.Context, adm *admin.Service, orgID string) (float64, error) {
			n, err := adm.DB.CountOrganizationMemberUsers(ctx, orgID, "", "")
			if err != nil {
				return 0, err
			}
			return float64(n), nil
		},
	},
}

type UsageReporterArgs struct{}

func (UsageReporterArgs) Kind() string { return "usage_reporter" }

type UsageReporterWorker struct {
	river.WorkerDefaults[UsageReporterArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work reports admin-database-derived billable usage metrics (currently seats) for every org with billing.
// Each metric is reported as a gauge for the current reporting period; the biller aggregates over the billing period.
// It is best-effort: a failure for one org or metric is logged and skipped so the rest still get reported.
func (w *UsageReporterWorker) Work(ctx context.Context, job *river.Job[UsageReporterArgs]) error {
	grain := w.admin.Biller.GetReportingGranularity()
	var granularity time.Duration
	switch grain {
	case billing.UsageReportingGranularityHour:
		granularity = time.Hour
	case billing.UsageReportingGranularityNone:
		w.logger.Debug("skipping usage reporting: no reporting granularity configured")
		return nil
	default:
		return fmt.Errorf("unsupported reporting granularity: %s", grain)
	}

	startTime := time.Now().UTC().Truncate(granularity)
	endTime := startTime.Add(granularity)

	orgIDs, err := w.admin.DB.FindOrganizationIDsWithBilling(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch orgs with billing: %w", err)
	}

	for _, orgID := range orgIDs {
		org, err := w.admin.DB.FindOrganization(ctx, orgID)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				continue
			}
			w.logger.Error("usage reporter: failed to find org", zap.String("org_id", orgID), zap.Error(err))
			continue
		}
		if org.BillingCustomerID == "" {
			continue
		}

		var usage []*billing.Usage
		for _, m := range orgUsageMetrics {
			val, err := m.collect(ctx, w.admin, orgID)
			if err != nil {
				w.logger.Error("usage reporter: failed to collect metric", zap.String("metric", m.name), zap.String("org_id", orgID), zap.Error(err))
				continue
			}
			usage = append(usage, &billing.Usage{
				CustomerID:     org.BillingCustomerID,
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
			w.logger.Error("usage reporter: failed to report usage", zap.String("org_id", orgID), zap.Error(err))
			continue
		}
	}

	return nil
}
