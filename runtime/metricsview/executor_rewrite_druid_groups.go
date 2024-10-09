package metricsview

import (
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

// rewriteDruidGroups rewrites the AST to always have GROUP BY in every SELECT node for Druid queries.
// This is needed to tap into code paths that ensure correct ordering of derived measures.
func (e *Executor) rewriteDruidGroups(ast *AST) error {
	if ast.dialect != drivers.DialectDruid {
		return nil
	}

	return e.rewriteDruidGroupsWalk(ast.Root)
}

func (e *Executor) rewriteDruidGroupsWalk(n *SelectNode) error {
	// Skip if already grouped
	if n.Group {
		return nil
	}

	// Recurse over FromSelect as it may contain sub queries with joins
	if n.FromSelect != nil {
		err := e.rewriteDruidGroupsWalk(n.FromSelect)
		if err != nil {
			return err
		}
	}

	// Skip if there is no sub query with JOIN
	// Druid requires GROUP BY and ANY_VALUE aggregate for measure if there is a subquery with JOIN
	if n.SpineSelect == nil && len(n.LeftJoinSelects) == 0 && n.JoinComparisonSelect == nil {
		return nil
	}

	// Rewrite
	n.Group = true
	for i, f := range n.MeasureFields {
		f.Expr = fmt.Sprintf("ANY_VALUE(%s)", f.Expr)
		n.MeasureFields[i] = f
	}

	// Recurse
	if n.SpineSelect != nil {
		err := e.rewriteDruidGroupsWalk(n.SpineSelect)
		if err != nil {
			return err
		}
	}
	for _, ljs := range n.LeftJoinSelects {
		err := e.rewriteDruidGroupsWalk(ljs)
		if err != nil {
			return err
		}
	}
	if n.JoinComparisonSelect != nil {
		err := e.rewriteDruidGroupsWalk(n.JoinComparisonSelect)
		if err != nil {
			return err
		}
	}

	return nil
}
