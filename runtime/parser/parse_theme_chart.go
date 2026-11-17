package parser

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/mazznoer/csscolorparser"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

// ThemeChartYAML is the raw structure of a theme's `chart` section.
// Its own properties apply to both modes; `light` and `dark` override them for one mode.
type ThemeChartYAML struct {
	ThemeChartStyleYAML `yaml:",inline"`
	Light               *ThemeChartStyleYAML `yaml:"light"`
	Dark                *ThemeChartStyleYAML `yaml:"dark"`
}

// ThemeChartStyleYAML is the chart styling for one mode.
// Numbers and booleans are pointers because zero and false are settings of their own: no stroke, no gridlines.
type ThemeChartStyleYAML struct {
	Extends    string `yaml:"extends"`
	FontFamily string `yaml:"font_family"`
	Text       struct {
		Title       ThemeChartTextRoleYAML `yaml:"title"`
		AxisLabel   ThemeChartTextRoleYAML `yaml:"axis_label"`
		AxisTitle   ThemeChartTextRoleYAML `yaml:"axis_title"`
		LegendLabel ThemeChartTextRoleYAML `yaml:"legend_label"`
		LegendTitle ThemeChartTextRoleYAML `yaml:"legend_title"`
		ValueLabel  ThemeChartTextRoleYAML `yaml:"value_label"`
	} `yaml:"text"`
	Grid struct {
		Measure  *bool     `yaml:"measure"`
		Category *bool     `yaml:"category"`
		Style    string    `yaml:"style"`
		Dash     []float64 `yaml:"dash"`
		Width    *float64  `yaml:"width"`
		Color    string    `yaml:"color"`
	} `yaml:"grid"`
	Axis struct {
		Line      *bool    `yaml:"line"`
		LineWidth *float64 `yaml:"line_width"`
		Ticks     *bool    `yaml:"ticks"`
		Color     string   `yaml:"color"`
	} `yaml:"axis"`
	Line struct {
		Width *float64 `yaml:"width"`
	} `yaml:"line"`
	Area struct {
		FillOpacity *float64 `yaml:"fill_opacity"`
		LineWidth   *float64 `yaml:"line_width"`
	} `yaml:"area"`
	Bar struct {
		Opacity      *float64 `yaml:"opacity"`
		CornerRadius *float64 `yaml:"corner_radius"`
	} `yaml:"bar"`
	Point struct {
		Size    *float64 `yaml:"size"`
		Opacity *float64 `yaml:"opacity"`
	} `yaml:"point"`
	Legend struct {
		Position string `yaml:"position"`
		Symbol   string `yaml:"symbol"`
	} `yaml:"legend"`
}

// ThemeChartTextRoleYAML styles one role text plays in a chart.
type ThemeChartTextRoleYAML struct {
	Size *float64 `yaml:"size"`
	// Weight is a CSS font weight, as a number (500) or a name (medium).
	Weight string `yaml:"weight"`
	Color  string `yaml:"color"`
}

// themeChartFontWeights maps the names of CSS font weights to their numbers.
var themeChartFontWeights = map[string]uint32{
	"thin":       100,
	"extralight": 200,
	"light":      300,
	"regular":    400,
	"normal":     400,
	"medium":     500,
	"semibold":   600,
	"bold":       700,
	"extrabold":  800,
	"black":      900,
}

var (
	themeChartGridStyles      = []string{"solid", "dashed", "dotted"}
	themeChartLegendPositions = []string{"top", "bottom", "left", "right"}
	themeChartLegendSymbols   = []string{"circle", "square", "line"}
)

