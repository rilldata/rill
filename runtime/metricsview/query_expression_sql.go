package metricsview

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ExpressionDialect controls the output format of ExpressionToSQL.
type ExpressionDialect int

const (
	// DialectJSON uses JSON-style formatting (double quotes for strings, square brackets for arrays).
	// This is the default for backwards compatibility.
	DialectJSON ExpressionDialect = iota
	// DialectDuckDB uses DuckDB SQL syntax (single quotes for strings, parentheses for IN lists).
	DialectDuckDB
)

// ExpressionToSQL converts an expression to a SQL string.
// The dialect parameter controls the output format. If not provided, defaults to DialectJSON for backwards compatibility.
func ExpressionToSQL(e *Expression, dialect ...ExpressionDialect) (string, error) {
	d := DialectJSON
	if len(dialect) > 0 {
		d = dialect[0]
	}
	b := exprSQLBuilder{Builder: &strings.Builder{}, dialect: d}
	err := b.writeExpression(e)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

type exprSQLBuilder struct {
	*strings.Builder
	dialect ExpressionDialect
}

func (b exprSQLBuilder) writeExpression(e *Expression) error {
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

func (b exprSQLBuilder) writeName(name string) error {
	if strings.Contains(name, `"`) {
		_, err := strings.NewReplacer(`"`, `""`).WriteString(b.Builder, name)
		return err
	}
	b.writeString(name)
	return nil
}

func (b exprSQLBuilder) writeValue(val any) error {
	if b.dialect == DialectDuckDB {
		return b.writeSQLValue(val)
	}
	// Default: JSON format
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

func (b exprSQLBuilder) writeSQLValue(val any) error {
	if val == nil {
		b.writeString("NULL")
		return nil
	}

	switch v := val.(type) {
	case string:
		// Escape single quotes by doubling them
		escaped := strings.ReplaceAll(v, "'", "''")
		b.writeByte('\'')
		b.writeString(escaped)
		b.writeByte('\'')
		return nil
	case bool:
		if v {
			b.writeString("TRUE")
		} else {
			b.writeString("FALSE")
		}
		return nil
	case []any:
		b.writeByte('(')
		for i, item := range v {
			if i > 0 {
				b.writeString(", ")
			}
			if err := b.writeSQLValue(item); err != nil {
				return err
			}
		}
		b.writeByte(')')
		return nil
	default:
		// For numbers and other types, use reflection to handle slices
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Slice {
			b.writeByte('(')
			for i := 0; i < rv.Len(); i++ {
				if i > 0 {
					b.writeString(", ")
				}
				if err := b.writeSQLValue(rv.Index(i).Interface()); err != nil {
					return err
				}
			}
			b.writeByte(')')
			return nil
		}
		// For numbers and other primitives, use fmt
		_, err := fmt.Fprintf(b.Builder, "%v", val)
		return err
	}
}

func (b exprSQLBuilder) writeSubquery(s *Subquery) error {
	_, err := b.WriteString("(SELECT ")
	if err != nil {
		return err
	}
	_, err = b.WriteString(s.Dimension.Name)
	if err != nil {
		return err
	}
	_, err = b.WriteString(" FROM metrics_view") // We don't know the actual metrics view name
	if err != nil {
		return err
	}
	if s.Where != nil {
		_, err := b.WriteString(" WHERE ")
		if err != nil {
			return err
		}
		err = b.writeExpression(s.Where)
		if err != nil {
			return err
		}
	}
	if s.Having != nil {
		_, err := b.WriteString(" HAVING ")
		if err != nil {
			return err
		}
		err = b.writeExpression(s.Having)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(")")
	return err
}

func (b exprSQLBuilder) writeCondition(cond *Condition) error {
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

func (b exprSQLBuilder) writeJoinedExpressions(exprs []*Expression, joiner string) error {
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

func (b exprSQLBuilder) writeBinaryCondition(exprs []*Expression, op Operator) error {
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
		b.writeString(" = ")
	case OperatorNeq:
		// Special case: "dim != NULL" should be written as "dim IS NOT NULL"
		if hasNilValue(right) {
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

func (b exprSQLBuilder) writeWrappedExpression(e *Expression) error {
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

func (b exprSQLBuilder) writeByte(v byte) {
	_ = b.WriteByte(v)
}

func (b exprSQLBuilder) writeString(s string) {
	_, _ = b.WriteString(s)
}

func hasNilValue(expr *Expression) bool {
	return expr != nil && expr.Value == nil && expr.Condition == nil && expr.Subquery == nil
}
