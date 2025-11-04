package executor

import (
	"context"
	"errors"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteTwoPhaseComparisons rewrites the query to query base time range first and then use the results to query the comparison time range.
func (e *Executor) rewriteTwoPhaseComparisons(ctx context.Context, qry *metricsview.Query, ast *metricsview.AST, ogLimit *int64) (bool, error) {
	if qry.Limit == nil || (*qry.Limit > e.instanceCfg.MetricsApproxComparisonTwoPhaseLimit) {
		return false, nil
	}

	// Skip if the criteria for a two-phase comparison are not met.
	if qry.ComparisonTimeRange == nil || len(qry.Sort) != 1 || len(qry.Dimensions) == 0 || len(qry.Dimensions) > 1 || len(qry.PivotOn) > 0 {
		return false, nil
	}

	// Find out what we're sorting by and also accumulate the underlying base measures
	sortField := qry.Sort[0]
	tr := qry.TimeRange
	sortCompare := false

	var bm []metricsview.Measure
	for _, qm := range qry.Measures {
		if qm.Compute == nil || (qm.Compute.ComparisonValue == nil && qm.Compute.ComparisonDelta == nil && qm.Compute.ComparisonRatio == nil && qm.Compute.PercentOfTotal == nil && qm.Compute.ComparisonTime == nil) {
			bm = append(bm, qm)
			continue
		}

		if qm.Name == sortField.Name {
			if qm.Compute.ComparisonValue != nil {
				sortCompare = true
				tr = qry.ComparisonTimeRange
				sortField.Name = qm.Compute.ComparisonValue.Measure
				continue
			}
			// only supports sorting on base or compare value and dims
			return false, nil
		}
	}

	// Build a query for the base time range
	baseQry := &metricsview.Query{
		MetricsView:         qry.MetricsView,
		Dimensions:          qry.Dimensions,
		Measures:            bm,
		PivotOn:             qry.PivotOn,
		Spine:               qry.Spine,
		Sort:                []metricsview.Sort{sortField},
		TimeRange:           tr,
		ComparisonTimeRange: nil,
		Where:               qry.Where,
		Having:              nil,
		Limit:               ogLimit,
		Offset:              qry.Offset,
		TimeZone:            qry.TimeZone,
		UseDisplayNames:     false,
		Rows:                false,
	} //exhaustruct:enforce

	// Execute the query for the base time range
	baseRes, err := e.Query(ctx, baseQry, nil)
	if err != nil {
		return false, err
	}
	defer baseRes.Close()

	sel, args, dimVals, err := e.olap.Dialect().SelectInlineResults(baseRes)
	if err != nil {
		if errors.Is(err, drivers.ErrOptimizationFailure) {
			return false, nil
		}
		return false, err
	}

	if len(dimVals) == 0 {
		return false, nil
	}

	// figure out node at which join comparison select is present
	n := ast.Root
	for {
		if n.JoinComparisonSelect != nil {
			break
		}
		n = n.FromSelect
		if n == nil {
			// no join comparison select found
			return false, nil
		}
	}

	base := &metricsview.SelectNode{
		Alias:     n.FromSelect.Alias,
		DimFields: n.FromSelect.DimFields,
		RawSelect: &metricsview.ExprNode{
			Expr: sel,
			Args: args,
		},
	}

	var comp *metricsview.SelectNode
	if !sortCompare {
		// sorting by base value - set the inline results as base node and add base results to comparison node where clause
		n.FromSelect = base
		comp = n.JoinComparisonSelect
	} else {
		// flip base for sorting on comparison value - set inline results as join comparison select and add base results to base node where clause
		base.Alias = n.JoinComparisonSelect.Alias
		base.DimFields = n.JoinComparisonSelect.DimFields
		n.JoinComparisonSelect = base
		comp = n.FromSelect
	}

	// Add the dimensions values as a "<dim> IN (<vals...>)" expression in the comparison query's WHERE clause.
	var inExpr *metricsview.Expression

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
	inExpr = &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator: metricsview.OperatorIn,
			Expressions: []*metricsview.Expression{
				{Name: qry.Dimensions[0].Name},
				{Value: vals},
			},
		},
	}
	if foundNil {
		inExpr = &metricsview.Expression{
			Condition: &metricsview.Condition{
				Operator: metricsview.OperatorOr,
				Expressions: []*metricsview.Expression{
					inExpr,
					{
						Condition: &metricsview.Condition{
							Operator: metricsview.OperatorEq,
							Expressions: []*metricsview.Expression{
								{Name: qry.Dimensions[0].Name},
								{Value: nil},
							},
						},
					},
				},
			},
		}
	}

	expr, args, err := ast.SQLForExpression(inExpr, comp, true, true)
	if err != nil {
		return false, fmt.Errorf("failed to compile 'having': %w", err)
	}
	res := &metricsview.ExprNode{
		Expr: expr,
		Args: args,
	}

	if comp.Where == nil {
		comp.Where = res
	} else {
		comp.Where = comp.Where.And(res.Expr, res.Args)
	}

	return true, nil
}
