package metricsresolver

import (
	"errors"
	"fmt"
	"strings"
)

func (ast *AST) buildExpression(e *Expression, having bool, n *MetricsSelect) (string, []any, error) {
	b := &expressionBuilder{
		ast:         ast,
		out:         &strings.Builder{},
		having:      having,
		metricsNode: n,
	}

	err := b.writeExpression(e)
	if err != nil {
		return "", nil, err
	}

	return b.out.String(), b.args, nil
}

type expressionBuilder struct {
	ast         *AST
	out         *strings.Builder
	args        []any
	having      bool
	metricsNode *MetricsSelect
}

func (b *expressionBuilder) writeExpression(e *Expression) error {
	if e == nil {
		return nil
	}
	if e.Name != "" {
		return b.writeName(e.Name)
	}
	if e.Value != nil {
		return b.writeValue(e.Value)
	}
	if e.Subquery != nil {
		return b.writeSubquery(e.Subquery)
	}
	if e.Condition != nil {
		return b.writeCondition(e.Condition)
	}
	return errors.New("invalid expression")
}

func (b *expressionBuilder) writeName(name string) error {
	expr, unnest, err := b.resolveName(name)
	if err != nil {
		return err
	}

	// writeName should not be called for names requiring unnesting
	if unnest {
		// TODO: Can be fixed with an EXISTS subquery
		return fmt.Errorf("cannot apply expression to dimension %q because it requires unnesting, which is not supported for expressions of this structure", name)
	}

	b.writeParenthesizedString(expr)
	return nil
}

func (b *expressionBuilder) writeValue(val any) error {
	b.writeString("?")
	b.args = append(b.args, val)
	return nil
}

func (b *expressionBuilder) writeSubquery(_ *Subquery) error {
	// TODO: Implement
	return fmt.Errorf("subqueries in expressions are not supported")
}

func (b *expressionBuilder) writeCondition(cond *Condition) error {
	switch cond.Operator {
	case OperatorEq:
		return b.writeBinaryCondition(cond.Expressions, " = ")
	case OperatorNeq:
		return b.writeBinaryCondition(cond.Expressions, " != ")
	case OperatorLt:
		return b.writeBinaryCondition(cond.Expressions, " < ")
	case OperatorLte:
		return b.writeBinaryCondition(cond.Expressions, " <= ")
	case OperatorGt:
		return b.writeBinaryCondition(cond.Expressions, " > ")
	case OperatorGte:
		return b.writeBinaryCondition(cond.Expressions, " >= ")
	case OperatorIn:
		return b.writeBinaryCondition(cond.Expressions, " IN ")
	case OperatorNin:
		return b.writeBinaryCondition(cond.Expressions, " NOT IN ")
	case OperatorIlike:
		return b.writeBinaryCondition(cond.Expressions, " LIKE ")
	case OperatorNilike:
		return b.writeBinaryCondition(cond.Expressions, " NOT LIKE ")
	case OperatorOr:
		return b.writeJoinedExpressions(cond.Expressions, " OR ")
	case OperatorAnd:
		return b.writeJoinedExpressions(cond.Expressions, " AND ")
	default:
		return fmt.Errorf("invalid expression operator %q", cond.Operator)
	}
}

func (b *expressionBuilder) writeBinaryCondition(exprs []*Expression, joiner string) error {
	if len(exprs) != 2 {
		return fmt.Errorf("binary condition must have exactly 2 expressions")
	}

	left := exprs[0]
	if left == nil {
		return fmt.Errorf("left expression is nil")
	}

	right := exprs[1]
	if right == nil {
		return fmt.Errorf("right expression is nil")
	}

	// TODO: Add unnest support
	// TODO: Add advanced LIKE/IN support

	// // Special handling for when both expressions are names
	// if left.Name != "" && right.Name != "" {
	// }

	// // Special handling for when the left expression is a name
	// if left.Name != "" {
	// 	leftExpr, unnest, err := b.resolveName(left.Name)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if !unnest {
	// 		b.writeParenthesizedString(leftExpr)

	// 	if unnest {
	// 		alias := b.ast.generateIdentifier()
	// 		b.ast.dialect.LateralUnnest(leftExpr, alias, left.Name)

	// 		b.writeString("EXISTS (SELECT 1 FROM ")
	// 		b.writeString()
	// 	}
	// }

	// Special handling for when the right expression is a name

	// EXISTS (SELECT 1 FROM LATERAL UNNEST(b) x(b) WHERE x.b = 20)

	// Fallback to generic expression building
	err := b.writeExpression(left)
	if err != nil {
		return err
	}
	b.writeString(joiner)
	err = b.writeExpression(right)
	if err != nil {
		return err
	}
	return nil
}

