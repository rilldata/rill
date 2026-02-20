package executor

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteQueryArrayUnnest rewrites IN/NIN operators on unnest dimensions to use
// array-contains functions (list_has_any for DuckDB, hasAny for ClickHouse).
func (e *Executor) rewriteQueryArrayUnnest(qry *metricsview.Query) {
	dialect := e.olap.Dialect()
	if dialect != drivers.DialectDuckDB && dialect != drivers.DialectClickHouse {
		return
	}

	rewriteExprArrayUnnest(qry.Where, e.metricsView)
	rewriteExprArrayUnnest(qry.Having, e.metricsView)
}

func rewriteExprArrayUnnest(expr *metricsview.Expression, mv *runtimev1.MetricsViewSpec) {
	if expr == nil {
		return
	}

	// Recurse into subquery's Where and Having
	if expr.Subquery != nil {
		rewriteExprArrayUnnest(expr.Subquery.Where, mv)
		rewriteExprArrayUnnest(expr.Subquery.Having, mv)
		return
	}

	if expr.Condition == nil {
		return
	}

	cond := expr.Condition

	// Recurse into sub-expressions for AND/OR
	if cond.Operator == metricsview.OperatorOr || cond.Operator == metricsview.OperatorAnd {
		for _, e := range cond.Expressions {
			rewriteExprArrayUnnest(e, mv)
		}
		return
	}

	// Check for IN/NIN with a dimension name on the left
	if cond.Operator != metricsview.OperatorIn && cond.Operator != metricsview.OperatorNin {
		return
	}

	if len(cond.Expressions) == 0 || cond.Expressions[0].Name == "" {
		return
	}

	// Look up the dimension to check if it's an unnest dimension
	if !isUnnestDimension(cond.Expressions[0].Name, mv) {
		return
	}

	// Rewrite the operator
	if cond.Operator == metricsview.OperatorIn {
		cond.Operator = metricsview.OperatorArrayContains
	} else {
		cond.Operator = metricsview.OperatorArrayNotContains
	}
}

func isUnnestDimension(name string, mv *runtimev1.MetricsViewSpec) bool {
	for _, dim := range mv.Dimensions {
		if dim.Name == name {
			return dim.Unnest
		}
	}
	return false
}
