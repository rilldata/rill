package metricssql

import (
	"context"
	"fmt"
	"time"

	"github.com/itlightning/dateparse"
	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/rilldata/rill/runtime/metricsview"
)

func ParseFilter(sql string) (*metricsview.Expression, error) {
	p := parser.New()
	p.SetSQLMode(mysql.ModeANSI | mysql.ModeANSIQuotes)
	sql = "SELECT * FROM tbl WHERE " + sql
	stmtNodes, _, err := p.ParseSQL(sql)
	if err != nil {
		return nil, err
	}

	if len(stmtNodes) != 1 {
		return nil, fmt.Errorf("invalid sql filter")
	}

	stmt, ok := stmtNodes[0].(*ast.SelectStmt)
	if !ok {
		return nil, fmt.Errorf("invalid sql filter")
	}
	return parseFilter(context.Background(), stmt.Where, nil, nil)
}

// note - context is optional, it is used to resolve time functions like time_range_start and time_range_end as they require context in which they are executed
func parseFilter(ctx context.Context, node, context ast.ExprNode, q *query) (*metricsview.Expression, error) {
	switch node := node.(type) {
	case *ast.ColumnNameExpr:
		col, err := parseColumnNameExpr(node)
		if err != nil {
			return nil, err
		}
		return &metricsview.Expression{
			Name: col,
		}, nil
	case *ast.BinaryOperationExpr:
		return parseBinaryOperation(ctx, node, q)
	case ast.ValueExpr:
		val, err := parseValueExpr(node)
		if err != nil {
			return nil, err
		}
		return &metricsview.Expression{
			Value: val,
		}, nil
	case *ast.IsNullExpr:
		return parseIsNullOperation(ctx, node, q)
	case *ast.IsTruthExpr:
		return parseIsTruthOperation(ctx, node, q)
	case *ast.ParenthesesExpr:
		return parseParentheses(ctx, node, q)
	case *ast.PatternInExpr:
		return parsePatternIn(ctx, node, q)
	case *ast.PatternLikeOrIlikeExpr:
		return parsePatternLikeOrIlike(ctx, node, q)
	case *ast.BetweenExpr:
		return parseBetween(ctx, node, q)
	case *ast.FuncCallExpr:
		return parseFuncCallInFilter(ctx, node, context, q)
	default:
		return nil, fmt.Errorf("metrics sql: unsupported type %T, expression %q", node, restore(node))
	}
}

func parseColumnNameExpr(in ast.Node) (string, error) {
	node, ok := in.(*ast.ColumnNameExpr)
	if !ok {
		return "", fmt.Errorf("metrics sql: expected column name expression")
	}
	if node.Name == nil {
		return "", fmt.Errorf("metrics sql: can only have dimension/measure name(s) in select list")
	}
	if node.Name.Schema.String() != "" || node.Name.Table.String() != "" {
		return "", fmt.Errorf("metrics sql: no alias or table reference is supported in column name. Found in `%s`", node.Name.String())
	}

	return node.Name.Name.O, nil
}

func parseBinaryOperation(ctx context.Context, node *ast.BinaryOperationExpr, q *query) (*metricsview.Expression, error) {
	left, err := parseFilter(ctx, node.L, node.R, q)
	if err != nil {
		return nil, err
	}

	right, err := parseFilter(ctx, node.R, node.L, q)
	if err != nil {
		return nil, err
	}

	return &metricsview.Expression{
		Condition: &metricsview.Condition{
			// The validation for allowed operators will be done by underlying AST builder
			Operator:    operator(node.Op),
			Expressions: []*metricsview.Expression{left, right},
		},
	}, nil
}

func parseIsNullOperation(ctx context.Context, node *ast.IsNullExpr, q *query) (*metricsview.Expression, error) {
	expr, err := parseFilter(ctx, node.Expr, nil, q)
	if err != nil {
		return nil, err
	}

	var op metricsview.Operator
	if node.Not {
		op = metricsview.OperatorNeq
	} else {
		op = metricsview.OperatorEq
	}
	return &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator: op,
			Expressions: []*metricsview.Expression{
				expr,
				{Value: nil},
			},
		},
	}, nil
}

