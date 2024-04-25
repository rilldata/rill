package metricssqlparser

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duration"
)

func (t *transformer) transformTimeRangeStart(ctx context.Context, node *ast.FuncCallExpr) (exprResult, error) {
	if len(node.Args) == 0 {
		return exprResult{}, fmt.Errorf("metrics sql: mandatory arg duration missing for time_range_start() function")
	}
	if len(node.Args) > 3 {
		return exprResult{}, fmt.Errorf("metrics sql: time_range_start() function expects at most 3 arguments")
	}
	// identify optional unit and colName args
	var unit int
	var colName string

	// identify column name
	if len(node.Args) == 3 {
		expr, err := t.transformExprNode(ctx, node.Args[2])
		if err != nil {
			return exprResult{}, err
		}
		var ok bool
		colName, ok = t.dimsToExpr[expr.expr]
		if !ok {
			return exprResult{}, fmt.Errorf("referenced columns %q is not a valid column", expr.expr)
		}
	}

	// identify unit
	if len(node.Args) == 1 {
		unit = 1
	} else {
		expr, err := t.transformExprNode(ctx, node.Args[1])
		if err != nil {
			return exprResult{}, err
		}
		i, err := strconv.ParseInt(expr.expr, 10, 64)
		if err != nil {
			return exprResult{}, err
		}
		unit = int(i)
	}

	expr, err := t.transformExprNode(ctx, node.Args[0])
	if err != nil {
		return exprResult{}, err
	}

	d, err := duration.ParseISO8601(strings.TrimSuffix(strings.TrimPrefix(expr.expr, "'"), "'"))
	if err != nil {
		return exprResult{}, fmt.Errorf("metrics sql: invalid ISO8601 duration %s", expr.expr)
	}

	watermark, col, err := t.getWatermark(ctx, colName)
	if err != nil {
		return exprResult{}, err
	}

	for i := 1; i <= unit; i++ {
		watermark = d.Sub(watermark)
	}
	return exprResult{
		expr:    fmt.Sprintf("'%s'", watermark.Format(time.RFC3339)),
		columns: []string{col},
		types:   []string{"DIMENSION"},
	}, nil
}

func (t *transformer) transformTimeRangeEnd(ctx context.Context, node *ast.FuncCallExpr) (exprResult, error) {
	if len(node.Args) == 0 {
		return exprResult{}, fmt.Errorf("metrics sql: mandatory arg duration missing for time_range_end() function")
	}
	if len(node.Args) > 2 {
		return exprResult{}, fmt.Errorf("metrics sql: time_range_end() function expects at most 2 arguments")
	}
	// identify optional colName arg
	var colName string
	if len(node.Args) == 2 {
		expr, err := t.transformExprNode(ctx, node.Args[1])
		if err != nil {
			return exprResult{}, err
		}
		var ok bool
		colName, ok = t.dimsToExpr[expr.expr]
		if !ok {
			return exprResult{}, fmt.Errorf("referenced columns %q is not a valid column", expr.expr)
		}
	}

	expr, err := t.transformExprNode(ctx, node.Args[0])
	if err != nil {
		return exprResult{}, err
	}

	d, err := duration.ParseISO8601(strings.TrimSuffix(strings.TrimPrefix(expr.expr, "'"), "'"))
	if err != nil {
		return exprResult{}, fmt.Errorf("metrics sql: invalid ISO8601 duration %s", expr.expr)
	}

	watermark, col, err := t.getWatermark(ctx, colName)
	if err != nil {
		return exprResult{}, err
	}

	var end time.Time
	if std, ok := d.(duration.StandardDuration); ok {
		end = std.EndTime(watermark)
	} else {
		end = watermark
	}

	return exprResult{
		expr:    fmt.Sprintf("'%s'", end.Format(time.RFC3339)),
		columns: []string{col},
		types:   []string{"DIMENSION"},
	}, nil
}

