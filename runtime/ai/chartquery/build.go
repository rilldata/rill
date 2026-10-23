package chartquery

import (
	"fmt"

	"github.com/rilldata/rill/runtime/metricsview"
)

// Budgets for the chart data. They bound the number of categories, rows and bytes a chart result may carry,
// since MCP hosts cap tool results and the widget renders every row it receives.
const (
	// MaxNominalValues caps the values of a nominal x axis, funnel stage or circular slice dimension.
	MaxNominalValues = 50
	// MaxColorValues caps the values of a color split.
	MaxColorValues = 20
	// MaxHeatmapAxisValues caps the values of each heatmap axis.
	MaxHeatmapAxisValues = 40
	// MaxRows caps the rows of the main query.
	MaxRows = 1500
	// MaxDataBytes caps the JSON size of the rows; rows are trimmed from the tail to fit.
	MaxDataBytes = 100 << 10
)

// Default limits applied when the spec does not set one. They match the frontend chart providers.
const (
	defaultNominalLimit = 20
	defaultColorLimit   = 10
	defaultSliceLimit   = 20
	defaultStageLimit   = 15
	defaultHeatmapLimit = 40
)

// Suffixes of the derived measures, shared with the frontend.
const (
	// PrevSuffix names the comparison period value of a measure.
	PrevSuffix = "_prev"
	// deltaSuffix names the absolute change of a measure used for delta sorting.
	deltaSuffix = "_delta"
)

// Plan is the set of queries that fetch a chart's data.
type Plan struct {
	Spec *Spec
	// TopN are pre-queries that select the values of a dimension (the x axis, color split, funnel stage or heatmap axes).
	// They run in order before the main query, and their values filter it and order its categories.
	TopN []TopNQuery
	// Main builds the main query from the selected domain values (keyed by field name).
	Main func(domain map[string][]any) *metricsview.Query
	// Extra are queries whose single row is returned as totals (donut and pie charts), keyed by name.
	Extra map[string]*metricsview.Query
	// RowLimit caps the rows returned from the main query; the query itself asks for one more row to detect truncation.
	RowLimit int
	// TimeField is the temporal field the main query is ordered by, newest first, when the chart has a time axis.
	// Ordering newest first makes the row cap and the byte budget drop the oldest rows; Execute restores ascending order.
	TimeField string
	// MeasureDomain lists the measure names to report as the domain of a "rill_measures" color split.
	MeasureDomain []string
}

// TopNQuery selects the values of one dimension.
type TopNQuery struct {
	// Channel is the spec channel whose values are selected (x, y, color or stage), as named in the input spec.
	Channel string `json:"channel"`
	// Field is the dimension whose values are selected.
	Field string `json:"field"`
	// Limit is the number of values to keep. The query asks for one more row to detect truncation.
	Limit int                `json:"limit"`
	Query *metricsview.Query `json:"query"`
}

// Build plans the queries for a chart spec.
func Build(s *Spec) (*Plan, error) {
	switch s.ChartType {
	case BarChart, LineChart, AreaChart, StackedBar, StackedBarNormalized:
		return buildCartesian(s)
	case DonutChart, PieChart:
		return buildCircular(s)
	case FunnelChart:
		return buildFunnel(s)
	case Heatmap:
		return buildHeatmap(s)
	case ComboChart:
		return buildCombo(s)
	case ScatterPlot:
		return buildScatter(s)
	default:
		return nil, fmt.Errorf("unsupported chart type %q", s.ChartType)
	}
}

