package rillv1

import (
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

	if tmp.Colors.Primary != "" {
		pc, err := csscolorparser.Parse(tmp.Colors.Primary)
		if err != nil {
			return nil, err
		}
		spec.PrimaryColor = toThemeColor(pc)
		spec.PrimaryColorRaw = tmp.Colors.Primary
	}

	if tmp.Colors.Secondary != "" {
		sc, err := csscolorparser.Parse(tmp.Colors.Secondary)
		if err != nil {
			return nil, err
		}
		spec.SecondaryColor = toThemeColor(sc)
		spec.SecondaryColorRaw = tmp.Colors.Secondary
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
