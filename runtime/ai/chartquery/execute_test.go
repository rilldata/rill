package chartquery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
)

// fakeRunner answers queries from a function and records the queries it received.
type fakeRunner struct {
	queries []*metricsview.Query
	answer  func(q *metricsview.Query) *Result
}

func (r *fakeRunner) Query(ctx context.Context, q *metricsview.Query) (*Result, error) {
	r.queries = append(r.queries, q)
	return r.answer(q), nil
}

func rows(field string, n int) []map[string]any {
	res := make([]map[string]any, n)
	for i := range res {
		res[i] = map[string]any{field: fmt.Sprintf("v%d", i), "bids": float64(100 - i)}
	}
	return res
}

func TestExecuteCartesian(t *testing.T) {
	spec, err := ParseSpec(BarChart, map[string]any{
		"metrics_view": "mv", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"},
		"comparison_time_range": map[string]any{"start": "2023-12-01T00:00:00Z", "end": "2024-01-01T00:00:00Z"},
		"color":                 map[string]any{"field": "device", "type": "nominal", "limit": 2},
		"x":                     map[string]any{"field": "advertiser", "type": "nominal", "limit": 3, "sort": "-y"},
		"y":                     map[string]any{"field": "bids", "type": "quantitative"},
	})
	require.NoError(t, err)
	plan, err := Build(spec)
	require.NoError(t, err)

	runner := &fakeRunner{answer: func(q *metricsview.Query) *Result {
		switch {
		case len(q.Dimensions) == 1 && q.Dimensions[0].Name == "advertiser":
			// One more row than the limit: the x axis is truncated.
			return &Result{Rows: rows("advertiser", 4)}
		case len(q.Dimensions) == 1 && q.Dimensions[0].Name == "device":
			return &Result{Rows: rows("device", 2)}
		default:
			return &Result{
				Rows: []map[string]any{
					{"advertiser": "v0", "device": "v0", "bids": float64(10), "bids_prev": float64(8)},
					{"advertiser": "v0", "device": "v1", "bids": float64(5), "bids_prev": nil},
				},
				Types:               map[string]string{"advertiser": "STRING", "device": "STRING", "bids": "INT64", "bids_prev": "INT64"},
				TimeRange:           &TimeRange{Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
				ComparisonTimeRange: &TimeRange{Start: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			}
		}
	}}

	out, err := Execute(t.Context(), plan, runner)
	require.NoError(t, err)

	// Two top-N queries then the main query, which filters to the selected values.
	require.Len(t, runner.queries, 3)
	main := runner.queries[2]
	require.Equal(t, []any{"v0", "v1", "v2"}, out.DomainValues["advertiser"])
	require.Equal(t, []any{"v0", "v1"}, out.DomainValues["device"])
	mainJSON, err := json.Marshal(main.Where)
	require.NoError(t, err)
	require.Contains(t, string(mainJSON), `{"name":"advertiser"},{"val":["v0","v1","v2"]}`)
	require.Contains(t, string(mainJSON), `{"name":"device"},{"val":["v0","v1"]}`)
	require.Len(t, main.Measures, 2)
	require.Equal(t, "bids_prev", main.Measures[1].Name)

	require.Len(t, out.Data, 2)
	require.Equal(t, "INT64", out.Types["bids"])
	require.Equal(t, map[string]int{"x": 3, "color": 2}, out.Limits)
	require.True(t, out.Truncated)
	require.False(t, out.RowsTruncated, "only a top-N was cut")
	require.Contains(t, out.TruncationWarning, `only the top 3 values of "advertiser" are shown`)
	require.NotContains(t, out.TruncationWarning, "rows")
	require.Equal(t, "2024-01-01T00:00:00Z", out.ResolvedTimeRange.Start.Format(time.RFC3339))
	require.NotNil(t, out.ResolvedComparisonTimeRange)
	require.Empty(t, out.TimeGrain)
}

func TestExecuteTemporalAndRowLimit(t *testing.T) {
	spec, err := ParseSpec(LineChart, map[string]any{
		"metrics_view": "mv", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z", "time_zone": "America/New_York"},
		"time_grain": "TIME_GRAIN_HOUR", "color": "primary",
		"x": map[string]any{"field": "__time", "type": "temporal"},
		"y": map[string]any{"field": "bids", "type": "quantitative"},
	})
	require.NoError(t, err)
	plan, err := Build(spec)
	require.NoError(t, err)

	// The query asks for the rows newest first (and one more than the cap), so the fake returns hours MaxRows down to 0.
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("EST", -5*3600))
	runner := &fakeRunner{answer: func(q *metricsview.Query) *Result {
		res := &Result{TimeRange: &TimeRange{Start: start, End: start.AddDate(0, 1, 0)}}
		for i := MaxRows; i >= 0; i-- {
			res.Rows = append(res.Rows, map[string]any{"__time": start.Add(time.Duration(i) * time.Hour), "bids": float64(i)})
		}
		return res
	}}

	out, err := Execute(t.Context(), plan, runner)
	require.NoError(t, err)
	require.Len(t, runner.queries, 1)
	require.Equal(t, "America/New_York", runner.queries[0].TimeZone)
	require.Equal(t, metricsview.TimeGrainHour, runner.queries[0].Dimensions[0].Compute.TimeFloor.Grain)
	require.Equal(t, []metricsview.Sort{{Name: "__time", Desc: true}}, runner.queries[0].Sort, "temporal charts are queried newest first")

	// The cap drops the oldest row (hour 0), then the oldest remaining bucket (hour 1) is dropped as possibly partial,
	// and the rows come back in ascending order: hours 2 to MaxRows.
	require.Equal(t, "TIME_GRAIN_HOUR", out.TimeGrain)
	require.Len(t, out.Data, MaxRows-1)
	require.Equal(t, "2024-01-01T07:00:00Z", out.Data[0]["__time"], "temporal values are RFC3339 in UTC")
	require.Equal(t, start.Add(MaxRows*time.Hour).UTC().Format(time.RFC3339), out.Data[len(out.Data)-1]["__time"], "the newest row survives")
	require.True(t, out.Truncated)
	require.True(t, out.RowsTruncated)
	require.Contains(t, out.TruncationWarning, "only the most recent 1499 periods (1499 rows) are shown")
	require.Contains(t, out.TruncationWarning, "use a coarser time_grain")
	require.Equal(t, "America/New_York", out.ResolvedTimeRange.TimeZone)
	require.Empty(t, out.Limits)
}

// TestExecuteTemporalKeepsNewestBuckets checks a row-truncated time series with a color split: the newest buckets survive,
// including a newest bucket that is genuinely partial, the oldest bucket (partial after the cut) is dropped whole,
// and the rows are returned in ascending time order with the order within each bucket kept.
func TestExecuteTemporalKeepsNewestBuckets(t *testing.T) {
	spec, err := ParseSpec(LineChart, map[string]any{
		"metrics_view": "mv", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-04-01T00:00:00Z"},
		"time_grain": "TIME_GRAIN_HOUR",
		"color":      map[string]any{"field": "device", "type": "nominal", "sort": []any{"a", "b"}},
		"x":          map[string]any{"field": "__time", "type": "temporal"},
		"y":          map[string]any{"field": "bids", "type": "quantitative"},
	})
	require.NoError(t, err)
	plan, err := Build(spec)
	require.NoError(t, err)

	// Newest first: bucket 0 has only device "a", buckets 1 to 750 have both devices; MaxRows+1 rows in total.
	newest := time.Date(2024, 3, 31, 23, 0, 0, 0, time.UTC)
	bucket := func(k int) time.Time { return newest.Add(-time.Duration(k) * time.Hour) }
	runner := &fakeRunner{answer: func(q *metricsview.Query) *Result {
		res := &Result{}
		res.Rows = append(res.Rows, map[string]any{"__time": bucket(0), "device": "a", "bids": float64(1)})
		for k := 1; k <= 750; k++ {
			for _, d := range []string{"a", "b"} {
				res.Rows = append(res.Rows, map[string]any{"__time": bucket(k), "device": d, "bids": float64(k)})
			}
		}
		require.Len(t, res.Rows, MaxRows+1)
		return res
	}}

	out, err := Execute(t.Context(), plan, runner)
	require.NoError(t, err)
	require.True(t, out.RowsTruncated)

	// The cap cuts bucket 750 to one row, so the whole bucket is dropped: buckets 749 down to 0 remain.
	require.Len(t, out.Data, 1+749*2)
	require.Equal(t, bucket(749).Format(time.RFC3339), out.Data[0]["__time"])
	require.Equal(t, "a", out.Data[0]["device"])
	require.Equal(t, "b", out.Data[1]["device"], "the order within a bucket is kept")
	require.Equal(t, bucket(0).Format(time.RFC3339), out.Data[len(out.Data)-1]["__time"], "the partial newest bucket survives")
	for _, row := range out.Data {
		require.NotEqual(t, bucket(750).Format(time.RFC3339), row["__time"], "the partial oldest bucket is dropped whole")
	}
	require.Contains(t, out.TruncationWarning, "only the most recent 750 periods (1499 rows) are shown")
}

func TestExecuteByteBudget(t *testing.T) {
	spec, err := ParseSpec(LineChart, map[string]any{
		"metrics_view": "mv", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"}, "color": "primary",
		"x": map[string]any{"field": "__time", "type": "temporal"},
		"y": map[string]any{"field": "bids", "type": "quantitative"},
	})
	require.NoError(t, err)
	plan, err := Build(spec)
	require.NoError(t, err)

	// 1000 rows of ~1 KiB each, newest first, exceed the 100 KiB budget without hitting the row limit.
	runner := &fakeRunner{answer: func(q *metricsview.Query) *Result {
		res := &Result{}
		for i := 999; i >= 0; i-- {
			res.Rows = append(res.Rows, map[string]any{"__time": fmt.Sprintf("2024-01-01T%03d", i), "bids": strings.Repeat("x", 1024)})
		}
		return res
	}}

	out, err := Execute(t.Context(), plan, runner)
	require.NoError(t, err)
	require.True(t, out.Truncated)
	require.True(t, out.RowsTruncated)
	require.Less(t, len(out.Data), 1000)
	require.NotEmpty(t, out.Data)
	data, err := json.Marshal(out.Data)
	require.NoError(t, err)
	require.LessOrEqual(t, len(data), MaxDataBytes)
	require.Contains(t, out.TruncationWarning, fmt.Sprintf("only the most recent %d periods (%d rows) are shown to fit the response size budget", len(out.Data), len(out.Data)))
	// The oldest rows are trimmed and the rest returned in ascending order, so the newest row is last.
	require.Equal(t, "2024-01-01T999", out.Data[len(out.Data)-1]["__time"])
	require.Less(t, out.Data[0]["__time"].(string), out.Data[1]["__time"].(string))
}

func TestExecuteDonutTotals(t *testing.T) {
	spec, err := ParseSpec(DonutChart, map[string]any{
		"metrics_view": "mv", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"},
		"color":   map[string]any{"field": "username", "type": "nominal", "limit": 2},
		"measure": map[string]any{"field": "commits", "type": "quantitative"},
	})
	require.NoError(t, err)
	plan, err := Build(spec)
	require.NoError(t, err)

	runner := &fakeRunner{answer: func(q *metricsview.Query) *Result {
		if len(q.Dimensions) == 0 {
			return &Result{Rows: []map[string]any{{"commits": float64(42)}}}
		}
		return &Result{Rows: []map[string]any{
			{"username": "alice", "commits": float64(20)},
			{"username": "bob", "commits": float64(12)},
		}}
	}}

	out, err := Execute(t.Context(), plan, runner)
	require.NoError(t, err)
	require.Len(t, runner.queries, 3)
	require.Equal(t, map[string]any{"commits": float64(42)}, out.Totals)
	require.Equal(t, []any{"alice", "bob"}, out.DomainValues["username"])
	require.Equal(t, map[string]int{"color": 2}, out.Limits)
	require.False(t, out.Truncated)
	require.Empty(t, out.TruncationWarning)
}

func TestExecuteMeasureDomainAndPinnedValues(t *testing.T) {
	spec, err := ParseSpec(StackedBar, map[string]any{
		"metrics_view": "mv", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"},
		"color": map[string]any{"field": "rill_measures", "type": "value"},
		"x":     map[string]any{"field": "weekday", "type": "nominal", "sort": []any{"Mon", "Tue"}},
		"y":     map[string]any{"field": "clicks", "type": "quantitative", "fields": []any{"clicks", "impressions"}},
	})
	require.NoError(t, err)
	plan, err := Build(spec)
	require.NoError(t, err)
	require.Empty(t, plan.TopN, "pinned values replace the top-N query")

	runner := &fakeRunner{answer: func(q *metricsview.Query) *Result { return &Result{} }}
	out, err := Execute(t.Context(), plan, runner)
	require.NoError(t, err)
	require.Len(t, runner.queries, 1)
	whereJSON, err := json.Marshal(runner.queries[0].Where)
	require.NoError(t, err)
	require.Contains(t, string(whereJSON), `{"name":"weekday"},{"val":["Mon","Tue"]}`)
	require.Equal(t, []any{"Mon", "Tue"}, out.DomainValues["weekday"])
	require.Equal(t, []any{"clicks", "impressions"}, out.DomainValues["rill_measures"])
	require.Empty(t, out.Limits, "pinned values apply no limit")
	require.NotNil(t, out.Data, "no rows is an empty slice, not nil")
	require.Empty(t, out.Data)
}

// TestExecuteLimitsHorizontal checks that the limit of a horizontal bar chart is reported on the input's y channel,
// which holds the dimension, and that the default limit is reported when the spec sets none.
func TestExecuteLimitsHorizontal(t *testing.T) {
	spec, err := ParseSpec(BarChart, map[string]any{
		"metrics_view": "mv", "time_range": map[string]any{"start": "2024-01-01T00:00:00Z", "end": "2024-02-01T00:00:00Z"}, "color": "primary",
		"x": map[string]any{"field": "bids", "type": "quantitative"},
		"y": map[string]any{"field": "advertiser", "type": "nominal", "sort": "-x"},
	})
	require.NoError(t, err)
	require.True(t, spec.Horizontal)
	plan, err := Build(spec)
	require.NoError(t, err)

	runner := &fakeRunner{answer: func(q *metricsview.Query) *Result { return &Result{} }}
	out, err := Execute(t.Context(), plan, runner)
	require.NoError(t, err)
	require.Equal(t, map[string]int{"y": defaultNominalLimit}, out.Limits)
}
