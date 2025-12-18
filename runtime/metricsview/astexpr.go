package metricsview

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// SQLForExpression generates a SQL expression for a query expression.
// pseudoHaving is true if the expression is allowed to reference measure expressions.
// visible is true if the expression is only allowed to reference dimensions and measures that are exposed by the security policy.
func (ast *AST) SQLForExpression(e *Expression, n *SelectNode, pseudoHaving, visible bool) (string, []any, error) {
	b := &sqlExprBuilder{
		ast:          ast,
		node:         n,
		pseudoHaving: pseudoHaving,
		visible:      visible,
		out:          &strings.Builder{},
	}

	err := b.writeExpression(e)
	if err != nil {
		return "", nil, err
	}

	return b.out.String(), b.args, nil
}

type sqlExprBuilder struct {
	ast          *AST
	node         *SelectNode
	pseudoHaving bool
	visible      bool
	out          *strings.Builder
	args         []any
}

// writeExpression writes the SQL expression for the given expression.
// The output is guaranteed to be wrapped in parentheses.
func (b *sqlExprBuilder) writeExpression(e *Expression) error {
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

func (b *sqlExprBuilder) writeName(name string) error {
	expr, unnest, _, err := b.sqlForName(name)
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

func (b *sqlExprBuilder) writeValue(val any) error {
	b.writeString("?")
	b.args = append(b.args, val)
	return nil
}

func (b *sqlExprBuilder) writeSubquery(sub *Subquery) error {
	// We construct a Query that combines the parent Query's contextual info with that of the Subquery.
	outer := b.ast.Query
	inner := &Query{
		MetricsView:         outer.MetricsView,
		Dimensions:          []Dimension{sub.Dimension},
		Measures:            sub.Measures,
		PivotOn:             nil,
		Spine:               nil,
		Sort:                nil,
		TimeRange:           outer.TimeRange,
		ComparisonTimeRange: outer.ComparisonTimeRange,
		Where:               sub.Where,
		Having:              sub.Having,
		Limit:               nil,
		Offset:              nil,
		TimeZone:            outer.TimeZone,
		UseDisplayNames:     false,
		Rows:                false,
	} //exhaustruct:enforce

	// Generate SQL for the subquery
	innerSecurity := b.ast.Security
	if !b.visible {
		innerSecurity = skipMetricsViewSecurity{}
	}
	innerAST, err := NewAST(b.ast.MetricsView, innerSecurity, inner, b.ast.Dialect)
	if err != nil {
		return fmt.Errorf("failed to create AST for subquery: %w", err)
	}
	sql, args, err := innerAST.SQL()
	if err != nil {
		return fmt.Errorf("failed to generate SQL for subquery: %w", err)
	}

	// Output: (SELECT <dimension> FROM (<subquery>))
	b.writeString("(SELECT ")
	b.writeString(b.ast.Dialect.EscapeIdentifier(sub.Dimension.Name))
	b.writeString(" FROM (")
	b.writeString(sql)
	b.writeString("))")
	b.args = append(b.args, args...)
	return nil
}

func (b *sqlExprBuilder) writeCondition(cond *Condition) error {
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

func (b *sqlExprBuilder) writeJoinedExpressions(exprs []*Expression, joiner string) error {
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

func (b *sqlExprBuilder) writeBinaryCondition(exprs []*Expression, op Operator) error {
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

	// Check there isn't an unnest on the right side
	if right.Name != "" {
		_, unnest, _, err := b.sqlForName(right.Name)
		if err != nil {
			return err
		}
		if unnest {
			return fmt.Errorf("cannot apply expression to dimension %q because it requires unnesting, which is only supported for the left side of an operation", right.Name)
		}
	}

	// Handle unnest on the left side
	if left.Name != "" {
		leftExpr, unnest, lookup, err := b.sqlForName(left.Name)
		if err != nil {
			return err
		}

		// If not unnested, write the expression as-is or if its a lookup rewrite as per dialect
		if !unnest {
			if lookup != nil {
				b.writeString(fmt.Sprintf("%s IN ", lookup.keyExpr))
				b.writeByte('(')
				ex, err := b.ast.Dialect.LookupSelectExpr(lookup.table, lookup.keyCol)
				if err != nil {
					return err
				}
				b.writeString(ex)
				b.writeString(" WHERE ")
				err = b.writeBinaryConditionInner(nil, right, lookup.valueCol, op)
				if err != nil {
					return err
				}
				b.writeString(")")
				return nil
			}

			return b.writeBinaryConditionInner(nil, right, leftExpr, op)
		}

		// Generate unnest join
		unnestTableAlias := b.ast.GenerateIdentifier()
		unnestFrom, tupleStyle, auto, err := b.ast.Dialect.LateralUnnest(leftExpr, unnestTableAlias, left.Name)
		if err != nil {
			return err
		}
		if auto {
			// Means the DB automatically unnests, so we can treat it as a normal value
			return b.writeBinaryConditionInner(nil, right, leftExpr, op)
		}
		var unnestColAlias string
		if tupleStyle {
			unnestColAlias = b.ast.Dialect.EscapeMember(unnestTableAlias, left.Name)
		} else {
			unnestColAlias = b.ast.Dialect.EscapeIdentifier(left.Name)
		}

		if !tupleStyle { // if tupleStyle, then we cannot refer to the column by table alias
			b.ast.unnests = append(b.ast.unnests, unnestFrom)
			return b.writeBinaryConditionInner(nil, right, unnestColAlias, op)
		}

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
		b.writeByte(')')
		return nil
	}

	// Handle netiher side is a name
	return b.writeBinaryConditionInner(left, right, "", op)
}

func (b *sqlExprBuilder) writeBinaryConditionInner(left, right *Expression, leftOverride string, op Operator) error {
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

	b.writeByte('(')

	if leftOverride != "" {
		b.writeParenthesizedString(leftOverride)
	} else {
		err := b.writeExpression(left)
		if err != nil {
			return err
		}
	}
	if hasNilValue(right) {
		// Special cases:
		// "dim = NULL" should be written as "dim IS NULL"
		// "dim != NULL" should be written as "dim IS NOT NULL"
		if op == OperatorEq {
			b.writeString(" IS NULL)")
			return nil
		} else if op == OperatorNeq {
			b.writeString(" IS NOT NULL)")
			return nil
		}
	}
	b.writeString(joiner)
	err := b.writeExpression(right)
	if err != nil {
		return err
	}

	b.writeByte(')')

	return nil
}

func (b *sqlExprBuilder) writeILikeCondition(left, right *Expression, leftOverride string, not bool) error {
	b.writeByte('(')

	if b.ast.Dialect.SupportsILike() {
		// Output: <left> [NOT] ILIKE <right>

		if leftOverride != "" {
			b.writeParenthesizedString(leftOverride)
		} else {
			err := b.writeExpression(left)
			if err != nil {
				return err
			}
		}

		if b.ast.Dialect.RequiresCastForLike() {
			b.writeString("::TEXT")
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

		if b.ast.Dialect.RequiresCastForLike() {
			b.writeString("::TEXT")
		}
	} else if b.ast.Dialect.SupportsRegexMatch() {
		if not {
			b.writeString(" NOT ")
		}
		b.writeString(b.ast.Dialect.GetRegexMatchFunction())
		b.writeByte('(')
		if leftOverride != "" {
			b.writeParenthesizedString(leftOverride)
		} else {
			err := b.writeExpression(left)
			if err != nil {
				return err
			}
		}
		b.writeString(", ")
		expr, err := convertLikeExpressionToRegexExpression(right)
		if err != nil {
			return fmt.Errorf("failed to convert LIKE expression to regex pattern: %w", err)
		}
		err = b.writeExpression(expr)
		if err != nil {
			return fmt.Errorf("failed to write regex pattern expression: %w", err)
		}
		b.writeByte(')')
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

		b.writeByte(')')

		return nil
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
		if b.ast.Dialect.RequiresCastForLike() {
			b.writeString("::TEXT")
		}
		b.writeByte(')')

		if not {
			b.writeString(" NOT LIKE ")
		} else {
			b.writeString(" LIKE ")
		}

		b.writeString("LOWER(")
		err := b.writeExpression(right)
		if err != nil {
			return err
		}
		if b.ast.Dialect.RequiresCastForLike() {
			b.writeString("::TEXT")
		}
		b.writeByte(')')
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

	b.writeByte(')')

	return nil
}

func (b *sqlExprBuilder) writeInCondition(left, right *Expression, leftOverride string, not bool) error {
	if right.Value != nil {
		vals, ok := right.Value.([]any)
		if !ok {
			return fmt.Errorf("the right value must be a list of values for an IN condition")
		}

		return b.writeInConditionForValues(left, leftOverride, vals, not)
	}

	b.writeByte('(')

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

	err := b.writeExpression(right)
	if err != nil {
		return err
	}

	b.writeByte(')')

	return nil
}

func (b *sqlExprBuilder) writeInConditionForValues(left *Expression, leftOverride string, vals []any, not bool) error {
	var hasNull, hasNonNull bool
	for _, v := range vals {
		if v == nil {
			hasNull = true
		} else {
			hasNonNull = true
		}
		if hasNull && hasNonNull {
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

	b.writeByte('(')

	if hasNonNull {
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
		var comma bool
		for _, val := range vals {
			if val == nil {
				continue
			}
			if comma {
				b.writeString(",?")
			} else {
				comma = true
				b.writeString("?")
			}
			b.args = append(b.args, val)
		}
		b.writeByte(')')
	}

	if hasNull {
		if hasNonNull {
			if not {
				b.writeString(" AND ")
			} else {
				b.writeString(" OR ")
			}
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

	b.writeByte(')')

	return nil
}

func (b *sqlExprBuilder) writeByte(v byte) {
	_ = b.out.WriteByte(v)
}

func (b *sqlExprBuilder) writeString(s string) {
	_, _ = b.out.WriteString(s)
}

func (b *sqlExprBuilder) writeParenthesizedString(s string) {
	_ = b.out.WriteByte('(')
	_, _ = b.out.WriteString(s)
	_ = b.out.WriteByte(')')
}

func (b *sqlExprBuilder) sqlForName(name string) (expr string, unnest bool, lookup *lookupMeta, err error) {
	// If node is nil, we are evaluating the expression against the underlying table.
	// In this case, we only allow filters to reference dimension names.
	if b.node == nil {
		// First, search for the dimension in the ASTs dimension fields (this also covers any computed dimension)
		for _, f := range b.ast.dimFields {
			if f.Name == name {
				// Note that we return "false" even though it may be an unnest dimension because it will already have been unnested since it's one of the dimensions included in the query.
				// So we can filter against it as if it's a normal dimension.
				return f.Expr, false, nil, nil
			}
		}

		// Second, search for the dimension in the metrics view's dimensions (since expressions are allowed to reference dimensions not included in the query)
		dim, err := b.ast.LookupDimension(name, b.visible)
		if err != nil {
			return "", false, nil, fmt.Errorf("invalid dimension reference %q: %w", name, err)
		}

		ex, err := b.ast.Dialect.MetricsViewDimensionExpression(dim)
		if err != nil {
			return "", false, nil, fmt.Errorf("invalid dimension reference %q: %w", name, err)
		}

		if dim.Unnest && dim.LookupTable != "" {
			return "", false, nil, fmt.Errorf("dimension %q is unnested and also has a lookup. This is not supported", name)
		}

		var lm *lookupMeta
		if dim.LookupTable != "" {
			var keyExpr string
			if dim.Column != "" {
				keyExpr = b.ast.Dialect.EscapeIdentifier(dim.Column)
			} else if dim.Expression != "" {
				keyExpr = dim.Expression
			} else {
				return "", false, nil, fmt.Errorf("dimension %q has a lookup table but no column or expression defined", name)
			}
			lm = &lookupMeta{
				table:    dim.LookupTable,
				keyExpr:  keyExpr,
				keyCol:   dim.LookupKeyColumn,
				valueCol: dim.LookupValueColumn,
			}
		}
		// Note: If dim.Unnest is true, we need to unnest it inside of the generated expression (because it's not part of the dimFields and therefore not unnested with a LATERAL JOIN).
		return ex, dim.Unnest, lm, nil
	}

	// Since node is not nil, we're in the context of a wrapped SELECT.
	// We only allow expressions against the node's dimensions and measures (not those in scope within sub-queries).

	// Check if it's a dimension name
	for _, f := range b.node.DimFields {
		if f.Name == name {
			// NOTE: We don't need to handle Unnest here because it's always applied at the innermost query (i.e. when node==nil).
			return f.Expr, false, nil, nil
		}
	}

	// Can't have expressions against a measure field unless it's a pseudo-HAVING clause (pseudo because we currently output it as a WHERE in an outer SELECT)
	if !b.pseudoHaving {
		return "", false, nil, fmt.Errorf("name %q in expression is not a dimension available in the current context", name)
	}

	// Check measure fields
	for _, f := range b.node.MeasureFields {
		if f.Name == name {
			return f.Expr, false, nil, nil
		}
	}

	return "", false, nil, fmt.Errorf("name %q in expression is not a dimension or measure available in the current context", name)
}

func convertLikeExpressionToRegexExpression(like *Expression) (*Expression, error) {
	val, ok := like.Value.(string)
	if !ok {
		return nil, fmt.Errorf("the pattern expression for regex match function must be a string value, got %T", like.Value)
	}
	// convert pattern to a case insensitive regex match pattern, e.g. "%foo%" becomes "^(?i).*foo.*$"
	pattern := strings.ReplaceAll(val, "%", ".*")
	pattern = fmt.Sprintf("^(?i)%s$", pattern)
	return &Expression{Value: pattern}, nil
}

type lookupMeta struct {
	table    string
	keyExpr  string
	keyCol   string
	valueCol string
}

// skipMetricsViewSecurity implements the MetricsViewSecurity interface in a way that allows all access.
type skipMetricsViewSecurity struct{}

var _ MetricsViewSecurity = skipMetricsViewSecurity{}

func (s skipMetricsViewSecurity) CanAccessField(field string) bool {
	return true
}

func (s skipMetricsViewSecurity) RowFilter() string {
	return ""
}

func (s skipMetricsViewSecurity) QueryFilter() *runtimev1.Expression {
	return nil
}