func (t *transformer) transformTimeRange(ctx context.Context, node *ast.FuncCallExpr) (exprResult, error) {
	if len(node.Args) == 0 {
		return exprResult{}, fmt.Errorf("metrics sql: mandatory arg duration missing for time_range() function")
	}
	if len(node.Args) > 3 {
		return exprResult{}, fmt.Errorf("metrics sql: time_range_end() function expects at most 3 arguments")
	}
	// identify optional unit and colName args
	var unit int
	var colName string

	// identify column name
	if len(node.Args) == 3 {
		expr, err := t.transformExprNode(ctx, node.Args[2])
		if err != nil {
			return exprResult{}, err
		}
		var ok bool
		colName, ok = t.dimsToExpr[expr.expr]
		if !ok {
			return exprResult{}, fmt.Errorf("referenced columns %q is not a valid column", expr.expr)
		}
	}

	// identify unit
	if len(node.Args) == 1 {
		unit = 1
	} else {
		expr, err := t.transformExprNode(ctx, node.Args[1])
		if err != nil {
			return exprResult{}, err
		}
		i, err := strconv.ParseInt(expr.expr, 10, 64)
		if err != nil {
			return exprResult{}, err
		}
		unit = int(i)
	}

	expr, err := t.transformExprNode(ctx, node.Args[0])
	if err != nil {
		return exprResult{}, err
	}

	d, err := duration.ParseISO8601(strings.TrimSuffix(strings.TrimPrefix(expr.expr, "'"), "'"))
	if err != nil {
		return exprResult{}, fmt.Errorf("metrics sql: invalid ISO8601 duration %s", expr.expr)
	}

	watermark, col, err := t.getWatermark(ctx, colName)
	if err != nil {
		return exprResult{}, err
	}

	var end time.Time
	if std, ok := d.(duration.StandardDuration); ok {
		end = std.EndTime(watermark)
	} else {
		end = watermark
	}

	for i := 1; i <= unit; i++ {
		watermark = d.Sub(watermark)
	}

	return exprResult{
		expr:    fmt.Sprintf("'%s/%s'", watermark.Format(time.RFC3339), end.Format(time.RFC3339)),
		columns: []string{col},
		types:   []string{"DIMENSION"},
	}, nil
}

func (t *transformer) getWatermark(ctx context.Context, colName string) (watermark time.Time, column string, err error) {
	olap, release, err := t.controller.AcquireOLAP(ctx, t.connector)
	if err != nil {
		return watermark, column, err
	}
	defer release()

	spec := t.metricsView.Spec
	var sql string
	if colName != "" {
		sql = fmt.Sprintf("SELECT MAX(%s) FROM %s ", olap.Dialect().EscapeIdentifier(colName), olap.Dialect().EscapeTable(spec.Database, spec.DatabaseSchema, spec.Table))
		column = colName
	} else if t.metricsView.Spec.WatermarkExpression != "" {
		sql = fmt.Sprintf("SELECT %s FROM %s", t.metricsView.Spec.WatermarkExpression, olap.Dialect().EscapeTable(spec.Database, spec.DatabaseSchema, spec.Table))
		// todo how to handle column name here
	} else if spec.TimeDimension != "" {
		sql = fmt.Sprintf("SELECT MAX(%s) FROM %s", olap.Dialect().EscapeIdentifier(spec.TimeDimension), olap.Dialect().EscapeTable(spec.Database, spec.DatabaseSchema, spec.Table))
		column = spec.TimeDimension
	} else {
		return watermark, column, fmt.Errorf("metrics sql: no watermark or time dimension found in metrics view")
	}
	result, err := olap.Execute(ctx, &drivers.Statement{Query: sql, Priority: t.priority})
	if err != nil {
		return watermark, column, err
	}
	defer result.Close()

	for result.Next() {
		if err := result.Scan(&watermark); err != nil {
			return watermark, column, fmt.Errorf("error scanning watermark: %w", err)
		}
	}
	if watermark.IsZero() {
		return watermark, column, fmt.Errorf("metrics sql: no watermark or time dimension found in metrics view")
	}
	return watermark, column, nil
}
