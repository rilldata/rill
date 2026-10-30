package chartquery

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/metricsview"
)

// Runner runs a metrics view query. It is implemented by the AI tools on top of the runtime's metrics resolver.
type Runner interface {
	Query(ctx context.Context, q *metricsview.Query) (*Result, error)
}

// Result is the outcome of one query.
type Result struct {
	// Rows are the result rows keyed by column name.
	Rows []map[string]any
	// Types maps column names to their type code without the CODE_ prefix (e.g. STRING, INT64, TIMESTAMP).
	Types map[string]string
	// TimeRange and ComparisonTimeRange are the time ranges the query resolved to.
	TimeRange           *TimeRange
	ComparisonTimeRange *TimeRange
}

// Output is the chart-ready data produced by Execute.
type Output struct {
	// TimeGrain is the TIME_GRAIN_* name used to floor temporal fields, or empty when none is used.
	TimeGrain string
	// Data are the rows of the main query keyed by field name.
	// Comparison values are under <measure>_prev and temporal values are RFC3339 UTC strings of the bucket start.
	Data []map[string]any
	// DomainValues lists the ordered values selected for each nominal axis, split, stage or heatmap axis,
	// and the measure names for a "rill_measures" color split.
	DomainValues map[string][]any
	// Limits is the effective top-N limit of each spec channel (x, y, color or stage, as named in the input spec)
	// after defaults and caps, so that a renderer can show the limit that was applied.
	Limits map[string]int
	// Totals holds the total of each measure for donut and pie charts.
	Totals map[string]any
	// Types maps the columns of the main query to their type codes.
	Types                       map[string]string
	ResolvedTimeRange           *TimeRange
	ResolvedComparisonTimeRange *TimeRange
	// Truncated reports whether any budget cut the data; TruncationWarning explains what was cut.
	// RowsTruncated reports whether the rows of the main query were cut (by the row cap or the byte budget) as opposed to a top-N cut;
	// a temporal chart then keeps its newest buckets, so its time axis is incomplete at the start.
	Truncated         bool
	RowsTruncated     bool
	TruncationWarning string
}

