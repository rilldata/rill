package chartquery

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

// Budgets for the facet channels of a Flint chart. Axes take the budget of a nominal x axis and series channels
// that of a color split (see the constants in build.go); small multiples stay small.
const (
	defaultFacetLimit = 6
	// MaxFacetValues caps the values of a column or row facet channel.
	MaxFacetValues = 8
)

// DetailChannel is a Rill extension to Flint's channels: a dimension bound to it sets the grain of the rows
// (one row per value) without being drawn, which charts of individual observations need on aggregated data.
// Chart types that have their own detail channel keep its meaning; for the others the widget drops it before compiling.
const DetailChannel = "detail"

//go:embed flint_catalog.json
var flintCatalogJSON []byte

// FlintCatalog lists the Flint chart types create_chart accepts, with the channels and properties of each.
// It is generated from the installed flint-chart package by web-mcp-app/scripts/generate-flint-catalog.mjs,
// so the server validates against the same compiler the widget renders with.
type FlintCatalog struct {
	FlintVersion string            `json:"flint_version"`
	ChartTypes   []*FlintChartType `json:"chart_types"`
}

// FlintChartType is a chart type of the catalog.
type FlintChartType struct {
	Name       string           `json:"name"`
	Channels   []string         `json:"channels"`
	Properties []*FlintProperty `json:"properties"`
}

// FlintProperty is a tunable property of a chart type (a key of chartProperties in a chart spec).
type FlintProperty struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	// Type is "continuous", "discrete" or "binary".
	Type    string                 `json:"type"`
	Min     *float64               `json:"min,omitempty"`
	Max     *float64               `json:"max,omitempty"`
	Step    *float64               `json:"step,omitempty"`
	Options []*FlintPropertyOption `json:"options,omitempty"`
	Default any                    `json:"default,omitempty"`
}

// FlintPropertyOption is a value of a discrete property.
type FlintPropertyOption struct {
	Value any    `json:"value"`
	Label string `json:"label"`
}

var flintCatalog = sync.OnceValue(func() *FlintCatalog {
	c := &FlintCatalog{}
	if err := json.Unmarshal(flintCatalogJSON, c); err != nil {
		panic(fmt.Errorf("chartquery: invalid flint_catalog.json: %w", err))
	}
	return c
})

// Catalog returns the embedded Flint chart catalog.
func Catalog() *FlintCatalog {
	return flintCatalog()
}

// ChartType returns the catalog entry of a chart type, or nil when the name is unknown.
func (c *FlintCatalog) ChartType(name string) *FlintChartType {
	for _, ct := range c.ChartTypes {
		if ct.Name == name {
			return ct
		}
	}
	return nil
}

// Names returns the chart type names in catalog order.
func (c *FlintCatalog) Names() []string {
	res := make([]string, len(c.ChartTypes))
	for i, ct := range c.ChartTypes {
		res[i] = ct.Name
	}
	return res
}

// AcceptsChannel reports whether a chart spec may bind the channel: one of the chart type's own channels, or the detail extension.
func (ct *FlintChartType) AcceptsChannel(channel string) bool {
	return channel == DetailChannel || slices.Contains(ct.Channels, channel)
}

// FieldKind is the role of a metrics view field in a chart: a measure, a dimension, or a time dimension.
type FieldKind string

const (
	FieldKindMeasure   FieldKind = "measure"
	FieldKindDimension FieldKind = "dimension"
	FieldKindTime      FieldKind = "time"
)

// FlintSpec is a parsed Flint chart spec (chartType and encodings) bound to a metrics view through the data block of create_chart.
// Unlike Spec, it is not normalized per chart type: the queries follow from the kinds of the bound fields alone.
type FlintSpec struct {
	ChartType           string
	MetricsView         string
	TimeRange           *TimeRange
	ComparisonTimeRange *TimeRange
	Where               *metricsview.Expression
	// TimeGrain floors the bound time fields. It is set only when one is bound, defaulting to day.
	TimeGrain metricsview.TimeGrain
	// Encodings are the channel bindings in the channel order of the chart type (the detail extension last);
	// a channel bound to an array of fields has one entry per field.
	Encodings []*FlintEncoding
}

// FlintEncoding is one field bound to a channel.
type FlintEncoding struct {
	Channel string
	Field   string
	Kind    FieldKind
	// Limit is the top-N limit the spec asks for on a dimension channel, or 0 for the default.
	Limit int
	// SortOrder is "ascending", "descending" or empty; SortBy names the channel or the field the values are sorted by.
	SortOrder string
	SortBy    string
}

