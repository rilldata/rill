package rillv1

import (
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type DashboardYAML struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string           `yaml:"title"`
	Columns    uint32           `yaml:"columns"`
	Gap        uint32           `yaml:"gap"`
	Components []struct {
		Chart  string `yaml:"chart"`
		X      uint32 `yaml:"x"`
		Y      uint32 `yaml:"y"`
		Width  uint32 `yaml:"width"`
		Height uint32 `yaml:"height"`
	} `yaml:"components"`
}

func (p *Parser) parseDashboard(node *Node) error {
	// Parse YAML
	tmp := &DashboardYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("charts cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("charts cannot have a connector")
	}

	if len(tmp.Components) == 0 {
		return errors.New(`at least one component must be provided`)
	}

	for _, component := range tmp.Components {
		if component.Chart == "" {
			return errors.New(`chart is mandatory for a component`)
		}
		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindChart, Name: component.Chart})
	}

	// Track dashboard
	r, err := p.insertResource(ResourceKindDashboard, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.DashboardSpec.Title = tmp.Title
	r.DashboardSpec.Columns = tmp.Columns
	r.DashboardSpec.Gap = tmp.Gap

	r.DashboardSpec.Components = make([]*runtimev1.DashboardComponent, len(tmp.Components))
	for i, component := range tmp.Components {
		r.DashboardSpec.Components[i] = &runtimev1.DashboardComponent{
			Chart:  component.Chart,
			X:      component.X,
			Y:      component.Y,
			Width:  component.Width,
			Height: component.Height,
		}
	}

	return nil
}
