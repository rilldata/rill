package docs

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mazznoer/csscolorparser"
	"github.com/rilldata/rill/runtime/parser"
)

// themeStylesheets are the frontend stylesheets that declare the default value of each theme variable, in cascade order.
var themeStylesheets = []string{
	"web-common/src/colors.css",
	"web-common/src/styles/chart-base.css",
	"web-common/src/app.css",
}

var (
	cssCommentRegexp  = regexp.MustCompile(`(?s)/\*.*?\*/`)
	cssRootRuleRegexp = regexp.MustCompile(`:root(\.dark)?\s*\{([^}]*)\}`)
	// cssBrandShadeRegexp matches the palettes that Rill generates from a theme's `primary` and `secondary` colors.
	cssBrandShadeRegexp = regexp.MustCompile(`^color-(?:theme-)?(primary|secondary|theme)-(?:light-|dark-)?(\d+)$`)
	cssColorMixRegexp   = regexp.MustCompile(`^color-mix\(in oklab, (.+) calc\(([\d.]+) \* 100%\), transparent\)$`)
	cssOKLCHRegexp      = regexp.MustCompile(`^oklch\(([\d.]+)(%?) ([\d.]+) ([\d.]+|none)\)$`)
)

// themeVariablesDoc returns the "Variables" section of the theme reference.
// The variables come from the parser's allow-list and their defaults from the frontend's stylesheets, so the section can't drift from either.
func themeVariablesDoc() (string, error) {
	styles := &themeStyles{light: make(map[string]string), dark: make(map[string]string)}
	for _, path := range themeStylesheets {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		styles.add(string(data))
	}

	var doc strings.Builder
	doc.WriteString("## Variables\n\n")
	doc.WriteString("The variables below can be set in `light`, `dark` or both. Each one can be set under `variables` or directly in the block, next to `primary`. The two forms are equivalent and can be mixed; if a variable is set in both places, the value under `variables` is used.\n\n")
	doc.WriteString("Colors can be hex values, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or `hsl()` values. A variable that a theme does not set keeps the default listed here, and a variable set only in `light` keeps its default in dark mode.")

	for _, group := range parser.ThemeVariableGroups {
		doc.WriteString(fmt.Sprintf("\n\n### %s\n\n%s\n\n", group.Title, group.Description))

		if group.Variables[0].Deprecated {
			doc.WriteString("| Deprecated name | Sets instead |\n| --- | --- |")
			for _, v := range group.Variables {
				replacement := "Nothing: it no longer has an effect"
				if len(v.ReplacedBy) > 0 {
					replacement = "`" + strings.Join(v.ReplacedBy, "`, `") + "`"
				}
				doc.WriteString(fmt.Sprintf("\n| `%s` | %s |", v.Name, replacement))
			}
			continue
		}

		if group.Variables[0].Unused {
			for i, v := range group.Variables {
				if i > 0 {
					doc.WriteString(", ")
				}
				doc.WriteString("`" + v.Name + "`")
			}
			continue
		}

		// The chart palettes are described as a whole, so their tables have no description column.
		described := group.Variables[0].Description != ""
		if described {
			doc.WriteString("| Variable | What it styles | Light default | Dark default |\n| --- | --- | --- | --- |")
		} else {
			doc.WriteString("| Variable | Light default | Dark default |\n| --- | --- | --- |")
		}
		for _, v := range group.Variables {
			light, err := styles.defaultFor(v, false)
			if err != nil {
				return "", fmt.Errorf("light default of theme variable %q: %w", v.Name, err)
			}
			dark, err := styles.defaultFor(v, true)
			if err != nil {
				return "", fmt.Errorf("dark default of theme variable %q: %w", v.Name, err)
			}
			if described {
				doc.WriteString(fmt.Sprintf("\n| `%s` | %s | %s | %s |", v.Name, v.Description, light, dark))
			} else {
				doc.WriteString(fmt.Sprintf("\n| `%s` | %s | %s |", v.Name, light, dark))
			}
		}
	}

	return doc.String(), nil
}

// themeStyles holds the custom properties declared on `:root` (light) and `:root.dark` (dark).
type themeStyles struct {
	light map[string]string
	dark  map[string]string
}

// add collects the custom properties from the `:root` and `:root.dark` rules of a stylesheet.
func (s *themeStyles) add(css string) {
	css = cssCommentRegexp.ReplaceAllString(css, "")
	for _, rule := range cssRootRuleRegexp.FindAllStringSubmatch(css, -1) {
		target := s.light
		if rule[1] != "" {
			target = s.dark
		}
		for decl := range strings.SplitSeq(rule[2], ";") {
			name, value, ok := strings.Cut(decl, ":")
			name = strings.TrimSpace(name)
			if !ok || !strings.HasPrefix(name, "--") {
				continue
			}
			// Collapse the whitespace of values that are wrapped over several lines.
			value = strings.Join(strings.Fields(value), " ")
			value = strings.ReplaceAll(strings.ReplaceAll(value, "( ", "("), " )", ")")
			target[strings.TrimPrefix(name, "--")] = value
		}
	}
}