// rawFlintData is the wire shape of the data block of a Flint chart call.
type rawFlintData struct {
	MetricsView         string                  `mapstructure:"metrics_view"`
	TimeRange           *rawTimeRange           `mapstructure:"time_range"`
	ComparisonTimeRange *rawTimeRange           `mapstructure:"comparison_time_range"`
	Where               *metricsview.Expression `mapstructure:"where"`
	TimeGrain           string                  `mapstructure:"time_grain"`
}

// rawFlintEncoding is the wire shape of an encoding object: Flint's own keys plus Rill's limit.
type rawFlintEncoding struct {
	Field     string `mapstructure:"field"`
	Type      string `mapstructure:"type"`
	Aggregate string `mapstructure:"aggregate"`
	SortOrder string `mapstructure:"sortOrder"`
	SortBy    string `mapstructure:"sortBy"`
	Scheme    string `mapstructure:"scheme"`
	Limit     int    `mapstructure:"limit"`
}

// flintEncodingKeys are the keys an encoding object may carry.
var flintEncodingKeys = []string{"field", "type", "aggregate", "sortOrder", "sortBy", "scheme", "limit"}

// ParseFlintSpec parses the chart_spec and data arguments of a Flint chart call.
// kinds classifies the metrics view's fields; a field the chart binds must be in it.
// It validates the chart type and channels against the catalog, and what the queries need from the data block.
func ParseFlintSpec(chartSpec, data map[string]any, kinds map[string]FieldKind) (*FlintSpec, error) {
	chartType, _ := chartSpec["chartType"].(string)
	ct := Catalog().ChartType(chartType)
	if ct == nil {
		return nil, fmt.Errorf("unknown chart type %q; the chart types are: %s", chartType, strings.Join(Catalog().Names(), ", "))
	}

	raw := &rawFlintData{}
	if err := mapstructureutil.WeakDecode(data, raw); err != nil {
		return nil, fmt.Errorf("invalid data: %w", err)
	}
	if raw.MetricsView == "" {
		return nil, fmt.Errorf("data must contain a 'metrics_view' field")
	}
	if raw.TimeRange == nil || raw.TimeRange.Start.IsZero() || raw.TimeRange.End.IsZero() {
		return nil, fmt.Errorf("data must contain a 'time_range' field with 'start' and 'end' properties")
	}

	encodings, ok := chartSpec["encodings"].(map[string]any)
	if !ok || len(encodings) == 0 {
		return nil, fmt.Errorf("chart_spec must contain an 'encodings' object that binds fields to channels")
	}
	for _, channel := range slices.Sorted(maps.Keys(encodings)) {
		if !ct.AcceptsChannel(channel) {
			return nil, fmt.Errorf("chart type %q does not accept channel %q; its channels are: %s", chartType, channel, strings.Join(ct.Channels, ", "))
		}
	}

	res := &FlintSpec{
		ChartType:   chartType,
		MetricsView: raw.MetricsView,
		TimeRange:   raw.TimeRange.toTimeRange(),
		Where:       raw.Where,
	}
	if raw.ComparisonTimeRange != nil && !raw.ComparisonTimeRange.Start.IsZero() && !raw.ComparisonTimeRange.End.IsZero() {
		res.ComparisonTimeRange = raw.ComparisonTimeRange.toTimeRange()
	}

	// Channels are visited in the order the chart type declares them, which keeps the plan stable regardless of the input's key order.
	channels := ct.Channels
	if !slices.Contains(channels, DetailChannel) {
		channels = append(slices.Clone(channels), DetailChannel)
	}
	for _, channel := range channels {
		value, ok := encodings[channel]
		if !ok {
			continue
		}
		entries, err := parseFlintEncoding(channel, value)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			kind, ok := kinds[e.Field]
			if !ok {
				return nil, fmt.Errorf("field %q on channel %q not found in metrics view %q", e.Field, channel, raw.MetricsView)
			}
			e.Kind = kind
			res.Encodings = append(res.Encodings, e)
		}
	}
	if len(res.Encodings) == 0 {
		return nil, fmt.Errorf("chart_spec must bind at least one field")
	}

	// The grain only matters when a time field is bound, in which case it defaults to day.
	if res.TimeField() != "" {
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

// parseFlintEncoding parses the value of a channel: a field name, an encoding object, or an array of either.
func parseFlintEncoding(channel string, value any) ([]*FlintEncoding, error) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil, fmt.Errorf("channel %q binds an empty field name", channel)
		}
		return []*FlintEncoding{{Channel: channel, Field: v}}, nil
	case map[string]any:
		for _, key := range slices.Sorted(maps.Keys(v)) {
			if !slices.Contains(flintEncodingKeys, key) {
				return nil, fmt.Errorf("channel %q has an unknown encoding key %q; the keys are: %s", channel, key, strings.Join(flintEncodingKeys, ", "))
			}
		}
		raw := &rawFlintEncoding{}
		if err := mapstructureutil.WeakDecode(v, raw); err != nil {
			return nil, fmt.Errorf("channel %q has an invalid encoding: %w", channel, err)
		}
		if raw.Field == "" {
			return nil, fmt.Errorf("channel %q must name a field", channel)
		}
		switch raw.SortOrder {
		case "", "ascending", "descending":
		default:
			return nil, fmt.Errorf("channel %q has an invalid sortOrder %q; use ascending or descending", channel, raw.SortOrder)
		}
		return []*FlintEncoding{{Channel: channel, Field: raw.Field, Limit: raw.Limit, SortOrder: raw.SortOrder, SortBy: raw.SortBy}}, nil
	case []any:
		if len(v) == 0 {
			return nil, fmt.Errorf("channel %q binds an empty array", channel)
		}
		var res []*FlintEncoding
		for _, item := range v {
			entries, err := parseFlintEncoding(channel, item)
			if err != nil {
				return nil, err
			}
			res = append(res, entries...)
		}
		return res, nil
	default:
		return nil, fmt.Errorf("channel %q must bind a field name, an encoding object, or an array of them, got %T", channel, value)
	}
}