// parseThemeChart parses the `chart` section of a theme.
// It returns nil if the section is absent or sets nothing.
func parseThemeChart(node *yaml.Node) (*runtimev1.ThemeChart, error) {
	if node.IsZero() || node.Tag == "!!null" {
		return nil, nil
	}

	// A typo in a property name would otherwise be dropped silently and the chart would keep its default.
	// yaml.Node.Decode has no strict mode, so the section goes through a decoder that does.
	data, err := yaml.Marshal(node)
	if err != nil {
		return nil, err
	}
	tmp := &ThemeChartYAML{}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(tmp); err != nil {
		return nil, newYAMLError(err)
	}

	base, err := tmp.ThemeChartStyleYAML.proto()
	if err != nil {
		return nil, err
	}

	res := &runtimev1.ThemeChart{}
	modes := []struct {
		name     string
		override *ThemeChartStyleYAML
		dst      **runtimev1.ThemeChartStyle
	}{
		{"light", tmp.Light, &res.Light},
		{"dark", tmp.Dark, &res.Dark},
	}
	for _, mode := range modes {
		style := proto.CloneOf(base)
		if mode.override != nil {
			override, err := mode.override.proto()
			if err != nil {
				return nil, fmt.Errorf("%s: %w", mode.name, err)
			}
			// proto.Merge appends lists, but a mode's dash pattern replaces the shared one.
			// So does its style: a shared dash pattern would otherwise outlive `style: solid`.
			if g := override.Grid; style.Grid != nil && g != nil && (g.Style != "" || len(g.Dash) > 0) {
				style.Grid.Dash = nil
			}
			proto.Merge(style, override)
		}
		*mode.dst = nilIfUnset(style)
	}

	if res.Light == nil && res.Dark == nil {
		return nil, nil
	}
	return res, nil
}

// proto validates the style and converts it to its proto form.
// Sections that set nothing are left nil.
func (s *ThemeChartStyleYAML) proto() (*runtimev1.ThemeChartStyle, error) {
	if strings.ContainsAny(s.FontFamily, ";{}<>\n") {
		return nil, fmt.Errorf("font_family: must be a list of font names, got %q", s.FontFamily)
	}

	err := errors.Join(
		validateThemeChartOneOf("grid.style", s.Grid.Style, themeChartGridStyles),
		validateThemeChartDash(s.Grid.Dash),
		validateThemeChartMin("grid.width", s.Grid.Width),
		validateThemeChartColor("grid.color", s.Grid.Color),
		validateThemeChartMin("axis.line_width", s.Axis.LineWidth),
		validateThemeChartColor("axis.color", s.Axis.Color),
		validateThemeChartMin("line.width", s.Line.Width),
		validateThemeChartOpacity("area.fill_opacity", s.Area.FillOpacity),
		validateThemeChartMin("area.line_width", s.Area.LineWidth),
		validateThemeChartOpacity("bar.opacity", s.Bar.Opacity),
		validateThemeChartMin("bar.corner_radius", s.Bar.CornerRadius),
		validateThemeChartMin("point.size", s.Point.Size),
		validateThemeChartOpacity("point.opacity", s.Point.Opacity),
		validateThemeChartOneOf("legend.position", s.Legend.Position, themeChartLegendPositions),
		validateThemeChartOneOf("legend.symbol", s.Legend.Symbol, themeChartLegendSymbols),
	)
	if err != nil {
		return nil, err
	}

	text := &runtimev1.ThemeChartText{}
	roles := []struct {
		name string
		role ThemeChartTextRoleYAML
		dst  **runtimev1.ThemeChartTextRole
	}{
		{"title", s.Text.Title, &text.Title},
		{"axis_label", s.Text.AxisLabel, &text.AxisLabel},
		{"axis_title", s.Text.AxisTitle, &text.AxisTitle},
		{"legend_label", s.Text.LegendLabel, &text.LegendLabel},
		{"legend_title", s.Text.LegendTitle, &text.LegendTitle},
		{"value_label", s.Text.ValueLabel, &text.ValueLabel},
	}
	for _, r := range roles {
		role, err := r.role.proto()
		if err != nil {
			return nil, fmt.Errorf("text.%s.%w", r.name, err)
		}
		*r.dst = role
	}

	return &runtimev1.ThemeChartStyle{
		Extends:    strings.TrimSpace(s.Extends),
		FontFamily: strings.TrimSpace(s.FontFamily),
		Text:       nilIfUnset(text),
		Grid: nilIfUnset(&runtimev1.ThemeChartGrid{
			Measure:  s.Grid.Measure,
			Category: s.Grid.Category,
			Style:    s.Grid.Style,
			Dash:     s.Grid.Dash,
			Width:    s.Grid.Width,
			Color:    s.Grid.Color,
		}),
		Axis: nilIfUnset(&runtimev1.ThemeChartAxis{
			Line:      s.Axis.Line,
			LineWidth: s.Axis.LineWidth,
			Ticks:     s.Axis.Ticks,
			Color:     s.Axis.Color,
		}),
		Line: nilIfUnset(&runtimev1.ThemeChartLine{Width: s.Line.Width}),
		Area: nilIfUnset(&runtimev1.ThemeChartArea{
			FillOpacity: s.Area.FillOpacity,
			LineWidth:   s.Area.LineWidth,
		}),
		Bar: nilIfUnset(&runtimev1.ThemeChartBar{
			Opacity:      s.Bar.Opacity,
			CornerRadius: s.Bar.CornerRadius,
		}),
		Point: nilIfUnset(&runtimev1.ThemeChartPoint{
			Size:    s.Point.Size,
			Opacity: s.Point.Opacity,
		}),
		Legend: nilIfUnset(&runtimev1.ThemeChartLegend{
			Position: s.Legend.Position,
			Symbol:   s.Legend.Symbol,
		}),
	}, nil
}

