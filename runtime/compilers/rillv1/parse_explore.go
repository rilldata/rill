package rillv1

import (
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type ExploreYAML struct {
	commonYAML  `yaml:",inline"`       // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title       string                 `yaml:"title"`
	Description string                 `yaml:"description"`
	MetricsView string                 `yaml:"metrics_view"`
	Dimensions  *NamesYAML             `yaml:"dimensions"`
	Measures    *NamesYAML             `yaml:"measures"`
	Theme       string                 `yaml:"theme"`
	TimeRanges  []ExploreTimeRangeYAML `yaml:"time_ranges"`
	TimeZones   []string               `yaml:"time_zones"`
	Presets     []*struct {
		Label               string     `yaml:"label"`
		Dimensions          *NamesYAML `yaml:"dimensions"`
		Measures            *NamesYAML `yaml:"measures"`
		TimeRange           string     `yaml:"time_range"`
		ComparisonMode      string     `yaml:"comparison_mode"`
		ComparisonDimension string     `yaml:"comparison_dimension"`
	} `yaml:"presets"`
}

// NamesYAML parses a list of names with support for a '*' scalar for all names,
// and support for a nested "exclude:" list for selecting all except the listed names.
//
// Note that '*' is represented by setting Exclude to true and leaving Names nil.
// (Because excluding nothing is the same as including everything.)
type NamesYAML struct {
	Names   []string
	Exclude bool
}

func (y *NamesYAML) UnmarshalYAML(v *yaml.Node) error {
	if v == nil {
		return nil
	}
	switch v.Kind {
	case yaml.ScalarNode:
		if v.Value == "*" {
			y.Exclude = true
			return nil
		}
		return fmt.Errorf("unexpected scalar %q", v.Value)
	case yaml.SequenceNode:
		y.Names = make([]string, len(v.Content))
		for i, c := range v.Content {
			if c.Kind != yaml.ScalarNode {
				return fmt.Errorf("unexpected non-string list entry on line %d", c.Line)
			}
			y.Names[i] = c.Value
		}
	case yaml.MappingNode:
		tmp := &struct {
			Exclude yaml.Node `yaml:"exclude"`
		}{}
		err := v.Decode(tmp)
		if err != nil {
			return err
		}
		if tmp.Exclude.IsZero() {
			return errors.New("expected '*', list of names, or `exclude` field")
		}

		// Exclude should also be '*' or a list of names.
		// For simpliciy, we can just recurse on it and invert the result.
		err = y.UnmarshalYAML(&tmp.Exclude)
		if err != nil {
			return fmt.Errorf("error parsing `exclude` field: %w", err)
		}
		y.Exclude = !y.Exclude // Oh the irony
	default:
		return fmt.Errorf("expected '*', list of names, or `exclude` field, got type %q", v.Kind)
	}
	return nil
}

func (y *NamesYAML) Safe() NamesYAML {
	if y == nil {
		// If not specified, default to '*' (include all).
		return NamesYAML{Names: nil, Exclude: true}
	}
	return *y
}

// ExploreTimeRangeYAML represents a time range in an ExploreYAML.
// It has a custom parser to support a mixed scalar and mapping structure.
// Example:
//
//	 time_ranges:
//		- P7D
//		- range: P30D
//		  comparison_offsets:
//		    - P30D
//		    - offset: P60D
//		      range: P90D
//
// The custom parsing is handled in UnmarshalYAML on this struct on an ExploreComparisonTimeRangeYAML.
type ExploreTimeRangeYAML struct {
	Range                string
	ComparisonTimeRanges []ExploreComparisonTimeRangeYAML
}

func (y *ExploreTimeRangeYAML) UnmarshalYAML(v *yaml.Node) error {
	if v == nil {
		return nil
	}
	switch v.Kind {
	case yaml.ScalarNode:
		y.Range = v.Value
	case yaml.MappingNode:
		tmp := &struct {
			Range             string                           `yaml:"range"`
			ComparisonOffsets []ExploreComparisonTimeRangeYAML `yaml:"comparison_offsets"`
		}{}
		err := v.Decode(tmp)
		if err != nil {
			return err
		}
		y.Range = tmp.Range
		y.ComparisonTimeRanges = tmp.ComparisonOffsets
	default:
		return fmt.Errorf("invalid time_range: should be a string or mapping, got kind %q", v.Kind)
	}
	return nil
}

// ExploreComparisonTimeRangeYAML is part of ExploreTimeRangeYAML. See its docstring.
type ExploreComparisonTimeRangeYAML struct {
	Offset string
	Range  string
}

func (y *ExploreComparisonTimeRangeYAML) UnmarshalYAML(v *yaml.Node) error {
	if v == nil {
		return nil
	}
	switch v.Kind {
	case yaml.ScalarNode:
		y.Offset = v.Value
	case yaml.MappingNode:
		tmp := &struct {
			Offset string `yaml:"offset"`
			Range  string `yaml:"range"`
		}{}
		err := v.Decode(tmp)
		if err != nil {
			return err
		}
		y.Offset = tmp.Offset
		y.Range = tmp.Range
	default:
		return fmt.Errorf("invalid comparison_offsets entry: should be a string or mapping, got kind %q", v.Kind)
	}
	return nil
}

var exploreComparisonModes = map[string]runtimev1.ExploreComparisonMode{
	"none":      runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE,
	"time":      runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_TIME,
	"dimension": runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_DIMENSION,
}

func (p *Parser) parseExplore(node *Node) error {
	// Parse YAML
	tmp := &ExploreYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("explores cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("explores cannot have a connector")
	}

	// Validate metrics_view
	if tmp.MetricsView == "" {
		return errors.New("metrics_view is required")
	}
	node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindMetricsView, Name: tmp.MetricsView})

	// Add theme to refs
	if tmp.Theme != "" {
		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindTheme, Name: tmp.Theme})
	}

	// Build and validate time ranges
	var timeRanges []*runtimev1.ExploreTimeRange
	for _, tr := range tmp.TimeRanges {
		if err := validateISO8601(tr.Range, false, false); err != nil {
			return fmt.Errorf("invalid time range %q: %w", tr.Range, err)
		}
		res := &runtimev1.ExploreTimeRange{Range: tr.Range}
		for _, ctr := range tr.ComparisonTimeRanges {
			if err := validateISO8601(ctr.Offset, false, false); err != nil {
				return fmt.Errorf("invalid comparison offset %q: %w", ctr.Offset, err)
			}
			if ctr.Range != "" {
				if err := validateISO8601(ctr.Range, false, false); err != nil {
					return fmt.Errorf("invalid comparison range %q: %w", ctr.Range, err)
				}
			}
			res.ComparisonTimeRanges = append(res.ComparisonTimeRanges, &runtimev1.ExploreComparisonTimeRange{
				Offset: ctr.Offset,
				Range:  ctr.Range,
			})
		}
		timeRanges = append(timeRanges, res)
	}

	// Validate time zones
	for _, tz := range tmp.TimeZones {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return err
		}
	}

	// Build and validate presets
	var presets []*runtimev1.ExplorePreset
	for _, p := range tmp.Presets {
		if p == nil {
			continue
		}

		if p.TimeRange != "" {
			if err := validateISO8601(p.TimeRange, false, false); err != nil {
				return fmt.Errorf("invalid time range %q: %w", p.TimeRange, err)
			}
		}

		mode := runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE
		if p.ComparisonMode != "" {
			var ok bool
			mode, ok = exploreComparisonModes[p.ComparisonMode]
			if !ok {
				return fmt.Errorf("invalid comparison mode %q (options: %s)", p.ComparisonMode, strings.Join(maps.Keys(exploreComparisonModes), ", "))
			}
		}

		if p.ComparisonDimension != "" && mode != runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_DIMENSION {
			return errors.New("can only set comparison_dimension when comparison_mode is 'dimension'")
		}

		presets = append(presets, &runtimev1.ExplorePreset{
			Label:               p.Label,
			Dimensions:          p.Dimensions.Safe().Names,
			DimensionsExclude:   p.Dimensions.Safe().Exclude,
			Measures:            p.Measures.Safe().Names,
			MeasuresExclude:     p.Measures.Safe().Exclude,
			TimeRange:           p.TimeRange,
			ComparisonMode:      mode,
			ComparisonDimension: p.ComparisonDimension,
		})
	}

	// Track explore
	r, err := p.insertResource(ResourceKindExplore, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.ExploreSpec.Title = tmp.Title
	r.ExploreSpec.Description = tmp.Description
	r.ExploreSpec.MetricsView = tmp.MetricsView
	r.ExploreSpec.Dimensions = tmp.Dimensions.Safe().Names
	r.ExploreSpec.DimensionsExclude = tmp.Dimensions.Safe().Exclude
	r.ExploreSpec.Measures = tmp.Measures.Safe().Names
	r.ExploreSpec.MeasuresExclude = tmp.Measures.Safe().Exclude
	r.ExploreSpec.Theme = tmp.Theme
	r.ExploreSpec.TimeRanges = timeRanges
	r.ExploreSpec.TimeZones = tmp.TimeZones
	r.ExploreSpec.Presets = presets

	return nil
}
