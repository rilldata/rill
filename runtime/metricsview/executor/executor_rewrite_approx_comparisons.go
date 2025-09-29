package executor

import (
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteApproxComparisons rewrites the AST to use a LEFT or RIGHT join instead of a FULL joins for comparisons,
// which enables more efficient query execution at the cost of some accuracy.
// ---- CTE rewrite ---- //
// Extracts out the base or comparison query into a CTE depending on the sort field.
// This is done to prevent running a group by query on comparison time range without a limit which can fail in some olap engines if dim cardinality is very high by adding filter in the join query to select only dimension values present in the CTE.
// This does cause CTE to be scanned twice but at least query will not fail.
func (e *Executor) rewriteApproxComparisons(ast *metricsview.AST, isMultiPhase bool) {
	if !e.instanceCfg.MetricsApproximateComparisons {
		return
	}

	_ = e.rewriteApproxComparisonsWalk(ast, ast.Root, isMultiPhase)
}

func (e *Executor) rewriteApproxComparisonsWalk(a *metricsview.AST, n *metricsview.SelectNode, isMultiPhase bool) bool {
	// If n is a comparison node, rewrite it
	var rewrote bool
	if n.JoinComparisonSelect != nil {
		rewrote = e.rewriteApproxComparisonNode(a, n, isMultiPhase)
	}

	// Recursively walk the base select.
	// NOTE: Probably doesn't matter, but should we walk the left join and comparison sub-selects?
	if n.FromSelect != nil {
		rewroteNested := e.rewriteApproxComparisonsWalk(a, n.FromSelect, isMultiPhase)
		rewrote = rewrote || rewroteNested
	}

	// If any node was rewritten, all parent nodes need to clear their offset (since it must only be applied once).
	if rewrote {
		n.Offset = nil
	}

	return rewrote
}

func (e *Executor) rewriteApproxComparisonNode(a *metricsview.AST, n *metricsview.SelectNode, isMultiPhase bool) bool {
	// Can only rewrite when sorting by exactly one field.
	if len(a.Root.OrderBy) != 1 {
		return false
	}
	sortField := a.Root.OrderBy[0]

	cteRewrite := e.instanceCfg.MetricsApproximateComparisonsCTE && !isMultiPhase
	if e.olap.Dialect() == drivers.DialectDruid && cteRewrite {
		// if there are unnests in the query, we can't rewrite the query for Druid
		// it fails with join on cte having multi value dimension, issue - https://github.com/apache/druid/issues/16896
		for _, dim := range n.FromSelect.DimFields {
			if dim.Unnest {
				return false
			}
		}
	}

	// Find out what we're sorting by
	var sortDim, sortBase, sortComparison, sortDelta bool
	var sortUnderlyingMeasure string
	if len(a.Root.OrderBy) > 0 {
		// Check if it's a measure
		for _, qm := range a.Query.Measures {
			if qm.Name != sortField.Name {
				continue
			}

			if qm.Compute != nil && qm.Compute.ComparisonValue != nil {
				sortComparison = true
				sortUnderlyingMeasure = qm.Compute.ComparisonValue.Measure
			} else if qm.Compute != nil && qm.Compute.ComparisonDelta != nil {
				sortDelta = true
				sortUnderlyingMeasure = qm.Compute.ComparisonDelta.Measure
			} else if qm.Compute != nil && qm.Compute.ComparisonRatio != nil {
				sortDelta = true
				sortUnderlyingMeasure = qm.Compute.ComparisonRatio.Measure
			} else {
				sortBase = true
				sortUnderlyingMeasure = qm.Name
			}

			break
		}

		if !sortBase && !sortComparison && !sortDelta {
			// It wasn't a measure. Check if it's a dimension.
			for _, qd := range a.Query.Dimensions {
				if qd.Name == sortField.Name {
					sortDim = true
					break
				}
			}
		}
	}

	// If sorting by a computed measure, we need to use the underlying measure name when pushing the order into the sub-select.
	if sortUnderlyingMeasure != "" {
		sortField.Name = sortUnderlyingMeasure
	}
	order := []metricsview.OrderFieldNode{sortField}

	// Note: All these cases are approximations in different ways.
	if sortBase {
		// We're sorting by a measure in FromSelect. We do a LEFT JOIN and push down the order/limit to it.
		// This should remain correct when the limit is lower than the number of rows in the base query.
		// The approximate part here is when the base query returns fewer rows than the limit, then dimension values that are only in the comparison query will be missing.
		n.JoinComparisonType = metricsview.JoinTypeLeft
		n.FromSelect.OrderBy = order
		n.FromSelect.Limit = a.Root.Limit
		n.FromSelect.Offset = a.Root.Offset

		if cteRewrite {
			// rewrite base query as CTE and use results from CTE in the comparison query
			// make FromSelect a CTE
			a.ConvertToCTE(n.FromSelect)

			// now change the JoinComparisonSelect WHERE clause to use selected dim values from CTE
			for _, dim := range n.JoinComparisonSelect.DimFields {
				dimName := a.Dialect.EscapeIdentifier(dim.Name)
				dimExpr := "(" + dim.Expr + ")" // wrap in parentheses to handle expressions
				n.JoinComparisonSelect.Where = n.JoinComparisonSelect.Where.And(fmt.Sprintf("%[1]s IS NULL OR %[1]s IN (SELECT %[2]q.%[3]s FROM %[2]q)", dimExpr, n.FromSelect.Alias, dimName), nil)
			}
		}
	} else if sortComparison {
		// We're sorting by a measure in JoinComparisonSelect. We can do a RIGHT JOIN and push down the order/limit to it.
		// This should remain correct when the limit is lower than the number of rows in the comparison query.
		// The approximate part here is when the comparison query returns fewer rows than the limit, then dimension values that are only in the base query will be missing.
		n.JoinComparisonType = metricsview.JoinTypeRight
		n.JoinComparisonSelect.OrderBy = order
		n.JoinComparisonSelect.Limit = a.Root.Limit
		n.JoinComparisonSelect.Offset = a.Root.Offset

		if cteRewrite {
			// rewrite comparison query as CTE and use results from CTE in the base query
			// make JoinComparisonSelect a CTE
			a.ConvertToCTE(n.JoinComparisonSelect)

			// now change the FromSelect WHERE clause to use selected dim values from CTE
			for _, dim := range n.FromSelect.DimFields {
				dimName := a.Dialect.EscapeIdentifier(dim.Name)
				dimExpr := "(" + dim.Expr + ")" // wrap in parentheses to handle expressions
				n.FromSelect.Where = n.FromSelect.Where.And(fmt.Sprintf("%[1]s IS NULL OR %[1]s IN (SELECT %[2]q.%[3]s FROM %[2]q)", dimExpr, n.JoinComparisonSelect.Alias, dimName), nil)
			}
		}
	} else if sortDim {
		// We're sorting by a dimension. We do a LEFT JOIN that only returns values present in the base query.
		// The approximate part here is that dimension values only present in the comparison query will be missing.
		n.JoinComparisonType = metricsview.JoinTypeLeft
		n.FromSelect.OrderBy = order
		n.FromSelect.Limit = a.Root.Limit
		n.FromSelect.Offset = a.Root.Offset

		if cteRewrite {
			// rewrite base query as CTE and use results from CTE in the comparison query
			// make FromSelect a CTE
			a.ConvertToCTE(n.FromSelect)

			// now change the JoinComparisonSelect WHERE clause to use selected dim values from CTE
			for _, dim := range n.JoinComparisonSelect.DimFields {
				dimName := a.Dialect.EscapeIdentifier(dim.Name)
				dimExpr := "(" + dim.Expr + ")" // wrap in parentheses to handle expressions
				n.JoinComparisonSelect.Where = n.JoinComparisonSelect.Where.And(fmt.Sprintf("%[1]s IS NULL OR %[1]s IN (SELECT %[2]q.%[3]s FROM %[2]q)", dimExpr, n.FromSelect.Alias, dimName), nil)
			}
		}
	} else if sortDelta {
		return false
	}
	// TODO: Good ideas for approx delta sorts?

	return true
}