// BuildFlint plans the queries for a Flint chart spec from the kinds of its bound fields:
// a top-N query for each dimension on an axis, series or facet channel, and a main query that groups by every bound dimension
// (time fields floored to the grain), selects every bound measure (with its comparison value when a comparison is set),
// filters to the top-N values and keeps the row cap.
func BuildFlint(s *FlintSpec) (*Plan, error) {
	measures := s.Measures()
	if len(measures) == 0 {
		return nil, fmt.Errorf("%s needs at least one measure bound to a channel", s.ChartType)
	}
	dims := s.Dimensions()
	timeFields := s.timeFields()

	plan := &Plan{Spec: s.executeSpec(), RowLimit: MaxRows}
	if len(timeFields) > 0 {
		plan.TimeField = timeFields[0]
	}

	// One top-N query per dimension channel with a budget; a dimension bound twice is selected on its first channel.
	topN := make(map[string]bool)
	for _, e := range dims {
		def, maximum, ok := flintChannelBudget(e.Channel)
		if !ok || topN[e.Field] {
			continue
		}
		topN[e.Field] = true
		limit := capLimit(e.Limit, def, maximum)
		plan.TopN = append(plan.TopN, TopNQuery{
			Channel: e.Channel,
			Field:   e.Field,
			Limit:   limit,
			Query: &metricsview.Query{
				MetricsView: s.MetricsView,
				Dimensions:  []metricsview.Dimension{{Name: e.Field}},
				Measures:    namedMeasures(measures),
				Sort:        []metricsview.Sort{s.topNSort(e, measures[0])},
				Where:       and(s.Where, nullExclusion(nominalField(e.Field))),
				TimeRange:   s.TimeRange.toQuery(),
				TimeZone:    s.TimeRange.TimeZone,
				Limit:       limitPtr(limit + 1),
			},
		})
	}

	// Temporal charts are queried newest first, so that the row cuts in Execute keep the most recent rows.
	plan.Main = func(domain map[string][]any) *metricsview.Query {
		q := &metricsview.Query{
			MetricsView:         s.MetricsView,
			Measures:            namedMeasures(measures),
			Where:               s.Where,
			TimeRange:           s.TimeRange.toQuery(),
			ComparisonTimeRange: s.ComparisonTimeRange.toQuery(),
			TimeZone:            s.TimeRange.TimeZone,
			Limit:               limitPtr(plan.RowLimit + 1),
		}
		if s.ComparisonTimeRange != nil {
			q.Measures = withPrevMeasures(measures)
		}
		for _, name := range timeFields {
			q.Dimensions = append(q.Dimensions, dimensionFor(temporalField(name), s.TimeGrain))
		}
		seen := make(map[string]bool)
		for _, e := range dims {
			if seen[e.Field] {
				continue
			}
			seen[e.Field] = true
			q.Dimensions = append(q.Dimensions, metricsview.Dimension{Name: e.Field})
			if topN[e.Field] {
				q.Where = and(q.Where, nullExclusion(nominalField(e.Field)), inFilter(nominalField(e.Field), domain))
			}
		}
		if plan.TimeField != "" {
			q.Sort = []metricsview.Sort{{Name: plan.TimeField, Desc: true}}
		} else {
			q.Sort = []metricsview.Sort{{Name: measures[0], Desc: true}}
		}
		return q
	}

	return plan, nil
}