func (b *expressionBuilder) writeJoinedExpressions(exprs []*Expression, joiner string) error {
	for i, e := range exprs {
		if i > 0 {
			b.writeString(joiner)
		}
		err := b.writeExpression(e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *expressionBuilder) writeString(s string) {
	_, _ = b.out.WriteString(s)
}

func (b *expressionBuilder) writeParenthesizedString(s string) {
	_ = b.out.WriteByte('(')
	_, _ = b.out.WriteString(s)
	_ = b.out.WriteByte(')')
}

func (b *expressionBuilder) resolveName(name string) (expr string, unnest bool, err error) {
	// If metricsNode is nil, we are evaluating the expression against the underlying table.
	// In this case, we only allow filters to reference dimension names.
	if b.metricsNode == nil {
		// First, search for the dimension in the ASTs dimension fields (this also covers any computed dimension)
		for _, f := range b.ast.dimFields {
			if f.Name == name {
				if f.Unnest {
					// Since it's unnested, we need to reference the unnested alias.
					// Note that we return "false" for "unnest" because it will already have been unnested since it's one of the dimensions included in the query,
					// so we can filter against it as if it's a normal dimension.
					return b.ast.expressionForMember(f.UnnestAlias, f.Name), false, nil
				}
				return f.Expr, false, nil
			}
		}

		// Second, search for the dimension in the metrics view's dimensions (since expressions are allowed to reference dimensions not included in the query)
		dim, err := b.ast.lookupDimension(name, true)
		if err != nil {
			return "", false, fmt.Errorf("invalid dimension reference %q: %w", name, err)
		}

		// Note: If dim.Unnest is true, we need to unnest it inside of the generated expression (because it's not part of the dimFields and therefore not unnested with a LATERAL JOIN).
		return b.ast.dialect.MetricsViewDimensionExpression(dim), dim.Unnest, nil
	}

	// Since metricsNode is not nil, we're in the context of a wrapped SELECT.
	// We only allow expressions against the node's dimensions and measures (not those in scope within sub-queries).

	// Check if it's a dimension name
	for _, f := range b.metricsNode.DimFields {
		if f.Name == name {
			// NOTE: We don't need to handle Unnest here because it's always applied at the innermost query (i.e. when metricsNode==nil).
			// TODO: When b.having==true, could it use the name instead of the expression?
			return f.Expr, false, nil
		}
	}

	// Can't have a WHERE clause against a measure field if it's a GROUP BY query
	if b.metricsNode.Group {
		return "", false, fmt.Errorf("name %q in expression is not a dimension available in the current context", name)
	}

	// Check measure fields
	for _, f := range b.metricsNode.MeasureFields {
		if f.Name == name {
			// TODO: When b.having==true, could it use the name instead of the expression?
			return f.Expr, false, nil
		}
	}

	return "", false, fmt.Errorf("name %q in expression is not a dimension or measure available in the current context", name)
}

// func (builder *ExpressionBuilder) buildLikeExpression(cond *runtimev1.Condition) (string, []any, error) {
// 	if len(cond.Exprs) != 2 {
// 		return "", nil, fmt.Errorf("like/not like expression should have exactly 2 sub expressions")
// 	}

// 	leftExpr, args, err := builder.buildExpression(cond.Exprs[0])
// 	if err != nil {
// 		return "", nil, err
// 	}

// 	rightExpr, subArgs, err := builder.buildExpression(cond.Exprs[1])
// 	if err != nil {
// 		return "", nil, err
// 	}
// 	args = append(args, subArgs...)

// 	notKeyword := ""
// 	if cond.Op == runtimev1.Operation_OPERATION_NLIKE {
// 		notKeyword = "NOT"
// 	}

// 	// identify if immediate identifier has unnest
// 	unnest := builder.identifierIsUnnest(cond.Exprs[0])

// 	var clause string
// 	// Build [NOT] len(list_filter("dim", x -> x ILIKE ?)) > 0
// 	if unnest && builder.dialect != drivers.DialectDruid && builder.dialect != drivers.DialectPinot {
// 		clause = fmt.Sprintf("%s len(list_filter((%s), x -> x ILIKE %s)) > 0", notKeyword, leftExpr, rightExpr)
// 	} else {
// 		if builder.dialect == drivers.DialectDruid || builder.dialect == drivers.DialectPinot {
// 			// Druid and Pinot does not support ILIKE
// 			clause = fmt.Sprintf("LOWER(%s) %s LIKE LOWER(CAST(%s AS VARCHAR))", leftExpr, notKeyword, rightExpr)
// 		} else {
// 			clause = fmt.Sprintf("(%s) %s ILIKE %s", leftExpr, notKeyword, rightExpr)
// 		}
// 	}

// 	// When you have "dim NOT ILIKE '...'", then NULL values are always excluded.
// 	// We need to explicitly include it.
// 	if cond.Op == runtimev1.Operation_OPERATION_NLIKE {
// 		clause += fmt.Sprintf(" OR (%s) IS NULL", leftExpr)
// 	}

// 	return clause, args, nil
// }

// func (builder *ExpressionBuilder) buildInExpression(cond *runtimev1.Condition) (string, []any, error) {
// 	if len(cond.Exprs) <= 1 {
// 		return "", nil, fmt.Errorf("in/not in expression should have at least 2 sub expressions")
// 	}

// 	leftExpr, args, err := builder.buildExpression(cond.Exprs[0])
// 	if err != nil {
// 		return "", nil, err
// 	}

// 	notKeyword := ""
// 	exclude := cond.Op == runtimev1.Operation_OPERATION_NIN
// 	if exclude {
// 		notKeyword = "NOT"
// 	}

// 	inHasNull := false
// 	var valClauses []string
// 	// Add to args, skipping nulls
// 	for _, subExpr := range cond.Exprs[1:] {
// 		if v, isVal := subExpr.Expression.(*runtimev1.Expression_Val); isVal {
// 			if _, isNull := v.Val.Kind.(*structpb.Value_NullValue); isNull {
// 				inHasNull = true
// 				continue // Handled later using "dim IS [NOT] NULL" clause
// 			}
// 		}
// 		inVal, subArgs, err := builder.buildExpression(subExpr)
// 		if err != nil {
// 			return "", nil, err
// 		}
// 		args = append(args, subArgs...)
// 		valClauses = append(valClauses, inVal)
// 	}

// 	// identify if immediate identifier has unnest
// 	unnest := builder.identifierIsUnnest(cond.Exprs[0])

// 	clauses := make([]string, 0)

// 	// If there were non-null args, add a "dim [NOT] IN (...)" clause
// 	if len(valClauses) > 0 {
// 		questionMarks := strings.Join(valClauses, ",")
// 		var clause string
// 		// Build [NOT] list_has_any("dim", ARRAY[?, ?, ...])
// 		if unnest && builder.dialect != drivers.DialectDruid {
// 			clause = fmt.Sprintf("%s list_has_any((%s), ARRAY[%s])", notKeyword, leftExpr, questionMarks)
// 		} else {
// 			clause = fmt.Sprintf("(%s) %s IN (%s)", leftExpr, notKeyword, questionMarks)
// 		}
// 		clauses = append(clauses, clause)
// 	}

// 	if inHasNull {
// 		// Add null check
// 		// NOTE: DuckDB doesn't handle NULL values in an "IN" expression. They must be checked with a "dim IS [NOT] NULL" clause.
// 		clauses = append(clauses, fmt.Sprintf("(%s) IS %s NULL", leftExpr, notKeyword))
// 	}
// 	var condsClause string
// 	if exclude {
// 		condsClause = strings.Join(clauses, " AND ")
// 	} else {
// 		condsClause = strings.Join(clauses, " OR ")
// 	}
// 	if exclude && !inHasNull && len(clauses) > 0 {
// 		// When you have "dim NOT IN (a, b, ...)", then NULL values are always excluded, even if NULL is not in the list.
// 		// E.g. this returns zero rows: "select * from (select 1 as a union select null as a) where a not in (1)"
// 		// We need to explicitly include it.
// 		condsClause += fmt.Sprintf(" OR (%s) IS NULL", leftExpr)
// 	}

// 	return condsClause, args, nil
// }
