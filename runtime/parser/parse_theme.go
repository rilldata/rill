package parser

import (
	"fmt"
	"maps"
	"regexp"
	"strings"

	"github.com/mazznoer/csscolorparser"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// ThemeYAML is the raw structure of a Theme for the UI in YAML (does not include common fields)
type ThemeYAML struct {
	commonYAML `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into Properties
	Colors     struct {
		Primary   string `yaml:"primary"`
		Secondary string `yaml:"secondary"`
	} `yaml:"colors"`
	Light *ThemeColors `yaml:"light"`
	Dark  *ThemeColors `yaml:"dark"`
}

// ThemeColors is the raw structure of a theme's `light` or `dark` block.
// CSS variables can be set under `variables` or directly in the block.
// See ThemeVariableGroups for the supported variables.
type ThemeColors struct {
	Primary       string            `yaml:"primary"`
	Secondary     string            `yaml:"secondary"`
	Variables     map[string]string `yaml:"variables"`
	FlatVariables map[string]string `yaml:",inline"`
}

// cssLengthRegexp matches the lengths accepted for theme variables: a non-negative number with a px, rem or em unit.
var cssLengthRegexp = regexp.MustCompile(`^(\d+(\.\d+)?|\.\d+)(px|rem|em)$`)

// parseTheme parses a theme definition and adds the resulting resource to p.Resources.
func (p *Parser) parseTheme(node *Node) error {
	tmp := &ThemeYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
	}

	spec, err := p.parseThemeYAML(tmp)
	if err != nil {
		return err
	}

	r, err := p.insertResource(ResourceKindTheme, node.Name, node.Paths, node.Tags, node.Metadata, node.Refs...)
	if err != nil {
		return err
	}

	r.ThemeSpec = spec

	return nil
}

func (p *Parser) parseThemeYAML(tmp *ThemeYAML) (*runtimev1.ThemeSpec, error) {
	spec := &runtimev1.ThemeSpec{}

	hasLegacyColors := tmp.Colors.Primary != "" || tmp.Colors.Secondary != ""
	hasCSSProperties := tmp.Light != nil || tmp.Dark != nil

	if hasLegacyColors && hasCSSProperties {
		return nil, fmt.Errorf("cannot use both legacy color properties (primary, secondary) and the new CSS property simultaneously")
	}

	if hasLegacyColors {
		if tmp.Colors.Primary != "" {
			pc, err := csscolorparser.Parse(tmp.Colors.Primary)
			if err != nil {
				return nil, fmt.Errorf("invalid primary color: %w", err)
			}
			spec.PrimaryColor = toThemeColor(pc)
			spec.PrimaryColorRaw = tmp.Colors.Primary
		}

		if tmp.Colors.Secondary != "" {
			sc, err := csscolorparser.Parse(tmp.Colors.Secondary)
			if err != nil {
				return nil, fmt.Errorf("invalid secondary color: %w", err)
			}
			spec.SecondaryColor = toThemeColor(sc)
			spec.SecondaryColorRaw = tmp.Colors.Secondary
		}

		return spec, nil
	}

	var err error
	if tmp.Light != nil {
		spec.Light, err = tmp.Light.validate()
		if err != nil {
			return nil, fmt.Errorf("invalid light theme: %w", err)
		}
	}

	if tmp.Dark != nil {
		spec.Dark, err = tmp.Dark.validate()
		if err != nil {
			return nil, fmt.Errorf("invalid dark theme: %w", err)
		}
	}

	return spec, nil
}

func toThemeColor(c csscolorparser.Color) *runtimev1.Color {
	return &runtimev1.Color{
		Red:   float32(c.R),
		Green: float32(c.G),
		Blue:  float32(c.B),
		Alpha: float32(c.A),
	}
}

func (t *ThemeColors) validate() (*runtimev1.ThemeColors, error) {
	res := &runtimev1.ThemeColors{
		Primary:   t.Primary,
		Secondary: t.Secondary,
		Variables: make(map[string]string),
	}

	if t.Primary != "" {
		_, err := csscolorparser.Parse(t.Primary)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for primary: %w", t.Primary, err)
		}
	}

	if t.Secondary != "" {
		_, err := csscolorparser.Parse(t.Secondary)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for secondary: %w", t.Secondary, err)
		}
	}

	// A variable set under `variables` takes precedence over the same variable set directly in the block.
	vars := make(map[string]string, len(t.FlatVariables)+len(t.Variables))
	maps.Copy(vars, t.FlatVariables)
	maps.Copy(vars, t.Variables)

	for k, v := range vars {
		def, ok := themeVariables[k]
		if !ok {
			return nil, fmt.Errorf("invalid CSS variable: %q", k)
		}

		if def.Length {
			v = strings.ToLower(strings.TrimSpace(v))
			if v == "0" {
				// A unitless zero is a valid length on its own, but not inside calc(), which the frontend uses to derive lengths from it.
				v = "0px"
			}
			if !cssLengthRegexp.MatchString(v) {
				return nil, fmt.Errorf("invalid value %q for CSS variable %q: must be a length in px, rem or em, such as 8px or 0.5rem", vars[k], k)
			}
		} else {
			_, err := csscolorparser.Parse(v)
			if err != nil {
				return nil, fmt.Errorf("invalid value %q for CSS variable %q: %w", v, k, err)
			}
		}

		switch {
		case k == "primary":
			// Only reachable through `variables`: it is equivalent to setting `primary` directly in the block.
			res.Primary = v
		case k == "secondary":
			res.Secondary = v
		case len(def.ReplacedBy) > 0:
			// Deprecated name: pass the value on under the new names, unless they are set explicitly.
			for _, newName := range def.ReplacedBy {
				if _, ok := vars[newName]; !ok {
					res.Variables[newName] = v
				}
			}
		default:
			res.Variables[k] = v
		}
	}

	return res, nil
}
