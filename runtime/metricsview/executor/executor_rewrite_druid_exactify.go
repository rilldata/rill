package executor

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteQueryDruidExactify applies an approach to get more accurate measure values for TopN queries in Druid (which are approximate).
// Even though we don't explicitly send TopN requests to Druid, it implicitly executes SQL queries that match certain criteria as TopN queries.
//
// The approach works by executing an inner query that returns the dimension values that we expect in the final result,
// and then adding those dimension values as a filter in the outer query. The specific filter in the second query leads to more accurate measure values being returned.
// For more details on this approach, see: https://druid.apache.org/docs/latest/querying/topnquery/#aliasing.
func (e *Executor) rewriteQueryDruidExactify(ctx context.Context, qry *metricsview.Query) error {
	// Check if it's enabled.
	if !e.instanceCfg.MetricsExactifyDruidTopN {
		return nil
	}

	// Only apply for Druid.
	if e.olap.Dialect() != drivers.DialectDruid {
		return nil
	}

	// Skip if the criteria for a Druid TopN query are not met.
	if len(qry.Dimensions) != 1 || len(qry.Sort) != 1 || qry.Limit == nil || *qry.Limit > 1000 || len(qry.PivotOn) > 0 {
		return nil
	}

	// Construct a new query that will just return the dimension values that we expect in the final result.
	inner := &metricsview.Query{
		MetricsView:         qry.MetricsView,
		Dimensions:          qry.Dimensions,
		Measures:            nil,
		PivotOn:             nil,
		Spine:               nil,
		Sort:                qry.Sort,
		TimeRange:           qry.TimeRange,
		ComparisonTimeRange: qry.ComparisonTimeRange,
		Where:               qry.Where,
		Having:              qry.Having,
		Limit:               qry.Limit,
		Offset:              qry.Offset,
		TimeZone:            qry.TimeZone,
		UseDisplayNames:     false,
		Rows:                false,
	} //exhaustruct:enforce

	// A TopN query can sort by a dimension or a measure.
	// If sorting by a measure, we also include that in the inner query.
	for _, qm := range qry.Measures {
		if qm.Name != inner.Sort[0].Name {
			continue
		}
		inner.Measures = append(inner.Measures, qm)
		break
	}

	// Build an AST for the inner query.
	ast, err := metricsview.NewAST(e.metricsView, e.security, inner, e.olap.Dialect())
	if err != nil {
		return fmt.Errorf("druid exactify: failed to build inner query AST: %w", err)
	}

	// Apply a limited subset of rewrites to the inner query.
	e.rewriteApproxComparisons(ast, false)

	// Generate the SQL for and execute the inner query.
	sql, args, err := ast.SQL()
	if err != nil {
		return err
	}
	res, err := e.olap.Query(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         e.priority,
		ExecutionTimeout: defaultInteractiveTimeout,
	})
	if err != nil {
		return err
	}
	defer res.Close()

	// Extract the dimension values returned from the inner query.
	var vals []any
	for res.Next() {
		var val any
		if len(inner.Measures) == 0 {
			if err := res.Scan(&val); err != nil {
				return fmt.Errorf("druid exactify: failed to scan value: %w", err)
			}
		} else {
			var tmp any // We're ignore the measure value, but we need to scan it.
			if err := res.Scan(&val, &tmp); err != nil {
				return fmt.Errorf("druid exactify: failed to scan value: %w", err)
			}
		}

		vals = append(vals, val)
	}
	err = res.Err()
	if err != nil {
		return err
	}

	// Add the dimensions values as a "<dim> IN (<vals...>)" expression in the outer query's WHERE clause.
	var inExpr *metricsview.Expression
	if len(vals) == 0 {
		inExpr = &metricsview.Expression{
			Value: false,
		}
	} else {
		inExpr = &metricsview.Expression{
			Condition: &metricsview.Condition{
				Operator: metricsview.OperatorIn,
				Expressions: []*metricsview.Expression{
					{Name: qry.Dimensions[0].Name},
					{Value: vals},
				},
			},
		}
	}

	if qry.Where == nil {
		qry.Where = inExpr
	} else {
		qry.Where = &metricsview.Expression{
			Condition: &metricsview.Condition{
				Operator: metricsview.OperatorAnd,
				Expressions: []*metricsview.Expression{
					qry.Where,
					inExpr,
				},
			},
		}
	}

	// Remove the limit from the outer query (the IN filter automatically limits the result size).
	qry.Limit = nil
	qry.Offset = nil

	return nil
}
