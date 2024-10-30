package metricssqlparser

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/itlightning/dateparse"
	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/format"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/rilldata/rill/runtime/metricsview"
)

func ParseSQLFilter(sql string) (*metricsview.Expression, error) {
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
	return parseFilter(context.Background(), stmt.Where, nil)
}

func parseFilter(ctx context.Context, node ast.ExprNode, q *query) (*metricsview.Expression, error) {
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
		return parseFuncCallInFilter(ctx, node, q)
	default:
		return nil, fmt.Errorf("metrics sql: unsupported expression %q", restore(node))
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
	left, err := parseFilter(ctx, node.L, q)
	if err != nil {
		return nil, err
	}

	right, err := parseFilter(ctx, node.R, q)
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
	expr, err := parseFilter(ctx, node.Expr, q)
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
	expr, err := parseFilter(ctx, node.Expr, q)
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
	expr, err := parseFilter(ctx, node.Expr, q)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func parsePatternIn(ctx context.Context, node *ast.PatternInExpr, q *query) (*metricsview.Expression, error) {
	if node.Sel != nil {
		return nil, fmt.Errorf("metrics sql: sub_query is not supported")
	}

	expr, err := parseFilter(ctx, node.Expr, q)
	if err != nil {
		return nil, err
	}

	var op metricsview.Operator
	if node.Not {
		op = metricsview.OperatorNin
	} else {
		op = metricsview.OperatorIn
	}
	values := make([]any, 0, len(node.List))
	for _, n := range node.List {
		val, err := parseValueExpr(n)
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}
	return &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator: op,
			Expressions: []*metricsview.Expression{
				expr,
				{Value: values},
			},
		},
	}, nil
}

func parsePatternLikeOrIlike(ctx context.Context, n *ast.PatternLikeOrIlikeExpr, q *query) (*metricsview.Expression, error) {
	if string(n.Escape) != "\\" {
		// druid supports it, duckdb and clickhouse do not
		return nil, fmt.Errorf("metrics sql: `ESCAPE` is not supported")
	}

	expr, err := parseFilter(ctx, n.Expr, q)
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
	expr, err := parseFilter(ctx, n.Expr, q)
	if err != nil {
		return nil, err
	}

	left, err := parseFilter(ctx, n.Left, q)
	if err != nil {
		return nil, err
	}

	right, err := parseFilter(ctx, n.Right, q)
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

func parseValueExpr(in ast.Node) (string, error) {
	node, ok := in.(ast.ValueExpr)
	if !ok {
		return "", fmt.Errorf("metrics sql: expected value expression, got %T", in)
	}
	var sb strings.Builder
	rctx := format.NewRestoreCtx(format.RestoreNameBackQuotes|format.RestoreStringWithoutCharset, &sb)
	if err := node.Restore(rctx); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func parseTimeUnitValueExpr(in ast.Node) (string, error) {
	node, ok := in.(*ast.TimeUnitExpr)
	if !ok {
		return "", fmt.Errorf("metrics sql: expected time_unit value expression, got %T", in)
	}
	return node.Unit.String(), nil
}

func parseFuncCallInFilter(ctx context.Context, node *ast.FuncCallExpr, q *query) (*metricsview.Expression, error) {
	switch node.FnName.L {
	case "time_range_start":
		if q == nil {
			return nil, fmt.Errorf("metrics sql: time_range_start function is only supported for metrics_sql")
		}
		return q.parseTimeRangeStart(ctx, node)
	case "time_range_end":
		if q == nil {
			return nil, fmt.Errorf("metrics sql: time_range_end function is only supported for metrics_sql")
		}
		return q.parseTimeRangeEnd(ctx, node)
	case "now":
		return &metricsview.Expression{
			Value: time.Now().Format(time.RFC3339),
		}, nil
	case "date_add", "date_sub": // ex : date_add(time, INTERVAL x UNIT)
		val, err := parseFilter(ctx, node.Args[0], q) // handling of time
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
		amt, err := strconv.Atoi(expr)
		if err != nil {
			return nil, fmt.Errorf("metrics sql: expected integer value in date_add/date_sub function")
		}

		expr, err = parseTimeUnitValueExpr(node.Args[2]) // handling of DAY
		if err != nil {
			return nil, err
		}

		var res time.Time
		if node.FnName.L == "date_add" {
			res, err = add(t, expr, amt)
		} else {
			res, err = sub(t, expr, amt)
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
