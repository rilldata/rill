package executor

import (
	"errors"

	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteQueryExactMeasures resolves which measures the query computes with their exact_expression and stores the set on the query (see metricsview.Query.ExactMeasures).
// It runs before any other rewrite, so that the AST and every query derived from this one (the totals of percent-of-total measures, the first phase of a two-phase comparison, the inner query of a Druid TopN, filter subqueries)
// copy the same set instead of resolving their own from the measures they keep.
// A query that already carries a set, such as the base query of a two-phase comparison, keeps it.
// It returns an error if the query computes any measure exactly while the instance has disabled exact measures (rill.metrics.exact_measures=false).
// It errors rather than ignoring the flags, so clients never present approximate values as exact.
func (e *Executor) rewriteQueryExactMeasures(qry *metricsview.Query) error {
	if qry.ExactMeasures == nil {
		qry.ExactMeasures = metricsview.ExactMeasureSet(e.metricsView, qry)
	}
	if len(qry.ExactMeasures) > 0 && !e.instanceCfg.MetricsExactMeasures {
		return errors.New("exact measures are disabled for this instance (rill.metrics.exact_measures=false)")
	}
	return nil
}
