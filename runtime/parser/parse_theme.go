package parser

import (
	"fmt"
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
	CSS *string `yaml:"css"`
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
	hasCSS := tmp.CSS != nil

	if hasLegacyColors && hasCSS {
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
	}

	if hasCSS {
		if err := validateCSS(*tmp.CSS); err != nil {
			return nil, fmt.Errorf("invalid CSS syntax: %w", err)
		}
		spec.Css = *tmp.CSS
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

// validateCSS performs basic CSS syntax validation
func validateCSS(css string) error {
	if strings.TrimSpace(css) == "" {
		return fmt.Errorf("CSS cannot be empty")
	}

	// Basic validation: check for balanced braces and semicolons
	// This is a simple validation - can be enhanced with a proper CSS parser later
	braceCount := 0
	parenCount := 0
	bracketCount := 0

	// Remove comments for validation
	css = removeCSSComments(css)

	for i, char := range css {
		switch char {
		case '{':
			braceCount++
		case '}':
			braceCount--
			if braceCount < 0 {
				return fmt.Errorf("unexpected closing brace '}' at position %d", i)
			}
		case '(':
			parenCount++
		case ')':
			parenCount--
			if parenCount < 0 {
				return fmt.Errorf("unexpected closing parenthesis ')' at position %d", i)
			}
		case '[':
			bracketCount++
		case ']':
			bracketCount--
			if bracketCount < 0 {
				return fmt.Errorf("unexpected closing bracket ']' at position %d", i)
			}
		}
	}

	if braceCount != 0 {
		return fmt.Errorf("unbalanced braces: %d unclosed brace(s)", braceCount)
	}
	if parenCount != 0 {
		return fmt.Errorf("unbalanced parentheses: %d unclosed parenthesis(es)", parenCount)
	}
	if bracketCount != 0 {
		return fmt.Errorf("unbalanced brackets: %d unclosed bracket(s)", bracketCount)
	}

	// Check for basic CSS structure (selector { property: value; })
	// This regex looks for at least one CSS rule
	cssRuleRegex := regexp.MustCompile(`[^{}]*\{[^{}]*:[^{}]*[;}]`)
	if !cssRuleRegex.MatchString(css) {
		return fmt.Errorf("CSS must contain at least one valid rule (selector { property: value; })")
	}

	return nil
}

// removeCSSComments removes CSS comments from the string
func removeCSSComments(css string) string {
	// Remove /* */ comments
	commentRegex := regexp.MustCompile(`/\*.*?\*/`)
	return commentRegex.ReplaceAllString(css, "")
}
