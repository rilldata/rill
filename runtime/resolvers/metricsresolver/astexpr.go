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

	if unnest {
		// We currently only handle unnest for the left expression in binary conditions (see writeBinaryCondition).
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
	case OperatorOr:
		return b.writeJoinedExpressions(cond.Expressions, " OR ")
	case OperatorAnd:
		return b.writeJoinedExpressions(cond.Expressions, " AND ")
	default:
		if !cond.Operator.Valid() {
			return fmt.Errorf("invalid expression operator %q", cond.Operator)
		}
		return b.writeBinaryCondition(cond.Expressions, cond.Operator)
	}
}

func (b *expressionBuilder) writeJoinedExpressions(exprs []*Expression, joiner string) error {
	if len(exprs) == 0 {
		return nil
	}

	b.writeByte('(')

	for i, e := range exprs {
		if i > 0 {
			b.writeString(joiner)
		}
		err := b.writeExpression(e)
		if err != nil {
			return err
		}
	}

	b.writeByte(')')

	return nil
}

func (b *expressionBuilder) writeBinaryCondition(exprs []*Expression, op Operator) error {
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

	// Check there isn't an unnest on the right side
	if right.Name != "" {
		_, unnest, err := b.resolveName(right.Name)
		if err != nil {
			return err
		}
		if unnest {
			return fmt.Errorf("cannot apply expression to dimension %q because it requires unnesting, which is only supported for the left side of an operation", right.Name)
		}
	}

	// Handle unnest on the left side
	if left.Name != "" {
		leftExpr, unnest, err := b.resolveName(left.Name)
		if err != nil {
			return err
		}

		// If not unnested, write the expression as-is
		if !unnest {
			return b.writeBinaryConditionInner(nil, right, leftExpr, op)
		}

		// Generate unnest join
		unnestTableAlias := b.ast.generateIdentifier()
		unnestFrom, ok, err := b.ast.dialect.LateralUnnest(leftExpr, unnestTableAlias, left.Name)
		if err != nil {
			return err
		}
		if !ok {
			// Means the DB automatically unnests, so we can treat it as a normal value
			return b.writeBinaryConditionInner(nil, right, leftExpr, op)
		}
		unnestColAlias := b.ast.expressionForMember(unnestTableAlias, left.Name)

		// Need to move "NOT" to outside of the subquery
		var not bool
		switch op {
		case OperatorNeq:
			op = OperatorEq
			not = true
		case OperatorNin:
			op = OperatorIn
			not = true
		case OperatorNilike:
			op = OperatorIlike
			not = true
		}

		// Output: [NOT] EXISTS (SELECT 1 FROM <unnestFrom> WHERE <unnestColAlias> <operator> <right>)
		if not {
			b.writeString("NOT ")
		}
		b.writeString("EXISTS (SELECT 1 FROM ")
		b.writeString(unnestFrom)
		b.writeString(" WHERE ")
		err = b.writeBinaryConditionInner(nil, right, unnestColAlias, op)
		if err != nil {
			return err
		}
		b.writeString(")")
		return nil
	}

	// Handle netiher side is a name
	return b.writeBinaryConditionInner(left, right, "", op)
}

func (b *expressionBuilder) writeBinaryConditionInner(left, right *Expression, leftOverride string, op Operator) error {
	var joiner string
	switch op {
	case OperatorEq:
		joiner = " = "
	case OperatorNeq:
		joiner = " != "
	case OperatorLt:
		joiner = " < "
	case OperatorLte:
		joiner = " <= "
	case OperatorGt:
		joiner = " > "
	case OperatorGte:
		joiner = " >= "
	case OperatorIlike:
		return b.writeILikeCondition(left, right, leftOverride, false)
	case OperatorNilike:
		return b.writeILikeCondition(left, right, leftOverride, true)
	case OperatorIn:
		return b.writeInCondition(left, right, leftOverride, false)
	case OperatorNin:
		return b.writeInCondition(left, right, leftOverride, true)
	default:
		return fmt.Errorf("invalid binary condition operator %q", op)
	}

	if leftOverride != "" {
		b.writeParenthesizedString(leftOverride)
	} else {
		err := b.writeExpression(left)
		if err != nil {
			return err
		}
	}
	b.writeString(joiner)
	err := b.writeExpression(right)
	if err != nil {
		return err
	}
	return nil
}

