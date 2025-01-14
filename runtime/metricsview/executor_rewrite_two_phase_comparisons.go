package metricsview

import (
	"context"
	"errors"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

// rewriteTwoPhaseComparisons rewrites the query to query base time range first and then use the results to query the comparison time range.
func (e *Executor) rewriteTwoPhaseComparisons(ctx context.Context, qry *Query, ast *AST, ogLimit *int64) (bool, error) {
	// Check if it's enabled.
	if !e.instanceCfg.MetricsApproxTwoPhaseComparisons {
		return false, nil
	}

	// Skip if the criteria for a two-phase comparison are not met.
	if qry.ComparisonTimeRange == nil || len(qry.Sort) != 1 || len(qry.Dimensions) == 0 || len(qry.Dimensions) > 1 || len(qry.PivotOn) > 0 {
		return false, nil
	}

	// Find out what we're sorting by and also accumulate the underlying base measure
	sortField := qry.Sort[0]

	var bm []Measure
	for _, qm := range qry.Measures {
		if qm.Compute == nil || (qm.Compute.ComparisonValue == nil && qm.Compute.ComparisonDelta == nil && qm.Compute.ComparisonRatio == nil && qm.Compute.PercentOfTotal == nil) {
			bm = append(bm, qm)
			continue
		}

		if qm.Name == sortField.Name {
			// only supported sorting on base value TODO extend for comparison value
			return false, nil
		}
	}

	// Build a query for the base time range
	baseQry := &Query{
		MetricsView:         qry.MetricsView,
		Dimensions:          qry.Dimensions,
		Measures:            bm,
		PivotOn:             qry.PivotOn,
		Spine:               qry.Spine,
		Sort:                qry.Sort,
		TimeRange:           qry.TimeRange,
		ComparisonTimeRange: nil,
		Where:               qry.Where,
		Having:              nil,
		Limit:               ogLimit,
		Offset:              qry.Offset,
		TimeZone:            qry.TimeZone,
		UseDisplayNames:     false,
	}

	// Execute the query for the base time range
	baseRes, err := e.Query(ctx, baseQry, nil)
	if err != nil {
		return false, err
	}
	defer baseRes.Close()

	sel, dimVals, err := e.olap.Dialect().SelectInlineResults(baseRes)
	if err != nil {
		if errors.Is(err, drivers.ErrOptimizationFailure) {
			return false, nil
		}
		return false, err
	}

	if len(dimVals) == 0 {
		return false, nil
	}

	base := &SelectNode{
		Alias:     ast.Root.FromSelect.Alias,
		DimFields: ast.Root.FromSelect.DimFields,
		RawSelect: sel,
	}

	ast.Root.FromSelect = base

	comp := ast.Root.JoinComparisonSelect

	// Add the dimensions values as a "<dim> IN (<vals...>)" expression in the outer query's WHERE clause.
	var inExpr *Expression

	// if any dim value is nil add condition with eq operator with nil value
	var vals []any
	foundNil := false
	for _, v := range dimVals {
		if v == nil {
			foundNil = true
		} else {
			vals = append(vals, v)
		}
	}
	inExpr = &Expression{
		Condition: &Condition{
			Operator: OperatorIn,
			Expressions: []*Expression{
				{Name: qry.Dimensions[0].Name},
				{Value: vals},
			},
		},
	}
	if foundNil {
		inExpr = &Expression{
			Condition: &Condition{
				Operator: OperatorOr,
				Expressions: []*Expression{
					inExpr,
					{
						Condition: &Condition{
							Operator: OperatorEq,
							Expressions: []*Expression{
								{Name: qry.Dimensions[0].Name},
								{Value: nil},
							},
						},
					},
				},
			},
		}
	}

	expr, args, err := ast.sqlForExpression(inExpr, comp, true, true)
	if err != nil {
		return false, fmt.Errorf("failed to compile 'having': %w", err)
	}
	res := &ExprNode{
		Expr: expr,
		Args: args,
	}

	if comp.Where == nil {
		comp.Where = res
	} else {
		comp.Where = comp.Where.and(res.Expr, res.Args)
	}

	return true, nil
}
