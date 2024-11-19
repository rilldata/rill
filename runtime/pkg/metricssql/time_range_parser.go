package metricssqlparser

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
)

func (q *query) parseTimeRangeStart(ctx context.Context, node *ast.FuncCallExpr) (*metricsview.Expression, error) {
	rt, unit, colName, err := parseTimeRangeArgs(node.Args)
	if err != nil {
		return nil, err
	}

	watermark, err := q.getWatermark(ctx, colName)
	if err != nil {
		return nil, err
	}

	minTime, err := q.getMinTime(ctx, colName)
	if err != nil {
		return nil, err
	}

	for i := 1; i <= unit; i++ {
		watermark, _, err = rt.Resolve(rilltime.ResolverContext{
			Now:        time.Now(),
			MinTime:    minTime,
			MaxTime:    watermark,
			FirstDay:   int(q.metricsViewSpec.FirstDayOfWeek),
			FirstMonth: int(q.metricsViewSpec.FirstMonthOfYear),
		})
		if err != nil {
			return nil, err
		}
	}

	return &metricsview.Expression{
		Value: watermark,
	}, nil
}

func (q *query) parseTimeRangeEnd(ctx context.Context, node *ast.FuncCallExpr) (*metricsview.Expression, error) {
	rt, unit, colName, err := parseTimeRangeArgs(node.Args)
	if err != nil {
		return nil, err
	}

	watermark, err := q.getWatermark(ctx, colName)
	if err != nil {
		return nil, err
	}

	minTime, err := q.getMinTime(ctx, colName)
	if err != nil {
		return nil, err
	}

	for i := 1; i <= unit; i++ {
		_, watermark, err = rt.Resolve(rilltime.ResolverContext{
			Now:        time.Now(),
			MinTime:    minTime,
			MaxTime:    watermark,
			FirstDay:   int(q.metricsViewSpec.FirstDayOfWeek),
			FirstMonth: int(q.metricsViewSpec.FirstMonthOfYear),
		})
		if err != nil {
			return nil, err
		}
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

func (q *query) getMinTime(ctx context.Context, colName string) (time.Time, error) {
	if colName == "" {
		colName = q.metricsViewSpec.TimeDimension
	}
	if colName == "" {
		// we cannot get min time without a time dimension or a column name specified. return a 0 time
		return time.Time{}, nil
	}

	olap, release, err := q.controller.AcquireOLAP(ctx, q.metricsViewSpec.Connector)
	if err != nil {
		return time.Time{}, err
	}
	defer release()

	sql := fmt.Sprintf("SELECT MIN(%s) FROM %s", olap.Dialect().EscapeIdentifier(colName), olap.Dialect().EscapeTable(q.metricsViewSpec.Database, q.metricsViewSpec.DatabaseSchema, q.metricsViewSpec.Table))
	result, err := olap.Execute(ctx, &drivers.Statement{Query: sql, Priority: q.priority})
	if err != nil {
		return time.Time{}, err
	}
	defer result.Close()

	var t time.Time
	for result.Next() {
		if err := result.Scan(&t); err != nil {
			return time.Time{}, fmt.Errorf("error scanning min time: %w", err)
		}
	}
	return t, nil
}

func parseTimeRangeArgs(args []ast.ExprNode) (*rilltime.RillTime, int, string, error) {
	if len(args) == 0 {
		return nil, 0, "", fmt.Errorf("metrics sql: mandatory arg duration missing for time_range_end() function")
	}
	if len(args) > 3 {
		return nil, 0, "", fmt.Errorf("metrics sql: time_range_end() function expects at most 3 arguments")
	}
	// identify optional args
	var (
		col  string
		unit int
		err  error
	)
	// identify unit
	if len(args) == 1 {
		unit = 1
	} else {
		val, err := parseValueExpr(args[1])
		if err != nil {
			return nil, 0, "", err
		}
		unit, err = strconv.Atoi(val)
		if err != nil {
			return nil, 0, "", err
		}
	}

	// identify column name
	if len(args) == 3 {
		col, err = parseColumnNameExpr(args[2])
		if err != nil {
			return nil, 0, "", err
		}
	}

	du, err := parseValueExpr(args[0])
	if err != nil {
		return nil, 0, "", err
	}

	rt, err := rilltime.Parse(strings.TrimSuffix(strings.TrimPrefix(du, "'"), "'"))
	if err != nil {
		return nil, 0, "", fmt.Errorf("metrics sql: invalid ISO8601 duration %s", du)
	}
	return rt, unit, col, nil
}
