package rillv1

import (
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"gopkg.in/yaml.v3"

	// Load IANA time zone data
	_ "time/tzdata"
)

// MetricsViewYAML is the raw structure of a MetricsView resource defined in YAML
type MetricsViewYAML struct {
	commonYAML         `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title              string           `yaml:"title"`
	DisplayName        string           `yaml:"display_name"` // Backwards compatibility
	Description        string           `yaml:"description"`
	Model              string           `yaml:"model"`
	Database           string           `yaml:"database"`
	DatabaseSchema     string           `yaml:"database_schema"`
	Table              string           `yaml:"table"`
	TimeDimension      string           `yaml:"timeseries"`
	Watermark          string           `yaml:"watermark"`
	SmallestTimeGrain  string           `yaml:"smallest_time_grain"`
	DefaultTimeRange   string           `yaml:"default_time_range"`
	AvailableTimeZones []string         `yaml:"available_time_zones"`
	FirstDayOfWeek     uint32           `yaml:"first_day_of_week"`
	FirstMonthOfYear   uint32           `yaml:"first_month_of_year"`
	DefaultTheme       string           `yaml:"default_theme"`
	Dimensions         []*struct {
		Name        string
		Label       string
		Column      string
		Expression  string
		Property    string // For backwards compatibility
		Description string
		Ignore      bool `yaml:"ignore"`
		Unnest      bool
		Uri         string
	}
	DefaultDimensions []string `yaml:"default_dimensions"`
	Measures          []*struct {
		Name                string
		Label               string
		Type                string
		Expression          string
		Window              *MetricsViewMeasureWindow
		Per                 MetricsViewFieldSelectorsYAML
		Requires            MetricsViewFieldSelectorsYAML
		Description         string
		FormatPreset        string `yaml:"format_preset"`
		FormatD3            string `yaml:"format_d3"`
		Ignore              bool   `yaml:"ignore"`
		ValidPercentOfTotal bool   `yaml:"valid_percent_of_total"`
	}
	DefaultMeasures []string `yaml:"default_measures"`
	Security        *struct {
		Access    string `yaml:"access"`
		RowFilter string `yaml:"row_filter"`
		Include   []*struct {
			Names     []string
			Condition string `yaml:"if"`
		}
		Exclude []*struct {
			Names     []string
			Condition string `yaml:"if"`
		}
	}
	DefaultComparison struct {
		Mode      string `yaml:"mode"`
		Dimension string `yaml:"dimension"`
	} `yaml:"default_comparison"`
	AvailableTimeRanges []AvailableTimeRange `yaml:"available_time_ranges"`
}

type AvailableTimeRange struct {
	Range             string
	ComparisonOffsets []AvailableComparisonOffset
}

type tmpAvailableTimeRange struct {
	Range             string                      `yaml:"range"`
	ComparisonOffsets []AvailableComparisonOffset `yaml:"comparison_offsets"`
}

func (t *AvailableTimeRange) UnmarshalYAML(v *yaml.Node) error {
	// This adds support for mixed definition
	// EG:
	// available_time_ranges:
	//   - P1W
	//   - range: P4W
	//     comparison_ranges ...
	if v == nil {
		return nil
	}

	switch v.Kind {
	case yaml.ScalarNode:
		t.Range = v.Value

	case yaml.MappingNode:
		// avoid infinite loop by using a separate struct
		tmp := &tmpAvailableTimeRange{}
		err := v.Decode(tmp)
		if err != nil {
			return err
		}
		t.Range = tmp.Range
		t.ComparisonOffsets = tmp.ComparisonOffsets

	default:
		return fmt.Errorf("available_time_range entry should be either a string or an object")
	}

	return nil
}

type AvailableComparisonOffset struct {
	Offset string
	Range  string
}

type tmpAvailableComparisonOffset struct {
	Offset string `yaml:"offset"`
	Range  string `yaml:"range"`
}

func (o *AvailableComparisonOffset) UnmarshalYAML(v *yaml.Node) error {
	// This adds support for mixed definition
	// EG:
	// comparison_offsets:
	//   - rill-PY
	//   - offset: rill-PM
	//     range: P2M
	if v == nil {
		return nil
	}

	switch v.Kind {
	case yaml.ScalarNode:
		o.Offset = v.Value

	case yaml.MappingNode:
		// avoid infinite loop by using a separate struct
		tmp := &tmpAvailableComparisonOffset{}
		err := v.Decode(tmp)
		if err != nil {
			return err
		}
		o.Offset = tmp.Offset
		o.Range = tmp.Range

	default:
		return fmt.Errorf("comparison_offsets entry should be either a string or an object")
	}

	return nil
}

type MetricsViewFieldSelectorYAML struct {
	Name       string
	TimeGrain  runtimev1.TimeGrain // Only for time dimensions
	Descending bool                // Only for sorting
}

func (f *MetricsViewFieldSelectorYAML) UnmarshalYAML(v *yaml.Node) error {
	if v == nil {
		return nil
	}

	switch v.Kind {
	case yaml.ScalarNode:
		f.Name = v.Value
	case yaml.MappingNode:
		// avoid infinite loop by using a separate struct
		tmp := &struct {
			Name      string
			TimeGrain string `yaml:"time_grain"`
		}{}
		err := v.Decode(tmp)
		if err != nil {
			return err
		}

		tg, err := parseTimeGrain(tmp.TimeGrain)
		if err != nil {
			return fmt.Errorf(`invalid "time_grain": %w`, err)
		}

		f.Name = tmp.Name
		f.TimeGrain = tg
	default:
		return fmt.Errorf("field reference should be either a string or an object")
	}

	return nil
}

type MetricsViewFieldSelectorsYAML []MetricsViewFieldSelectorYAML

func (f *MetricsViewFieldSelectorsYAML) UnmarshalYAML(v *yaml.Node) error {
	if v == nil {
		return nil
	}

	switch v.Kind {
	case yaml.ScalarNode:
		*f = []MetricsViewFieldSelectorYAML{{Name: v.Value}}
	case yaml.SequenceNode:
		res := make([]MetricsViewFieldSelectorYAML, len(v.Content))
		for i, n := range v.Content {
			var tmp MetricsViewFieldSelectorYAML
			err := n.Decode(&tmp)
			if err != nil {
				return err
			}
			res[i] = tmp
		}
		*f = res
	default:
		return fmt.Errorf("field references should be a name or a list")
	}

	return nil
}

type MetricsViewMeasureWindow struct {
	Partition bool
	Order     []MetricsViewFieldSelectorYAML
	OrderTime bool // Preset for ordering by only the time dimension
	Frame     string
}

func (f *MetricsViewMeasureWindow) UnmarshalYAML(v *yaml.Node) error {
	if v == nil {
		return nil
	}

	switch v.Kind {
	case yaml.ScalarNode:
		switch strings.ToLower(v.Value) {
		case "time", "true":
			f.Partition = true
			f.OrderTime = true
		case "all":
			f.Partition = false
		default:
			return fmt.Errorf(`invalid window type %q`, v.Value)
		}
	case yaml.MappingNode:
		// Avoid infinite loop by using a separate struct
		tmp := &struct {
			Partition *bool
			Order     *MetricsViewFieldSelectorsYAML
			Frame     string
		}{}
		err := v.Decode(tmp)
		if err != nil {
			return err
		}

		// Let partition default to true
		f.Partition = true
		if tmp.Partition != nil {
			f.Partition = *tmp.Partition
		}

		if tmp.Order != nil {
			f.Order = *tmp.Order
		} else {
			// If order is not specified, default to ordering by time if it's a partitioned window
			f.OrderTime = f.Partition
		}

		f.Frame = tmp.Frame
	default:
		return fmt.Errorf("measure window should be either a string or an object")
	}

	return nil
}

var comparisonModesMap = map[string]runtimev1.MetricsViewSpec_ComparisonMode{
	"":          runtimev1.MetricsViewSpec_COMPARISON_MODE_UNSPECIFIED,
	"none":      runtimev1.MetricsViewSpec_COMPARISON_MODE_NONE,
	"time":      runtimev1.MetricsViewSpec_COMPARISON_MODE_TIME,
	"dimension": runtimev1.MetricsViewSpec_COMPARISON_MODE_DIMENSION,
}

var validComparisonModes = []string{"none", "time", "dimension"}

const (
	nameIsMeasure   uint8 = 1
	nameIsDimension uint8 = 2
)

// parseMetricsView parses a metrics view (dashboard) definition and adds the resulting resource to p.Resources.
func (p *Parser) parseMetricsView(node *Node) error {
	// Parse YAML
	tmp := &MetricsViewYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Backwards compatibility
	if tmp.DisplayName != "" && tmp.Title == "" {
		tmp.Title = tmp.DisplayName
	}

	var table string
	if tmp.Table == "" {
		table = tmp.Model
	} else if tmp.Model == "" {
		table = tmp.Table
	} else {
		return fmt.Errorf(`cannot set both the "model" field and the "table" field`)
	}
	if table == "" {
		return fmt.Errorf(`must set a value for either the "model" field or the "table" field`)
	}

	smallestTimeGrain, err := parseTimeGrain(tmp.SmallestTimeGrain)
	if err != nil {
		return fmt.Errorf(`invalid "smallest_time_grain": %w`, err)
	}

	if tmp.DefaultTimeRange != "" {
		err := validateISO8601(tmp.DefaultTimeRange, false, false)
		if err != nil {
			return fmt.Errorf(`invalid "default_time_range": %w`, err)
		}
	}

	for _, tz := range tmp.AvailableTimeZones {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return err
		}
	}

	names := make(map[string]uint8)
	names[strings.ToLower(tmp.TimeDimension)] = nameIsDimension

	for i, dim := range tmp.Dimensions {
		if dim == nil || dim.Ignore {
			continue
		}

		// Backwards compatibility
		if dim.Property != "" && dim.Column == "" {
			dim.Column = dim.Property
		}

		// Backwards compatibility
		if dim.Name == "" {
			if dim.Column == "" {
				dim.Name = fmt.Sprintf("dimension_%d", i)
			} else {
				dim.Name = dim.Column
			}
		}

		if (dim.Column == "" && dim.Expression == "") || (dim.Column != "" && dim.Expression != "") {
			return fmt.Errorf("exactly one of column or expression should be set for dimension: %q", dim.Name)
		}

		lower := strings.ToLower(dim.Name)
		if _, ok := names[lower]; ok {
			return fmt.Errorf("found duplicate dimension or measure name %q", dim.Name)
		}
		names[lower] = nameIsDimension
	}

	for _, dimension := range tmp.DefaultDimensions {
		if v, ok := names[strings.ToLower(dimension)]; !ok || v != nameIsDimension {
			return fmt.Errorf(`dimension %q referenced in "default_dimensions" not found`, dimension)
		}
	}

	measures := make([]*runtimev1.MetricsViewSpec_MeasureV2, 0, len(tmp.Measures))
	for i, measure := range tmp.Measures {
		if measure == nil || measure.Ignore {
			continue
		}

		// Backwards compatibility
		if measure.Name == "" {
			measure.Name = fmt.Sprintf("measure_%d", i)
		}

		lower := strings.ToLower(measure.Name)
		if _, ok := names[lower]; ok {
			return fmt.Errorf("found duplicate dimension or measure name %q", measure.Name)
		}
		names[lower] = nameIsMeasure

		if measure.FormatPreset != "" && measure.FormatD3 != "" {
			return fmt.Errorf(`cannot set both "format_preset" and "format_d3" for a measure`)
		}

		var perDimensions []*runtimev1.MetricsViewSpec_DimensionSelector
		for _, per := range measure.Per {
			typ, ok := names[strings.ToLower(per.Name)]
			if !ok || typ != nameIsDimension {
				return fmt.Errorf(`per dimension %q not found`, per.Name)
			}
			perDimensions = append(perDimensions, &runtimev1.MetricsViewSpec_DimensionSelector{
				Name:      per.Name,
				TimeGrain: per.TimeGrain,
			})
		}

		var requiredDimensions []*runtimev1.MetricsViewSpec_DimensionSelector
		var referencedMeasures []string
		for _, ref := range measure.Requires {
			typ, ok := names[strings.ToLower(ref.Name)]

			// All dimensions have already been parsed, so we know for sure if it's a dimension
			if ok && typ == nameIsDimension {
				requiredDimensions = append(requiredDimensions, &runtimev1.MetricsViewSpec_DimensionSelector{
					Name:      ref.Name,
					TimeGrain: ref.TimeGrain,
				})
				continue
			}

			// If not a dimension, we assume it's a measure and validate after the loop (when all measures have been seen)
			referencedMeasures = append(referencedMeasures, ref.Name)
		}

		var window *runtimev1.MetricsViewSpec_MeasureWindow
		if measure.Window != nil {
			// Build order list
			var order []*runtimev1.MetricsViewSpec_DimensionSelector
			if measure.Window.OrderTime && tmp.TimeDimension != "" {
				order = append(order, &runtimev1.MetricsViewSpec_DimensionSelector{
					Name: tmp.TimeDimension,
				})
			}
			for _, o := range measure.Window.Order {
				typ, ok := names[strings.ToLower(o.Name)]
				if !ok || typ != nameIsDimension {
					return fmt.Errorf(`order dimension %q not found`, o.Name)
				}

				order = append(order, &runtimev1.MetricsViewSpec_DimensionSelector{
					Name:      o.Name,
					TimeGrain: o.TimeGrain,
					Desc:      o.Descending,
				})
			}

			// Add items in order list to requiredDimensions
			for _, o := range order {
				found := false
				for _, rd := range requiredDimensions {
					if strings.EqualFold(rd.Name, o.Name) {
						found = true
						break
					}
				}
				if !found {
					requiredDimensions = append(requiredDimensions, &runtimev1.MetricsViewSpec_DimensionSelector{
						Name:      o.Name,
						TimeGrain: o.TimeGrain,
					})
				}
			}

			// Build window
			window = &runtimev1.MetricsViewSpec_MeasureWindow{
				Partition:       measure.Window.Partition,
				OrderBy:         order,
				FrameExpression: measure.Window.Frame,
			}
		}

		var typ runtimev1.MetricsViewSpec_MeasureType
		switch strings.ToLower(measure.Type) {
		case "":
			typ = runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE
			if len(referencedMeasures) > 0 || len(perDimensions) > 0 {
				typ = runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED
			}
		case "simple":
			typ = runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE
			if len(referencedMeasures) > 0 || len(perDimensions) > 0 {
				return fmt.Errorf(`measure type "simple" cannot have "per" or "requires" fields`)
			}
		case "derived":
			typ = runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED
		case "time_comparison":
			typ = runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON
		default:
			return fmt.Errorf(`invalid measure type %q (allowed values: simple, derived, time_comparison)`, measure.Type)
		}

		measures = append(measures, &runtimev1.MetricsViewSpec_MeasureV2{
			Name:                measure.Name,
			Expression:          measure.Expression,
			Type:                typ,
			Window:              window,
			PerDimensions:       perDimensions,
			RequiredDimensions:  requiredDimensions,
			ReferencedMeasures:  referencedMeasures,
			Label:               measure.Label,
			Description:         measure.Description,
			FormatPreset:        measure.FormatPreset,
			FormatD3:            measure.FormatD3,
			ValidPercentOfTotal: measure.ValidPercentOfTotal,
		})
	}
	if len(measures) == 0 {
		return fmt.Errorf("must define at least one measure")
	}

	// Validate referenced measures now that all measures have been seen
	for _, m := range measures {
		for _, ref := range m.ReferencedMeasures {
			if typ, ok := names[strings.ToLower(ref)]; !ok || typ != nameIsMeasure {
				return fmt.Errorf(`referenced measure %q not found`, ref)
			}
		}
	}

	for _, measure := range tmp.DefaultMeasures {
		if v, ok := names[strings.ToLower(measure)]; !ok || v != nameIsMeasure {
			return fmt.Errorf(`measure %q referenced in "default_dimensions" not found`, measure)
		}
	}

	tmp.DefaultComparison.Mode = strings.ToLower(tmp.DefaultComparison.Mode)
	if _, ok := comparisonModesMap[tmp.DefaultComparison.Mode]; !ok {
		return fmt.Errorf("invalid mode: %q. allowed values: %s", tmp.DefaultComparison.Mode, strings.Join(validComparisonModes, ","))
	}
	if tmp.DefaultComparison.Dimension != "" {
		if v, ok := names[strings.ToLower(tmp.DefaultComparison.Dimension)]; !ok && v != nameIsDimension {
			return fmt.Errorf("default comparison dimension %q doesn't exist", tmp.DefaultComparison.Dimension)
		}
	}

	if tmp.AvailableTimeRanges != nil {
		for _, r := range tmp.AvailableTimeRanges {
			err := validateISO8601(r.Range, false, false)
			if err != nil {
				return fmt.Errorf("invalid range in available_time_ranges: %w", err)
			}

			for _, o := range r.ComparisonOffsets {
				err := validateISO8601(o.Offset, false, false)
				if err != nil {
					return fmt.Errorf("invalid offset in comparison_offsets: %w", err)
				}

				if o.Range != "" {
					err := validateISO8601(o.Range, false, false)
					if err != nil {
						return fmt.Errorf("invalid range in comparison_offsets: %w", err)
					}
				}
			}
		}
	}

	if tmp.Security != nil {
		templateData := TemplateData{
			Environment: p.Environment,
			User: map[string]interface{}{
				"name":   "dummy",
				"email":  "mock@example.org",
				"domain": "example.org",
				"groups": []interface{}{"all"},
				"admin":  false,
			},
		}

		if tmp.Security.Access != "" {
			access, err := ResolveTemplate(tmp.Security.Access, templateData)
			if err != nil {
				return fmt.Errorf(`invalid 'security': 'access' templating is not valid: %w`, err)
			}
			_, err = EvaluateBoolExpression(access)
			if err != nil {
				return fmt.Errorf(`invalid 'security': 'access' expression error: %w`, err)
			}
		}

		if tmp.Security.RowFilter != "" {
			_, err := ResolveTemplate(tmp.Security.RowFilter, templateData)
			if err != nil {
				return fmt.Errorf(`invalid 'security': 'row_filter' templating is not valid: %w`, err)
			}
		}

		if len(tmp.Security.Include) > 0 && len(tmp.Security.Exclude) > 0 {
			return errors.New("invalid 'security': only one of 'include' and 'exclude' can be specified")
		}
		if tmp.Security.Include != nil {
			for _, include := range tmp.Security.Include {
				if include == nil || len(include.Names) == 0 || include.Condition == "" {
					return fmt.Errorf("invalid 'security': 'include' fields must have a valid 'if' condition and 'names' list")
				}
				seen := make(map[string]bool)
				for _, name := range include.Names {
					lower := strings.ToLower(name)
					if seen[lower] {
						return fmt.Errorf("invalid 'security': 'include' property %q is duplicated", name)
					}
					seen[lower] = true
					if _, ok := names[lower]; !ok {
						return fmt.Errorf("invalid 'security': 'include' property %q does not exists in dimensions or measures list", name)
					}
				}
				cond, err := ResolveTemplate(include.Condition, templateData)
				if err != nil {
					return fmt.Errorf(`invalid 'security': 'if' condition templating for field %q is not valid: %w`, include.Names, err)
				}
				_, err = EvaluateBoolExpression(cond)
				if err != nil {
					return fmt.Errorf(`invalid 'security': 'if' condition for field %q not evaluating to a boolean: %w`, include.Names, err)
				}
			}
		}
		if tmp.Security.Exclude != nil {
			for _, exclude := range tmp.Security.Exclude {
				if exclude == nil || len(exclude.Names) == 0 || exclude.Condition == "" {
					return fmt.Errorf("invalid 'security': 'exclude' fields must have a valid 'if' condition and 'names' list")
				}
				seen := make(map[string]bool)
				for _, name := range exclude.Names {
					lower := strings.ToLower(name)
					if seen[lower] {
						return fmt.Errorf("invalid 'security': 'exclude' property %q is duplicated", name)
					}
					seen[lower] = true
					if _, ok := names[lower]; !ok {
						return fmt.Errorf("invalid 'security': 'exclude' property %q does not exists in dimensions or measures list", name)
					}
				}
				cond, err := ResolveTemplate(exclude.Condition, templateData)
				if err != nil {
					return fmt.Errorf(`invalid 'security': 'if' condition templating for field %q is not valid: %w`, exclude.Names, err)
				}
				_, err = EvaluateBoolExpression(cond)
				if err != nil {
					return fmt.Errorf(`invalid 'security': 'if' condition for field %q not evaluating to a boolean: %w`, exclude.Names, err)
				}
			}
		}
	}

	node.Refs = append(node.Refs, ResourceName{Name: table})
	if tmp.DefaultTheme != "" {
		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindTheme, Name: tmp.DefaultTheme})
	}

	r, err := p.insertResource(ResourceKindMetricsView, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.
	spec := r.MetricsViewSpec

	spec.Connector = node.Connector
	spec.Database = tmp.Database
	spec.DatabaseSchema = tmp.DatabaseSchema
	spec.Table = table
	spec.Title = tmp.Title
	spec.Description = tmp.Description
	spec.TimeDimension = tmp.TimeDimension
	spec.WatermarkExpression = tmp.Watermark
	spec.SmallestTimeGrain = smallestTimeGrain
	spec.DefaultTimeRange = tmp.DefaultTimeRange
	spec.AvailableTimeZones = tmp.AvailableTimeZones
	spec.FirstDayOfWeek = tmp.FirstDayOfWeek
	spec.FirstMonthOfYear = tmp.FirstMonthOfYear
	spec.DefaultTheme = tmp.DefaultTheme

	for _, dim := range tmp.Dimensions {
		if dim == nil || dim.Ignore {
			continue
		}

		spec.Dimensions = append(spec.Dimensions, &runtimev1.MetricsViewSpec_DimensionV2{
			Name:        dim.Name,
			Column:      dim.Column,
			Expression:  dim.Expression,
			Label:       dim.Label,
			Description: dim.Description,
			Unnest:      dim.Unnest,
			Uri:         dim.Uri,
		})
	}
	spec.DefaultDimensions = tmp.DefaultDimensions

	spec.Measures = measures
	spec.DefaultMeasures = tmp.DefaultMeasures

	spec.DefaultComparisonMode = comparisonModesMap[tmp.DefaultComparison.Mode]
	if tmp.DefaultComparison.Dimension != "" {
		spec.DefaultComparisonDimension = tmp.DefaultComparison.Dimension
	}

	if tmp.AvailableTimeRanges != nil {
		for _, r := range tmp.AvailableTimeRanges {
			t := &runtimev1.MetricsViewSpec_AvailableTimeRange{
				Range: r.Range,
			}
			if r.ComparisonOffsets != nil {
				t.ComparisonOffsets = make([]*runtimev1.MetricsViewSpec_AvailableComparisonOffset, len(r.ComparisonOffsets))
				for i, o := range r.ComparisonOffsets {
					t.ComparisonOffsets[i] = &runtimev1.MetricsViewSpec_AvailableComparisonOffset{
						Offset: o.Offset,
						Range:  o.Range,
					}
				}
			}
			spec.AvailableTimeRanges = append(spec.AvailableTimeRanges, t)
		}
	}

	if tmp.Security != nil {
		if spec.Security == nil {
			spec.Security = &runtimev1.MetricsViewSpec_SecurityV2{}
		}
		spec.Security.Access = tmp.Security.Access
		spec.Security.RowFilter = tmp.Security.RowFilter
		// validation has been done above, only one of these will be set
		if tmp.Security.Include != nil {
			for _, include := range tmp.Security.Include {
				spec.Security.Include = append(spec.Security.Include, &runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
					Condition: include.Condition,
					Names:     include.Names,
				})
			}
		}
		if tmp.Security.Exclude != nil {
			for _, exclude := range tmp.Security.Exclude {
				spec.Security.Exclude = append(spec.Security.Exclude, &runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
					Condition: exclude.Condition,
					Names:     exclude.Names,
				})
			}
		}
	}

	return nil
}

// parseTimeGrain parses a YAML time grain string
func parseTimeGrain(s string) (runtimev1.TimeGrain, error) {
	switch strings.ToLower(s) {
	case "":
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, nil
	case "ms", "millisecond":
		return runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND, nil
	case "s", "second":
		return runtimev1.TimeGrain_TIME_GRAIN_SECOND, nil
	case "min", "minute":
		return runtimev1.TimeGrain_TIME_GRAIN_MINUTE, nil
	case "h", "hour":
		return runtimev1.TimeGrain_TIME_GRAIN_HOUR, nil
	case "d", "day":
		return runtimev1.TimeGrain_TIME_GRAIN_DAY, nil
	case "w", "week":
		return runtimev1.TimeGrain_TIME_GRAIN_WEEK, nil
	case "month":
		return runtimev1.TimeGrain_TIME_GRAIN_MONTH, nil
	case "q", "quarter":
		return runtimev1.TimeGrain_TIME_GRAIN_QUARTER, nil
	case "y", "year":
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR, nil
	default:
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, fmt.Errorf("invalid time grain %q", s)
	}
}

// validateISO8601 is a wrapper around duration.ParseISO8601 with additional validation:
// a) that the duration does not have seconds granularity,
// b) if onlyStandard is true, that the duration does not use any of the Rill-specific extensions (such as year-to-date).
// c) if onlySingular is true, that the duration does not consist of more than one component (e.g. P2Y is valid, P2Y3M is not).
func validateISO8601(isoDuration string, onlyStandard, onlyOneComponent bool) error {
	d, err := duration.ParseISO8601(isoDuration)
	if err != nil {
		return err
	}

	sd, ok := d.(duration.StandardDuration)
	if !ok {
		if onlyStandard {
			return fmt.Errorf("only standard durations are allowed")
		}
		return nil
	}

	if sd.Second != 0 {
		return fmt.Errorf("durations with seconds are not allowed")
	}

	if onlyOneComponent {
		n := 0
		if sd.Year != 0 {
			n++
		}
		if sd.Month != 0 {
			n++
		}
		if sd.Week != 0 {
			n++
		}
		if sd.Day != 0 {
			n++
		}
		if sd.Hour != 0 {
			n++
		}
		if sd.Minute != 0 {
			n++
		}
		if sd.Second != 0 {
			n++
		}
		if n > 1 {
			return fmt.Errorf("only one component is allowed")
		}
	}

	return nil
}
