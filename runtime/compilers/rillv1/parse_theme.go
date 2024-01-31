package rillv1

import (
	"context"

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
func (p *Parser) parseTheme(ctx context.Context, node *Node) error {
	tmp := &ThemeYAML{}
	// TODO: get from defaults
	if node.YAML != nil {
		if err := node.YAML.Decode(tmp); err != nil {
			return pathError{path: node.YAMLPath, err: newYAMLError(err)}
		}
	}

	// Parse the colors now to get the parse error before inserting resource
	var pc csscolorparser.Color
	hasPc := false
	var sc csscolorparser.Color
	hasSc := false
	var err error
	if tmp.Colors.Primary != "" {
		pc, err = csscolorparser.Parse(tmp.Colors.Primary)
		if err != nil {
			return err
		}
		hasPc = true
	}
	if tmp.Colors.Secondary != "" {
		sc, err = csscolorparser.Parse(tmp.Colors.Secondary)
		if err != nil {
			return err
		}
		hasSc = true
	}

	r, err := p.insertResource(ResourceKindTheme, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}

	if hasPc {
		r.ThemeSpec.PrimaryColor = toThemeColor(pc)
	}
	if hasSc {
		r.ThemeSpec.SecondaryColor = toThemeColor(sc)
	}

	return nil
}

func toThemeColor(c csscolorparser.Color) *runtimev1.Color {
	return &runtimev1.Color{
		Red:   float32(c.R),
		Green: float32(c.G),
		Blue:  float32(c.B),
		Alpha: float32(c.A),
	}
}
