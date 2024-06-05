package metricsview

import "fmt"

// rewriteQueryLimit rewrites the query limit to enforce system limits.
// For unlimited queries, it adds a limit just above the system limit. The result reader should then error if the cap is exceeded.
func (e *Executor) rewriteQueryLimit(qry *Query, export bool) error {
	limitCap := e.queryLimitCap(export)

	// No magic if there is no cap
	if limitCap == 0 {
		return nil
	}

	// If no limit on the query, set the limit to +1 of the cap. The result reader should then error if the cap is exceeded.
	if qry.Limit == nil {
		limitCap++
		qry.Limit = &limitCap
		return nil
	}

	// If the limit exceeds the cap, we error immediately
	if *qry.Limit > limitCap {
		return fmt.Errorf("query limit of %d rows exceeds the system cap of %d rows", *qry.Limit, limitCap)
	}

	return nil
}

// queryLimitCap returns the system limit for the given query type.
func (e *Executor) queryLimitCap(export bool) int64 {
	if export {
		return e.instanceCfg.DownloadRowLimit
	}
	return e.instanceCfg.InteractiveSQLRowLimit
}
