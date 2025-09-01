package metricssqlparser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
)

func (q *query) parseTimeRangeStart(ctx context.Context, node *ast.FuncCallExpr, timeDimNode *ast.ColumnNameExpr) (*metricsview.Expression, error) {
	rillTime, err := parseTimeRangeArgs(node.Args)
	if err != nil {
		return nil, err
	}

	timeDim := "" // Default to empty string if no time dimension is provided
	if timeDimNode != nil {
		timeDim = timeDimNode.Name.Name.O
	}

	ts, err := q.executor.Timestamps(ctx, timeDim)
	if err != nil {
		return nil, err
	}

	watermark, _, _ := rillTime.Eval(rilltime.EvalOptions{
		Now:        time.Now(),
		MinTime:    ts.Min,
		MaxTime:    ts.Max,
		Watermark:  ts.Watermark,
		FirstDay:   int(q.metricsViewSpec.FirstDayOfWeek),
		FirstMonth: int(q.metricsViewSpec.FirstMonthOfYear),
	})

	return &metricsview.Expression{
		Value: watermark,
	}, nil
}

func (q *query) parseTimeRangeEnd(ctx context.Context, node *ast.FuncCallExpr, timeDimNode *ast.ColumnNameExpr) (*metricsview.Expression, error) {
	rillTime, err := parseTimeRangeArgs(node.Args)
	if err != nil {
		return nil, err
	}

	timeDim := "" // Default to empty string if no time dimension is provided
	if timeDimNode != nil {
		timeDim = timeDimNode.Name.Name.O
	}

	ts, err := q.executor.Timestamps(ctx, timeDim)
	if err != nil {
		return nil, err
	}

	_, watermark, _ := rillTime.Eval(rilltime.EvalOptions{
		Now:        time.Now(),
		MinTime:    ts.Min,
		MaxTime:    ts.Max,
		Watermark:  ts.Watermark,
		FirstDay:   int(q.metricsViewSpec.FirstDayOfWeek),
		FirstMonth: int(q.metricsViewSpec.FirstMonthOfYear),
	})

	return &metricsview.Expression{
		Value: watermark,
	}, nil
}

func parseTimeRangeArgs(args []ast.ExprNode) (*rilltime.Expression, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("metrics sql: time_range_start/time_range_end expects exactly one arg")
	}
	var err error

	du, err := parseValueExpr(args[0])
	if err != nil {
		return nil, err
	}

	rt, err := rilltime.Parse(strings.TrimSuffix(strings.TrimPrefix(du, "'"), "'"), rilltime.ParseOptions{})
	if err != nil {
		return nil, fmt.Errorf("metrics sql: invalid ISO8601 duration %s", du)
	}
	return rt, nil
}
