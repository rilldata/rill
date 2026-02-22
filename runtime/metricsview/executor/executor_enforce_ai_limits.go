package executor

import (
	"fmt"

	"github.com/rilldata/rill/runtime/metricsview"
)

// enforceAILimits checks if the query adheres to instance specific AI limits. Should be called after time ranges have been resolved for the query.
func (e *Executor) enforceAILimits(qry *metricsview.Query) error {
	if !e.aiQuery {
		return nil
	}

	if qry.TimeRange.IsZero() {
		return fmt.Errorf("a valid time range is required for AI queries")
	}

	if e.instanceCfg.AIMaxTimeRangeDays <= 0 {
		return nil
	}

	days := qry.TimeRange.End.Sub(qry.TimeRange.Start).Hours() / 24
	if days > float64(e.instanceCfg.AIMaxTimeRangeDays) {
		return fmt.Errorf("time range for AI queries cannot exceed %d days", e.instanceCfg.AIMaxTimeRangeDays)
	}

	return nil
}
