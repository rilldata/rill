package metricsview

// rewriteComparisonJoins rewrites the AST to use a LEFT or RIGHT join instead of a FULL join when safe given the query's sorting.
func (e *Executor) rewriteComparisonJoins(ast *AST) {
	if !e.instanceCfg.MetricsApproximateComparisons {
		return
	}

	_ = e.rewriteComparisonJoinsWalk(ast, ast.Root)
	return
}

func (e *Executor) rewriteComparisonJoinsWalk(a *AST, n *SelectNode) bool {
	// If n is a comparison node, rewrite it
	var rewrote bool
	if n.JoinComparisonSelect != nil {
		rewrote = e.rewriteComparisonNode(a, n)
	}

	// Recursively walk the base select.
	// NOTE: Probably doesn't matter, but should we walk the left join and comparison sub-selects?
	if n.FromSelect != nil {
		rewroteNested := e.rewriteComparisonJoinsWalk(a, n.FromSelect)
		rewrote = rewrote || rewroteNested
	}

	// If any node was rewritten, all parent nodes need to clear their offset (since it must only be applied once).
	if rewrote {
		n.Offset = nil
	}

	return rewrote
}

func (e *Executor) rewriteComparisonNode(a *AST, n *SelectNode) bool {
	// Can only rewrite when sorting by exactly one field.
	if len(a.Root.OrderBy) != 1 {
		return false
	}
	sortField := a.Root.OrderBy[0]

	// We also support doing approximate comparisons to support further optimizations when the accuracy of comparisons sorting is not critical.
	approx := e.instanceCfg.MetricsApproximateComparisons

	// Find out what we're sorting by
	var sortDim, sortBase, sortComparison, sortDelta bool
	if len(a.Root.OrderBy) > 0 {
		// Check if it's a measure
		for _, qm := range a.query.Measures {
			if qm.Name != sortField.Name {
				continue
			}

			if qm.Compute != nil && qm.Compute.ComparisonValue != nil {
				sortComparison = true
			} else if qm.Compute != nil && qm.Compute.ComparisonDelta != nil {
				sortDelta = true
			} else if qm.Compute != nil && qm.Compute.ComparisonRatio != nil {
				sortDelta = true
			} else {
				sortBase = true
			}

			break
		}

		if !sortBase && !sortComparison && !sortDelta {
			// It wasn't a measure. Check if it's a dimension.
			for _, qd := range a.query.Dimensions {
				if qd.Name == sortField.Name {
					sortDim = true
					break
				}
			}
		}
	}

	if sortBase {
		// We're sorting by a measure in FromSelect. We can do a LEFT JOIN and push down the order/limit to it.
		n.JoinComparisonType = JoinTypeLeft
		n.FromSelect.OrderBy = a.Root.OrderBy
		n.FromSelect.Limit = a.Root.Limit
		n.FromSelect.Offset = a.Root.Offset
	} else if sortComparison {
		// We're sorting by a measure in JoinComparisonSelect. We can do a RIGHT JOIN and push down the order/limit to it.
		n.JoinComparisonType = JoinTypeRight
		n.JoinComparisonSelect.OrderBy = a.Root.OrderBy
		n.JoinComparisonSelect.Limit = a.Root.Limit
		n.JoinComparisonSelect.Offset = a.Root.Offset
	} else if approx && sortDim {
		// We're sorting by a dimension.
		// For correct results, we need to do a FULL JOIN since a dimension value may only be present in one of the base or comparison sub-queries.
		// But for approximate results, we can do a LEFT JOIN that only returns values present in the base query.
		n.JoinComparisonType = JoinTypeLeft
		n.FromSelect.OrderBy = a.Root.OrderBy
		n.FromSelect.Limit = a.Root.Limit
		n.FromSelect.Offset = a.Root.Offset
	}
	// TODO: Good ideas for approx delta sorts?

	return true
}
