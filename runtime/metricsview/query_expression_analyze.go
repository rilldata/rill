package metricsview

import "slices"

// AnalyzeExpressionFields analyzes a metrics expression and returns the field names referenced in it.
func AnalyzeExpressionFields(e *Expression) []string {
	fields := make(map[string]bool)
	analyzeExpressionFieldsInner(e, fields)
	res := make([]string, 0, len(fields))
	for k := range fields {
		res = append(res, k)
	}
	slices.Sort(res)
	return res
}

func analyzeExpressionFieldsInner(e *Expression, res map[string]bool) {
	if e == nil {
		return
	}

	if e.Name != "" {
		res[e.Name] = true
	}

	if e.Condition != nil {
		for _, expr := range e.Condition.Expressions {
			analyzeExpressionFieldsInner(expr, res)
		}
	}

	if e.Subquery != nil {
		analyzeExpressionFieldsInner(e.Subquery.Where, res)
		analyzeExpressionFieldsInner(e.Subquery.Having, res)
		for _, m := range e.Subquery.Measures {
			res[m.Name] = true
		}
		res[e.Subquery.Dimension.Name] = true
	}
}