// buildCartesian mirrors CartesianChartProvider: a top-N query for a nominal x axis (sorted by the measure, or by its
// period-over-period delta), a top-N query for the color split, and a main query that filters to those values,
// floors temporal x values to the grain and adds a <measure>_prev column for each measure when a comparison is set.
func buildCartesian(s *Spec) (*Plan, error) {
	measures := s.Y.measures()
	if len(measures) == 0 {
		return nil, fmt.Errorf("%s requires a quantitative y field", s.ChartType)
	}
	if s.X == nil || s.X.Field == "" {
		return nil, fmt.Errorf("%s requires an x field", s.ChartType)
	}
	multiMeasure := len(s.Y.Fields) > 0
	comparison := s.ComparisonTimeRange != nil

	plan := &Plan{Spec: s, RowLimit: MaxRows}

	// Resolve the x sort against the vertical layout. Delta sorts need a comparison, otherwise they fall back to the measure.
	xSort := s.X.Sort
	if xSort == "" || len(s.X.SortValues) > 0 {
		xSort = "-y"
	}
	if !comparison {
		switch xSort {
		case "y_delta":
			xSort = "y"
		case "-y_delta":
			xSort = "-y"
		}
	}
	var topNSort, mainSort metricsview.Sort
	var deltaSort bool
	switch xSort {
	case "x", "-x":
		topNSort = metricsview.Sort{Name: s.X.Field, Desc: xSort == "-x"}
		mainSort = topNSort
	case "y_delta", "-y_delta":
		deltaSort = true
		topNSort = metricsview.Sort{Name: measures[0] + deltaSuffix, Desc: xSort == "-y_delta"}
		mainSort = metricsview.Sort{Name: measures[0], Desc: xSort == "-y_delta"}
	default:
		topNSort = metricsview.Sort{Name: measures[0], Desc: xSort == "-y"}
		mainSort = topNSort
	}

	colorField := s.Color
	if colorField != nil && (colorField.Field == "" || colorField.Field == MeasuresField || colorField.Type == TypeValue || multiMeasure) {
		colorField = nil
	}
	if multiMeasure && s.Color != nil {
		plan.MeasureDomain = measures
	}

	// The x axis of a horizontal bar chart is the input's y channel (see Spec.Horizontal).
	xChannel := "x"
	if s.Horizontal {
		xChannel = "y"
	}
	if s.X.isNominal() && len(s.X.SortValues) == 0 {
		limit := capLimit(s.X.Limit, defaultNominalLimit, MaxNominalValues)
		q := &metricsview.Query{
			MetricsView: s.MetricsView,
			Dimensions:  []metricsview.Dimension{{Name: s.X.Field}},
			Measures:    namedMeasures(measures),
			Sort:        []metricsview.Sort{topNSort},
			Where:       and(s.Where, nullExclusion(s.X)),
			TimeRange:   s.TimeRange.toQuery(),
			TimeZone:    s.TimeRange.TimeZone,
			Limit:       limitPtr(limit + 1),
		}
		if deltaSort {
			q.Measures = append(q.Measures, metricsview.Measure{
				Name:    measures[0] + deltaSuffix,
				Compute: &metricsview.MeasureCompute{ComparisonDelta: &metricsview.MeasureComputeComparisonDelta{Measure: measures[0]}},
			})
			q.ComparisonTimeRange = s.ComparisonTimeRange.toQuery()
		}
		plan.TopN = append(plan.TopN, TopNQuery{Channel: xChannel, Field: s.X.Field, Limit: limit, Query: q})
	}

	if colorField != nil && len(colorField.SortValues) == 0 {
		limit := capLimit(colorField.Limit, defaultColorLimit, MaxColorValues)
		plan.TopN = append(plan.TopN, TopNQuery{
			Channel: "color",
			Field:   colorField.Field,
			Limit:   limit,
			Query: &metricsview.Query{
				MetricsView: s.MetricsView,
				Dimensions:  []metricsview.Dimension{{Name: colorField.Field}},
				Measures:    namedMeasures(measures),
				Sort:        []metricsview.Sort{{Name: measures[0], Desc: true}},
				Where:       and(s.Where, nullExclusion(colorField)),
				TimeRange:   s.TimeRange.toQuery(),
				TimeZone:    s.TimeRange.TimeZone,
				Limit:       limitPtr(limit + 1),
			},
		})
	}

	// Temporal charts are queried newest first, so that the row cuts in Execute keep the most recent rows.
	if s.X.isTemporal() {
		plan.TimeField = s.X.Field
	}
	plan.Main = func(domain map[string][]any) *metricsview.Query {
		q := &metricsview.Query{
			MetricsView:         s.MetricsView,
			Dimensions:          []metricsview.Dimension{dimensionFor(s.X, s.TimeGrain)},
			Measures:            namedMeasures(measures),
			Where:               and(s.Where, nullExclusion(s.X), inFilter(s.X, domain)),
			TimeRange:           s.TimeRange.toQuery(),
			ComparisonTimeRange: s.ComparisonTimeRange.toQuery(),
			TimeZone:            s.TimeRange.TimeZone,
			Limit:               limitPtr(plan.RowLimit + 1),
		}
		if comparison {
			q.Measures = withPrevMeasures(measures)
		}
		if s.X.isTemporal() {
			q.Sort = []metricsview.Sort{{Name: s.X.Field, Desc: true}}
		} else {
			q.Sort = []metricsview.Sort{mainSort}
		}
		if colorField != nil {
			q.Dimensions = append(q.Dimensions, metricsview.Dimension{Name: colorField.Field})
			q.Where = and(q.Where, inFilter(colorField, domain))
			q.Sort = append(q.Sort, metricsview.Sort{Name: colorField.Field})
		}
		return q
	}

	return plan, nil
}

