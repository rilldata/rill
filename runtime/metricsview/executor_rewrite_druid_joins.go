package metricsview

import (
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

// rewriteDruidJoins rewrites the AST to avoid large joins for Druid queries because Druid requires that one part of a JOIN fits in memory.
// It may lead to some joins becoming approximate.
// It does not rewrite limits that are already set.
func (e *Executor) rewriteDruidJoins(ast *AST) error {
	if ast.dialect != drivers.DialectDruid {
		return nil
	}

	if !e.instanceCfg.MetricsApproximateComparisons {
		return fmt.Errorf("approximate comparisons must be enabled for Druid queries")
	}

	limit := int64(100_000) // Druid's limit
	if ast.Root.Limit != nil {
		limit = *ast.Root.Limit * 2 // Double the provided limit
	}

	return e.rewriteDruidJoinsWalk(ast.Root, &limit)
}

func (e *Executor) rewriteDruidJoinsWalk(n *SelectNode, limit *int64) error {
	// Skip if doesn't have subqueries
	if n.FromSelect == nil {
		return nil
	}

	// Skip and recurse if it doesn't have JOINs
	if n.LeftJoinSelects == nil && n.JoinComparisonSelect == nil {
		return e.rewriteDruidJoinsWalk(n.FromSelect, limit)
	}

	// Apply limits and recurse
	applyLimit(n.FromSelect, limit)
	err := e.rewriteDruidJoinsWalk(n.FromSelect, limit)
	if err != nil {
		return err
	}

	for _, ljs := range n.LeftJoinSelects {
		applyLimit(ljs, limit)

		err := e.rewriteDruidJoinsWalk(ljs, limit)
		if err != nil {
			return err
		}
	}
	if n.JoinComparisonSelect != nil {
		applyLimit(n.JoinComparisonSelect, limit)

		err := e.rewriteDruidJoinsWalk(n.JoinComparisonSelect, limit)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyLimit(n *SelectNode, limit *int64) {
	if n.Limit == nil {
		n.Limit = limit
	}
}