func parseIsTruthOperation(ctx context.Context, node *ast.IsTruthExpr, q *query) (*metricsview.Expression, error) {
	expr, err := parseFilter(ctx, node.Expr, nil, q)
	if err != nil {
		return nil, err
	}

	var op metricsview.Operator
	if node.Not {
		op = metricsview.OperatorNeq
	} else {
		op = metricsview.OperatorEq
	}
	return &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator: op,
			Expressions: []*metricsview.Expression{
				expr,
				{Value: "TRUE"},
			},
		},
	}, nil
}

func parseParentheses(ctx context.Context, node *ast.ParenthesesExpr, q *query) (*metricsview.Expression, error) {
	expr, err := parseFilter(ctx, node.Expr, nil, q)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func parsePatternIn(ctx context.Context, node *ast.PatternInExpr, q *query) (*metricsview.Expression, error) {
	expr, err := parseFilter(ctx, node.Expr, nil, q)
	if err != nil {
		return nil, err
	}

	var op metricsview.Operator
	if node.Not {
		op = metricsview.OperatorNin
	} else {
		op = metricsview.OperatorIn
	}

	var right *metricsview.Expression
	if node.Sel != nil {
		subquery, err := parseInSubquery(ctx, node.Sel, q)
		if err != nil {
			return nil, err
		}
		right = &metricsview.Expression{Subquery: subquery}
	} else {
		values := make([]any, 0, len(node.List))
		for _, n := range node.List {
			val, err := parseValueExpr(n)
			if err != nil {
				return nil, err
			}
			values = append(values, val)
		}
		right = &metricsview.Expression{Value: values}
	}

	return &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator:    op,
			Expressions: []*metricsview.Expression{expr, right},
		},
	}, nil
}

