package executor

import (
	"errors"

	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteQueryExactMeasures resolves which measures the query computes with their exact_expression and writes the result back into the query's Exact flags (see metricsview.FlagExactMeasures).
// It runs before any other rewrite, so that the AST and every query derived from this one (the totals of percent-of-total measures, the first phase of a two-phase comparison, the inner query of a Druid TopN, filter subqueries)
// resolve the same precision from the measures they copy, without the set being passed along.
// It returns an error if the query computes any measure exactly while the instance has disabled exact measures (rill.metrics.exact_measures=false).
// It errors rather than ignoring the flags, so clients never present approximate values as exact.
func (e *Executor) rewriteQueryExactMeasures(qry *metricsview.Query) error {
	if !metricsview.FlagExactMeasures(e.metricsView, qry) {
		return nil
	}
	if !e.instanceCfg.MetricsExactMeasures {
		return errors.New("exact measures are disabled for this instance (rill.metrics.exact_measures=false)")
	}
	return nil
}
