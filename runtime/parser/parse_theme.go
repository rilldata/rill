package parser

import (
	"fmt"

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

type ThemeColors struct {
	Primary   string            `yaml:"primary"`
	Secondary string            `yaml:"secondary"`
	Variables map[string]string `yaml:",inline"`
}


// deprecatedCSSVariables maps deprecated variable names to their new semantic equivalents.
// Old themes using these names will continue to work through automatic mapping.
var deprecatedCSSVariables = map[string][]string{
	"surface":           {"surface-background", "surface-base"},
	"background":        {"surface-subtle"},
	"card":              {"surface-card"},
	"card-foreground":   {"fg-secondary"},
	"muted":             {"surface-muted"},
	"muted-foreground":  {"fg-muted"},
	"accent":            {"popover-accent"},
	"accent-foreground": {"fg-accent"},
	"ring":              {"ring-focus"},
	"foreground":        {"fg-primary"},
}

var allowedCSSVariables = map[string]bool{
	// Primary theme colors
	"primary":   true,
	"secondary": true,

	
	// Surface semantic variables
	"surface-base":       true,
	"surface-subtle":     true,
	"surface-background": true,
	"surface-hover":      true,
	"surface-active":     true,
	"surface-overlay":    true,
	"surface-muted":      true,
	"surface-card":       true,

	// Foreground semantic variables
	"fg-primary":   true,
	"fg-secondary": true,
	"fg-tertiary":  true,
	"fg-inverse":   true,
	"fg-muted":     true,
	"fg-disabled":  true,
	"fg-accent":    true,

	// Accent semantic variables
	"accent-primary":          true,
	"accent-primary-action":   true,
	"accent-secondary":        true,
	"accent-secondary-action": true,

	// Icon semantic variables
	"icon-default":  true,
	"icon-muted":    true,
	"icon-disabled": true,
	"icon-accent":   true,

	// Border and input
	"border": true,
	"input":  true,

	// Ring (focus states)
	"ring-focus":  true,
	"ring-offset": true,

	// Tooltip
	"tooltip": true,

	// Destructive actions
	"destructive":            true,
	"destructive-foreground": true,

	// Popover
	"popover":            true,
	"popover-accent":     true,
	"popover-foreground": true,
	"popover-footer":     true,

	// Non-deprecated misc
	"radius": true,

	// Deprecated but still allowed (mapped automatically)
	"ring":                 true,
	"surface":              true,
	"background":           true,
	"foreground":           true,
	"card":                 true,
	"card-foreground":      true,
	"muted":                true,
	"muted-foreground":     true,
	"accent":               true,
	"accent-foreground":    true,
	"primary-foreground":   true,
	"secondary-foreground": true,

	// Sequential palette (9 colors)
	"color-sequential-1": true,
	"color-sequential-2": true,
	"color-sequential-3": true,
	"color-sequential-4": true,
	"color-sequential-5": true,
	"color-sequential-6": true,
	"color-sequential-7": true,
	"color-sequential-8": true,
	"color-sequential-9": true,

	// Diverging palette (11 colors)
	"color-diverging-1":  true,
	"color-diverging-2":  true,
	"color-diverging-3":  true,
	"color-diverging-4":  true,
	"color-diverging-5":  true,
	"color-diverging-6":  true,
	"color-diverging-7":  true,
	"color-diverging-8":  true,
	"color-diverging-9":  true,
	"color-diverging-10": true,
	"color-diverging-11": true,

	// Qualitative palette (24 colors)
	"color-qualitative-1":  true,
	"color-qualitative-2":  true,
	"color-qualitative-3":  true,
	"color-qualitative-4":  true,
	"color-qualitative-5":  true,
	"color-qualitative-6":  true,
	"color-qualitative-7":  true,
	"color-qualitative-8":  true,
	"color-qualitative-9":  true,
	"color-qualitative-10": true,
	"color-qualitative-11": true,
	"color-qualitative-12": true,
	"color-qualitative-13": true,
	"color-qualitative-14": true,
	"color-qualitative-15": true,
	"color-qualitative-16": true,
	"color-qualitative-17": true,
	"color-qualitative-18": true,
	"color-qualitative-19": true,
	"color-qualitative-20": true,
	"color-qualitative-21": true,
	"color-qualitative-22": true,
	"color-qualitative-23": true,
	"color-qualitative-24": true,
}

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

	r, err := p.insertResource(ResourceKindTheme, node.Name, node.Paths, node.Refs...)
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
	// Create a new map for the final variables with deprecated names mapped to new ones
	finalVariables := make(map[string]string)

	for k, v := range t.Variables {
		if !allowedCSSVariables[k] {
			return nil, fmt.Errorf("invalid CSS variable: %q", k)
		}
		_, err := csscolorparser.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for CSS variable %q: %w", v, k, err)
		}

		// Check if this is a deprecated variable that should be mapped
		if newNames, isDeprecated := deprecatedCSSVariables[k]; isDeprecated {
			for _, newName := range newNames {
				// Only map to the new name if it's not already explicitly set
				if _, alreadySet := t.Variables[newName]; !alreadySet {
					finalVariables[newName] = v
				}
			}
			// Don't output the deprecated name
		} else {
			finalVariables[k] = v
		}
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

	return &runtimev1.ThemeColors{
		Primary:   t.Primary,
		Secondary: t.Secondary,
		Variables: finalVariables,
	}, nil
}
