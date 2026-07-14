package metricsview

import (
	"time"

	"github.com/rilldata/rill/runtime/pkg/rilltime"
)

// ResolveMaxQueryTimeRange resolves a metrics view's max_query_time_range property to a duration relative to now.
// Returns 0 for empty or unparseable input.
func ResolveMaxQueryTimeRange(maxQueryTimeRange string, now time.Time) time.Duration {
	if maxQueryTimeRange == "" {
		return 0
	}
	expr, err := rilltime.Parse(maxQueryTimeRange, rilltime.ParseOptions{})
	if err != nil {
		return 0
	}
	start, end, _ := expr.Eval(rilltime.EvalOptions{
		Now:       now,
		MinTime:   now,
		MaxTime:   now,
		Watermark: now,
	})
	return end.Sub(start)
}
