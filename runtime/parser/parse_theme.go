package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/mazznoer/csscolorparser"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
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
		if strings.TrimSpace(*tmp.CSS) == "" {
			return nil, fmt.Errorf("CSS cannot be empty")
		}

		sanitizedCSS, err := sanitizeCSS(*tmp.CSS)
		if err != nil {
			return nil, fmt.Errorf("invalid CSS syntax: %w", err)
		}
		spec.Css = sanitizedCSS
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

// Dangerous patterns to filter out that could lead to an XSS attack
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)javascript:`),
	regexp.MustCompile(`(?i)expression\s*\(`),
	regexp.MustCompile(`(?i)@import`),
	regexp.MustCompile(`(?i)document\.`),
	regexp.MustCompile(`(?i)eval\s*\(`),
	regexp.MustCompile(`(?i)vbscript:`),
	regexp.MustCompile(`(?i)data:\s*text/html`),
}

// Allowed URL schemes
var allowedSchemes = regexp.MustCompile(`^(https?|data:image|#)`)

func sanitizeCSS(c string) (string, error) {
	p := css.NewParser(parse.NewInput(bytes.NewBufferString(c)), false)
	out := ""

	for {
		gt, _, data := p.Next()
		dataStr := string(data)

		if gt == css.ErrorGrammar {
			break
		}

		out += dataStr

		switch gt {
		case css.CommentGrammar:
			// ignore comments
		case css.AtRuleGrammar, css.BeginAtRuleGrammar, css.QualifiedRuleGrammar, css.BeginRulesetGrammar, css.DeclarationGrammar, css.CustomPropertyGrammar:
			// Check if there is an import. We would need to check the files for xss, so we need to block it for now.
			if (gt == css.AtRuleGrammar || gt == css.BeginAtRuleGrammar) && strings.HasPrefix(strings.ToLower(strings.TrimSpace(dataStr)), "@import") {
				return "", fmt.Errorf("imports not allowed: %q", dataStr)
			}

			if gt == css.DeclarationGrammar || gt == css.CustomPropertyGrammar {
				out += ":"
			}

			for _, val := range p.Values() {
				valData := string(val.Data)

				// Check for dangerous patterns in values
				for _, pattern := range dangerousPatterns {
					if pattern.MatchString(valData) {
						return "", fmt.Errorf("disallowed css value: %q", valData)
					}
				}

				// Special handling for URL values
				if (val.TokenType == css.URLToken || val.TokenType == css.FunctionToken) && strings.HasPrefix(strings.ToLower(valData), "url(") {
					// Extract URL and validate
					url := strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(valData), ")"), "url(")
					// Remove quotes if present
					url = strings.Trim(url, `"'`)
					if url != "" && allowedSchemes.MatchString(url) {
						return "", fmt.Errorf("invalid URL: %q", url)
					}
				}

				out += valData
			}

			switch gt {
			case css.BeginAtRuleGrammar, css.BeginRulesetGrammar:
				out += "{"
			case css.AtRuleGrammar, css.DeclarationGrammar, css.CustomPropertyGrammar:
				out += ";"
			case css.QualifiedRuleGrammar:
				out += ","
			default:
			}
		case css.EndAtRuleGrammar, css.EndRulesetGrammar:
			if strings.TrimSpace(dataStr) != "}" {
				return "", fmt.Errorf("unbalanced braces")
			}
		default:
		}
	}

	if p.Err() != nil && !errors.Is(p.Err(), io.EOF) {
		return "", p.Err()
	}

	return out, nil
}
