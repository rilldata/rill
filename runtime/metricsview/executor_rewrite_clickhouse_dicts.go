package metricsview

import (
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

type lookupMeta struct {
	table    string
	keyExpr  string
	keyCol   string
	valueCol string
}

// Rewrites filter on dictionary columns to use the key column instead of the value column. For example, if the dictionary value column is "country" and key column is "country_key",
// then the filter will be rewritten to country_key = 'US' instead of country = 'USA'.
func (e *Executor) rewriteClickhouseDictFilters(qry *Query) map[string]*lookupMeta {
	if e.olap.Dialect() != drivers.DialectClickHouse {
		return nil
	}

	dictMeta := make(map[string]*lookupMeta)
	for _, dim := range e.metricsView.Dimensions {
		if dim.Lookup != nil {
			dictMeta[dim.Name] = &lookupMeta{
				table:    dim.Lookup.Table,
				keyExpr:  dim.Column,
				keyCol:   dim.Lookup.KeyColumn,
				valueCol: dim.Lookup.ValueColumn,
			}
		}
	}
	if qry.Where != nil && len(dictMeta) > 0 {
		e.handleExpression(qry.Where, dictMeta)
	}
	return dictMeta
}

func (e *Executor) handleExpression(expr *Expression, dictMeta map[string]*lookupMeta) {
	if expr == nil || expr.Condition == nil {
		return
	}

	if expr.Condition.Operator == OperatorIn || expr.Condition.Operator == OperatorEq || expr.Condition.Operator == OperatorNeq {
		// expecting expression list to contain identifier in first position and then the values in subsequent positions
		exprs := expr.Condition.Expressions
		if len(exprs) < 2 {
			return
		}

		lkpMeta := dictMeta[exprs[0].Name]
		if lkpMeta == nil {
			return
		}

		subquery := &Subquery{
			RawSQL: fmt.Sprintf("SELECT %s FROM dictionary(%s) WHERE %s IN ", e.olap.Dialect().EscapeIdentifier(lkpMeta.keyCol), e.olap.Dialect().EscapeIdentifier(lkpMeta.table), e.olap.Dialect().EscapeIdentifier(lkpMeta.valueCol)),
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
				Identifier: lkpMeta.keyExpr,
			},
			{
				Subquery: subquery,
			},
		}
	} else {
		for _, ex := range expr.Condition.Expressions {
			e.handleExpression(ex, dictMeta)
		}
	}
}

// rewriteClickhouseDictGroupBys for dictionary dimension group by ID column, so if the dimension is dictGet(”,”, DICT_ID) AS DIM_NAME the group by DIM_ID to prevent lookup for each value
// also wrap SELECT inside a subquery and group by looked up value (DIM_NAME) again to merge duplicate groups if lookup in not injective i.e. multiple DIM_ID can map to a single DIM_NAME
func (e *Executor) rewriteClickhouseDictGroupBys(ast *AST, dictMeta map[string]*lookupMeta) {
	if len(dictMeta) == 0 {
		return
	}

	// handle CTEs first
	for _, cte := range ast.CTEs {
		e.rewriteClickhouseDictGroupBySelect(ast, cte, dictMeta)
	}

	// now the root node
	e.rewriteClickhouseDictGroupBySelect(ast, ast.Root, dictMeta)
}

func (e *Executor) rewriteClickhouseDictGroupBySelect(ast *AST, n *SelectNode, dictMeta map[string]*lookupMeta) {
	if n == nil {
		return
	}

	if n.Group {
		wrap := false
		for i := range n.DimFields {
			if lookup, ok := dictMeta[n.DimFields[i].Name]; ok {
				n.DimFields[i].GroupByIdentifier = lookup.keyExpr
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
		e.rewriteClickhouseDictGroupBySelect(ast, n.FromSelect, dictMeta)
	}

	if n.JoinComparisonSelect != nil {
		e.rewriteClickhouseDictGroupBySelect(ast, n.JoinComparisonSelect, dictMeta)
	}

	if n.SpineSelect != nil {
		e.rewriteClickhouseDictGroupBySelect(ast, n.SpineSelect, dictMeta)
	}

	for _, s := range n.LeftJoinSelects {
		e.rewriteClickhouseDictGroupBySelect(ast, s, dictMeta)
	}

	for _, s := range n.CrossJoinSelects {
		e.rewriteClickhouseDictGroupBySelect(ast, s, dictMeta)
	}
}
