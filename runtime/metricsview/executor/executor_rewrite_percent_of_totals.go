package executor

import (
	"context"
	"errors"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
)

func (e *Executor) rewritePercentOfTotals(ctx context.Context, qry *metricsview.Query) error {
	var measures []metricsview.Measure
	var measureIndices []int
	for i, measure := range qry.Measures {
		if measure.Compute != nil && measure.Compute.PercentOfTotal != nil {
			measures = append(measures, metricsview.Measure{
				Name: measure.Compute.PercentOfTotal.Measure,
			})
			measureIndices = append(measureIndices, i)
		}
	}

	if len(measures) == 0 {
		return nil
	}

	totalsQry := &metricsview.Query{
		MetricsView:         qry.MetricsView,
		Dimensions:          nil,
		Measures:            measures,
		PivotOn:             nil,
		Spine:               nil,
		Sort:                nil,
		TimeRange:           qry.TimeRange,
		ComparisonTimeRange: nil,
		Where:               qry.Where,
		Having:              nil, // 'having' should only apply after totals are calculated
		Limit:               nil,
		Offset:              nil,
		TimeZone:            qry.TimeZone,
		UseDisplayNames:     false,
		Rows:                false,
	} //exhaustruct:enforce

	// Build an AST for the totals query.
	ast, err := metricsview.NewAST(e.metricsView, e.security, totalsQry, e.olap.Dialect())
	if err != nil {
		return fmt.Errorf("percent of totals: failed to build the totals query AST: %w", err)
	}

	// Apply a limited subset of rewrites to the inner query.
	e.rewriteApproxComparisons(ast, false)
	err = e.rewriteQueryDruidExactify(ctx, totalsQry)
	if err != nil {
		return err
	}

	// Generate the SQL for and execute the totals query.
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

	if !res.Next() {
		return errors.New("query returned no results")
	}

	vals := make([]any, len(measures))
	for i := range vals {
		vals[i] = new(float64)
	}
	err = res.Scan(vals...)
	if err != nil {
		return err
	}

	for i, measure := range measures {
		v, ok := vals[i].(*float64)
		if !ok {
			return fmt.Errorf("%q is not a valid number", measure.Name)
		}

		qry.Measures[measureIndices[i]].Compute.PercentOfTotal.Total = v
	}

	return nil
}
