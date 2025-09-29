package executor

import (
	"fmt"

	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteQueryEnforceCaps rewrites the query to enforce system limits.
// It works by adding a limit just above the system cap.
// It returns a cap that the result reader should check and error on if exceeded.
// If it returns 0, no cap should be enforced.
func (e *Executor) rewriteQueryEnforceCaps(qry *metricsview.Query) (int64, error) {
	limitCap := e.instanceCfg.InteractiveSQLRowLimit

	// No magic if there is no cap
	if limitCap == 0 {
		return 0, nil
	}

	// If no limit on the query, set the limit to +1 of the cap. The result reader should then error if the cap is exceeded.
	if qry.Limit == nil {
		tmp := limitCap + 1
		qry.Limit = &tmp
		return limitCap, nil
	}

	// If the limit exceeds the cap, we error immediately
	if *qry.Limit > limitCap {
		return 0, fmt.Errorf("query limit of %d rows exceeds the system cap of %d rows", *qry.Limit, limitCap)
	}

	return 0, nil
}
