// Package chartquery turns a Rill chart configuration (the argument of the create_chart AI tool) into the metrics view
// queries that fetch its data, and runs them within fixed budgets.
//
// The query shapes mirror the frontend chart data providers in web-common/src/features/components/charts
// (top-N pre-queries for nominal axes and the color split, a main query with time flooring and comparison values,
// a total query for circular charts), so that a chart rendered outside the Rill app shows the same data as inside it.
//
// The package is pure: it builds metricsview.Query values and runs them through a caller-provided Runner.
package chartquery

import (
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

// Chart types supported by create_chart.
const (
	BarChart             = "bar_chart"
	LineChart            = "line_chart"
	AreaChart            = "area_chart"
	StackedBar           = "stacked_bar"
	StackedBarNormalized = "stacked_bar_normalized"
	DonutChart           = "donut_chart"
	PieChart             = "pie_chart"
	FunnelChart          = "funnel_chart"
	Heatmap              = "heatmap"
	ComboChart           = "combo_chart"
	ScatterPlot          = "scatter_plot"
)

// Field types used in chart specs.
const (
	TypeNominal      = "nominal"
	TypeTemporal     = "temporal"
	TypeQuantitative = "quantitative"
	TypeValue        = "value"
)

// MeasuresField is the pseudo field that colors a chart by measure instead of by a dimension.
const MeasuresField = "rill_measures"

// Spec is a parsed and normalized chart configuration.
// Cartesian charts are normalized to the vertical layout (dimension on X, measure on Y),
// which is the layout the queries are written against; Horizontal records that the input had them swapped.
type Spec struct {
	ChartType           string
	MetricsView         string
	TimeRange           *TimeRange
	ComparisonTimeRange *TimeRange
	Where               *metricsview.Expression
	// TimeGrain is the grain used to floor temporal fields. It is set only when a temporal field is used,
	// defaulting to day when the spec does not name one.
	TimeGrain metricsview.TimeGrain

	X          *Field
	Y          *Field
	Y1         *Field
	Y2         *Field
	Color      *Field
	ColorValue string // Static color (e.g. "primary") when color is not a field
	Measure    *Field
	Stage      *Field
	Dimension  *Field
	Size       *Field

	BreakdownMode string
	Horizontal    bool
}

// Field is a field configuration in a chart spec.
type Field struct {
	Field    string
	Type     string
	Fields   []string
	Limit    int
	ShowNull bool
	// Sort is the channel-relative sort ("x", "-y", "-y_delta", "-color", "-measure", ...), or empty.
	Sort string
	// SortValues is the explicit value order when the spec sorts by a list of values; it replaces the top-N query.
	SortValues []string
}

// TimeRange is an absolute time range with an optional IANA time zone.
type TimeRange struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	TimeZone string    `json:"time_zone,omitempty"`
}

// rawSpec is the wire shape of a chart spec. Unknown keys (visual options such as zeroBasedOrigin) are ignored.
type rawSpec struct {
	MetricsView         string                  `mapstructure:"metrics_view"`
	TimeRange           *rawTimeRange           `mapstructure:"time_range"`
	ComparisonTimeRange *rawTimeRange           `mapstructure:"comparison_time_range"`
	Where               *metricsview.Expression `mapstructure:"where"`
	TimeGrain           string                  `mapstructure:"time_grain"`
	X                   *rawField               `mapstructure:"x"`
	Y                   *rawField               `mapstructure:"y"`
	Y1                  *rawField               `mapstructure:"y1"`
	Y2                  *rawField               `mapstructure:"y2"`
	Color               any                     `mapstructure:"color"`
	Measure             *rawField               `mapstructure:"measure"`
	Stage               *rawField               `mapstructure:"stage"`
	Dimension           *rawField               `mapstructure:"dimension"`
	Size                *rawField               `mapstructure:"size"`
	BreakdownMode       string                  `mapstructure:"breakdownMode"`
}

type rawTimeRange struct {
	Start    time.Time `mapstructure:"start"`
	End      time.Time `mapstructure:"end"`
	TimeZone string    `mapstructure:"time_zone"`
}

type rawField struct {
	Field    string   `mapstructure:"field"`
	Type     string   `mapstructure:"type"`
	Fields   []string `mapstructure:"fields"`
	Limit    int      `mapstructure:"limit"`
	ShowNull bool     `mapstructure:"showNull"`
	Sort     any      `mapstructure:"sort"`
}

