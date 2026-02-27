package executor

import (
	"fmt"

	"github.com/rilldata/rill/runtime/metricsview"
)

// enforceQueryLimits checks that the query adheres to any limits specified in the QueryLimits. This should be called after time_range is resolved.
func (e *Executor) enforceQueryLimits(qry *metricsview.Query) error {
	if qry.QueryLimits == nil {
		return nil
	}

	if qry.QueryLimits.RequireTimeRange && (qry.TimeRange == nil || qry.TimeRange.IsZero()) {
		return fmt.Errorf("a valid time_range should be specified for the query")
	}

	// if require_time_range not set and time range is not specified, we skip the max time range check
	if qry.QueryLimits.MaxTimeRangeDays <= 0 || qry.TimeRange == nil || qry.TimeRange.IsZero() {
		return nil
	}

	days := qry.TimeRange.End.Sub(qry.TimeRange.Start).Hours() / 24
	if days > float64(qry.QueryLimits.MaxTimeRangeDays) {
		return fmt.Errorf("time range for query cannot exceed %d days, this can be adjusted using rill.ai.max_time_range_days env var", qry.QueryLimits.MaxTimeRangeDays)
	}

	return nil
}
