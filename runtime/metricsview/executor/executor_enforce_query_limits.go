package executor

import (
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/metricsview"
)

// enforceQueryLimits checks that the query adheres to any limits specified in the QueryLimits or on the metrics view spec.
// This should be called after time_range is resolved.
func (e *Executor) enforceQueryLimits(qry *metricsview.Query) error {
	if qry.QueryLimits != nil && qry.QueryLimits.RequireTimeRange && (qry.TimeRange == nil || qry.TimeRange.IsZero()) {
		return fmt.Errorf("a valid time_range should be specified for the query")
	}

	if err := e.enforceMaxTimeRange(qry, qry.TimeRange); err != nil {
		return err
	}
	return e.enforceMaxTimeRange(qry, qry.ComparisonTimeRange)
}

// enforceMaxTimeRange returns nil if tr fits within the configured cap, else an error.
// Both QueryLimits.MaxTimeRangeDays (AI env var path) and the metrics view's max_query_time_range are checked;
// the more restrictive of the two wins.
func (e *Executor) enforceMaxTimeRange(qry *metricsview.Query, tr *metricsview.TimeRange) error {
	if tr == nil || tr.IsZero() {
		return nil
	}

	var maxDur time.Duration
	var capErr error

	if qry.QueryLimits != nil && qry.QueryLimits.MaxTimeRangeDays > 0 {
		d := time.Duration(qry.QueryLimits.MaxTimeRangeDays) * 24 * time.Hour
		maxDur = d
		capErr = fmt.Errorf("time range for query cannot exceed %d days, configured via the rill.ai.max_time_range_days env var", qry.QueryLimits.MaxTimeRangeDays)
	}

	if e.metricsView != nil && e.metricsView.MaxQueryTimeRange != "" {
		d := metricsview.ResolveMaxQueryTimeRange(e.metricsView.MaxQueryTimeRange, time.Now())
		if d > 0 && (maxDur == 0 || d < maxDur) {
			maxDur = d
			capErr = fmt.Errorf("time range for query cannot exceed %s, configured via the metrics view's max_query_time_range property", e.metricsView.MaxQueryTimeRange)
		}
	}

	if maxDur <= 0 {
		return nil
	}
	if tr.End.Sub(tr.Start) > maxDur {
		return capErr
	}
	return nil
}
