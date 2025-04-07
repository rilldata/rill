package metricsview

import (
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

// Rewrites filter on dictionary columns to use the key column instead of the value column. For example, if the dictionary value column is "country" and key column is "country_key",
// then the filter will be rewritten to country_key = 'US' instead of country = 'USA'.
func (e *Executor) rewriteClickhouseDictFilters(qry *Query) {
	if e.olap.Dialect() != drivers.DialectClickHouse {
		return
	}

	if qry.Where != nil {
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
		if len(dictMeta) > 0 {
			e.handleExpression(qry.Where, dictMeta)
		}
	}
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
				// could not find expected expression values, don't rewrite
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

type lookupMeta struct {
	table    string
	keyExpr  string
	keyCol   string
	valueCol string
}