func parseInSubquery(ctx context.Context, n ast.ExprNode, q *query) (*metricsview.Subquery, error) {
	// Extract the subquery SELECT statement
	subquery, ok := n.(*ast.SubqueryExpr)
	if !ok {
		return nil, fmt.Errorf("metrics sql: subquery expression type %T not supported", n)
	}
	sel, ok := subquery.Query.(*ast.SelectStmt)
	if !ok {
		return nil, fmt.Errorf("metrics sql: subquery type %T is not supported", subquery.Query)
	}

	// You can do a lot of stuff in a SELECT statement. Check it doesn't do anything we don't support.
	switch {
	case sel.Kind != ast.SelectStmtKindSelect:
		return nil, fmt.Errorf("metrics sql: subquery of kind %s is not supported", sel.Kind.String())
	case sel.SelectStmtOpts != nil && sel.SelectStmtOpts.Distinct:
		return nil, fmt.Errorf("metrics sql: subquery with DISTINCT is not supported")
	case len(sel.WindowSpecs) > 0:
		return nil, fmt.Errorf("metrics sql: subquery with window specifications is not supported")
	case sel.OrderBy != nil:
		return nil, fmt.Errorf("metrics sql: subquery with ORDER BY is not supported")
	case sel.Limit != nil:
		return nil, fmt.Errorf("metrics sql: subquery with LIMIT is not supported")
	case sel.LockInfo != nil:
		return nil, fmt.Errorf("metrics sql: subquery with lock info is not supported")
	case len(sel.TableHints) > 0:
		return nil, fmt.Errorf("metrics sql: subquery with table hints is not supported")
	case sel.IsInBraces:
		return nil, fmt.Errorf("metrics sql: subquery with braces is not supported")
	case sel.WithBeforeBraces:
		return nil, fmt.Errorf("metrics sql: subquery with WITH before braces is not supported")
	case sel.QueryBlockOffset != 0:
		return nil, fmt.Errorf("metrics sql: subquery with query block offset is not supported")
	case sel.SelectIntoOpt != nil:
		return nil, fmt.Errorf("metrics sql: subquery with SELECT INTO is not supported")
	case sel.AfterSetOperator != nil:
		return nil, fmt.Errorf("metrics sql: subquery with set operations is not supported")
	case len(sel.Lists) > 0:
		return nil, fmt.Errorf("metrics sql: subquery with row expressions is not supported")
	case sel.With != nil:
		return nil, fmt.Errorf("metrics sql: subquery with WITH clause is not supported")
	case sel.AsViewSchema:
		return nil, fmt.Errorf("metrics sql: subquery as view schema is not supported")
	}

	// Validate the FROM clause is a plain `FROM metrics_view`
	if sel.From == nil || sel.From.TableRefs == nil {
		return nil, fmt.Errorf("metrics sql: subquery must have a FROM clause")
	}
	if sel.From.TableRefs.Right != nil {
		return nil, fmt.Errorf("metrics sql: subquery with JOIN is not supported")
	}
	tblSrc, ok := sel.From.TableRefs.Left.(*ast.TableSource)
	if !ok {
		return nil, fmt.Errorf("metrics sql: subquery must select from a metrics view")
	}
	tbl, ok := tblSrc.Source.(*ast.TableName)
	if !ok {
		return nil, fmt.Errorf("metrics sql: subquery must select from a metrics view")
	}
	tblName := tbl.Name.String()
	if q != nil && q.q.MetricsView != tblName {
		return nil, fmt.Errorf("metrics sql: subquery must select from the metrics view %q", q.q.MetricsView)
	}

	// Parse the selected dimension
	if sel.Fields == nil || len(sel.Fields.Fields) != 1 {
		return nil, fmt.Errorf("metrics sql: subquery must select exactly one dimension (note: you can still reference other measures in the HAVING expression)")
	}
	field := sel.Fields.Fields[0]
	if field.WildCard != nil || field.Expr == nil {
		return nil, fmt.Errorf("metrics sql: subquery must select a specific field, not a wildcard")
	}
	dim, err := parseColumnNameExpr(field.Expr)
	if err != nil {
		return nil, err
	}

	// GROUP BY is optional, but if provided, check it matches the dimension
	if sel.GroupBy != nil {
		if sel.GroupBy.Rollup {
			return nil, fmt.Errorf("metrics sql: subquery with ROLLUP in GROUP BY is not supported")
		}
		if len(sel.GroupBy.Items) != 1 {
			return nil, fmt.Errorf("metrics sql: subquery must group by exactly one dimension")
		}
		groupByName, err := parseColumnNameExpr(sel.GroupBy.Items[0])
		if err != nil {
			return nil, fmt.Errorf("metrics sql: failed to parse GROUP BY expression in subquery: %w", err)
		}
		if groupByName != dim {
			return nil, fmt.Errorf("metrics sql: subquery must group by the same dimension %q as the subquery's SELECT clause", dim)
		}
	}

	// Parse the WHERE and HAVING clauses
	var where, having *metricsview.Expression
	if sel.Where != nil {
		var err error
		where, err = parseFilter(ctx, sel.Where, nil, q)
		if err != nil {
			return nil, err
		}
	}
	if sel.Having != nil && sel.Having.Expr != nil {
		var err error
		having, err = parseFilter(ctx, sel.Having.Expr, nil, q)
		if err != nil {
			return nil, err
		}
	}

	// Extract measures from the HAVING clause.
	//
	// NOTE: This makes a couple assumptions that are currently viable, but may not hold in the future.
	// Specifically that a) computed measures aren't used here, b) dimensions are not referenced in the HAVING.
	// In the future, we may want to move this into the handling of HAVING inside the metricsview package, so that you don't need to select measures that are only used in HAVING.
	var measures []metricsview.Measure
	if having != nil {
		fields := metricsview.AnalyzeExpressionFields(having)
		for _, field := range fields {
			measures = append(measures, metricsview.Measure{Name: field})
		}
	}

	// Success
	return &metricsview.Subquery{
		Dimension: metricsview.Dimension{Name: dim},
		Measures:  measures,
		Where:     where,
		Having:    having,
	}, nil
}

func parsePatternLikeOrIlike(ctx context.Context, n *ast.PatternLikeOrIlikeExpr, q *query) (*metricsview.Expression, error) {
	if string(n.Escape) != "\\" {
		// druid supports it, duckdb and clickhouse do not
		return nil, fmt.Errorf("metrics sql: `ESCAPE` is not supported")
	}

	expr, err := parseFilter(ctx, n.Expr, nil, q)
	if err != nil {
		return nil, err
	}

	var op metricsview.Operator
	if n.Not {
		op = metricsview.OperatorNilike
	} else {
		op = metricsview.OperatorIlike
	}

	pattern, err := parseValueExpr(n.Pattern)
	if err != nil {
		return nil, err
	}

	return &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator: op,
			Expressions: []*metricsview.Expression{
				expr,
				{Value: pattern},
			},
		},
	}, nil
}