// buildCircular mirrors CircularChartProvider: a top-N query for the slice dimension sorted by the measure,
// a main query filtered to those values, and a total query for the center label and percent-of-total tooltip.
func buildCircular(s *Spec) (*Plan, error) {
	if s.Measure == nil || s.Measure.Field == "" {
		return nil, fmt.Errorf("%s requires a measure field", s.ChartType)
	}
	measure := s.Measure.Field
	color := s.Color
	if color != nil && color.Field == "" {
		color = nil
	}

	plan := &Plan{Spec: s, RowLimit: MaxRows, Extra: map[string]*metricsview.Query{}}

	var sort []metricsview.Sort
	if color != nil {
		limit := capLimit(color.Limit, defaultSliceLimit, MaxNominalValues)
		switch color.Sort {
		case "color", "-color":
			sort = []metricsview.Sort{{Name: color.Field, Desc: color.Sort == "-color"}}
		case "measure":
			sort = []metricsview.Sort{{Name: measure}}
		default:
			sort = []metricsview.Sort{{Name: measure, Desc: true}}
		}
		if len(color.SortValues) == 0 {
			plan.TopN = append(plan.TopN, TopNQuery{
				Channel: "color",
				Field:   color.Field,
				Limit:   limit,
				Query: &metricsview.Query{
					MetricsView: s.MetricsView,
					Dimensions:  []metricsview.Dimension{{Name: color.Field}},
					Measures:    namedMeasures([]string{measure}),
					Sort:        sort,
					Where:       and(s.Where, nullExclusion(color)),
					TimeRange:   s.TimeRange.toQuery(),
					TimeZone:    s.TimeRange.TimeZone,
					Limit:       limitPtr(limit + 1),
				},
			})
		}
		plan.RowLimit = limit
	}

	plan.Extra["total"] = &metricsview.Query{
		MetricsView: s.MetricsView,
		Measures:    namedMeasures([]string{measure}),
		Where:       and(s.Where, nullExclusion(color)),
		TimeRange:   s.TimeRange.toQuery(),
		TimeZone:    s.TimeRange.TimeZone,
	}

	plan.Main = func(domain map[string][]any) *metricsview.Query {
		q := &metricsview.Query{
			MetricsView: s.MetricsView,
			Measures:    namedMeasures([]string{measure}),
			Where:       s.Where,
			Sort:        sort,
			TimeRange:   s.TimeRange.toQuery(),
			TimeZone:    s.TimeRange.TimeZone,
			Limit:       limitPtr(plan.RowLimit + 1),
		}
		if color != nil {
			q.Dimensions = []metricsview.Dimension{{Name: color.Field}}
			q.Where = and(s.Where, inFilter(color, domain))
		}
		return q
	}

	return plan, nil
}

