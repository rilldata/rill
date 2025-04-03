package metricsview

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

type lookupMeta struct {
	table  string
	column string
	key    string
}

// Rewrites filter on dictionary columns to use the key column instead of the value column. For example, if the dictionary value column is "country" and key column is "country_key",
// then the filter will be rewritten to country_key = 'US' instead of country = 'USA'.
func (e *Executor) rewriteClickhouseDictFilters(qry *Query) map[string]*lookupMeta {
	if e.olap.Dialect() != drivers.DialectClickHouse {
		return nil
	}

	dictLookups := make(map[string]*lookupMeta)
	for _, dim := range e.metricsView.Dimensions {
		if strings.HasPrefix(dim.Expression, "dictGet") {
			tbl, col, key := drivers.ParseClickhouseDictGet(dim.Expression)
			dictLookups[dim.Name] = &lookupMeta{
				table:  tbl,
				column: col,
				key:    key,
			}
		}
	}
	if qry.Where != nil && len(dictLookups) > 0 {
		e.handleExpression(qry.Where, dictLookups)
	}
	return dictLookups
}

func (e *Executor) handleExpression(expr *Expression, dictLookups map[string]*lookupMeta) {
	if expr == nil || expr.Condition == nil {
		return
	}

	if expr.Condition.Operator == OperatorIn || expr.Condition.Operator == OperatorEq || expr.Condition.Operator == OperatorNeq {
		// expecting expression list to contain identifier in first position and then the values in subsequent positions
		exprs := expr.Condition.Expressions
		if len(exprs) < 2 {
			return
		}

		if dictLookups[exprs[0].Name] == nil {
			return
		}

		lkpMeta := dictLookups[exprs[0].Name]

		subquery := &Subquery{
			RawSQL: fmt.Sprintf("SELECT %s FROM dictionary(%s) WHERE %s IN ", e.olap.Dialect().EscapeIdentifier(lkpMeta.key), e.olap.Dialect().EscapeIdentifier(lkpMeta.table), e.olap.Dialect().EscapeIdentifier(lkpMeta.column)),
		}

		subquery.RawSQL += "("
		for i := 1; i < len(exprs); i++ {
			if exprs[i].Value != nil {
				if i == 1 {
					// quoting numeric values also works fine as clickhouse casts them if required
					subquery.RawSQL += e.olap.Dialect().EscapeStringValue(fmt.Sprintf("%v", exprs[i].Value))
				} else {
					subquery.RawSQL += ", " + e.olap.Dialect().EscapeStringValue(fmt.Sprintf("%v", exprs[i].Value))
				}
			} else {
				// could find expected expression values, don't rewrite
				return
			}
		}
		subquery.RawSQL += ")"

		expr.Condition.Expressions = []*Expression{
			{
				Identifier: lkpMeta.key,
			},
			{
				Subquery: subquery,
			},
		}
	} else {
		for _, ex := range expr.Condition.Expressions {
			e.handleExpression(ex, dictLookups)
		}
	}
}

// wrap select inside subquery and group by looked up value again to prevent duplicate groups if lookup in not injective
func (e *Executor) rewriteClickhouseDictGroupBys(ast *AST, dictLookups map[string]*lookupMeta) {
	if len(dictLookups) == 0 {
		return
	}

	// handle CTEs first
	for _, cte := range ast.CTEs {
		e.rewriteClickhouseDictGroupBySelect(ast, cte, dictLookups)
	}

	// now the root node
	e.rewriteClickhouseDictGroupBySelect(ast, ast.Root, dictLookups)
}

func (e *Executor) rewriteClickhouseDictGroupBySelect(ast *AST, n *SelectNode, dictLookups map[string]*lookupMeta) {
	if n == nil {
		return
	}

	if n.Group {
		wrap := false
		for i := range n.DimFields {
			if lookup, ok := dictLookups[n.DimFields[i].Name]; ok {
				n.DimFields[i].GroupByIdentifier = lookup.key
				wrap = true
			} else {
				n.DimFields[i].GroupByIdentifier = ast.dimFields[i].Name
			}
		}
		if wrap {
			n.GroupByIdentifier = true
			ast.wrapSelect(n, ast.generateIdentifier())
			n = n.FromSelect
		}
	}

	if n.FromSelect != nil {
		e.rewriteClickhouseDictGroupBySelect(ast, n.FromSelect, dictLookups)
	}

	if n.JoinComparisonSelect != nil {
		e.rewriteClickhouseDictGroupBySelect(ast, n.JoinComparisonSelect, dictLookups)
	}

	if n.SpineSelect != nil {
		e.rewriteClickhouseDictGroupBySelect(ast, n.SpineSelect, dictLookups)
	}

	for _, s := range n.LeftJoinSelects {
		e.rewriteClickhouseDictGroupBySelect(ast, s, dictLookups)
	}

	for _, s := range n.CrossJoinSelects {
		e.rewriteClickhouseDictGroupBySelect(ast, s, dictLookups)
	}
}
