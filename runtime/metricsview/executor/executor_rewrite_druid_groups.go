package executor

import (
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteDruidGroups rewrites the AST to always have GROUP BY in every SELECT node for Druid queries.
// This is needed to tap into code paths that ensure correct ordering of derived measures.
func (e *Executor) rewriteDruidGroups(ast *metricsview.AST) error {
	if ast.Dialect != drivers.DialectDruid {
		return nil
	}

	return e.rewriteDruidGroupsWalk(ast.Root)
}

func (e *Executor) rewriteDruidGroupsWalk(n *metricsview.SelectNode) error {
	var hasJoins bool

	// Recurse
	if n.FromSelect != nil {
		err := e.rewriteDruidGroupsWalk(n.FromSelect)
		if err != nil {
			return err
		}
	}
	if n.SpineSelect != nil {
		err := e.rewriteDruidGroupsWalk(n.SpineSelect)
		if err != nil {
			return err
		}
		hasJoins = true
	}
	for _, ljs := range n.LeftJoinSelects {
		err := e.rewriteDruidGroupsWalk(ljs)
		if err != nil {
			return err
		}
		hasJoins = true
	}
	if n.JoinComparisonSelect != nil {
		err := e.rewriteDruidGroupsWalk(n.JoinComparisonSelect)
		if err != nil {
			return err
		}
		hasJoins = true
	}

	if !hasJoins {
		// Skip if there is no sub query with JOIN, Druid requires GROUP BY and ANY_VALUE aggregate for measures if there is a subquery with JOIN
		return nil
	}

	// Rewrite
	n.Group = true
	for i, f := range n.MeasureFields {
		f.Expr = fmt.Sprintf("ANY_VALUE(%s)", f.Expr)
		n.MeasureFields[i] = f
	}

	return nil
}