// defaultFor returns the default of a theme variable as Markdown.
func (s *themeStyles) defaultFor(v parser.ThemeVariable, dark bool) (string, error) {
	name := v.Name
	if name == "primary" || name == "secondary" {
		// The brand colors are not custom properties themselves: the default is the middle shade of Rill's own palette.
		name = fmt.Sprintf("color-%s-500", name)
	}

	value, ok := s.lookup(name, dark)
	if !ok {
		return "", errors.New("not declared in the stylesheets")
	}
	if v.Length {
		return "`" + value + "`", nil
	}

	c, shade, err := s.resolve(value, dark)
	if err != nil {
		return "", err
	}
	if shade != "" && name == v.Name {
		return shade, nil
	}

	res := fmt.Sprintf("`#%02x%02x%02x`", toByte(c.R), toByte(c.G), toByte(c.B))
	if c.A < 1 {
		res += fmt.Sprintf(" at %s%% opacity", strconv.FormatFloat(math.Round(c.A*1000)/10, 'f', -1, 64))
	}
	return res, nil
}

// lookup returns the declared value of a custom property in light or dark mode.
func (s *themeStyles) lookup(name string, dark bool) (string, bool) {
	if dark {
		if v, ok := s.dark[name]; ok {
			return v, true
		}
	}
	v, ok := s.light[name]
	return v, ok
}

// resolve evaluates a CSS color value, following var() references.
// If the value is derived from the palette of a brand color, it also returns a description of the shade.
func (s *themeStyles) resolve(value string, dark bool) (csscolorparser.Color, string, error) {
	if inner, ok := strings.CutPrefix(value, "var("); ok && strings.HasSuffix(value, ")") {
		name, fallback, hasFallback := strings.Cut(strings.TrimSuffix(inner, ")"), ",")
		name = strings.TrimPrefix(strings.TrimSpace(name), "--")

		next, ok := s.lookup(name, dark)
		if !ok {
			if !hasFallback {
				return csscolorparser.Color{}, "", fmt.Errorf("custom property %q is not declared", name)
			}
			next = strings.TrimSpace(fallback)
		}

		c, shade, err := s.resolve(next, dark)
		if err != nil {
			return csscolorparser.Color{}, "", err
		}
		if m := cssBrandShadeRegexp.FindStringSubmatch(name); m != nil && ok {
			brand := m[1]
			if brand == "theme" {
				// `color-theme-*` is an alias of the primary palette.
				brand = "primary"
			}
			shade = fmt.Sprintf("Shade %s of `%s`", m[2], brand)
		}
		return c, shade, nil
	}

	if m := cssColorMixRegexp.FindStringSubmatch(value); m != nil {
		c, shade, err := s.resolve(m[1], dark)
		if err != nil {
			return csscolorparser.Color{}, "", err
		}
		alpha, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			return csscolorparser.Color{}, "", err
		}
		c.A *= alpha
		return c, shade, nil
	}

	if m := cssOKLCHRegexp.FindStringSubmatch(value); m != nil {
		var lch [3]float64
		for i, raw := range []string{m[1], m[3], m[4]} {
			if raw == "none" {
				// Grays have no hue.
				continue
			}
			f, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return csscolorparser.Color{}, "", err
			}
			lch[i] = f
		}
		if m[2] == "%" {
			lch[0] /= 100
		}
		return oklchToSRGB(lch[0], lch[1], lch[2]), "", nil
	}

	c, err := csscolorparser.Parse(value)
	if err != nil {
		return csscolorparser.Color{}, "", fmt.Errorf("unsupported CSS value %q: %w", value, err)
	}
	return c, "", nil
}

// oklchToSRGB converts an OKLCH color to sRGB, clamping it to the sRGB gamut.
// See https://bottosson.github.io/posts/oklab/ for the conversion.
func oklchToSRGB(l, c, h float64) csscolorparser.Color {
	a := c * math.Cos(h*math.Pi/180)
	b := c * math.Sin(h*math.Pi/180)

	// Cone responses, which OKLab stores as cube roots.
	lms := [3]float64{
		l + 0.3963377774*a + 0.2158037573*b,
		l - 0.1055613458*a - 0.0638541728*b,
		l - 0.0894841775*a - 1.2914855480*b,
	}
	for i, v := range lms {
		lms[i] = v * v * v
	}

	rgb := [3]float64{
		4.0767416621*lms[0] - 3.3077115913*lms[1] + 0.2309699292*lms[2],
		-1.2684380046*lms[0] + 2.6097574011*lms[1] - 0.3413193965*lms[2],
		-0.0041960863*lms[0] - 0.7034186147*lms[1] + 1.7076147010*lms[2],
	}
	for i, v := range rgb {
		// Apply the sRGB transfer function to the linear value.
		if v > 0.0031308 {
			v = 1.055*math.Pow(v, 1/2.4) - 0.055
		} else {
			v *= 12.92
		}
		rgb[i] = math.Min(math.Max(v, 0), 1)
	}

	return csscolorparser.Color{R: rgb[0], G: rgb[1], B: rgb[2], A: 1}
}

func toByte(v float64) uint8 {
	return uint8(math.Round(v * 255))
}