// buildFunnel mirrors FunnelChartProvider: in dimension mode, a top-N query for the stage dimension sorted by the measure
// and a main query filtered to those stages; in measures mode, a single query that returns every measure as a stage.
func buildFunnel(s *Spec) (*Plan, error) {
	measures := s.Measure.measures()
	if len(measures) == 0 {
		return nil, fmt.Errorf("%s requires a measure field", s.ChartType)
	}
	multiMeasure := s.BreakdownMode == "measures" || (s.Stage == nil && len(measures) > 1)

	plan := &Plan{Spec: s, RowLimit: MaxRows}

	if multiMeasure {
		plan.RowLimit = 1
		plan.Main = func(map[string][]any) *metricsview.Query {
			return &metricsview.Query{
				MetricsView: s.MetricsView,
				Measures:    namedMeasures(measures),
				Where:       s.Where,
				TimeRange:   s.TimeRange.toQuery(),
				TimeZone:    s.TimeRange.TimeZone,
				Limit:       limitPtr(plan.RowLimit + 1),
			}
		}
		return plan, nil
	}

	if s.Stage == nil || s.Stage.Field == "" {
		return nil, fmt.Errorf("%s requires a stage field in dimension mode", s.ChartType)
	}
	measure := measures[0]
	limit := capLimit(s.Stage.Limit, defaultStageLimit, MaxNominalValues)
	sort := []metricsview.Sort{{Name: measure, Desc: s.Stage.Sort != "y"}}
	plan.RowLimit = limit

	if len(s.Stage.SortValues) == 0 {
		plan.TopN = append(plan.TopN, TopNQuery{
			Channel: "stage",
			Field:   s.Stage.Field,
			Limit:   limit,
			Query: &metricsview.Query{
				MetricsView: s.MetricsView,
				Dimensions:  []metricsview.Dimension{{Name: s.Stage.Field}},
				Measures:    namedMeasures([]string{measure}),
				Sort:        sort,
				Where:       and(s.Where, nullExclusion(s.Stage)),
				TimeRange:   s.TimeRange.toQuery(),
				TimeZone:    s.TimeRange.TimeZone,
				Limit:       limitPtr(limit + 1),
			},
		})
	}

	plan.Main = func(domain map[string][]any) *metricsview.Query {
		return &metricsview.Query{
			MetricsView: s.MetricsView,
			Dimensions:  []metricsview.Dimension{{Name: s.Stage.Field}},
			Measures:    namedMeasures([]string{measure}),
			Where:       and(s.Where, nullExclusion(s.Stage), inFilter(s.Stage, domain)),
			Sort:        sort,
			TimeRange:   s.TimeRange.toQuery(),
			TimeZone:    s.TimeRange.TimeZone,
			Limit:       limitPtr(plan.RowLimit + 1),
		}
	}

	return plan, nil
}

// buildHeatmap mirrors HeatmapChartProvider: a top-N query per nominal axis sorted by the color measure,
// and a main query over both axes filtered to the selected values.
func buildHeatmap(s *Spec) (*Plan, error) {
	if s.Color == nil || s.Color.Field == "" {
		return nil, fmt.Errorf("%s requires a quantitative color field", s.ChartType)
	}
	if s.X == nil || s.X.Field == "" || s.Y == nil || s.Y.Field == "" {
		return nil, fmt.Errorf("%s requires x and y fields", s.ChartType)
	}
	measure := s.Color.Field

	plan := &Plan{Spec: s, RowLimit: MaxRows}

	axisSort := func(axis *Field) metricsview.Sort {
		switch axis.Sort {
		case "x", "-x":
			return metricsview.Sort{Name: s.X.Field, Desc: axis.Sort == "-x"}
		case "y", "-y":
			return metricsview.Sort{Name: s.Y.Field, Desc: axis.Sort == "-y"}
		case "color":
			return metricsview.Sort{Name: measure}
		default:
			return metricsview.Sort{Name: measure, Desc: true}
		}
	}
	for _, axis := range []struct {
		channel string
		field   *Field
	}{{"x", s.X}, {"y", s.Y}} {
		if !axis.field.isNominal() || len(axis.field.SortValues) > 0 {
			continue
		}
		limit := capLimit(axis.field.Limit, defaultHeatmapLimit, MaxHeatmapAxisValues)
		plan.TopN = append(plan.TopN, TopNQuery{
			Channel: axis.channel,
			Field:   axis.field.Field,
			Limit:   limit,
			Query: &metricsview.Query{
				MetricsView: s.MetricsView,
				Dimensions:  []metricsview.Dimension{{Name: axis.field.Field}},
				Measures:    namedMeasures([]string{measure}),
				Sort:        []metricsview.Sort{axisSort(axis.field)},
				Where:       and(s.Where, nullExclusion(axis.field)),
				TimeRange:   s.TimeRange.toQuery(),
				TimeZone:    s.TimeRange.TimeZone,
				Limit:       limitPtr(limit + 1),
			},
		})
	}

	if s.X.isTemporal() {
		plan.TimeField = s.X.Field
	}
	plan.Main = func(domain map[string][]any) *metricsview.Query {
		q := &metricsview.Query{
			MetricsView: s.MetricsView,
			Dimensions:  []metricsview.Dimension{dimensionFor(s.X, s.TimeGrain), dimensionFor(s.Y, s.TimeGrain)},
			Measures:    namedMeasures([]string{measure}),
			Where:       and(s.Where, inFilter(s.X, domain), inFilter(s.Y, domain)),
			TimeRange:   s.TimeRange.toQuery(),
			TimeZone:    s.TimeRange.TimeZone,
			Limit:       limitPtr(plan.RowLimit + 1),
		}
		// Sorting x descending keeps the newest buckets when a temporal heatmap is truncated;
		// for a nominal x the widget orders the cells by the top-N domain, so the row order is irrelevant.
		q.Sort = []metricsview.Sort{{Name: s.X.Field, Desc: true}}
		return q
	}

	return plan, nil
}

