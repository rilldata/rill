package executor

import (
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/duration"
)

// enforceQueryLimits checks that the query adheres to any limits specified in the QueryLimits or on the metrics view spec.
// This should be called after time_range is resolved.
func (e *Executor) enforceQueryLimits(qry *metricsview.Query) error {
	if qry.QueryLimits != nil && qry.QueryLimits.RequireTimeRange && (qry.TimeRange == nil || qry.TimeRange.IsZero()) {
		return fmt.Errorf("a valid time_range should be specified for the query")
	}

	if qry.TimeRange == nil || qry.TimeRange.IsZero() {
		return nil
	}

	maxRange, errMsg := e.maxTimeRange(qry)
	if maxRange <= 0 {
		return nil
	}

	if err := checkTimeRangeWithinCap(qry.TimeRange, maxRange, errMsg); err != nil {
		return err
	}
	if qry.ComparisonTimeRange != nil && !qry.ComparisonTimeRange.IsZero() {
		if err := checkTimeRangeWithinCap(qry.ComparisonTimeRange, maxRange, errMsg); err != nil {
			return err
		}
	}
	return nil
}

// maxTimeRange returns the effective cap on the query's time range and a pre-formatted error message
// describing where the cap was configured. Returns 0 if no cap applies.
//
// An explicit caller-provided QueryLimits.MaxTimeRangeDays takes precedence over the metrics view's
// max_query_time_range spec property — this matters for the AI tool path which tightens the cap via
// the rill.ai.max_time_range_days env var.
func (e *Executor) maxTimeRange(qry *metricsview.Query) (time.Duration, string) {
	if qry.QueryLimits != nil && qry.QueryLimits.MaxTimeRangeDays > 0 {
		days := qry.QueryLimits.MaxTimeRangeDays
		return time.Duration(days) * 24 * time.Hour,
			fmt.Sprintf("time range for query cannot exceed %d days, configured via the rill.ai.max_time_range_days env var", days)
	}
	if e.metricsView != nil && e.metricsView.MaxQueryTimeRange != "" {
		d, err := duration.ParseISO8601(e.metricsView.MaxQueryTimeRange)
		if err != nil {
			return 0, ""
		}
		native, ok := d.EstimateNative()
		if !ok || native <= 0 {
			return 0, ""
		}
		return native,
			fmt.Sprintf("time range for query cannot exceed %s, configured via the metrics view's max_query_time_range property", e.metricsView.MaxQueryTimeRange)
	}
	return 0, ""
}

func checkTimeRangeWithinCap(tr *metricsview.TimeRange, maxRange time.Duration, errMsg string) error {
	if tr.End.Sub(tr.Start) > maxRange {
		return errors.New(errMsg)
	}
	return nil
}