// proto validates the text role and converts it to its proto form.
// It returns nil if the role sets nothing.
// Its errors start with the name of the offending property.
func (r ThemeChartTextRoleYAML) proto() (*runtimev1.ThemeChartTextRole, error) {
	res := &runtimev1.ThemeChartTextRole{Size: r.Size, Color: r.Color}

	if r.Size != nil && *r.Size <= 0 {
		return nil, fmt.Errorf("size: must be greater than 0, got %v", *r.Size)
	}

	if r.Weight != "" {
		weight, ok := themeChartFontWeights[strings.ToLower(r.Weight)]
		if !ok {
			n, err := strconv.ParseUint(r.Weight, 10, 32)
			if err != nil || n < 100 || n > 900 || n%100 != 0 {
				return nil, fmt.Errorf("weight: must be a multiple of 100 from 100 to 900 or a weight name such as \"medium\", got %q", r.Weight)
			}
			weight = uint32(n)
		}
		res.Weight = &weight
	}

	if err := validateThemeChartColor("color", r.Color); err != nil {
		return nil, err
	}

	return nilIfUnset(res), nil
}

// nilIfUnset returns nil for a message that sets nothing, so that an empty section is absent from the spec.
func nilIfUnset[M interface {
	comparable
	proto.Message
}](m M) M {
	set := false
	m.ProtoReflect().Range(func(protoreflect.FieldDescriptor, protoreflect.Value) bool {
		set = true
		return false
	})
	if !set {
		var unset M
		return unset
	}
	return m
}

func validateThemeChartColor(name, value string) error {
	if value == "" {
		return nil
	}
	if _, err := csscolorparser.Parse(value); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func validateThemeChartOneOf(name, value string, allowed []string) error {
	if value == "" || slices.Contains(allowed, value) {
		return nil
	}
	return fmt.Errorf("%s: must be one of %s, got %q", name, strings.Join(allowed, ", "), value)
}

func validateThemeChartMin(name string, value *float64) error {
	if value == nil || *value >= 0 {
		return nil
	}
	return fmt.Errorf("%s: must not be negative, got %v", name, *value)
}

func validateThemeChartOpacity(name string, value *float64) error {
	if value == nil || (*value >= 0 && *value <= 1) {
		return nil
	}
	return fmt.Errorf("%s: must be between 0 and 1, got %v", name, *value)
}

// validateThemeChartDash checks a dash pattern: lengths that draw something.
func validateThemeChartDash(dash []float64) error {
	if len(dash) == 0 {
		return nil
	}
	drawn := false
	for _, length := range dash {
		if length < 0 {
			return fmt.Errorf("grid.dash: lengths must not be negative, got %v", length)
		}
		drawn = drawn || length > 0
	}
	if !drawn {
		return errors.New("grid.dash: needs a length greater than 0")
	}
	return nil
}