// buildCombo mirrors ComboChartProvider: a top-N query for a nominal x axis sorted by the first measure,
// and a main query with both measures.
func buildCombo(s *Spec) (*Plan, error) {
	var measures []string
	for _, f := range []*Field{s.Y1, s.Y2} {
		if f != nil && f.Field != "" && f.Type == TypeQuantitative {
			measures = append(measures, f.Field)
		}
	}
	if len(measures) == 0 {
		return nil, fmt.Errorf("%s requires quantitative y1 and y2 fields", s.ChartType)
	}
	if s.X == nil || s.X.Field == "" {
		return nil, fmt.Errorf("%s requires an x field", s.ChartType)
	}

	plan := &Plan{Spec: s, RowLimit: MaxRows}
	if s.Color != nil && s.Color.Field != "" {
		plan.MeasureDomain = measures
	}

	var sort metricsview.Sort
	switch s.X.Sort {
	case "x", "-x":
		sort = metricsview.Sort{Name: s.X.Field, Desc: s.X.Sort == "-x"}
	case "y":
		sort = metricsview.Sort{Name: measures[0]}
	default:
		sort = metricsview.Sort{Name: measures[0], Desc: true}
	}

	if s.X.isNominal() {
		limit := capLimit(s.X.Limit, defaultNominalLimit, MaxNominalValues)
		plan.RowLimit = limit
		if len(s.X.SortValues) == 0 {
			plan.TopN = append(plan.TopN, TopNQuery{
				Channel: "x",
				Field:   s.X.Field,
				Limit:   limit,
				Query: &metricsview.Query{
					MetricsView: s.MetricsView,
					Dimensions:  []metricsview.Dimension{{Name: s.X.Field}},
					Measures:    namedMeasures(measures[:1]),
					Sort:        []metricsview.Sort{sort},
					Where:       and(s.Where, nullExclusion(s.X)),
					TimeRange:   s.TimeRange.toQuery(),
					TimeZone:    s.TimeRange.TimeZone,
					Limit:       limitPtr(limit + 1),
				},
			})
		}
	}

	if s.X.isTemporal() {
		plan.TimeField = s.X.Field
	}
	plan.Main = func(domain map[string][]any) *metricsview.Query {
		q := &metricsview.Query{
			MetricsView: s.MetricsView,
			Dimensions:  []metricsview.Dimension{dimensionFor(s.X, s.TimeGrain)},
			Measures:    namedMeasures(measures),
			Where:       and(s.Where, nullExclusion(s.X), inFilter(s.X, domain)),
			TimeRange:   s.TimeRange.toQuery(),
			TimeZone:    s.TimeRange.TimeZone,
			Limit:       limitPtr(plan.RowLimit + 1),
		}
		if s.X.isTemporal() {
			q.Sort = []metricsview.Sort{{Name: s.X.Field, Desc: true}}
		} else {
			q.Sort = []metricsview.Sort{sort}
		}
		return q
	}

	return plan, nil
}

