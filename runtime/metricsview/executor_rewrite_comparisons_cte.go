package metricsview

import "fmt"

// rewriteComparisonsCTE extracts out the base or comparison query into a CTE depending on the sort field.
// This is done to enable more efficient query execution. The CTE ise used as FromSelect or JoinComparisonSelect in the AST and also used in corresponding other join query to filter out dimension values that are not present in the base query.
func (e *Executor) rewriteComparisonsCTE(ast *AST) {
	if !e.instanceCfg.MetricsComparisonsCTE {
		return
	}
	if ast.Root.JoinComparisonSelect != nil {
		_ = e.rewriteComparisonsCTENode(ast, ast.Root)
	}
}

func (e *Executor) rewriteComparisonsCTENode(a *AST, n *SelectNode) bool {
	// Can only rewrite when sorting by exactly one field.
	if len(a.Root.OrderBy) != 1 {
		return false
	}
	sortField := a.Root.OrderBy[0]

	// Find out what we're sorting by
	var sortDim, sortBase, sortComparison, sortDelta bool
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

	// if sorting by a dimension or base measure then make the FromSelect a CTE and use it in the comparison query
	if sortDim || sortBase {
		// make FromSelect a CTE and change FromSelect to use the CTE
		cte := n.FromSelect
		n.CTEs = append(n.CTEs, cte)
		n.FromSelect = nil
		n.FromTable = &cte.Alias

		// now change the JoinComparisonSelect WHERE clause to use selected dim values from CTE
		for _, dim := range cte.DimFields {
			var dimExpr string
			if dim.Expr != "" {
				dimExpr = dim.Expr
			} else {
				dimExpr = dim.Name
			}
			n.JoinComparisonSelect.Where = n.JoinComparisonSelect.Where.and(fmt.Sprintf("%[1]s IS NULL OR %[1]s IN (SELECT %[2]q.%[3]q FROM %[2]q)", dimExpr, cte.Alias, dim.Name), nil)
		}
	} else if sortComparison {
		// make JoinComparisonSelect a CTE and change JoinComparisonSelect to use the CTE
		cte := n.JoinComparisonSelect
		n.CTEs = append(n.CTEs, cte)
		n.JoinComparisonSelect = nil
		n.JoinComparisonTable = &cte.Alias

		// now change the FromSelect WHERE clause to use selected dim values from CTE
		for _, dim := range cte.DimFields {
			var dimExpr string
			if dim.Expr != "" {
				dimExpr = dim.Expr
			} else {
				dimExpr = dim.Name
			}
			n.FromSelect.Where = n.FromSelect.Where.and(fmt.Sprintf("%[1]s IS NULL OR %[1]s IN (SELECT %[2]q.%[3]q FROM %[2]q)", dimExpr, cte.Alias, dim.Name), nil)
		}
	} else {
		return false
	}

	return true
}