func (b *expressionBuilder) writeILikeCondition(left, right *Expression, leftOverride string, not bool) error {
	if not {
		b.writeByte('(')
	}

	if b.ast.dialect.SupportsILike() {
		// Output: <left> [NOT] ILIKE <right>

		if leftOverride != "" {
			b.writeParenthesizedString(leftOverride)
		} else {
			err := b.writeExpression(left)
			if err != nil {
				return err
			}
		}

		if not {
			b.writeString(" NOT ILIKE ")
		} else {
			b.writeString(" ILIKE ")
		}

		err := b.writeExpression(right)
		if err != nil {
			return err
		}
	} else {
		// Output: LOWER(<left>) [NOT] LIKE LOWER(<right>)

		b.writeString("LOWER(")
		if leftOverride != "" {
			b.writeString(leftOverride)
		} else {
			err := b.writeExpression(left)
			if err != nil {
				return err
			}
		}
		b.writeString(")")

		if not {
			b.writeString(" NOT ILIKE ")
		} else {
			b.writeString(" ILIKE ")
		}

		b.writeString("LOWER(")
		err := b.writeExpression(right)
		if err != nil {
			return err
		}
		b.writeString(")")
	}

	// When you have "dim NOT ILIKE <val>", then NULL values are always excluded. We need to explicitly include it.
	if not {
		b.writeString(" OR ")
		if leftOverride != "" {
			b.writeParenthesizedString(leftOverride)
		} else {
			err := b.writeExpression(left)
			if err != nil {
				return err
			}
		}
		b.writeString(" IS NULL")
	}

	// Closes the parens opened at the start
	if not {
		b.writeByte(')')
	}

	return nil
}

func (b *expressionBuilder) writeInCondition(left, right *Expression, leftOverride string, not bool) error {
	if right.Value == nil {
		return fmt.Errorf("the right expression must be a value for an IN condition")
	}
	vals, ok := right.Value.([]any)
	if !ok {
		return fmt.Errorf("the right expression must be a list of values for an IN condition")
	}

	var hasNull bool
	for _, v := range vals {
		if v == nil {
			hasNull = true
			break
		}
	}

	if len(vals) == 0 {
		if not {
			b.writeString("TRUE")
		} else {
			b.writeString("FALSE")
		}
		return nil
	}

	wrapParens := not || hasNull
	if wrapParens {
		b.writeByte('(')
	}

	if leftOverride != "" {
		b.writeParenthesizedString(leftOverride)
	} else {
		err := b.writeExpression(left)
		if err != nil {
			return err
		}
	}

	if not {
		b.writeString(" NOT IN ")
	} else {
		b.writeString(" IN ")
	}

	b.writeByte('(')
	for i := 0; i < len(vals); i++ {
		if i == 0 {
			b.writeString("?")
		} else {
			b.writeString(",?")
		}
	}
	b.writeByte(')')
	b.args = append(b.args, vals...)

	if hasNull {
		if not {
			b.writeString(" AND ")
		} else {
			b.writeString(" OR ")
		}

		if leftOverride != "" {
			b.writeParenthesizedString(leftOverride)
		} else {
			err := b.writeExpression(left)
			if err != nil {
				return err
			}
		}

		if not {
			b.writeString(" IS NOT NULL")
		} else {
			b.writeString(" IS NULL")
		}
	}

	// When you have "dim NOT IN (...)", then NULL values are always excluded. We need to explicitly include it.
	if not && !hasNull {
		b.writeString(" OR ")
		if leftOverride != "" {
			b.writeParenthesizedString(leftOverride)
		} else {
			err := b.writeExpression(left)
			if err != nil {
				return err
			}
		}
		b.writeString(" IS NULL")
	}

	if wrapParens {
		b.writeByte(')')
	}

	return nil
}

func (b *expressionBuilder) writeByte(v byte) {
	_ = b.out.WriteByte(v)
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
