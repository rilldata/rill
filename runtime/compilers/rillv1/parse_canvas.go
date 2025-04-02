package rillv1

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type CanvasYAML struct {
	commonYAML           `yaml:",inline"`       // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	DisplayName          string                 `yaml:"display_name"`
	Title                string                 `yaml:"title"` // Deprecated: use display_name
	Banner               string                 `yaml:"banner"`
	MaxWidth             uint32                 `yaml:"max_width"`
	GapX                 uint32                 `yaml:"gap_x"`
	GapY                 uint32                 `yaml:"gap_y"`
	Theme                yaml.Node              `yaml:"theme"` // Name (string) or inline theme definition (map)
	AllowCustomTimeRange *bool                  `yaml:"allow_custom_time_range"`
	TimeRanges           []ExploreTimeRangeYAML `yaml:"time_ranges"`
	TimeZones            []string               `yaml:"time_zones"`
	Filters              struct {
		Enable *bool `yaml:"enable"`
	}
	Defaults *struct {
		TimeRange           string `yaml:"time_range"`
		ComparisonMode      string `yaml:"comparison_mode"`
		ComparisonDimension string `yaml:"comparison_dimension"`
	} `yaml:"defaults"`
	Variables []*ComponentVariableYAML `yaml:"variables"`
	Rows      []*struct {
		Height *string `yaml:"height"`
		Items  []*struct {
			Width           *string              `yaml:"width"`
			Component       string               `yaml:"component"` // Name of an externally defined component
			InlineComponent map[string]yaml.Node `yaml:",inline"`   // Any other properties are considered an inline component definition
		} `yaml:"items"`
	}
	Security *SecurityPolicyYAML `yaml:"security"`
}

func (p *Parser) parseCanvas(node *Node) error {
	// Parse YAML
	tmp := &CanvasYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("canvases cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("canvases cannot have a connector")
	}

	// Display name backwards compatibility
	if tmp.Title != "" && tmp.DisplayName == "" {
		tmp.DisplayName = tmp.Title
	}

	// Set default for AllowCustomTimeRange to true if not provided
	allowCustomTimeRange := true
	if tmp.AllowCustomTimeRange != nil {
		allowCustomTimeRange = *tmp.AllowCustomTimeRange
	}

	// Parse theme if present.
	// If it returns a themeSpec, it will be inserted as a separate resource later in this function.
	themeName, themeSpec, err := p.parseThemeRef(&tmp.Theme)
	if err != nil {
		return err
	}
	if themeName != "" && themeSpec == nil {
		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindTheme, Name: themeName})
	}

	// Build and validate time ranges
	var timeRanges []*runtimev1.ExploreTimeRange
	for _, tr := range tmp.TimeRanges {
		if _, err := rilltime.Parse(tr.Range, rilltime.ParseOptions{}); err != nil {
			return fmt.Errorf("invalid time range %q: %w", tr.Range, err)
		}
		res := &runtimev1.ExploreTimeRange{Range: tr.Range}
		for _, ctr := range tr.ComparisonTimeRanges {
			err = rilltime.ParseCompatibility(ctr.Range, ctr.Offset)
			if err != nil {
				return err
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

	// Parse variable definitions.
	var variables []*runtimev1.ComponentVariable
	if len(tmp.Variables) > 0 {
		variables = make([]*runtimev1.ComponentVariable, len(tmp.Variables))
	}
	for i, v := range tmp.Variables {
		variables[i], err = v.Proto()
		if err != nil {
			return fmt.Errorf("invalid variable at index %d: %w", i, err)
		}
	}

	// Parse rows and items.
	// Items have position and size, and either reference an externally defined component by name or define a component inline.
	var rows []*runtimev1.CanvasRow
	var inlineComponentDefs []*componentDef // Track inline component definitions so we can insert them after we have validated all components
	for i, row := range tmp.Rows {
		if row == nil {
			return fmt.Errorf("row at index %d is empty", i)
		}

		var height *uint32
		var heightUnit string
		if row.Height != nil {
			v, u, err := parseItemSize(*row.Height)
			if err != nil {
				return fmt.Errorf("invalid height for row %d: %w", i, err)
			}
			if v != 0 && u != "px" {
				return fmt.Errorf("invalid height unit %q for row %d: unit must be 'px'", u, i)
			}
			height = &v
			heightUnit = u
		}

		var items []*runtimev1.CanvasItem
		for j, item := range row.Items {
			if item == nil {
				return fmt.Errorf("item %d in row %d is empty", j, i)
			}

			var width *uint32
			var widthUnit string
			if item.Width != nil {
				v, u, err := parseItemSize(*item.Width)
				if err != nil {
					return fmt.Errorf("invalid width for item %d in row %d: %w", j, i, err)
				}
				if u != "" {
					return fmt.Errorf("invalid width unit %q for item %d in row %d: 'width' cannot have a unit", u, j, i)
				}
				width = &v
				widthUnit = u
			}

			// Validate that exactly one of Component and InlineComponent are set
			if item.Component == "" && len(item.InlineComponent) == 0 {
				return fmt.Errorf("item %d in row %d is missing a component definition", j, i)
			}
			if item.Component != "" && len(item.InlineComponent) > 0 {
				return fmt.Errorf("item %d in row %d has properties incompatible with 'component'", j, i)
			}

			// Parse inline component definition if present and assign into item.Component
			var definedInCanvs bool
			if len(item.InlineComponent) > 0 {
				name, def, err := p.parseCanvasInlineComponent(node.Name, i, j, item.InlineComponent)
				if err != nil {
					return fmt.Errorf("invalid component for item %d in row %d: %w", j, i, err)
				}

				item.Component = name
				inlineComponentDefs = append(inlineComponentDefs, def)
				definedInCanvs = true
			}

			items = append(items, &runtimev1.CanvasItem{
				Component:       item.Component,
				DefinedInCanvas: definedInCanvs,
				Width:           width,
				WidthUnit:       widthUnit,
			})

			node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindComponent, Name: item.Component})
		}

		rows = append(rows, &runtimev1.CanvasRow{
			Height:     height,
			HeightUnit: heightUnit,
			Items:      items,
		})
	}

	// Build and validate presets
	var defaultPreset *runtimev1.CanvasPreset
	if tmp.Defaults != nil {
		if tmp.Defaults.TimeRange != "" {
			if _, err := rilltime.Parse(tmp.Defaults.TimeRange, rilltime.ParseOptions{}); err != nil {
				return fmt.Errorf("invalid time range %q: %w", tmp.Defaults.TimeRange, err)
			}
		}

		mode := runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE
		if tmp.Defaults.ComparisonMode != "" {
			var ok bool
			mode, ok = exploreComparisonModes[tmp.Defaults.ComparisonMode]
			if !ok {
				return fmt.Errorf("invalid comparison mode %q (options: %s)", tmp.Defaults.ComparisonMode, strings.Join(maps.Keys(exploreComparisonModes), ", "))
			}
		}

		if tmp.Defaults.ComparisonDimension != "" && mode != runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_DIMENSION {
			return errors.New("can only set comparison_dimension when comparison_mode is 'dimension'")
		}

		defaultPreset = &runtimev1.CanvasPreset{
			TimeRange:           pointerIfNotEmpty(tmp.Defaults.TimeRange),
			ComparisonMode:      mode,
			ComparisonDimension: pointerIfNotEmpty(tmp.Defaults.ComparisonDimension),
		}
	}

	// Parse security rules
	rules, err := tmp.Security.Proto()
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if rule.GetAccess() == nil {
			return fmt.Errorf("the 'canvas' resource type only supports 'access' security rules")
		}
	}

	// Track canvas
	r, err := p.insertResource(ResourceKindCanvas, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.CanvasSpec.DisplayName = tmp.DisplayName
	if r.CanvasSpec.DisplayName == "" {
		r.CanvasSpec.DisplayName = ToDisplayName(node.Name)
	}
	r.CanvasSpec.Banner = tmp.Banner
	r.CanvasSpec.MaxWidth = tmp.MaxWidth
	r.CanvasSpec.GapX = tmp.GapX
	r.CanvasSpec.GapY = tmp.GapY
	r.CanvasSpec.Theme = themeName
	r.CanvasSpec.AllowCustomTimeRange = allowCustomTimeRange
	r.CanvasSpec.TimeRanges = timeRanges
	r.CanvasSpec.TimeZones = tmp.TimeZones
	r.CanvasSpec.FiltersEnabled = true
	if tmp.Filters.Enable != nil {
		r.CanvasSpec.FiltersEnabled = *tmp.Filters.Enable
	}
	r.CanvasSpec.DefaultPreset = defaultPreset
	r.CanvasSpec.EmbeddedTheme = themeSpec
	r.CanvasSpec.Variables = variables
	r.CanvasSpec.Rows = rows
	r.CanvasSpec.SecurityRules = rules

	// Track inline components
	for _, def := range inlineComponentDefs {
		r, err := p.insertResource(ResourceKindComponent, def.name, node.Paths, def.refs...)
		if err != nil {
			// Normally we could return the error, but we can't do that here because we've already inserted the canvas.
			// Since the component has been validated with insertDryRun in parseCanvasItemComponent, this error should never happen in practice.
			// So let's panic.
			panic(err)
		}
		r.ComponentSpec = def.spec
	}

	return nil
}

