package executor

import (
	"fmt"

	"github.com/rilldata/rill/runtime/metricsview"
)

// enforceQueryLimits checks that the query adheres to any limits specified in the QueryLimit. This should be called after time_range is resolved.
func (e *Executor) enforceQueryLimits(qry *metricsview.Query) error {
	if qry.QueryLimit.RequireTimeRange && qry.TimeRange.IsZero() {
		return fmt.Errorf("a valid time_range should be specified for the query")
	}

	if qry.QueryLimit.MaxTimeRangeDays <= 0 {
		return nil
	}

	days := qry.TimeRange.End.Sub(qry.TimeRange.Start).Hours() / 24
	if days > float64(qry.QueryLimit.MaxTimeRangeDays) {
		return fmt.Errorf("time range for query cannot exceed %d days, this can be adjusted using rill.ai.max_time_range_days env var", qry.QueryLimit.MaxTimeRangeDays)
	}

	return nil
}
