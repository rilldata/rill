package executor

import (
	"context"
	"errors"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"go.uber.org/zap"
)

// rewriteTwoPhaseComparisons rewrites the query to query base time range first and then use the results to query the comparison time range.
// It returns false without an error when the query is not eligible for the rewrite; the reason is logged at debug level.
func (e *Executor) rewriteTwoPhaseComparisons(ctx context.Context, qry *metricsview.Query, ast *metricsview.AST, ogLimit *int64) (bool, error) {
	// Not a comparison query, so there is nothing to rewrite.
	if qry.ComparisonTimeRange == nil {
		return false, nil
	}

	if qry.Limit == nil || *qry.Limit > e.instanceCfg.MetricsApproxComparisonTwoPhaseLimit {
		e.logTwoPhaseComparisonsSkipped(ctx, "the query has no limit or its limit exceeds rill.metrics.approximate_comparisons_two_phase_limit")
		return false, nil
	}

	// Skip if the criteria for a two-phase comparison are not met.
	if len(qry.Sort) != 1 || len(qry.Dimensions) != 1 || len(qry.PivotOn) > 0 {
		e.logTwoPhaseComparisonsSkipped(ctx, "the query must have exactly one dimension, exactly one sort field and no pivot")
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
			e.logTwoPhaseComparisonsSkipped(ctx, "sorting by a computed measure other than a comparison value is not supported")
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
		QueryLimits:         qry.QueryLimits,
		UnusedFields:        nil,
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
			e.logTwoPhaseComparisonsSkipped(ctx, err.Error())
			return false, nil
		}
		return false, err
	}

	if len(dimVals) == 0 {
		e.logTwoPhaseComparisonsSkipped(ctx, "the base query returned no rows")
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
			e.logTwoPhaseComparisonsSkipped(ctx, "the query has no comparison join")
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

// logTwoPhaseComparisonsSkipped logs why a comparison query is not rewritten to two phases.
// It logs at debug level because many comparison queries are legitimately not eligible, but the reason matters when investigating slow comparison queries.
func (e *Executor) logTwoPhaseComparisonsSkipped(ctx context.Context, reason string) {
	logger, err := e.rt.InstanceLogger(ctx, e.instanceID)
	if err != nil {
		return
	}
	logger.Debug("skipping two-phase comparison query", zap.String("table", e.metricsView.Table), zap.String("reason", reason))
}
