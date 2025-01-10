package metricssqlparser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
)

func (q *query) parseTimeRangeStart(ctx context.Context, node *ast.FuncCallExpr) (*metricsview.Expression, error) {
	rillTime, colName, err := parseTimeRangeArgs(node.Args)
	if err != nil {
		return nil, err
	}

	watermark, err := q.getWatermark(ctx, colName)
	if err != nil {
		return nil, err
	}

	if colName == "" {
		colName = q.metricsViewSpec.TimeDimension
	}
	minTime, err := q.executor.MinTime(ctx, colName)
	if err != nil {
		return nil, err
	}

	watermark, _, err = rillTime.Eval(rilltime.EvalOptions{
		Now:        time.Now(),
		MinTime:    minTime,
		MaxTime:    watermark,
		FirstDay:   int(q.metricsViewSpec.FirstDayOfWeek),
		FirstMonth: int(q.metricsViewSpec.FirstMonthOfYear),
	})
	if err != nil {
		return nil, err
	}

	return &metricsview.Expression{
		Value: watermark,
	}, nil
}

func (q *query) parseTimeRangeEnd(ctx context.Context, node *ast.FuncCallExpr) (*metricsview.Expression, error) {
	rillTime, colName, err := parseTimeRangeArgs(node.Args)
	if err != nil {
		return nil, err
	}

	watermark, err := q.getWatermark(ctx, colName)
	if err != nil {
		return nil, err
	}

	if colName == "" {
		colName = q.metricsViewSpec.TimeDimension
	}
	minTime, err := q.executor.MinTime(ctx, colName)
	if err != nil {
		return nil, err
	}

	_, watermark, err = rillTime.Eval(rilltime.EvalOptions{
		Now:        time.Now(),
		MinTime:    minTime,
		MaxTime:    watermark,
		FirstDay:   int(q.metricsViewSpec.FirstDayOfWeek),
		FirstMonth: int(q.metricsViewSpec.FirstMonthOfYear),
	})
	if err != nil {
		return nil, err
	}

	return &metricsview.Expression{
		Value: watermark,
	}, nil
}

func (q *query) getWatermark(ctx context.Context, colName string) (watermark time.Time, err error) {
	olap, release, err := q.controller.AcquireOLAP(ctx, q.metricsViewSpec.Connector)
	if err != nil {
		return watermark, err
	}
	defer release()

	var sql string
	if colName != "" {
		sql = fmt.Sprintf("SELECT MAX(%s) FROM %s ", olap.Dialect().EscapeIdentifier(colName), olap.Dialect().EscapeTable(q.metricsViewSpec.Database, q.metricsViewSpec.DatabaseSchema, q.metricsViewSpec.Table))
	} else if q.metricsViewSpec.WatermarkExpression != "" {
		sql = fmt.Sprintf("SELECT %s FROM %s", q.metricsViewSpec.WatermarkExpression, olap.Dialect().EscapeTable(q.metricsViewSpec.Database, q.metricsViewSpec.DatabaseSchema, q.metricsViewSpec.Table))
		// todo how to handle column name here
	} else if q.metricsViewSpec.TimeDimension != "" {
		sql = fmt.Sprintf("SELECT MAX(%s) FROM %s", olap.Dialect().EscapeIdentifier(q.metricsViewSpec.TimeDimension), olap.Dialect().EscapeTable(q.metricsViewSpec.Database, q.metricsViewSpec.DatabaseSchema, q.metricsViewSpec.Table))
	} else {
		return watermark, fmt.Errorf("metrics sql: no watermark or time dimension found in metrics view")
	}
	result, err := olap.Execute(ctx, &drivers.Statement{Query: sql, Priority: q.priority})
	if err != nil {
		return watermark, err
	}
	defer result.Close()

	for result.Next() {
		if err := result.Scan(&watermark); err != nil {
			return watermark, fmt.Errorf("error scanning watermark: %w", err)
		}
	}
	if watermark.IsZero() {
		return watermark, fmt.Errorf("metrics sql: no watermark or time dimension found in metrics view")
	}
	return watermark, nil
}

func parseTimeRangeArgs(args []ast.ExprNode) (*rilltime.Expression, string, error) {
	if len(args) == 0 {
		return nil, "", fmt.Errorf("metrics sql: mandatory arg duration missing for time_range_end() function")
	}
	if len(args) > 2 {
		return nil, "", fmt.Errorf("metrics sql: time_range_end() function expects at most 2 arguments")
	}
	// identify optional args
	var (
		col string
		err error
	)

	// identify column name
	if len(args) == 2 {
		col, err = parseColumnNameExpr(args[1])
		if err != nil {
			return nil, "", err
		}
	}

	du, err := parseValueExpr(args[0])
	if err != nil {
		return nil, "", err
	}

	rt, err := rilltime.Parse(strings.TrimSuffix(strings.TrimPrefix(du, "'"), "'"))
	if err != nil {
		return nil, "", fmt.Errorf("metrics sql: invalid ISO8601 duration %s", du)
	}
	return rt, col, nil
}