// Execute runs the plan's queries in order and assembles the chart data within the budgets.
func Execute(ctx context.Context, plan *Plan, runner Runner) (*Output, error) {
	out := &Output{
		TimeGrain:    plan.Spec.TimeGrainName(),
		DomainValues: make(map[string][]any),
		Limits:       make(map[string]int, len(plan.TopN)),
	}
	var warnings []string

	// Select the values of each top-N dimension. The spec may also pin the values explicitly.
	domain := make(map[string][]any)
	for _, f := range plan.Spec.fieldsWithSortValues() {
		values := make([]any, len(f.SortValues))
		for i, v := range f.SortValues {
			values[i] = v
		}
		domain[f.Field] = values
		out.DomainValues[f.Field] = values
	}
	for _, topN := range plan.TopN {
		out.Limits[topN.Channel] = topN.Limit
		res, err := runner.Query(ctx, topN.Query)
		if err != nil {
			return nil, fmt.Errorf("top-N query for %q failed: %w", topN.Field, err)
		}
		rows := res.Rows
		if len(rows) > topN.Limit {
			rows = rows[:topN.Limit]
			out.Truncated = true
			warnings = append(warnings, fmt.Sprintf("only the top %d values of %q are shown", topN.Limit, topN.Field))
		}
		values := make([]any, len(rows))
		for i, row := range rows {
			values[i] = normalizeValue(row[topN.Field])
		}
		domain[topN.Field] = values
		out.DomainValues[topN.Field] = values
	}
	if plan.Spec.Color != nil && len(plan.MeasureDomain) > 0 {
		values := make([]any, len(plan.MeasureDomain))
		for i, m := range plan.MeasureDomain {
			values[i] = m
		}
		out.DomainValues[plan.Spec.Color.Field] = values
	}

	// Run the main query. Temporal charts are queried newest first (see Plan.TimeField), so the row cuts below keep the most recent rows.
	res, err := runner.Query(ctx, plan.Main(domain))
	if err != nil {
		return nil, fmt.Errorf("chart query failed: %w", err)
	}
	rows := res.Rows
	if len(rows) > plan.RowLimit {
		rows = rows[:plan.RowLimit]
		out.RowsTruncated = true
	}
	for _, row := range rows {
		for k, v := range row {
			row[k] = normalizeValue(v)
		}
	}
	out.Data = rows
	out.Types = res.Types
	out.ResolvedTimeRange = res.TimeRange
	out.ResolvedComparisonTimeRange = res.ComparisonTimeRange
	if out.ResolvedTimeRange != nil && plan.Spec.TimeRange != nil {
		out.ResolvedTimeRange.TimeZone = plan.Spec.TimeRange.TimeZone
	}

	// Trim rows from the tail until the data fits the byte budget.
	data, err := json.Marshal(out.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chart data: %w", err)
	}
	overBudget := len(data) > MaxDataBytes
	for len(data) > MaxDataBytes && len(out.Data) > 1 {
		keep := len(out.Data) * MaxDataBytes / len(data)
		if keep >= len(out.Data) {
			keep = len(out.Data) - 1
		}
		out.Data = out.Data[:max(keep, 1)]
		data, err = json.Marshal(out.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal chart data: %w", err)
		}
		out.RowsTruncated = true
	}

	// The rows of a temporal chart arrive newest first. When rows were cut, the oldest bucket may be partial
	// (a time series with a color split has several rows per bucket), so it is dropped unless it is the only one.
	// The rows are then returned in ascending time order, keeping the order within each bucket.
	var periods int
	if plan.TimeField != "" {
		buckets := timeBuckets(out.Data, plan.TimeField)
		if out.RowsTruncated && len(buckets) > 1 {
			buckets = buckets[:len(buckets)-1]
		}
		slices.Reverse(buckets)
		out.Data = slices.Concat(buckets...)
		periods = len(buckets)
	}
	if out.RowsTruncated {
		out.Truncated = true
		warning := fmt.Sprintf("the data was limited to %d rows", len(out.Data))
		if plan.TimeField != "" {
			warning = fmt.Sprintf("only the most recent %d periods (%d rows) are shown", periods, len(out.Data))
		}
		if overBudget {
			warning += " to fit the response size budget"
		}
		warnings = append(warnings, warning)
	}

	// Data is never nil, so that the result carries data: [] when there are no rows and a renderer can tell an empty chart from missing data.
	if out.Data == nil {
		out.Data = []map[string]any{}
	}

	// Run the extra queries (totals).
	for name, q := range plan.Extra {
		res, err := runner.Query(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("%s query failed: %w", name, err)
		}
		if len(res.Rows) == 0 {
			continue
		}
		if out.Totals == nil {
			out.Totals = make(map[string]any)
		}
		for _, m := range q.Measures {
			out.Totals[m.Name] = normalizeValue(res.Rows[0][m.Name])
		}
	}

	if out.Truncated {
		remedy := "lower the field limits or add filters"
		if out.TimeGrain != "" {
			remedy = "use a coarser time_grain, lower the field limits, or add filters"
		}
		out.TruncationWarning = fmt.Sprintf("The chart data is incomplete: %s. To see the complete picture, %s.", strings.Join(warnings, "; "), remedy)
	}
	return out, nil
}

// timeBuckets splits rows that are ordered by a time field into runs of rows with the same time value.
func timeBuckets(rows []map[string]any, field string) [][]map[string]any {
	var res [][]map[string]any
	for i, row := range rows {
		if i == 0 || row[field] != rows[i-1][field] {
			res = append(res, nil)
		}
		res[len(res)-1] = append(res[len(res)-1], row)
	}
	return res
}

// fieldsWithSortValues returns the fields whose values are pinned by the spec.
func (s *Spec) fieldsWithSortValues() []*Field {
	var res []*Field
	for _, f := range []*Field{s.X, s.Y, s.Color, s.Stage} {
		if f != nil && f.Field != "" && len(f.SortValues) > 0 {
			res = append(res, f)
		}
	}
	return res
}

// normalizeValue makes a value JSON-friendly for the widget: timestamps become RFC3339 UTC strings.
func normalizeValue(v any) any {
	switch t := v.(type) {
	case time.Time:
		t = t.UTC()
		if t.Nanosecond() != 0 {
			return t.Format(time.RFC3339Nano)
		}
		return t.Format(time.RFC3339)
	case *time.Time:
		if t == nil {
			return nil
		}
		return normalizeValue(*t)
	default:
		return v
	}
}
