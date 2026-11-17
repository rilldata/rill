package parser

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strconv"
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
	Light  *ThemeColors     `yaml:"light"`
	Dark   *ThemeColors     `yaml:"dark"`
	Font   *ThemeFontYAML   `yaml:"font"`
	Chrome *ThemeChromeYAML `yaml:"chrome"`
}

type ThemeColors struct {
	Primary   string            `yaml:"primary"`
	Secondary string            `yaml:"secondary"`
	Variables map[string]string `yaml:",inline"`
}

// ThemeFontYAML is the raw structure of a theme's font.
type ThemeFontYAML struct {
	Family string `yaml:"family"`
	Files  []struct {
		Path   string `yaml:"path"`
		Weight string `yaml:"weight"` // A single weight like "400", or a range like "100 900" for a variable font
		Style  string `yaml:"style"`  // "normal" or "italic"
	} `yaml:"files"`
}

// ThemeChromeYAML is the raw structure of a theme's chrome tokens.
// The lengths are in pixels, with an optional "px" suffix.
type ThemeChromeYAML struct {
	BorderWidth string `yaml:"border_width"`
	FontSize    string `yaml:"font_size"`
	FontWeight  string `yaml:"font_weight"`
}

// themeFontFamilyRegexp matches a single font family name.
// It excludes quotes, commas and other CSS syntax, so that the name can be used in a font-family declaration as is.
var themeFontFamilyRegexp = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N} ._-]*$`)

// themeFontGenericFamilies are the CSS keywords that name a font the browser picks, which can't be provided by files.
var themeFontGenericFamilies = map[string]bool{
	"serif":         true,
	"sans-serif":    true,
	"monospace":     true,
	"cursive":       true,
	"fantasy":       true,
	"system-ui":     true,
	"ui-serif":      true,
	"ui-sans-serif": true,
	"ui-monospace":  true,
	"ui-rounded":    true,
}

// themeFontFileExtensions are the font file formats that browsers can load.
var themeFontFileExtensions = map[string]bool{
	".woff2": true,
	".woff":  true,
	".ttf":   true,
	".otf":   true,
}

// maxThemeFontFiles limits the number of files a viewer's browser downloads for a theme.
const maxThemeFontFiles = 16

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

	// Dimension styling
	"dimension":            true,
	"dimension-foreground": true,
	"dimension-border":     true,

	// Measure styling
	"measure":            true,
	"measure-foreground": true,
	"measure-border":     true,

	// Tooltip
	"tooltip": true,

	// Destructive actions
	"destructive":            true,
	"destructive-foreground": true,

	// KPI delta coloring
	"kpi-positive": true,
	"kpi-negative": true,

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

	r, err := p.insertResource(ResourceKindTheme, node.Name, node.Paths, node.Tags, node.Metadata, node.Refs...)
	if err != nil {
		return err
	}

	r.ThemeSpec = spec

	return nil
}

func (p *Parser) parseThemeYAML(tmp *ThemeYAML) (*runtimev1.ThemeSpec, error) {
	spec := &runtimev1.ThemeSpec{}

	if tmp.Font != nil {
		// Same default as in rill.yaml, for when rill.yaml is missing or invalid.
		publicPaths := []string{"public"}
		if p.RillYAML != nil {
			publicPaths = p.RillYAML.PublicPaths
		}

		font, err := tmp.Font.validate(publicPaths)
		if err != nil {
			return nil, fmt.Errorf("invalid font: %w", err)
		}
		spec.Font = font
	}

	if tmp.Chrome != nil {
		chrome, err := tmp.Chrome.validate()
		if err != nil {
			return nil, fmt.Errorf("invalid chrome: %w", err)
		}
		spec.Chrome = chrome
	}

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

// validate checks the font and converts it to its spec.
// The font files must be inside one of publicPaths, since those are the only project files served to viewers.
func (f *ThemeFontYAML) validate(publicPaths []string) (*runtimev1.ThemeFont, error) {
	family := strings.TrimSpace(f.Family)
	if family == "" {
		return nil, errors.New(`"family" is required`)
	}
	if len(family) > 64 || !themeFontFamilyRegexp.MatchString(family) {
		return nil, fmt.Errorf("invalid family %q: must be a single font family name of letters, digits, spaces, dots, hyphens and underscores", f.Family)
	}

	if len(f.Files) > 0 && themeFontGenericFamilies[strings.ToLower(family)] {
		return nil, fmt.Errorf("the generic family %q is provided by the browser and can't have files", family)
	}
	if len(f.Files) > maxThemeFontFiles {
		return nil, fmt.Errorf("at most %d files are supported, got %d", maxThemeFontFiles, len(f.Files))
	}

	font := &runtimev1.ThemeFont{Family: family}
	for i, file := range f.Files {
		// Same normalized form as the paths that the runtime's assets endpoint serves.
		filePath := strings.TrimPrefix(path.Clean("/"+strings.ReplaceAll(file.Path, `\`, "/")), "/")
		if file.Path == "" || filePath == "" {
			return nil, fmt.Errorf(`files[%d]: "path" is required`, i)
		}

		public := false
		for _, publicPath := range publicPaths {
			dir := strings.Trim(path.Clean("/"+strings.ReplaceAll(publicPath, `\`, "/")), "/")
			if dir != "" && strings.HasPrefix(filePath, dir+"/") {
				public = true
				break
			}
		}
		if !public {
			return nil, fmt.Errorf("files[%d]: path %q must be inside a public path (%s)", i, file.Path, strings.Join(publicPaths, ", "))
		}

		if !themeFontFileExtensions[strings.ToLower(path.Ext(filePath))] {
			return nil, fmt.Errorf("files[%d]: path %q must be a .woff2, .woff, .ttf or .otf file", i, file.Path)
		}

		// A single weight, or the lowest and highest weight of a variable font.
		weights := strings.Fields(file.Weight)
		if len(weights) == 0 {
			weights = []string{"400"}
		}
		if len(weights) == 1 {
			weights = append(weights, weights[0])
		}
		weightMin, minErr := strconv.ParseUint(weights[0], 10, 32)
		weightMax, maxErr := strconv.ParseUint(weights[1], 10, 32)
		if len(weights) > 2 || minErr != nil || maxErr != nil || weightMin < 1 || weightMax > 1000 || weightMin > weightMax {
			return nil, fmt.Errorf(`files[%d]: invalid weight %q: must be a number between 1 and 1000, or a range such as "100 900" for a variable font`, i, file.Weight)
		}

		var italic bool
		switch strings.ToLower(file.Style) {
		case "", "normal":
		case "italic":
			italic = true
		default:
			return nil, fmt.Errorf(`files[%d]: invalid style %q: must be "normal" or "italic"`, i, file.Style)
		}

		font.Files = append(font.Files, &runtimev1.ThemeFontFile{
			Path:      filePath,
			WeightMin: uint32(weightMin),
			WeightMax: uint32(weightMax),
			Italic:    italic,
		})
	}

	return font, nil
}

// validate checks the chrome tokens and converts them to their spec.
func (c *ThemeChromeYAML) validate() (*runtimev1.ThemeChrome, error) {
	chrome := &runtimev1.ThemeChrome{}

	if c.BorderWidth != "" {
		v, err := parseThemeNumber(strings.TrimSuffix(c.BorderWidth, "px"), 0, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid border_width %q: must be between 0px and 8px", c.BorderWidth)
		}
		chrome.BorderWidthPx = &v
	}

	if c.FontSize != "" {
		v, err := parseThemeNumber(strings.TrimSuffix(c.FontSize, "px"), 10, 20)
		if err != nil {
			return nil, fmt.Errorf("invalid font_size %q: must be between 10px and 20px", c.FontSize)
		}
		chrome.FontSizePx = &v
	}

	if c.FontWeight != "" {
		v, err := parseThemeNumber(c.FontWeight, 100, 900)
		if err != nil {
			return nil, fmt.Errorf("invalid font_weight %q: must be a number between 100 and 900", c.FontWeight)
		}
		chrome.FontWeight = &v
	}

	return chrome, nil
}

// parseThemeNumber parses a whole number in the range from minVal to maxVal.
func parseThemeNumber(s string, minVal, maxVal uint32) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return 0, err
	}
	if v < uint64(minVal) || v > uint64(maxVal) {
		return 0, fmt.Errorf("%d is out of range", v)
	}
	return uint32(v), nil
}
