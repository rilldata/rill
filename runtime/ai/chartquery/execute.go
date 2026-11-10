package chartquery

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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
	// DomainValues lists the ordered values selected for each top-N dimension,
	// followed by the label of the "Other" slice when one was added (see Plan.Other).
	DomainValues map[string][]any
	// Limits is the effective top-N limit of each dimension after defaults and caps,
	// keyed by the channel of the chart spec (x, y, color, column, ...) and then by field,
	// so that a renderer can show the limit that was applied.
	Limits map[string]map[string]int
	// Totals holds the total of each measure for pie and donut charts.
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
		Limits:       make(map[string]map[string]int, len(plan.TopN)),
	}

	// Select the values of each top-N dimension.
	domain := make(map[string][]any)
	var cut []TopNQuery // The top-N queries that cut values
	for _, topN := range plan.TopN {
		if out.Limits[topN.Channel] == nil {
			out.Limits[topN.Channel] = make(map[string]int)
		}
		out.Limits[topN.Channel][topN.Field] = topN.Limit
		res, err := runner.Query(ctx, topN.Query)
		if err != nil {
			return nil, fmt.Errorf("top-N query for %q failed: %w", topN.Field, err)
		}
		rows := res.Rows
		if len(rows) > topN.Limit {
			rows = rows[:topN.Limit]
			out.Truncated = true
			cut = append(cut, topN)
		}
		values := make([]any, len(rows))
		for i, row := range rows {
			values[i] = normalizeValue(row[topN.Field])
		}
		domain[topN.Field] = values
		out.DomainValues[topN.Field] = values
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

	// Trim rows from the tail until the data fits the byte budget, which the rows share with the domain values:
	// those repeat the values of the top-N dimensions, which may be long.
	// Data that does not fit even with a single row fails the chart: no cut makes it fit, and a chart without rows would misrepresent the data.
	domainData, err := json.Marshal(out.DomainValues)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chart domain values: %w", err)
	}
	budget := MaxDataBytes - len(domainData)
	data, err := json.Marshal(out.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chart data: %w", err)
	}
	overBudget := len(data) > budget
	for len(data) > budget && len(out.Data) > 1 {
		keep := len(out.Data) * budget / len(data)
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
	if len(data) > budget {
		return nil, fmt.Errorf("the chart data exceeds the response size budget of %d bytes even when cut to a single row, so a bound dimension holds values too long to chart; bind dimensions with shorter values, or lower their limits", MaxDataBytes)
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
	var rowsWarning string
	if out.RowsTruncated {
		out.Truncated = true
		rowsWarning = fmt.Sprintf("the data was limited to %d rows", len(out.Data))
		if plan.TimeField != "" {
			rowsWarning = fmt.Sprintf("only the most recent %d periods (%d rows) are shown", periods, len(out.Data))
		}
		if overBudget {
			rowsWarning += " to fit the response size budget"
		}
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

	// Combine the slices the top-N of a pie or donut chart cut into an "Other" slice.
	var other string
	if plan.Other != nil && slices.ContainsFunc(cut, func(q TopNQuery) bool { return q.Field == plan.Other.Field }) {
		other, err = addOtherSlice(ctx, plan.Other, domain, out, runner)
		if err != nil {
			return nil, err
		}
	}

	if out.Truncated {
		var warnings []string
		for _, q := range cut {
			if other != "" && q.Field == plan.Other.Field {
				warnings = append(warnings, fmt.Sprintf("only the top %d values of %q are shown individually, and the rest are combined into %q", q.Limit, q.Field, other))
			} else {
				warnings = append(warnings, fmt.Sprintf("only the top %d values of %q are shown", q.Limit, q.Field))
			}
		}
		if rowsWarning != "" {
			warnings = append(warnings, rowsWarning)
		}
		remedy := "lower the field limits or add filters"
		if out.TimeGrain != "" {
			remedy = "use a coarser time_grain, lower the field limits, or add filters"
		}
		out.TruncationWarning = fmt.Sprintf("The chart data is incomplete: %s. To see the complete picture, %s.", strings.Join(warnings, "; "), remedy)
	}
	return out, nil
}

// addOtherSlice appends an "Other" row to a pie or donut chart's data for the slice values the top-N query cut, sized by the measure over them,
// and appends its label to the domain values; it returns the label, or an empty string when it added no slice.
// The label is "Other", parenthesized as often as needed to differ from the shown values.
// The slice is only added when the measure adds up: the value over the cut values must equal the total minus the shown slices.
// A sum or a count does; an average, a ratio or a distinct count generally does not, and a remainder of it would size a meaningless slice.
// The check runs on the data, since the metrics view does not declare which measures add up.
func addOtherSlice(ctx context.Context, q *OtherQuery, domain map[string][]any, out *Output, runner Runner) (string, error) {
	res, err := runner.Query(ctx, q.Query(domain))
	if err != nil {
		return "", fmt.Errorf("other query failed: %w", err)
	}
	if len(res.Rows) == 0 {
		return "", nil
	}
	value := normalizeValue(res.Rows[0][q.Measure])
	rest, ok := number(value)
	if !ok || rest <= 0 {
		return "", nil
	}
	total, ok := number(out.Totals[q.Measure])
	if !ok {
		return "", nil
	}
	var shown float64
	for _, row := range out.Data {
		v, ok := number(row[q.Measure])
		if !ok {
			return "", nil
		}
		shown += v
	}
	// The tolerance absorbs the rounding of floating-point sums, which the database adds up in a different order per query.
	if math.Abs(shown+rest-total) > 1e-6*math.Max(math.Abs(total), 1) {
		return "", nil
	}

	label := "Other"
	for slices.Contains(domain[q.Field], any(label)) {
		label = "(" + label + ")"
	}
	out.Data = append(out.Data, map[string]any{q.Field: label, q.Measure: value})
	out.DomainValues[q.Field] = append(slices.Clone(out.DomainValues[q.Field]), label)
	return label, nil
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

// number returns a numeric value of a result row as a float64, and whether it is numeric.
func number(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	case int:
		return float64(n), true
	default:
		return 0, false
	}
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
