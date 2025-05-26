package metricsview

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func ExpressionToSQL(e *Expression) (string, error) {
	b := exprBuilder{
		Builder: &strings.Builder{},
	}
	err := b.writeExpression(e, "")
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func ExpressionToExportString(e *Expression) (string, error) {
	b := exprBuilder{
		Builder:   &strings.Builder{},
		forExport: true,
	}
	err := b.writeExpression(e, ", ")
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

type exprBuilder struct {
	forExport bool
	*strings.Builder
}

func (b exprBuilder) writeExpression(e *Expression, joiner string) error {
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
		return b.writeCondition(e.Condition, joiner)
	}
	return errors.New("invalid expression")
}

func (b exprBuilder) writeName(name string) error {
	if strings.Contains(name, `"`) {
		_, err := strings.NewReplacer(`"`, `""`).WriteString(b.Builder, name)
		return err
	}
	b.writeString(name)
	return nil
}

func (b exprBuilder) writeValue(val any) error {
	// In case of Non SQL for array for more the 10 values we need  print 10 values + N-10 More
	if arr, ok := val.([]any); ok && b.forExport {
		limit := 10
		n := len(arr)
		count := n
		if n > limit {
			count = limit
		}

		for i := 0; i < count; i++ {
			if i > 0 {
				b.WriteString(", ")
			}
			res, err := json.Marshal(arr[i])
			if err != nil {
				return err
			}
			b.WriteString(string(res))
		}

		if n > limit {
			fmt.Fprintf(b, " + %d More", n-limit)
		}

		return nil
	}

	// Else: for non array values
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

func (b exprBuilder) writeSubquery(subquery *Subquery) error {
	if b.forExport {
		// for Export String we have to write like measure_abc > 0 measure_abc < 1 by dim_xyz so joiner is space
		err := b.writeCondition(subquery.Having.Condition, " ")
		if err != nil {
			return err
		}
		_, err = b.WriteString(" by " + subquery.Dimension.Name)
		return err
	}
	_, err := b.WriteString("<subquery>")
	return err
}

func (b exprBuilder) writeCondition(cond *Condition, joiner string) error {
	switch cond.Operator {
	case OperatorOr:
		if b.forExport {
			return b.writeJoinedExpressions(cond.Expressions, joiner)
		}
		return b.writeJoinedExpressions(cond.Expressions, " OR ")
	case OperatorAnd:
		if b.forExport {
			return b.writeJoinedExpressions(cond.Expressions, joiner)
		}
		return b.writeJoinedExpressions(cond.Expressions, " AND ")
	default:
		if !cond.Operator.Valid() {
			return fmt.Errorf("invalid expression operator %q", cond.Operator)
		}
		return b.writeBinaryCondition(cond.Expressions, cond.Operator, joiner)
	}
}

func (b exprBuilder) writeJoinedExpressions(exprs []*Expression, joiner string) error {
	if len(exprs) == 0 {
		return nil
	}

	for i, e := range exprs {
		if i > 0 {
			b.writeString(joiner)
		}
		err := b.writeWrappedExpression(e, joiner)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b exprBuilder) writeBinaryCondition(exprs []*Expression, op Operator, joiner string) error {
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

	// For Export String if right is Subquery just print the subquery and return
	if b.forExport && right.Subquery != nil {
		return b.writeSubquery(right.Subquery)
	}

	err := b.writeWrappedExpression(left, joiner)
	if err != nil {
		return err
	}

	switch op {
	case OperatorEq:
		// Special case: "dim = NULL" should be written as "dim IS NULL"
		if hasNilValue(right) {
			if b.forExport {
				b.writeString(" = NULL")
				return nil
			}
			b.writeString(" IS NULL")
			return nil
		}
		b.writeString(" = ")
	case OperatorNeq:
		// Special case: "dim != NULL" should be written as "dim IS NOT NULL"
		if hasNilValue(right) {
			if b.forExport {
				b.writeString(" != NULL")
				return nil
			}
			b.writeString(" IS NOT NULL")
			return nil
		}
		b.writeString(" != ")
	case OperatorLt:
		b.writeString(" < ")
	case OperatorLte:
		b.writeString(" <= ")
	case OperatorGt:
		b.writeString(" > ")
	case OperatorGte:
		b.writeString(" >= ")
	case OperatorIn:
		if b.forExport {
			b.writeString(" = ")
		} else {
			b.writeString(" IN ")
		}
	case OperatorNin:
		if b.forExport {
			b.writeString(" != ")
		} else {
			b.writeString(" NOT IN ")
		}
	case OperatorIlike:
		if b.forExport {
			b.writeString(" = ")
		} else {
			b.writeString(" ILIKE ")
		}
	case OperatorNilike:
		if b.forExport {
			b.writeString(" != ")
		} else {
			b.writeString(" NOT ILIKE ")
		}
	default:
		return fmt.Errorf("invalid binary condition operator %q", op)
	}

	err = b.writeWrappedExpression(right, joiner)
	if err != nil {
		return err
	}

	return nil
}

func (b exprBuilder) writeWrappedExpression(e *Expression, joiner string) error {
	if b.forExport {
		return b.writeExpression(e, joiner)
	}
	if e.Condition != nil {
		b.writeByte('(')
	}
	err := b.writeExpression(e, joiner)
	if err != nil {
		return err
	}
	if e.Condition != nil {
		b.writeByte(')')
	}
	return nil
}

func (b exprBuilder) writeByte(v byte) {
	_ = b.WriteByte(v)
}

func (b exprBuilder) writeString(s string) {
	_, _ = b.WriteString(s)
}

func hasNilValue(expr *Expression) bool {
	return expr != nil && expr.Value == nil && expr.Condition == nil && expr.Subquery == nil
}
