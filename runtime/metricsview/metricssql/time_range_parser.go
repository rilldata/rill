package metricssql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pingcap/tidb/pkg/parser/ast"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

func (q *query) parseTimeRangeStart(ctx context.Context, node *ast.FuncCallExpr, timeDimNode *ast.ColumnNameExpr) (*metricsview.Expression, error) {
	rillTime, err := parseTimeRangeArgs(node.Args, q.metricsViewSpec)
	if err != nil {
		return nil, err
	}

	timeDim := "" // Default to empty string if no time dimension is provided
	if timeDimNode != nil {
		timeDim = timeDimNode.Name.Name.O
	}

	if q.opts == nil || q.opts.GetTimestamps == nil {
		return nil, fmt.Errorf("metrics sql: not able to resolve dynamic time expressions in this context")
	}

	ts, err := q.opts.GetTimestamps(ctx, q.metricsView, timeDim)
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
	rillTime, err := parseTimeRangeArgs(node.Args, q.metricsViewSpec)
	if err != nil {
		return nil, err
	}

	timeDim := "" // Default to empty string if no time dimension is provided
	if timeDimNode != nil {
		timeDim = timeDimNode.Name.Name.O
	}

	if q.opts == nil || q.opts.GetTimestamps == nil {
		return nil, fmt.Errorf("metrics sql: not able to resolve dynamic time expressions in this context")
	}

	ts, err := q.opts.GetTimestamps(ctx, q.metricsView, timeDim)
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

func parseTimeRangeArgs(args []ast.ExprNode, mv *runtimev1.MetricsViewSpec) (*rilltime.Expression, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("metrics sql: time_range_start/time_range_end expects exactly one arg")
	}
	var err error

	duVal, err := parseValueExpr(args[0])
	if err != nil {
		return nil, err
	}
	du, ok := duVal.(string)
	if !ok {
		return nil, fmt.Errorf("metrics sql: expected string for duration, got %T", duVal)
	}

	rt, err := rilltime.Parse(strings.TrimSuffix(strings.TrimPrefix(du, "'"), "'"), rilltime.ParseOptions{
		SmallestGrain: timeutil.TimeGrainFromAPI(mv.SmallestTimeGrain),
	})
	if err != nil {
		return nil, fmt.Errorf("metrics sql: invalid ISO8601 duration %s", du)
	}
	return rt, nil
}
