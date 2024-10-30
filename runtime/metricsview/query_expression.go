package metricsview

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func ExpressionToString(e *Expression) (string, error) {
	b := exprStrBuilder{Builder: &strings.Builder{}}
	err := b.writeExpression(e)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

type exprStrBuilder struct {
	*strings.Builder
}

func (b exprStrBuilder) writeExpression(e *Expression) error {
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

func (b exprStrBuilder) writeName(name string) error {
	if strings.Contains(name, `"`) {
		_, err := strings.NewReplacer(`"`, `""`).WriteString(b.Builder, name)
		return err
	}
	b.writeString(name)
	return nil
}

func (b exprStrBuilder) writeValue(val any) error {
	res, err := json.Marshal(val)
	if err != nil {
		return err
	}
	if len(res) > 0 && res[len(res)-1] == '\n' {
		res = res[:len(res)-1]
	}
	_, err = b.WriteString(string(res))
	return err
}

func (b exprStrBuilder) writeSubquery(_ *Subquery) error {
	_, err := b.WriteString("<subquery>")
	return err
}

func (b exprStrBuilder) writeCondition(cond *Condition) error {
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

func (b exprStrBuilder) writeJoinedExpressions(exprs []*Expression, joiner string) error {
	if len(exprs) == 0 {
		return nil
	}

	for i, e := range exprs {
		if i > 0 {
			b.writeString(joiner)
		}
		err := b.writeWrappedExpression(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b exprStrBuilder) writeBinaryCondition(exprs []*Expression, op Operator) error {
	// Backwards compatibility: For IN and NIN, the right hand side may be a flattened list of values, not a single list.
	if op == OperatorIn || op == OperatorNin {
		if len(exprs) == 2 {
			rhs := exprs[1]
			typ := reflect.TypeOf(rhs.Value)
			isListVal := typ != nil && typ.Kind() == reflect.Slice
			if rhs.Name == "" && !isListVal && rhs.Condition == nil && rhs.Subquery == nil {
				// Convert the right hand side to a list
				exprs[1] = &Expression{Value: []any{rhs.Value}}
			}
		}
		if len(exprs) > 2 {
			vals := make([]any, 0, len(exprs)-1)
			for _, e := range exprs[1:] {
				vals = append(vals, e.Value)
			}
			exprs = []*Expression{exprs[0], {Value: vals}}
		}
	}

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

	err := b.writeWrappedExpression(left)
	if err != nil {
		return err
	}

	switch op {
	case OperatorEq:
		// Special case: "dim = NULL" should be written as "dim IS NULL"
		if hasNilValue(right) {
			b.writeString(" IS NULL")
			return nil
		}
		b.writeString("=")
	case OperatorNeq:
		// Special case: "dim != NULL" should be written as "dim IS NOT NULL"
		if hasNilValue(right) {
			b.writeString(" IS NOT NULL")
			return nil
		}
		b.writeString("!=")
	case OperatorLt:
		b.writeString("<")
	case OperatorLte:
		b.writeString("<=")
	case OperatorGt:
		b.writeString(">")
	case OperatorGte:
		b.writeString(">=")
	case OperatorIn:
		b.writeString(" IN ")
	case OperatorNin:
		b.writeString(" NOT IN ")
	case OperatorIlike:
		b.writeString(" ILIKE ")
	case OperatorNilike:
		b.writeString(" NOT ILIKE ")
	default:
		return fmt.Errorf("invalid binary condition operator %q", op)
	}

	err = b.writeWrappedExpression(right)
	if err != nil {
		return err
	}

	return nil
}

func (b exprStrBuilder) writeWrappedExpression(e *Expression) error {
	if e.Condition != nil {
		b.writeByte('(')
	}
	err := b.writeExpression(e)
	if err != nil {
		return err
	}
	if e.Condition != nil {
		b.writeByte(')')
	}
	return nil
}

func (b exprStrBuilder) writeByte(v byte) {
	_ = b.WriteByte(v)
}

func (b exprStrBuilder) writeString(s string) {
	_, _ = b.WriteString(s)
}

func hasNilValue(expr *Expression) bool {
	return expr != nil && expr.Value == nil && expr.Condition == nil && expr.Subquery == nil
}
