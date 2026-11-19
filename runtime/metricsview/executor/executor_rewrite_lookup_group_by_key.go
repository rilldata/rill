package executor

import (
	"context"
	"slices"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// rewriteLookupGroupByKey rewrites the grouped selects against the underlying table whose dimensions include a lookup dimension with lookup_group_by_key.
// Such a select is split in two: the inner select groups by the lookup keys and computes partial aggregates,
// and the outer select resolves the lookup values once per key, merges the partial aggregates and regroups by the resolved values.
// For a non-injective dictionary this merges the keys that share a value, which a plain GROUP BY on the key would not.
//
// The outer select keeps the sort, limit and offset. They can never move into the inner select: a value's measures merge several keys,
// so the top keys are not the top values. The rewrite must therefore run after every other AST rewrite, in particular the limit pushdowns,
// which would otherwise consider the inner select eligible because the partial aggregates reuse the measure names.
// A select with a measure the dialect cannot split is left untouched.
// Exact aggregates return the same results either way; approximate ones (such as quantile or topK) may differ slightly after merging partial states,
// and count(DISTINCT) is always evaluated with uniqExact regardless of the count_distinct_implementation setting.
func (e *Executor) rewriteLookupGroupByKey(ctx context.Context, ast *metricsview.AST) {
	// Skip the walk for metrics views without a lookup_group_by_key dimension.
	if ast.Query.Rows || !slices.ContainsFunc(ast.MetricsView.Dimensions, (*runtimev1.MetricsViewSpec_Dimension).GetLookupGroupByKey) {
		return
	}

	var applied int
	var skipped []string
	rewriteLookupGroupByKeyWalk(ast, ast.Root, &applied, &skipped)
	if applied == 0 && len(skipped) == 0 {
		return
	}

	trace.SpanFromContext(ctx).SetAttributes(
		attribute.Int("lookup_group_by_key.applied", applied),
		attribute.StringSlice("lookup_group_by_key.skipped", skipped),
	)
}

// rewriteLookupGroupByKeyWalk visits the sub-selects before the node itself, so the inner select created for a node is never visited.
func rewriteLookupGroupByKeyWalk(ast *metricsview.AST, n *metricsview.SelectNode, applied *int, skipped *[]string) {
	if n == nil || n.RawSelect != nil {
		return
	}

	rewriteLookupGroupByKeyWalk(ast, n.FromSelect, applied, skipped)
	rewriteLookupGroupByKeyWalk(ast, n.SpineSelect, applied, skipped)
	for _, ljs := range n.LeftJoinSelects {
		rewriteLookupGroupByKeyWalk(ast, ljs, applied, skipped)
	}
	for _, cjs := range n.CrossJoinSelects {
		rewriteLookupGroupByKeyWalk(ast, cjs, applied, skipped)
	}
	rewriteLookupGroupByKeyWalk(ast, n.JoinComparisonSelect, applied, skipped)

	if n.FromTable == nil || !n.Group {
		return
	}

	ok, reason := rewriteLookupGroupByKeyNode(ast, n)
	if ok {
		*applied++
	} else if reason != "" {
		*skipped = append(*skipped, reason)
	}
}

// rewriteLookupGroupByKeyNode rewrites one grouped select against the underlying table.
// It returns whether the select was rewritten and, when it had a dimension with lookup_group_by_key but was left untouched, the reason why.
func rewriteLookupGroupByKeyNode(ast *metricsview.AST, n *metricsview.SelectNode) (bool, string) {
	// Find the dimensions to group by key.
	// A field is only eligible if it still selects the dimension's lookup expression as is:
	// a field renamed by another rewrite is not a dimension of the metrics view, and a field computed from the dimension (such as a time floor) must pass through.
	flagged := make(map[int]*runtimev1.MetricsViewSpec_Dimension)
	for i, f := range n.DimFields {
		if f.Unnest {
			continue
		}
		dim, err := ast.LookupDimension(f.Name, false)
		if err != nil || dim.LookupTable == "" || !dim.GetLookupGroupByKey() {
			continue
		}
		expr, err := ast.Dialect.MetricsViewDimensionExpression(dim)
		if err != nil || expr != f.Expr {
			continue
		}
		flagged[i] = dim
	}
	if len(flagged) == 0 {
		return false, ""
	}

	innerAlias := ast.GenerateIdentifier()

	// Split the measures into partial aggregates for the inner select and merge expressions for the outer select.
	// The inner select drops treat_nulls_as because it applies to the merged value, not to the partial state.
	taken := make(map[string]bool) // Field names used in the inner select.
	innerMeasures := make([]metricsview.FieldNode, 0, len(n.MeasureFields))
	outerMeasures := make([]metricsview.FieldNode, 0, len(n.MeasureFields))
	for _, f := range n.MeasureFields {
		partials, final, ok := ast.Dialect.SplitAggregateExpression(f.Expr, f.Name, innerAlias)
		if !ok {
			return false, "measure_not_splittable:" + f.Name
		}
		for _, p := range partials {
			if taken[p.Alias] {
				return false, "alias_conflict:" + p.Alias
			}
			taken[p.Alias] = true
			innerMeasures = append(innerMeasures, metricsview.FieldNode{Name: p.Alias, Expr: p.Expr})
		}
		outerMeasures = append(outerMeasures, metricsview.FieldNode{
			Name:        f.Name,
			DisplayName: f.DisplayName,
			Expr:        final,
			TreatNullAs: f.TreatNullAs,
		})
	}

	// The other dimensions pass through the inner select under their own names, which may collide with a partial aggregate alias like "m__1".
	// keyFields maps a key expression to the inner field exposing it, so that a dimension on the raw key and several lookups on the same key share one field.
	keyFields := make(map[string]string)
	for i, f := range n.DimFields {
		if _, ok := flagged[i]; ok {
			continue
		}
		if taken[f.Name] {
			return false, "alias_conflict:" + f.Name
		}
		taken[f.Name] = true
		if !f.Unnest {
			keyFields[f.Expr] = f.Name
		}
	}

	// A flagged dimension becomes its key in the inner select and the lookup over that key in the outer select.
	innerDims := make([]metricsview.FieldNode, 0, len(n.DimFields))
	outerDims := make([]metricsview.FieldNode, 0, len(n.DimFields))
	for i, f := range n.DimFields {
		dim, ok := flagged[i]
		if !ok {
			innerDims = append(innerDims, f)
			outerDims = append(outerDims, metricsview.FieldNode{
				Name:        f.Name,
				DisplayName: f.DisplayName,
				Expr:        ast.Dialect.EscapeMemberAlias(innerAlias, f.Name),
			})
			continue
		}

		keyExpr := metricsview.LookupKeyExpr(ast.Dialect, dim)
		keyAlias, found := keyFields[keyExpr]
		// A default expression may reference the key column by name, which only resolves if the inner select exposes the key under that name.
		if found && dim.LookupDefaultExpression != "" && keyAlias != dim.Column {
			found = false
		}
		if !found {
			// The name is only taken when the query also groups by another dimension named like the key column, which is rare enough to skip the rewrite.
			keyAlias = lookupKeyAlias(dim)
			if taken[keyAlias] {
				return false, "key_alias_conflict:" + dim.Name
			}
			taken[keyAlias] = true
			keyFields[keyExpr] = keyAlias
			innerDims = append(innerDims, metricsview.FieldNode{Name: keyAlias, Expr: keyExpr})
		}

		expr, err := ast.Dialect.LookupExpr(dim.LookupTable, dim.LookupValueColumn, ast.Dialect.EscapeMemberAlias(innerAlias, keyAlias), dim.LookupDefaultExpression)
		if err != nil {
			return false, "lookup_expression:" + dim.Name
		}
		outerDims = append(outerDims, metricsview.FieldNode{
			Name:        f.Name,
			DisplayName: f.DisplayName,
			Expr:        expr,
		})
	}

	// The inner select keeps the table, the unnests and the filters; the outer select keeps the alias, the CTE status, the sort, the limit and the offset.
	inner := *n
	inner.Alias = innerAlias
	inner.IsCTE = false
	inner.DimFields = innerDims
	inner.MeasureFields = innerMeasures
	inner.Having = nil
	inner.OrderBy = nil
	inner.Limit = nil
	inner.Offset = nil

	n.FromTable = nil
	n.FromSelect = &inner
	n.Unnests = nil
	n.Where = nil
	n.TimeWhere = nil
	n.DimFields = outerDims
	n.MeasureFields = outerMeasures

	return true, ""
}

// lookupKeyAlias returns the name under which a sub-select exposes a lookup dimension's key.
// It is the key column's own name when the dimension targets a column, which keeps a lookup_default_expression that references the column working.
func lookupKeyAlias(dim *runtimev1.MetricsViewSpec_Dimension) string {
	if dim.Column != "" {
		return dim.Column
	}
	return dim.Name + "__key"
}
