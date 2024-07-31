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
	"github.com/rilldata/rill/runtime/pkg/duration"
)

func (q *query) parseTimeRangeStart(ctx context.Context, node *ast.FuncCallExpr) (*metricsview.Expression, error) {
	d, unit, colName, err := q.parseTimeRangeArgs(node.Args)
	if err != nil {
		return nil, err
	}

	watermark, err := q.getWatermark(ctx, colName)
	if err != nil {
		return nil, err
	}

	if t, ok := d.(duration.TruncToDateDuration); ok {
		watermark = t.SubWithUnit(watermark, unit)
	} else {
		for i := 1; i <= unit; i++ {
			watermark = d.Sub(watermark)
		}
	}
	return &metricsview.Expression{
		Value: watermark,
	}, nil
}

func (q *query) parseTimeRangeEnd(ctx context.Context, node *ast.FuncCallExpr) (*metricsview.Expression, error) {
	d, unit, colName, err := q.parseTimeRangeArgs(node.Args)
	if err != nil {
		return nil, err
	}

	watermark, err := q.getWatermark(ctx, colName)
	if err != nil {
		return nil, err
	}

	if t, ok := d.(duration.TruncToDateDuration); ok {
		watermark = t.SubWithUnit(watermark, unit-1)
	} else {
		for i := 1; i < unit; i++ {
			watermark = d.Sub(watermark)
		}
	}

	var end time.Time
	if std, ok := d.(duration.StandardDuration); ok {
		end = std.EndTime(watermark)
	} else {
		end = watermark
	}

	return &metricsview.Expression{
		Value: end,
	}, nil
}

func (q *query) parseTimeRangeArgs(args []ast.ExprNode) (duration.Duration, int, string, error) {
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
		val, err := q.parseValueExpr(args[1])
		if err != nil {
			return nil, 0, "", err
		}
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, 0, "", err
		}
		unit = int(i)
	}

	// identify column name
	if len(args) == 3 {
		col, _, err = q.parseColumnNameExpr(args[2])
		if err != nil {
			return nil, 0, "", err
		}
	}

	du, err := q.parseValueExpr(args[0])
	if err != nil {
		return nil, 0, "", err
	}

	d, err := duration.ParseISO8601(strings.TrimSuffix(strings.TrimPrefix(du, "'"), "'"))
	if err != nil {
		return nil, 0, "", fmt.Errorf("metrics sql: invalid ISO8601 duration %s", du)
	}
	return d, unit, col, nil
}

func (q *query) getWatermark(ctx context.Context, colName string) (watermark time.Time, err error) {
	olap, release, err := q.controller.AcquireOLAP(ctx, q.metricsView.Spec.Connector)
	if err != nil {
		return watermark, err
	}
	defer release()

	spec := q.metricsView.Spec
	var sql string
	if colName != "" {
		sql = fmt.Sprintf("SELECT MAX(%s) FROM %s ", olap.Dialect().EscapeIdentifier(colName), olap.Dialect().EscapeTable(spec.Database, spec.DatabaseSchema, spec.Table))
	} else if q.metricsView.Spec.WatermarkExpression != "" {
		sql = fmt.Sprintf("SELECT %s FROM %s", q.metricsView.Spec.WatermarkExpression, olap.Dialect().EscapeTable(spec.Database, spec.DatabaseSchema, spec.Table))
		// todo how to handle column name here
	} else if spec.TimeDimension != "" {
		sql = fmt.Sprintf("SELECT MAX(%s) FROM %s", olap.Dialect().EscapeIdentifier(spec.TimeDimension), olap.Dialect().EscapeTable(spec.Database, spec.DatabaseSchema, spec.Table))
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
