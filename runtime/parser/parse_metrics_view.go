package parser

import (
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"

	// Load IANA time zone data
	_ "time/tzdata"
)

// MetricsViewYAML is the raw structure of a MetricsView resource defined in YAML
type MetricsViewYAML struct {
	commonYAML        `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Parent            string           `yaml:"parent"` // Parent metrics view, if any
	DisplayName       string           `yaml:"display_name"`
	Title             string           `yaml:"title"` // Deprecated: use display_name
	Description       string           `yaml:"description"`
	AIInstructions    string           `yaml:"ai_instructions"`
	Model             string           `yaml:"model"`
	Database          string           `yaml:"database"`
	DatabaseSchema    string           `yaml:"database_schema"`
	Table             string           `yaml:"table"`
	TimeDimension     string           `yaml:"timeseries"`
	Watermark         string           `yaml:"watermark"`
	SmallestTimeGrain string           `yaml:"smallest_time_grain"`
	FirstDayOfWeek    uint32           `yaml:"first_day_of_week"`
	FirstMonthOfYear  uint32           `yaml:"first_month_of_year"`
	Dimensions        []*struct {
		Name                    string
		DisplayName             string `yaml:"display_name"`
		Label                   string // Deprecated: use display_name
		Description             string
		Column                  string
		Expression              string
		Property                string // For backwards compatibility
		Ignore                  bool   `yaml:"ignore"` // Deprecated
		Unnest                  bool
		URI                     string
		LookupTable             string `yaml:"lookup_table"`
		LookupKeyColumn         string `yaml:"lookup_key_column"`
		LookupValueColumn       string `yaml:"lookup_value_column"`
		LookupDefaultExpression string `yaml:"lookup_default_expression"`
		SmallestTimeGrain       string `yaml:"smallest_time_grain"`
	}
	Measures []*struct {
		Name                string
		DisplayName         string `yaml:"display_name"`
		Label               string // Deprecated: use display_name
		Description         string
		Type                string
		Expression          string
		Window              *MetricsViewMeasureWindow
		Per                 MetricsViewFieldSelectorsYAML
		Requires            MetricsViewFieldSelectorsYAML
		FormatPreset        string         `yaml:"format_preset"`
		FormatD3            string         `yaml:"format_d3"`
		FormatD3Locale      map[string]any `yaml:"format_d3_locale"`
		Ignore              bool           `yaml:"ignore"` // Deprecated
		ValidPercentOfTotal bool           `yaml:"valid_percent_of_total"`
		TreatNullsAs        string         `yaml:"treat_nulls_as"`
	}
	ParentDimensions *FieldSelectorYAML `yaml:"parent_dimensions"` // used when Parent is set
	ParentMeasures   *FieldSelectorYAML `yaml:"parent_measures"`   // used when Parent is set
	Annotations      []*struct {
		Name           string             `yaml:"name"`
		Model          string             `yaml:"model"`
		Database       string             `yaml:"database"`
		DatabaseSchema string             `yaml:"database_schema"`
		Table          string             `yaml:"table"`
		Connector      string             `yaml:"connector"`
		Measures       *FieldSelectorYAML `yaml:"measures"`
	} `yaml:"annotations"`
	Security *SecurityPolicyYAML
	Cache    struct {
		Enabled *bool  `yaml:"enabled"`
		KeySQL  string `yaml:"key_sql"`
		KeyTTL  string `yaml:"key_ttl"`
	} `yaml:"cache"`
	Explore *struct {
		Skip                 bool                   `yaml:"skip"`
		Name                 string                 `yaml:"name"` // Name of the explore, defaults to the metrics view name
		DisplayName          string                 `yaml:"display_name"`
		Description          string                 `yaml:"description"`
		Banner               string                 `yaml:"banner"`
		Theme                yaml.Node              `yaml:"theme"` // Name (string) or inline theme definition (map)
		TimeRanges           []ExploreTimeRangeYAML `yaml:"time_ranges"`
		TimeZones            []string               `yaml:"time_zones"` // Single time zone or list of time zones
		LockTimeZone         bool                   `yaml:"lock_time_zone"`
		AllowCustomTimeRange *bool                  `yaml:"allow_custom_time_range"`
		Defaults             *struct {
			Dimensions          *FieldSelectorYAML `yaml:"dimensions"`
			Measures            *FieldSelectorYAML `yaml:"measures"`
			TimeRange           string             `yaml:"time_range"`
			ComparisonMode      string             `yaml:"comparison_mode"`
			ComparisonDimension string             `yaml:"comparison_dimension"`
		} `yaml:"defaults"`
		Embeds struct {
			HidePivot bool `yaml:"hide_pivot"`
		} `yaml:"embeds"`
	} `yaml:"explore"`

	// DEPRECATED FIELDS
	DefaultTimeRange   string   `yaml:"default_time_range"`
	AvailableTimeZones []string `yaml:"available_time_zones"`
	DefaultTheme       string   `yaml:"default_theme"`
	DefaultDimensions  []string `yaml:"default_dimensions"`
	DefaultMeasures    []string `yaml:"default_measures"`
	DefaultComparison  struct {
		Mode      string `yaml:"mode"`
		Dimension string `yaml:"dimension"`
	} `yaml:"default_comparison"`
	AvailableTimeRanges []ExploreTimeRangeYAML `yaml:"available_time_ranges"`
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

var comparisonModesMap = map[string]runtimev1.ExploreComparisonMode{
	"":          runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_UNSPECIFIED,
	"none":      runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE,
	"time":      runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_TIME,
	"dimension": runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_DIMENSION,
}

var validComparisonModes = []string{"none", "time", "dimension"}

const (
	nameIsMeasure   uint8 = 1
	nameIsDimension uint8 = 2
)

// parseMetricsView parses a metrics view definition and adds the resulting resource to p.Resources.
func (p *Parser) parseMetricsView(node *Node) error {
	// Parse YAML
	tmp := &MetricsViewYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Backwards compatibility
	if tmp.Title != "" && tmp.DisplayName == "" {
		tmp.DisplayName = tmp.Title
	}

	if tmp.Table != "" && tmp.Model != "" {
		return fmt.Errorf(`cannot set both the "model" field and the "table" field`)
	}
	if tmp.Table == "" && tmp.Model == "" && tmp.Parent == "" {
		return fmt.Errorf(`must set a value for either the "model", "table" or "parent" field`)
	}

	smallestTimeGrain, err := parseTimeGrain(tmp.SmallestTimeGrain)
	if err != nil {
		return fmt.Errorf(`invalid "smallest_time_grain": %w`, err)
	}
	if smallestTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED && smallestTimeGrain < runtimev1.TimeGrain_TIME_GRAIN_SECOND {
		return errors.New(`"smallest_time_grain" must be at least "second"`)
	}

	if tmp.DefaultTimeRange != "" {
		_, err := rilltime.Parse(tmp.DefaultTimeRange, rilltime.ParseOptions{})
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

	if tmp.Parent != "" {
		if len(tmp.Dimensions) > 0 || len(tmp.Measures) > 0 {
			return fmt.Errorf("cannot define dimensions or measures in a derived metrics view, use parent_dimensions and parent_measures to select from parent %q", tmp.Parent)
		}
		if tmp.Database != "" || tmp.DatabaseSchema != "" || tmp.Table != "" || tmp.Model != "" {
			return fmt.Errorf("cannot set data source in a derived metrics view (parent %q)", tmp.Parent)
		}
		if tmp.Cache.Enabled != nil || tmp.Cache.KeySQL != "" || tmp.Cache.KeyTTL != "" {
			return fmt.Errorf("cannot set cache in a derived metrics view (parent %q)", tmp.Parent)
		}
		// disallow deprecated fields in derived metrics views
		if tmp.DefaultTimeRange != "" || tmp.DefaultTheme != "" || len(tmp.DefaultDimensions) > 0 || len(tmp.DefaultMeasures) > 0 || tmp.DefaultComparison.Mode != "" || tmp.DefaultComparison.Dimension != "" {
			return fmt.Errorf("cannot set defaults in derived metrics view (parent %q), defaults can be set under explore key", tmp.Parent)
		}
		if len(tmp.AvailableTimeRanges) > 0 || len(tmp.AvailableTimeZones) > 0 {
			return fmt.Errorf("cannot set available time ranges or time zones in derived metrics view (parent %q), use explore key", tmp.Parent)
		}

		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindMetricsView, Name: tmp.Parent})
	} else if tmp.ParentDimensions != nil || tmp.ParentMeasures != nil {
		return fmt.Errorf("parent_dimensions and parent_measures can only be set in derived metrics views, use dimensions and measures instead")
	}

	names := make(map[string]uint8)
	names[strings.ToLower(tmp.TimeDimension)] = nameIsDimension
	timeDimSeenInDimList := false

	dimensions := make([]*runtimev1.MetricsViewSpec_Dimension, 0, len(tmp.Dimensions))
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

		// Backwards compatibility
		if dim.Label != "" && dim.DisplayName == "" {
			dim.DisplayName = dim.Label
		}

		// When display name is not provided, we derive a human-friendly one from the dimension name
		if dim.DisplayName == "" {
			dim.DisplayName = ToDisplayName(dim.Name)
		}

		// The "column" and "expression" properties are mutually exclusive
		if (dim.Column == "" && dim.Expression == "") || (dim.Column != "" && dim.Expression != "") {
			return fmt.Errorf("exactly one of column or expression should be set for dimension: %q", dim.Name)
		}

		// Validate the lookup table fields
		if dim.LookupTable != "" || dim.LookupKeyColumn != "" || dim.LookupValueColumn != "" {
			if dim.LookupTable == "" || dim.LookupKeyColumn == "" || dim.LookupValueColumn == "" {
				return fmt.Errorf("all lookup fields should be defined (lookup_table, lookup_key_column and lookup_value_column should be defined")
			}
			if strings.Contains(dim.Expression, "dictGet") {
				return fmt.Errorf("dictGet expression and lookup fields cannot be used together")
			}
		}

		// Validate the dimension name is unique
		lower := strings.ToLower(dim.Name)
		if _, ok := names[lower]; ok {
			// allow time dimension to be defined in the dimensions list once
			if strings.EqualFold(lower, tmp.TimeDimension) {
				if timeDimSeenInDimList {
					return fmt.Errorf("time dimension %q defined multiple times", tmp.TimeDimension)
				} else if dim.Name != tmp.TimeDimension {
					return fmt.Errorf("dimension name %q does not match the case of time dimension %q", dim.Name, tmp.TimeDimension)
				}
				timeDimSeenInDimList = true
			} else {
				return fmt.Errorf("found duplicate dimension or measure name %q", dim.Name)
			}
		}
		names[lower] = nameIsDimension

		smallestTimeGrain, err := parseTimeGrain(dim.SmallestTimeGrain)
		if err != nil {
			return fmt.Errorf(`invalid "smallest_time_grain" for dimension %q: %w`, dim.Name, err)
		}
		if smallestTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED && smallestTimeGrain < runtimev1.TimeGrain_TIME_GRAIN_SECOND {
			return fmt.Errorf(`invalid "smallest_time_grain" for dimension %q: must be at least "second"`, dim.Name)
		}

		// Dimension is valid, add to the list
		dimensions = append(dimensions, &runtimev1.MetricsViewSpec_Dimension{
			Name:                    dim.Name,
			DisplayName:             dim.DisplayName,
			Description:             dim.Description,
			Column:                  dim.Column,
			Expression:              dim.Expression,
			Unnest:                  dim.Unnest,
			Uri:                     dim.URI,
			LookupTable:             dim.LookupTable,
			LookupKeyColumn:         dim.LookupKeyColumn,
			LookupValueColumn:       dim.LookupValueColumn,
			LookupDefaultExpression: dim.LookupDefaultExpression,
			SmallestTimeGrain:       smallestTimeGrain,
		})
	}

	for _, dimension := range tmp.DefaultDimensions {
		if v, ok := names[strings.ToLower(dimension)]; !ok || v != nameIsDimension {
			return fmt.Errorf(`dimension %q referenced in "default_dimensions" not found`, dimension)
		}
	}

	measures := make([]*runtimev1.MetricsViewSpec_Measure, 0, len(tmp.Measures))
	for i, measure := range tmp.Measures {
		if measure == nil || measure.Ignore {
			continue
		}

		// Backwards compatibility
		if measure.Name == "" {
			measure.Name = fmt.Sprintf("measure_%d", i)
		}

		// Backwards compatibility
		if measure.Label != "" && measure.DisplayName == "" {
			measure.DisplayName = measure.Label
		}

		if measure.DisplayName == "" {
			measure.DisplayName = ToDisplayName(measure.Name)
		}

		lower := strings.ToLower(measure.Name)
		if _, ok := names[lower]; ok {
			return fmt.Errorf("found duplicate dimension or measure name %q", measure.Name)
		}
		names[lower] = nameIsMeasure

		if measure.FormatPreset != "" && measure.FormatD3 != "" {
			return fmt.Errorf(`cannot set both "format_preset" and "format_d3" for a measure`)
		}

		var formatD3Locale *structpb.Struct
		if measure.FormatD3Locale != nil {
			if measure.FormatD3 == "" {
				return fmt.Errorf(`"format_d3_locale" can only be set if "format_d3" is set`)
			}

			formatD3Locale, err = structpb.NewStruct(measure.FormatD3Locale)
			if err != nil {
				return fmt.Errorf(`invalid "format_d3_locale": %w`, err)
			}
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

		measures = append(measures, &runtimev1.MetricsViewSpec_Measure{
			Name:                measure.Name,
			DisplayName:         measure.DisplayName,
			Description:         measure.Description,
			Expression:          measure.Expression,
			Type:                typ,
			Window:              window,
			PerDimensions:       perDimensions,
			RequiredDimensions:  requiredDimensions,
			ReferencedMeasures:  referencedMeasures,
			FormatPreset:        measure.FormatPreset,
			FormatD3:            measure.FormatD3,
			FormatD3Locale:      formatD3Locale,
			ValidPercentOfTotal: measure.ValidPercentOfTotal,
			TreatNullsAs:        measure.TreatNullsAs,
		})
	}
	if len(measures) == 0 && tmp.Parent == "" {
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

	// 0 is default and type is uint32
	if tmp.FirstDayOfWeek > 7 {
		return fmt.Errorf("invalid first day of week %d, must be between 1 and 7", tmp.FirstDayOfWeek)
	}

	// 0 is default and type is uint32
	if tmp.FirstMonthOfYear > 12 {
		return fmt.Errorf("invalid first month of year %d, must be between 1 and 12", tmp.FirstMonthOfYear)
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
			_, err := rilltime.Parse(r.Range, rilltime.ParseOptions{})
			if err != nil {
				return fmt.Errorf("invalid range in available_time_ranges: %w", err)
			}

			for _, o := range r.ComparisonTimeRanges {
				err = rilltime.ParseCompatibility(o.Range, o.Offset)
				if err != nil {
					return err
				}
			}
		}
	}

	// Gather all lookup table names
	var lookupTableNames map[string]bool
	for _, dim := range tmp.Dimensions {
		if dim != nil && dim.LookupTable != "" {
			if lookupTableNames == nil {
				lookupTableNames = make(map[string]bool)
			}
			lookupTableNames[dim.LookupTable] = true
		}
	}

	securityRules, err := tmp.Security.Proto()
	if err != nil {
		return err
	}

	if tmp.Model != "" {
		// Not setting Kind because for backwards compatibility, it may actually be a source or an external table.
		node.Refs = append(node.Refs, ResourceName{Name: tmp.Model})
	}
	if tmp.Table != "" {
		// By convention, if the table name matches a source or model name we add a DAG link.
		// We may want to remove this at some point, but the cases where it would not be desired are very rare.
		// Not setting Kind so that inference kicks in.
		node.Refs = append(node.Refs, ResourceName{Name: tmp.Table})
	}

	// Attempt to link the lookup tables in the DAG in case they are models.
	// If they are not models, the upstream logic for refs will filter them out.
	for lookupTable := range lookupTableNames {
		// Not setting Kind so that inference kicks in.
		node.Refs = append(node.Refs, ResourceName{Name: lookupTable})
	}

	if tmp.DefaultTheme != "" {
		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindTheme, Name: tmp.DefaultTheme})
	}

	// Add annotations as refs to the end of the metrics view.
	for _, annotation := range tmp.Annotations {
		if annotation == nil {
			continue
		}

		if tmp.Table != "" && tmp.Model != "" {
			return fmt.Errorf(`cannot set both the "model" field and the "table" field for annotation`)
		}
		if tmp.Table == "" && tmp.Model == "" {
			return fmt.Errorf(`must set a value for either the "model" field or the "table" field for annotation`)
		}
		if annotation.Name == "" {
			if annotation.Model != "" {
				annotation.Name = annotation.Model
			} else {
				annotation.Name = annotation.Table
			}
		}

		if annotation.Model != "" {
			// Not setting Kind because for backwards compatibility, it may actually be a source or an external table.
			node.Refs = append(node.Refs, ResourceName{Name: annotation.Model})
		} else if annotation.Table != "" {
			// By convention, if the table name matches a source or model name we add a DAG link.
			// We may want to remove this at some point, but the cases where it would not be desired are very rare.
			// Not setting Kind so that inference kicks in.
			node.Refs = append(node.Refs, ResourceName{Name: annotation.Table})
		}
	}

	securityRefs, err := inferRefsFromSecurityRules(securityRules)
	if err != nil {
		return err
	}
	node.Refs = append(node.Refs, securityRefs...)

	var cacheTTLDuration time.Duration
	if tmp.Cache.KeyTTL != "" {
		cacheTTLDuration, err = time.ParseDuration(tmp.Cache.KeyTTL)
		if err != nil {
			return fmt.Errorf(`invalid "cache.key_ttl": %w`, err)
		}
	}

	// validate and insert inline explore, if true and no error is returned from the method then an explore resource is created so no error should be returned after this point
	skipExplore, exploreRes, err := p.parseAndInsertInlineExplore(tmp, node.Name, node.Paths)
	if err != nil {
		return fmt.Errorf("failed to parse inline explore: %w", err)
	}

	// insert metrics view resource immediately after parsing the inline explore as it inserts the explore resource so we should not return an error now
	r, err := p.insertResource(ResourceKindMetricsView, node.Name, node.Paths, node.Refs...)
	if err != nil {
		// If we fail to insert the metrics view, we must delete the inline explore if it was created.
		if exploreRes != nil {
			panic(fmt.Sprintf("failed to insert metrics view %q, but inline explore was created: %s", node.Name, exploreRes.Name))
		}
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.
	spec := r.MetricsViewSpec

	spec.Parent = tmp.Parent
	spec.Connector = node.Connector
	spec.Database = tmp.Database
	spec.DatabaseSchema = tmp.DatabaseSchema
	spec.Table = tmp.Table
	spec.Model = tmp.Model
	spec.DisplayName = tmp.DisplayName
	if spec.DisplayName == "" {
		spec.DisplayName = ToDisplayName(node.Name)
	}
	spec.Description = tmp.Description
	spec.AiInstructions = tmp.AIInstructions
	spec.TimeDimension = tmp.TimeDimension
	spec.WatermarkExpression = tmp.Watermark
	spec.SmallestTimeGrain = smallestTimeGrain
	spec.FirstDayOfWeek = tmp.FirstDayOfWeek
	spec.FirstMonthOfYear = tmp.FirstMonthOfYear

	spec.Dimensions = dimensions
	spec.Measures = measures

	// if time dimension is not defined in the dimensions list but is defined in the `timeseries` key, we prepend it to the dimensions list here
	if !timeDimSeenInDimList && tmp.TimeDimension != "" {
		spec.Dimensions = append([]*runtimev1.MetricsViewSpec_Dimension{
			{
				Name:        tmp.TimeDimension,
				Column:      tmp.TimeDimension,
				DisplayName: ToDisplayName(tmp.TimeDimension),
			},
		}, spec.Dimensions...)
	}

	for _, annotation := range tmp.Annotations {
		if annotation == nil {
			continue
		}
		var annotationMeasuresSelector *runtimev1.FieldSelector
		annotationMeasures, ok := annotation.Measures.TryResolve()
		if !ok {
			annotationMeasuresSelector = annotation.Measures.Proto()
		}

		spec.Annotations = append(spec.Annotations, &runtimev1.MetricsViewSpec_Annotation{
			Name:             annotation.Name,
			Model:            annotation.Model,
			Database:         annotation.Database,
			DatabaseSchema:   annotation.DatabaseSchema,
			Table:            annotation.Table,
			Connector:        annotation.Connector,
			Measures:         annotationMeasures,
			MeasuresSelector: annotationMeasuresSelector,
		})
	}

	// Parse the dimensions and measures selectors
	if tmp.Parent != "" {
		spec.ParentDimensions = tmp.ParentDimensions.Proto()
		spec.ParentMeasures = tmp.ParentMeasures.Proto()
	}

	spec.SecurityRules = securityRules
	spec.CacheEnabled = tmp.Cache.Enabled
	spec.CacheKeySql = tmp.Cache.KeySQL
	spec.CacheKeyTtlSeconds = int64(cacheTTLDuration.Seconds())

	// When version is greater than 0 or inline explore is defined or skip explore set to true, we skip creating a default explore resource. Application should set version to 0 now to enable automatic explore emission.
	if node.Version > 0 || skipExplore {
		return nil
	}

	refs := []ResourceName{{Kind: ResourceKindMetricsView, Name: node.Name}}
	if tmp.DefaultTheme != "" {
		refs = append(refs, ResourceName{Kind: ResourceKindTheme, Name: tmp.DefaultTheme})
	}
	e, err := p.insertResource(ResourceKindExplore, node.Name, node.Paths, refs...)
	if err != nil {
		// We mustn't error because we have already emitted one resource.
		// Since this probably means an explore has been defined separately, we can just ignore this error.
		return nil
	}

	e.ExploreSpec.DisplayName = spec.DisplayName
	e.ExploreSpec.Description = spec.Description
	e.ExploreSpec.MetricsView = node.Name
	for _, dim := range spec.Dimensions {
		e.ExploreSpec.Dimensions = append(e.ExploreSpec.Dimensions, dim.Name)
	}
	for _, m := range spec.Measures {
		e.ExploreSpec.Measures = append(e.ExploreSpec.Measures, m.Name)
	}
	if tmp.Parent != "" {
		e.ExploreSpec.DimensionsSelector = &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}}
		e.ExploreSpec.MeasuresSelector = &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}}
	}
	e.ExploreSpec.Theme = tmp.DefaultTheme
	for _, tr := range tmp.AvailableTimeRanges {
		res := &runtimev1.ExploreTimeRange{Range: tr.Range}
		for _, ctr := range tr.ComparisonTimeRanges {
			res.ComparisonTimeRanges = append(res.ComparisonTimeRanges, &runtimev1.ExploreComparisonTimeRange{
				Offset: ctr.Offset,
				Range:  ctr.Range,
			})
		}
		e.ExploreSpec.TimeRanges = append(e.ExploreSpec.TimeRanges, res)
	}
	e.ExploreSpec.TimeZones = tmp.AvailableTimeZones

	var presetDimensionsSelector, presetMeasuresSelector *runtimev1.FieldSelector
	if len(tmp.DefaultDimensions) == 0 {
		presetDimensionsSelector = &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}}
	}
	if len(tmp.DefaultMeasures) == 0 {
		presetMeasuresSelector = &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}}
	}
	var tr *string
	if tmp.DefaultTimeRange != "" {
		tr = &tmp.DefaultTimeRange
	}
	var compareDim *string
	if tmp.DefaultComparison.Dimension != "" {
		compareDim = &tmp.DefaultComparison.Dimension
	}
	e.ExploreSpec.DefaultPreset = &runtimev1.ExplorePreset{
		Dimensions:          tmp.DefaultDimensions,
		DimensionsSelector:  presetDimensionsSelector,
		Measures:            tmp.DefaultMeasures,
		MeasuresSelector:    presetMeasuresSelector,
		TimeRange:           tr,
		ComparisonMode:      comparisonModesMap[tmp.DefaultComparison.Mode],
		ComparisonDimension: compareDim,
	}
	// Backwards compatibility: explore parser will default to true so also emit true on the emitted explore spec
	e.ExploreSpec.AllowCustomTimeRange = true
	e.ExploreSpec.DefinedInMetricsView = true

	return nil
}

// parseAndInsertInlineExplore parses and validates the inline explore definition in a metrics view YAML. It returns true if automatic explore emission should be skipped, false otherwise.
func (p *Parser) parseAndInsertInlineExplore(tmp *MetricsViewYAML, mvName string, mvPaths []string) (bool, *Resource, error) {
	if tmp.Explore == nil {
		return false, nil, nil
	}
	if tmp.Explore.Skip {
		return true, nil, nil
	}

	if tmp.DefaultTimeRange != "" || len(tmp.AvailableTimeZones) > 0 || tmp.DefaultTheme != "" || len(tmp.DefaultDimensions) > 0 || len(tmp.DefaultMeasures) > 0 || tmp.DefaultComparison.Mode != "" || tmp.DefaultComparison.Dimension != "" || len(tmp.AvailableTimeRanges) > 0 {
		return false, nil, fmt.Errorf("setting defaults or available time zones or ranges under metrics view is deprecated, set them under explore key")
	}

	var timeRanges []*runtimev1.ExploreTimeRange
	for _, tr := range tmp.Explore.TimeRanges {
		if _, err := rilltime.Parse(tr.Range, rilltime.ParseOptions{}); err != nil {
			return false, nil, fmt.Errorf("invalid time range %q: %w", tr.Range, err)
		}
		res := &runtimev1.ExploreTimeRange{Range: tr.Range}
		for _, ctr := range tr.ComparisonTimeRanges {
			err := rilltime.ParseCompatibility(ctr.Range, ctr.Offset)
			if err != nil {
				return false, nil, err
			}
			res.ComparisonTimeRanges = append(res.ComparisonTimeRanges, &runtimev1.ExploreComparisonTimeRange{
				Offset: ctr.Offset,
				Range:  ctr.Range,
			})
		}
		timeRanges = append(timeRanges, res)
	}

	// Validate time zones
	for _, tz := range tmp.Explore.TimeZones {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return false, nil, err
		}
	}

	// Build and validate presets
	var defaultPreset *runtimev1.ExplorePreset
	if tmp.Explore.Defaults != nil {
		if tmp.Explore.Defaults.TimeRange != "" {
			if _, err := rilltime.Parse(tmp.Explore.Defaults.TimeRange, rilltime.ParseOptions{}); err != nil {
				return false, nil, fmt.Errorf("invalid time range %q: %w", tmp.Explore.Defaults.TimeRange, err)
			}
		}

		mode := runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE
		if tmp.Explore.Defaults.ComparisonMode != "" {
			var ok bool
			mode, ok = exploreComparisonModes[tmp.Explore.Defaults.ComparisonMode]
			if !ok {
				return false, nil, fmt.Errorf("invalid comparison mode %q (options: %s)", tmp.Explore.Defaults.ComparisonMode, strings.Join(maps.Keys(exploreComparisonModes), ", "))
			}
		}

		if tmp.Explore.Defaults.ComparisonDimension != "" && mode != runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_DIMENSION {
			return false, nil, errors.New("can only set comparison_dimension when comparison_mode is 'dimension'")
		}

		var presetDimensionsSelector *runtimev1.FieldSelector
		presetDimensions, ok := tmp.Explore.Defaults.Dimensions.TryResolve()
		if !ok {
			presetDimensionsSelector = tmp.Explore.Defaults.Dimensions.Proto()
		}

		var presetMeasuresSelector *runtimev1.FieldSelector
		presetMeasures, ok := tmp.Explore.Defaults.Measures.TryResolve()
		if !ok {
			presetMeasuresSelector = tmp.Explore.Defaults.Measures.Proto()
		}

		var tr *string
		if tmp.Explore.Defaults.TimeRange != "" {
			tr = &tmp.Explore.Defaults.TimeRange
		}
		var compareDim *string
		if tmp.Explore.Defaults.ComparisonDimension != "" {
			compareDim = &tmp.Explore.Defaults.ComparisonDimension
		}
		defaultPreset = &runtimev1.ExplorePreset{
			Dimensions:          presetDimensions,
			DimensionsSelector:  presetDimensionsSelector,
			Measures:            presetMeasures,
			MeasuresSelector:    presetMeasuresSelector,
			TimeRange:           tr,
			ComparisonMode:      mode,
			ComparisonDimension: compareDim,
		}
	}

	// Set default for AllowCustomTimeRange to true if not provided
	allowCustomTimeRange := true
	if tmp.Explore.AllowCustomTimeRange != nil {
		allowCustomTimeRange = *tmp.Explore.AllowCustomTimeRange
	}

	refs := []ResourceName{{Kind: ResourceKindMetricsView, Name: mvName}}
	// Parse theme if present.
	// If it returns a themeSpec, it will be inserted as a separate resource later in this function.
	themeName, themeSpec, err := p.parseThemeRef(&tmp.Explore.Theme)
	if err != nil {
		return false, nil, err
	}
	if themeName != "" && themeSpec == nil {
		refs = append(refs, ResourceName{Kind: ResourceKindTheme, Name: themeName})
	}

	// before inserting inline explore, dry run inserting the parent metrics view resource to ensure that the explore can be inserted
	err = p.insertDryRun(ResourceKindMetricsView, mvName)
	if err != nil {
		return false, nil, fmt.Errorf("failed to dry run inserting metrics view %q: %w", mvName, err)
	}

	name := mvName
	if tmp.Explore.Name != "" {
		name = tmp.Explore.Name
	}
	// Track explore
	r, err := p.insertResource(ResourceKindExplore, name, mvPaths, refs...)
	if err != nil {
		return false, nil, err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.
	r.ExploreSpec.DisplayName = tmp.Explore.DisplayName
	if r.ExploreSpec.DisplayName == "" {
		r.ExploreSpec.DisplayName = ToDisplayName(name)
	}
	r.ExploreSpec.Description = tmp.Explore.Description
	r.ExploreSpec.MetricsView = mvName
	r.ExploreSpec.Banner = tmp.Explore.Banner
	r.ExploreSpec.DimensionsSelector = &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}}
	r.ExploreSpec.MeasuresSelector = &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}}
	r.ExploreSpec.Theme = themeName
	r.ExploreSpec.EmbeddedTheme = themeSpec
	r.ExploreSpec.TimeRanges = timeRanges
	r.ExploreSpec.TimeZones = tmp.Explore.TimeZones
	r.ExploreSpec.DefaultPreset = defaultPreset
	r.ExploreSpec.EmbedsHidePivot = tmp.Explore.Embeds.HidePivot
	r.ExploreSpec.LockTimeZone = tmp.Explore.LockTimeZone
	r.ExploreSpec.AllowCustomTimeRange = allowCustomTimeRange
	r.ExploreSpec.DefinedInMetricsView = true

	return true, r, nil
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

var validationTemplateData = TemplateData{
	Environment: "dev",
	User: map[string]interface{}{
		"name":   "dummy",
		"email":  "mock@example.org",
		"domain": "example.org",
		"groups": []interface{}{"all"},
		"admin":  false,
	},
	Resolve: func(ref ResourceName) (string, error) {
		return ref.Name, nil
	},
}

// parseNamesYAML parses a []string or a '*' denoting "all names" from a YAML node.
func parseNamesYAML(n yaml.Node) (names []string, all bool, err error) {
	switch n.Kind {
	case yaml.ScalarNode:
		if n.Value == "*" {
			all = true
			return
		}
		err = fmt.Errorf("unexpected scalar %q", n.Value)
	case yaml.SequenceNode:
		names = make([]string, len(n.Content))
		for i, c := range n.Content {
			if c.Kind != yaml.ScalarNode {
				err = fmt.Errorf("unexpected non-string list entry on line %d", c.Line)
				return
			}
			names[i] = c.Value
		}
	default:
		err = fmt.Errorf("invalid field names %v", n)
	}
	return
}

// inferRefsFromSecurityRules infers resource references from security rules.
func inferRefsFromSecurityRules(rules []*runtimev1.SecurityRule) ([]ResourceName, error) {
	var refs []ResourceName
	for _, r := range rules {
		// RowFilter rules are the only rules that can reference external data (since they execute inside the OLAP instead of in the in-memory expression engine).
		if r == nil {
			continue
		}
		rowFilter := r.GetRowFilter()
		if rowFilter == nil {
			continue
		}

		meta, err := AnalyzeTemplate(rowFilter.Sql)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'sql' in row_filter security rule: %w`, err)
		}

		refs = append(refs, meta.Refs...)
	}
	// No need to deduplicate because that's done upstream when the resource is inserted.
	return refs, nil
}
