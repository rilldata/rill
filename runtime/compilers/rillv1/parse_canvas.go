package rillv1

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type CanvasYAML struct {
	commonYAML  `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	DisplayName string                 `yaml:"display_name"`
	Title       string                 `yaml:"title"` // Deprecated: use display_name
	MaxWidth    uint32                 `yaml:"max_width"`
	GapX        uint32                 `yaml:"gap_x"`
	GapY        uint32                 `yaml:"gap_y"`
	Theme       yaml.Node              `yaml:"theme"` // Name (string) or inline theme definition (map)
	TimeRanges  []ExploreTimeRangeYAML `yaml:"time_ranges"`
	TimeZones   []string               `yaml:"time_zones"`
	Defaults    *struct {
		TimeRange           string `yaml:"time_range"`
		ComparisonMode      string `yaml:"comparison_mode"`
		ComparisonDimension string `yaml:"comparison_dimension"`
	} `yaml:"defaults"`
	Variables []*ComponentVariableYAML `yaml:"variables"`
	Items     []*struct {
		Component yaml.Node `yaml:"component"` // Can be a name (string) or inline component definition (map)
		X         *uint32   `yaml:"x"`
		Y         *uint32   `yaml:"y"`
		Width     *uint32   `yaml:"width"`
		Height    *uint32   `yaml:"height"`
	} `yaml:"items"`
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

	// Parse items.
	// Each item can either reference an externally defined component by name or define a component inline.
	var items []*runtimev1.CanvasItem
	var inlineComponentDefs []*componentDef
	for i, item := range tmp.Items {
		if item == nil {
			return fmt.Errorf("item at index %d is nil", i)
		}

		component, inlineComponentDef, err := p.parseCanvasItemComponent(node.Name, i, item.Component)
		if err != nil {
			return fmt.Errorf("invalid component at index %d: %w", i, err)
		}

		// Track inline component definitions so we can insert them after we have validated all components
		if inlineComponentDef != nil {
			inlineComponentDefs = append(inlineComponentDefs, inlineComponentDef)
		}

		items = append(items, &runtimev1.CanvasItem{
			Component:       component,
			DefinedInCanvas: inlineComponentDef != nil,
			X:               item.X,
			Y:               item.Y,
			Width:           item.Width,
			Height:          item.Height,
		})

		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindComponent, Name: component})
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

	if r.CanvasSpec.DisplayName == "" {
		r.CanvasSpec.DisplayName = ToDisplayName(node.Name)
	}
	r.CanvasSpec.MaxWidth = tmp.MaxWidth
	r.CanvasSpec.GapX = tmp.GapX
	r.CanvasSpec.GapY = tmp.GapY
	r.CanvasSpec.Theme = themeName
	r.CanvasSpec.TimeRanges = timeRanges
	r.CanvasSpec.TimeZones = tmp.TimeZones
	r.CanvasSpec.DefaultPreset = defaultPreset
	r.CanvasSpec.EmbeddedTheme = themeSpec
	r.CanvasSpec.Variables = variables
	r.CanvasSpec.Items = items
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

// parseCanvasItemComponent parses a canvas item's "component" property.
// It may be a string (name of an externally defined component) or an inline component definition.
func (p *Parser) parseCanvasItemComponent(canvasName string, idx int, n yaml.Node) (string, *componentDef, error) {
	if n.Kind == yaml.ScalarNode {
		var name string
		err := n.Decode(&name)
		if err != nil {
			return "", nil, err
		}
		return name, nil, nil
	}

	if n.Kind != yaml.MappingNode {
		return "", nil, errors.New("expected a component name or inline declaration")
	}

	tmp := &ComponentYAML{}
	err := n.Decode(tmp)
	if err != nil {
		return "", nil, err
	}

	spec, refs, err := p.parseComponentYAML(tmp)
	if err != nil {
		return "", nil, err
	}

	spec.DefinedInCanvas = true

	name := fmt.Sprintf("%s--component-%d", canvasName, idx)

	err = p.insertDryRun(ResourceKindComponent, name)
	if err != nil {
		name = fmt.Sprintf("%s--component-%d-%s", canvasName, idx, uuid.New())
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

func pointerIfNotEmpty(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