// flintChannelBudget returns the default and maximum top-N limit of a dimension channel, and whether the channel has a budget at all.
// Axes take the budget of a nominal x axis, series channels that of a color split, facets stay small;
// detail and the remaining channels are bounded by the row cap instead.
func flintChannelBudget(channel string) (def, maximum int, ok bool) {
	switch channel {
	case "x", "y":
		return defaultNominalLimit, MaxNominalValues, true
	case "color", "group", "shape", "strokeDash":
		return defaultColorLimit, MaxColorValues, true
	case "column", "row":
		return defaultFacetLimit, MaxFacetValues, true
	default:
		return 0, 0, false
	}
}

// topNSort resolves the sort of a dimension's top-N query: sortBy names a channel or a field, defaulting to the first measure;
// a sort by the dimension itself orders its values, anything else orders by that measure. sortOrder flips the direction,
// which defaults to descending for measures and ascending for the dimension's own values.
func (s *FlintSpec) topNSort(e *FlintEncoding, firstMeasure string) metricsview.Sort {
	target := firstMeasure
	if e.SortBy != "" {
		if f := s.FieldOn(e.SortBy); f != nil {
			target = f.Field
		} else if s.fieldNamed(e.SortBy) != nil {
			target = e.SortBy
		}
	}
	sort := metricsview.Sort{Name: target, Desc: true}
	if target == e.Field {
		sort.Desc = false
	} else if s.fieldNamed(target).Kind != FieldKindMeasure {
		// Sorting by another dimension is not expressible in a single-dimension query; fall back to the measure.
		sort = metricsview.Sort{Name: firstMeasure, Desc: true}
	}
	switch e.SortOrder {
	case "ascending":
		sort.Desc = false
	case "descending":
		sort.Desc = true
	}
	return sort
}

// FieldOn returns the first field bound to a channel, or nil.
func (s *FlintSpec) FieldOn(channel string) *FlintEncoding {
	for _, e := range s.Encodings {
		if e.Channel == channel {
			return e
		}
	}
	return nil
}

// fieldNamed returns the first binding of a field, or nil.
func (s *FlintSpec) fieldNamed(name string) *FlintEncoding {
	for _, e := range s.Encodings {
		if e.Field == name {
			return e
		}
	}
	return nil
}

// Fields returns the names of the metrics view fields the chart binds, in channel order and without duplicates.
func (s *FlintSpec) Fields() []string {
	var res []string
	seen := make(map[string]bool)
	for _, e := range s.Encodings {
		if !seen[e.Field] {
			seen[e.Field] = true
			res = append(res, e.Field)
		}
	}
	return res
}

// Measures returns the bound measures in channel order, without duplicates.
func (s *FlintSpec) Measures() []string {
	var res []string
	seen := make(map[string]bool)
	for _, e := range s.Encodings {
		if e.Kind == FieldKindMeasure && !seen[e.Field] {
			seen[e.Field] = true
			res = append(res, e.Field)
		}
	}
	return res
}

// Dimensions returns the bindings of dimensions other than time fields, in channel order. A dimension bound to several channels appears once per channel.
func (s *FlintSpec) Dimensions() []*FlintEncoding {
	var res []*FlintEncoding
	for _, e := range s.Encodings {
		if e.Kind == FieldKindDimension {
			res = append(res, e)
		}
	}
	return res
}

// TimeField returns the first bound time field, or an empty string.
func (s *FlintSpec) TimeField() string {
	if fields := s.timeFields(); len(fields) > 0 {
		return fields[0]
	}
	return ""
}

func (s *FlintSpec) timeFields() []string {
	var res []string
	for _, e := range s.Encodings {
		if e.Kind == FieldKindTime && !slices.Contains(res, e.Field) {
			res = append(res, e.Field)
		}
	}
	return res
}

// TimeGrainName returns the TIME_GRAIN_* name of the resolved grain, or an empty string when no time field is bound.
func (s *FlintSpec) TimeGrainName() string {
	return s.executeSpec().TimeGrainName()
}

// executeSpec is the Spec view of a Flint spec that Execute reads: the time grain and the query scope; it binds no fields.
func (s *FlintSpec) executeSpec() *Spec {
	return &Spec{
		ChartType:           s.ChartType,
		MetricsView:         s.MetricsView,
		TimeRange:           s.TimeRange,
		ComparisonTimeRange: s.ComparisonTimeRange,
		Where:               s.Where,
		TimeGrain:           s.TimeGrain,
	}
}

func nominalField(name string) *Field {
	return &Field{Field: name, Type: TypeNominal}
}

func temporalField(name string) *Field {
	return &Field{Field: name, Type: TypeTemporal}
}