func parseBetween(ctx context.Context, n *ast.BetweenExpr, q *query) (*metricsview.Expression, error) {
	expr, err := parseFilter(ctx, n.Expr, nil, q)
	if err != nil {
		return nil, err
	}

	left, err := parseFilter(ctx, n.Left, n.Expr, q)
	if err != nil {
		return nil, err
	}

	right, err := parseFilter(ctx, n.Right, n.Expr, q)
	if err != nil {
		return nil, err
	}
	return &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator: metricsview.OperatorAnd,
			Expressions: []*metricsview.Expression{
				{
					Condition: &metricsview.Condition{
						Operator:    metricsview.OperatorGte,
						Expressions: []*metricsview.Expression{expr, left},
					},
				},
				{
					Condition: &metricsview.Condition{
						Operator:    metricsview.OperatorLte,
						Expressions: []*metricsview.Expression{expr, right},
					},
				},
			},
		},
	}, nil
}

func parseValueExpr(in ast.Node) (any, error) {
	// Extract underlying value
	node, ok := in.(ast.ValueExpr)
	if !ok {
		return "", fmt.Errorf("metrics sql: expected value expression, got %T", in)
	}
	val := node.GetValue()

	// Handle a couple types that we prefer simplified
	switch actual := val.(type) {
	case int64:
		val = int(actual) // Cast to plain int
	case uint64:
		val = int(actual) // Cast to plain int
	}

	return val, nil
}

func parseTimeUnitValueExpr(in ast.Node) (string, error) {
	node, ok := in.(*ast.TimeUnitExpr)
	if !ok {
		return "", fmt.Errorf("metrics sql: expected time_unit value expression, got %T", in)
	}
	return node.Unit.String(), nil
}

func parseFuncCallInFilter(ctx context.Context, node *ast.FuncCallExpr, context ast.ExprNode, q *query) (*metricsview.Expression, error) {
	switch node.FnName.L {
	case "time_range_start":
		if q == nil {
			return nil, fmt.Errorf("metrics sql: time_range_start function is only supported for metrics_sql")
		}
		timeDim, ok := context.(*ast.ColumnNameExpr)
		if !ok {
			return nil, fmt.Errorf("metrics sql: time_range_start function requires a valid time dimension")
		}
		return q.parseTimeRangeStart(ctx, node, timeDim)
	case "time_range_end":
		if q == nil {
			return nil, fmt.Errorf("metrics sql: time_range_end function is only supported for metrics_sql")
		}
		timeDim, ok := context.(*ast.ColumnNameExpr)
		if !ok {
			return nil, fmt.Errorf("metrics sql: time_range_end function requires a valid time dimension")
		}
		return q.parseTimeRangeEnd(ctx, node, timeDim)
	case "now":
		return &metricsview.Expression{
			Value: time.Now().Format(time.RFC3339),
		}, nil
	case "date_add", "date_sub": // ex : date_add(time, INTERVAL x UNIT)
		val, err := parseFilter(ctx, node.Args[0], nil, q) // handling of time
		if err != nil {
			return nil, err
		}
		timeStr, ok := val.Value.(string)
		if !ok {
			return nil, fmt.Errorf("metrics sql: expected time value in date_add/date_sub function")
		}
		t, err := dateparse.ParseAny(timeStr)
		if err != nil {
			return nil, err
		}

		expr, err := parseValueExpr(node.Args[1]) // handling of x
		if err != nil {
			return nil, err
		}
		amt, ok := expr.(int)
		if !ok {
			return nil, fmt.Errorf("metrics sql: expected integer value in date_add/date_sub function")
		}

		timeUnit, err := parseTimeUnitValueExpr(node.Args[2]) // handling of DAY
		if err != nil {
			return nil, err
		}

		var res time.Time
		if node.FnName.L == "date_add" {
			res, err = add(t, timeUnit, amt)
		} else {
			res, err = sub(t, timeUnit, amt)
		}
		if err != nil {
			return nil, err
		}
		return &metricsview.Expression{
			Value: res.Format(time.RFC3339),
		}, nil
	default:
		return nil, fmt.Errorf("metrics sql: function `%s` not supported in where clause", node.FnName.L)
	}
}