// parseCanvasInlineComponent parses an inline component definition in a canvas item.
func (p *Parser) parseCanvasInlineComponent(canvasName string, rowIdx, itemIdx int, props map[string]yaml.Node) (string, *componentDef, error) {
	var n yaml.Node
	err := n.Encode(props)
	if err != nil {
		return "", nil, fmt.Errorf("invalid component for item %d in row %d: %w", itemIdx, rowIdx, err)
	}

	tmp := &ComponentYAML{}
	err = n.Decode(tmp)
	if err != nil {
		return "", nil, err
	}

	spec, refs, err := p.parseComponentYAML(tmp)
	if err != nil {
		return "", nil, err
	}

	spec.DefinedInCanvas = true

	name := fmt.Sprintf("%s--component-%d-%d", canvasName, rowIdx, itemIdx)

	err = p.insertDryRun(ResourceKindComponent, name)
	if err != nil {
		name = fmt.Sprintf("%s--component-%d-%d-%s", canvasName, rowIdx, itemIdx, uuid.New())
		err = p.insertDryRun(ResourceKindComponent, name)
		if err != nil {
			return "", nil, err
		}
	}

	def := &componentDef{
		name: name,
		refs: refs,
		spec: spec,
	}

	return name, def, nil
}

type componentDef struct {
	name string
	refs []ResourceName
	spec *runtimev1.ComponentSpec
}

// itemSizeRegex is used for parseItemSize.
var itemSizeRegex = regexp.MustCompile(`^(\d+)\s*(.*)$`)

// parseItemSize parses a string of the format "<int><space?><unit?>".
// Examples: "100", "100px", "100 px".
func parseItemSize(s string) (uint32, string, error) {
	if s == "" {
		return 0, "", nil
	}

	matches := itemSizeRegex.FindStringSubmatch(s)
	if matches == nil {
		return 0, "", fmt.Errorf("invalid size %q", s)
	}

	size, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		return 0, "", fmt.Errorf("invalid size %q: %w", s, err)
	}

	return uint32(size), matches[2], nil
}

func pointerIfNotEmpty(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