// ParseSpec parses the chart_type and spec arguments of create_chart.
// It validates only what the queries need; field existence is checked by the caller against the metrics view.
func ParseSpec(chartType string, spec map[string]any) (*Spec, error) {
	switch chartType {
	case BarChart, LineChart, AreaChart, StackedBar, StackedBarNormalized, DonutChart, PieChart, FunnelChart, Heatmap, ComboChart, ScatterPlot:
	default:
		return nil, fmt.Errorf("unsupported chart type %q", chartType)
	}

	raw := &rawSpec{}
	if err := mapstructureutil.WeakDecode(spec, raw); err != nil {
		return nil, fmt.Errorf("invalid chart spec: %w", err)
	}
	if raw.MetricsView == "" {
		return nil, fmt.Errorf("spec must contain a 'metrics_view' field")
	}
	if raw.TimeRange == nil || raw.TimeRange.Start.IsZero() || raw.TimeRange.End.IsZero() {
		return nil, fmt.Errorf("spec must contain a 'time_range' field with 'start' and 'end' properties")
	}

	res := &Spec{
		ChartType:     chartType,
		MetricsView:   raw.MetricsView,
		TimeRange:     raw.TimeRange.toTimeRange(),
		Where:         raw.Where,
		X:             raw.X.toField(),
		Y:             raw.Y.toField(),
		Y1:            raw.Y1.toField(),
		Y2:            raw.Y2.toField(),
		Measure:       raw.Measure.toField(),
		Stage:         raw.Stage.toField(),
		Dimension:     raw.Dimension.toField(),
		Size:          raw.Size.toField(),
		BreakdownMode: raw.BreakdownMode,
	}
	if raw.ComparisonTimeRange != nil && !raw.ComparisonTimeRange.Start.IsZero() && !raw.ComparisonTimeRange.End.IsZero() {
		res.ComparisonTimeRange = raw.ComparisonTimeRange.toTimeRange()
	}

	// Color is either a field configuration or a static color string (which may also be a funnel color mode).
	switch c := raw.Color.(type) {
	case nil:
	case string:
		res.ColorValue = c
	case map[string]any:
		f := &rawField{}
		if err := mapstructureutil.WeakDecode(c, f); err != nil {
			return nil, fmt.Errorf("invalid color field: %w", err)
		}
		res.Color = f.toField()
	default:
		return nil, fmt.Errorf("invalid color: expected an object or a string, got %T", raw.Color)
	}

	// Horizontal bar charts put the measure on x and the dimension on y. Normalize to the vertical layout,
	// rewriting channel-relative sorts so that they keep referring to the same field.
	switch chartType {
	case BarChart, StackedBar, StackedBarNormalized:
		if res.X != nil && res.X.Type == TypeQuantitative {
			res.X, res.Y = res.Y, res.X
			res.X.swapSortChannel()
			res.Y.swapSortChannel()
			res.Horizontal = true
		}
	}

	// Resolve the time grain: it only matters when a temporal field is used, in which case it defaults to day.
	if res.hasTemporalField() {
		grain, err := parseTimeGrain(raw.TimeGrain)
		if err != nil {
			return nil, err
		}
		if grain == metricsview.TimeGrainUnspecified {
			grain = metricsview.TimeGrainDay
		}
		res.TimeGrain = grain
	}

	return res, nil
}

func (r *rawTimeRange) toTimeRange() *TimeRange {
	return &TimeRange{Start: r.Start, End: r.End, TimeZone: r.TimeZone}
}

func (r *rawField) toField() *Field {
	if r == nil {
		return nil
	}
	f := &Field{
		Field:    r.Field,
		Type:     r.Type,
		Fields:   r.Fields,
		Limit:    r.Limit,
		ShowNull: r.ShowNull,
	}
	switch s := r.Sort.(type) {
	case string:
		f.Sort = s
	case []any:
		for _, v := range s {
			f.SortValues = append(f.SortValues, fmt.Sprintf("%v", v))
		}
	case []string:
		f.SortValues = s
	}
	return f
}

// sortChannelTranspose rewrites channel-relative sort values when the x and y channels are swapped.
var sortChannelTranspose = map[string]string{
	"x": "y", "-x": "-y", "y": "x", "-y": "-x",
	"x_delta": "y_delta", "-x_delta": "-y_delta", "y_delta": "x_delta", "-y_delta": "-x_delta",
}

func (f *Field) swapSortChannel() {
	if f == nil {
		return
	}
	if s, ok := sortChannelTranspose[f.Sort]; ok {
		f.Sort = s
	}
}

func (f *Field) isNominal() bool  { return f != nil && f.Field != "" && f.Type == TypeNominal }
func (f *Field) isTemporal() bool { return f != nil && f.Field != "" && f.Type == TypeTemporal }

// measures returns the measures named by a quantitative field: the multi-measure list plus the primary field.
func (f *Field) measures() []string {
	if f == nil {
		return nil
	}
	var res []string
	seen := make(map[string]bool)
	for _, m := range f.Fields {
		if m != "" && !seen[m] {
			seen[m] = true
			res = append(res, m)
		}
	}
	if f.Field != "" && !seen[f.Field] {
		res = append(res, f.Field)
	}
	return res
}

func (s *Spec) hasTemporalField() bool {
	for _, f := range []*Field{s.X, s.Y, s.Y1, s.Y2, s.Color, s.Measure, s.Stage, s.Dimension, s.Size} {
		if f.isTemporal() {
			return true
		}
	}
	return false
}

// Fields returns the names of the metrics view fields the chart references, in a stable order and without duplicates.
func (s *Spec) Fields() []string {
	var res []string
	seen := make(map[string]bool)
	add := func(name string) {
		if name == "" || name == MeasuresField || seen[name] {
			return
		}
		seen[name] = true
		res = append(res, name)
	}
	for _, f := range []*Field{s.X, s.Y, s.Y1, s.Y2, s.Measure, s.Stage, s.Dimension, s.Size, s.Color} {
		if f == nil {
			continue
		}
		add(f.Field)
		for _, m := range f.Fields {
			add(m)
		}
	}
	return res
}

// TimeGrainName returns the TIME_GRAIN_* name of the resolved grain, or an empty string when no temporal field is used.
func (s *Spec) TimeGrainName() string {
	if s.TimeGrain == metricsview.TimeGrainUnspecified {
		return ""
	}
	return "TIME_GRAIN_" + strings.ToUpper(string(s.TimeGrain))
}

// parseTimeGrain maps a TIME_GRAIN_* name (as used in the chart spec) to a metricsview grain.
func parseTimeGrain(s string) (metricsview.TimeGrain, error) {
	if s == "" {
		return metricsview.TimeGrainUnspecified, nil
	}
	g := metricsview.TimeGrain(strings.ToLower(strings.TrimPrefix(s, "TIME_GRAIN_")))
	if !g.Valid() || g == metricsview.TimeGrainUnspecified {
		return "", fmt.Errorf("invalid time_grain %q", s)
	}
	return g, nil
}
