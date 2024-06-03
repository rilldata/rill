package metricsview

// rewriteApproximateComparisons rewrites the AST to avoid large joins for comparison queries.
// The result is faster queries at the cost of
// This consists of avoiding large joins by applying a limit to one side of the comparison join and changing the join type to a LEFT or RIGHT join instead of a FULL join.
func (e *Executor) rewriteApproximateComparisons(ast *AST) error {
	if !e.instanceCfg.MetricsApproximateComparisons {
		return nil
	}

	return e.rewriteApproximateComparisonsWalk(ast, ast.Root)
}

func (e *Executor) rewriteApproximateComparisonsWalk(a *AST, n *SelectNode) error {
	// If n is a comparison node, rewrite it
	if n.JoinComparisonSelect != nil {
		err := e.rewriteApproximateComparisonNode(a, n)
		if err != nil {
			return err
		}
	}

	// Recursively walk the base select.
	// NOTE: Probably doesn't matter, but should we walk the left join and comparison sub-selects?
	if n.FromSelect != nil {
		err := e.rewriteApproximateComparisonsWalk(a, n.FromSelect)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Executor) rewriteApproximateComparisonNode(a *AST, n *SelectNode) error {
	// If the first sort field is a measure in the query
	var sortBase, sortComparison, sortDelta bool
	if len(a.Root.OrderBy) > 1 {
		target := a.Root.OrderBy[0].Name
		for _, qm := range a.query.Measures {
			if qm.Name != target {
				continue
			}

			if qm.Compute == nil {
				sortBase = true
			} else if qm.Compute.ComparisonValue != nil {
				sortComparison = true
			} else if qm.Compute.ComparisonDelta != nil || qm.Compute.ComparisonRatio != nil {
				sortDelta = true
			}

			break
		}
	}

	if !sortBase && !sortComparison && !sortDelta {
		return nil
	}

	if sortBase || sortDelta {
		n.JoinComparisonType = "LEFT OUTER"
	} else if sortComparison {
		n.JoinComparisonType = "RIGHT OUTER"
	}

	approximationLimit := a.Root.Limit
	if approximationLimit != nil && *approximationLimit < 100 && sortDelta {
		tmp := int64(100)
		approximationLimit = &tmp
	}

	n.OrderBy = orderByValidSubset(a.Root.OrderBy, n)
	n.Limit = approximationLimit
	n.Offset = a.Root.Offset

	return nil
}

// orderByValidSubset returns a all or a subset of fields in fs that are valid for ordering in n.
func orderByValidSubset(fs []OrderFieldNode, n *SelectNode) []OrderFieldNode {
	res := make([]OrderFieldNode, 0, len(fs))

	for _, f := range fs {
		found := false
		for _, m := range n.MeasureFields {
			if f.Name == m.Name {
				found = true
				break
			}
		}
		if !found {
			for _, d := range n.DimFields {
				if f.Name == d.Name {
					found = true
					break
				}
			}
		}
		if !found {
			return res
		}
		res = append(res, f)
	}

	return res
}

// joinType := "FULL"
// if !q.Exact {
// 	deltaComparison := q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA ||
// 		q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA

// 	approximationLimit := int(q.Limit)
// 	if q.Limit != 0 && q.Limit < 100 && deltaComparison {
// 		approximationLimit = 100
// 	}

// 	if q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE || deltaComparison {
// 		joinType = "LEFT OUTER"
// 		baseLimitClause = subQueryOrderByClause
// 		if approximationLimit > 0 {
// 			baseLimitClause += fmt.Sprintf(" LIMIT %d", approximationLimit)
// 		}
// 	} else if q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE {
// 		joinType = "RIGHT OUTER"
// 		comparisonLimitClause = subQueryOrderByClause
// 		if approximationLimit > 0 {
// 			comparisonLimitClause += fmt.Sprintf(" LIMIT %d", approximationLimit)
// 		}
// 	}
// }
