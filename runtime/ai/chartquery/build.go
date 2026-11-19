// Package chartquery turns a Flint chart spec bound to a metrics view (the arguments external MCP clients pass to create_chart and chart_data)
// into the metrics view queries that fetch its data, and runs them within fixed budgets.
//
// The queries follow from the kinds of the bound fields:
// a top-N query for each dimension on an axis, series or facet channel,
// a main query that groups by every bound dimension (time fields floored to the grain) and keeps the selected values,
// and a total and an "Other" query for a pie or donut chart (see BuildFlint).
//
// The package is pure: it builds metricsview.Query values and runs them through a caller-provided Runner.
package chartquery

import (
	"time"

	"github.com/rilldata/rill/runtime/metricsview"
)

// Budgets for the chart data. They bound the number of categories, rows and bytes a chart result may carry,
// since MCP hosts cap tool results and the widget renders every row it receives.
const (
	// MaxNominalValues caps the values of a dimension on an axis.
	MaxNominalValues = 50
	// MaxColorValues caps the values of a dimension on a series channel (color, group, shape or strokeDash).
	MaxColorValues = 20
	// MaxFacetValues caps the values of a dimension on a facet channel (column or row).
	MaxFacetValues = 8
	// MaxRows caps the rows of the main query.
	MaxRows = 1500
	// MaxDataBytes caps the JSON size of the chart data, the rows and the domain values together; rows are trimmed from the tail to fit.
	// It leaves room for the rest of a result under the size at which hosts stop passing a result through
	// (Claude writes one over about 150,000 characters to a file, which the widget never receives).
	MaxDataBytes = 100 << 10
)

// Default limits applied when the spec does not set one (see flintChannelBudget).
// Small multiples stay small, and so does the angular axis of a polar chart, whose labels crowd sooner than along a straight axis.
const (
	defaultNominalLimit = 20
	defaultColorLimit   = 10
	defaultFacetLimit   = 6
	defaultPolarLimit   = 10
)

// PrevSuffix names the comparison period value of a measure; the widget reads the same suffix.
const PrevSuffix = "_prev"

// Plan is the set of queries that fetch a chart's data.
type Plan struct {
	// Spec is the chart spec the queries were planned for.
	Spec *FlintSpec
	// TopN are pre-queries that select the values of a dimension on an axis, series or facet channel.
	// They run in order before the main query, and their values filter it and order its categories.
	TopN []TopNQuery
	// Main builds the main query from the selected domain values (keyed by field name).
	Main func(domain map[string][]any) *metricsview.Query
	// Extra are queries whose single row is returned as totals (pie and donut charts), keyed by name.
	Extra map[string]*metricsview.Query
	// Other aggregates the values of a pie or donut chart's slice dimension that its top-N query cut, for an "Other" slice (see Execute).
	Other *OtherQuery
	// RowLimit caps the rows returned from the main query; the query itself asks for one more row to detect truncation.
	RowLimit int
	// TimeField is the temporal field the main query is ordered by, newest first, when the chart has a time axis.
	// Ordering newest first makes the row cap and the byte budget drop the oldest rows; Execute restores ascending order.
	TimeField string
}

// TopNQuery selects the values of one dimension.
type TopNQuery struct {
	// Channel is the channel of the chart spec whose values are selected (x, y, color, column, ...).
	Channel string `json:"channel"`
	// Field is the dimension whose values are selected.
	Field string `json:"field"`
	// Limit is the number of values to keep. The query asks for one more row to detect truncation.
	Limit int                `json:"limit"`
	Query *metricsview.Query `json:"query"`
}

// OtherQuery aggregates the measure of a pie or donut chart over the values of its slice dimension that the top-N query cut.
// Execute adds the result as an "Other" slice when the measure adds up to the chart's total (the "total" query of Plan.Extra).
type OtherQuery struct {
	// Field is the slice dimension and Measure the measure the slices are sized by.
	Field   string
	Measure string
	// Query builds the query from the selected domain values, which it leaves out.
	Query func(domain map[string][]any) *metricsview.Query
}

// TimeRange is an absolute time range with an optional IANA time zone.
type TimeRange struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	TimeZone string    `json:"time_zone,omitempty"`
}

func (tr *TimeRange) toQuery() *metricsview.TimeRange {
	if tr == nil {
		return nil
	}
	return &metricsview.TimeRange{Start: tr.Start, End: tr.End}
}

// capLimit applies the default and the cap to a spec limit.
func capLimit(limit, def, maximum int) int {
	if limit <= 0 {
		limit = def
	}
	return min(limit, maximum)
}

func limitPtr(n int) *int64 {
	v := int64(n)
	return &v
}

func namedMeasures(names []string) []metricsview.Measure {
	res := make([]metricsview.Measure, len(names))
	for i, n := range names {
		res[i] = metricsview.Measure{Name: n}
	}
	return res
}

// withPrevMeasures returns each measure followed by its comparison period value under <measure>_prev.
func withPrevMeasures(names []string) []metricsview.Measure {
	res := make([]metricsview.Measure, 0, 2*len(names))
	for _, n := range names {
		res = append(res,
			metricsview.Measure{Name: n},
			metricsview.Measure{Name: n + PrevSuffix, Compute: &metricsview.MeasureCompute{ComparisonValue: &metricsview.MeasureComputeComparisonValue{Measure: n}}},
		)
	}
	return res
}

// nullExclusion returns a filter that drops the null values of a dimension.
func nullExclusion(field string) *metricsview.Expression {
	return &metricsview.Expression{Condition: &metricsview.Condition{
		Operator:    metricsview.OperatorNin,
		Expressions: []*metricsview.Expression{{Name: field}, {Value: []any{nil}}},
	}}
}

// inFilter returns a filter that keeps only the selected values of a dimension, or nil when none were selected.
func inFilter(field string, domain map[string][]any) *metricsview.Expression {
	values, ok := domain[field]
	if !ok || len(values) == 0 {
		return nil
	}
	return &metricsview.Expression{Condition: &metricsview.Condition{
		Operator:    metricsview.OperatorIn,
		Expressions: []*metricsview.Expression{{Name: field}, {Value: values}},
	}}
}

// and combines filters with AND, skipping nil ones.
func and(exprs ...*metricsview.Expression) *metricsview.Expression {
	var res []*metricsview.Expression
	for _, e := range exprs {
		if e != nil {
			res = append(res, e)
		}
	}
	switch len(res) {
	case 0:
		return nil
	case 1:
		return res[0]
	default:
		return &metricsview.Expression{Condition: &metricsview.Condition{Operator: metricsview.OperatorAnd, Expressions: res}}
	}
}