// buildScatter mirrors ScatterPlotChartProvider: a top-N query for the color split sorted by the y measure,
// and a main query with the x, y and size measures per point dimension (and color).
func buildScatter(s *Spec) (*Plan, error) {
	var measures []string
	var dims []*Field
	for _, f := range []*Field{s.X, s.Y} {
		switch {
		case f == nil || f.Field == "":
			return nil, fmt.Errorf("%s requires x and y fields", s.ChartType)
		case f.Type == TypeQuantitative:
			measures = append(measures, f.Field)
		case f.Type == TypeTemporal:
			dims = append(dims, f)
		}
	}
	if s.Size != nil && s.Size.Field != "" && s.Size.Type == TypeQuantitative {
		measures = append(measures, s.Size.Field)
	}
	if len(measures) == 0 {
		return nil, fmt.Errorf("%s requires quantitative x and y fields", s.ChartType)
	}
	if s.Dimension.isNominal() {
		dims = append(dims, s.Dimension)
	}
	color := s.Color
	if color != nil && color.Field == "" {
		color = nil
	}
	if color != nil {
		dims = append(dims, color)
	}

	plan := &Plan{Spec: s, RowLimit: MaxRows}

	var sort []metricsview.Sort
	if s.Y != nil && s.Y.Type == TypeQuantitative {
		sort = []metricsview.Sort{{Name: s.Y.Field, Desc: true}}
	}

	if color != nil && len(color.SortValues) == 0 {
		limit := capLimit(color.Limit, defaultColorLimit, MaxColorValues)
		plan.TopN = append(plan.TopN, TopNQuery{
			Channel: "color",
			Field:   color.Field,
			Limit:   limit,
			Query: &metricsview.Query{
				MetricsView: s.MetricsView,
				Dimensions:  []metricsview.Dimension{{Name: color.Field}},
				Measures:    namedMeasures(measures),
				Sort:        sort,
				Where:       and(s.Where, nullExclusion(color)),
				TimeRange:   s.TimeRange.toQuery(),
				TimeZone:    s.TimeRange.TimeZone,
				Limit:       limitPtr(limit + 1),
			},
		})
	}

	plan.Main = func(domain map[string][]any) *metricsview.Query {
		q := &metricsview.Query{
			MetricsView: s.MetricsView,
			Measures:    namedMeasures(measures),
			Where:       and(s.Where, nullExclusion(s.Dimension), inFilter(color, domain)),
			Sort:        sort,
			TimeRange:   s.TimeRange.toQuery(),
			TimeZone:    s.TimeRange.TimeZone,
			Limit:       limitPtr(plan.RowLimit + 1),
		}
		for _, d := range dims {
			q.Dimensions = append(q.Dimensions, dimensionFor(d, s.TimeGrain))
		}
		return q
	}

	return plan, nil
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

func (tr *TimeRange) toQuery() *metricsview.TimeRange {
	if tr == nil {
		return nil
	}
	return &metricsview.TimeRange{Start: tr.Start, End: tr.End}
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

// dimensionFor returns the query dimension for a field, flooring temporal fields to the grain.
func dimensionFor(f *Field, grain metricsview.TimeGrain) metricsview.Dimension {
	if f.isTemporal() && grain != metricsview.TimeGrainUnspecified {
		return metricsview.Dimension{
			Name:    f.Field,
			Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: f.Field, Grain: grain}},
		}
	}
	return metricsview.Dimension{Name: f.Field}
}

// nullExclusion returns a filter that drops null values of a nominal field, unless the spec asks to show them.
func nullExclusion(f *Field) *metricsview.Expression {
	if !f.isNominal() || f.ShowNull {
		return nil
	}
	return &metricsview.Expression{Condition: &metricsview.Condition{
		Operator:    metricsview.OperatorNin,
		Expressions: []*metricsview.Expression{{Name: f.Field}, {Value: []any{nil}}},
	}}
}

// inFilter returns a filter that keeps only the selected values of a field, or nil when none were selected.
func inFilter(f *Field, domain map[string][]any) *metricsview.Expression {
	if f == nil || f.Field == "" {
		return nil
	}
	values, ok := domain[f.Field]
	if !ok || len(values) == 0 {
		return nil
	}
	return &metricsview.Expression{Condition: &metricsview.Condition{
		Operator:    metricsview.OperatorIn,
		Expressions: []*metricsview.Expression{{Name: f.Field}, {Value: values}},
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
