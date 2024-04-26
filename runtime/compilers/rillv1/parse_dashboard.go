package rillv1

import (
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gopkg.in/yaml.v3"
)

type DashboardYAML struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string           `yaml:"title"`
	Columns    uint32           `yaml:"columns"`
	Gap        uint32           `yaml:"gap"`
	Components []*struct {
		Component yaml.Node `yaml:"component"` // Can be a name (string) or inline component definition (map)
		X         uint32    `yaml:"x"`
		Y         uint32    `yaml:"y"`
		Width     uint32    `yaml:"width"`
		Height    uint32    `yaml:"height"`
		FontSize  uint32    `yaml:"font_size"`
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
		return fmt.Errorf("dashboards cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("dashboards cannot have a connector")
	}

	// Ensure there's at least one component
	if len(tmp.Components) == 0 {
		return errors.New(`at least one component must be provided`)
	}

	// Parse components.
	// Components may either reference an externally defined component by name or be defined inline.
	var inlineDefs []*componentDef
	dashboardComponents := make([]*runtimev1.DashboardComponent, len(tmp.Components))
	for i, c := range tmp.Components {
		name, inlineDef, err := p.parseDashboardComponent(node.Name, i, c.Component)
		if err != nil {
			return fmt.Errorf("invalid component at index %d: %w", i, err)
		}

		// Track inline component definitions so we can insert them after we have validated all components
		if inlineDef != nil {
			inlineDefs = append(inlineDefs, inlineDef)
		}

		dashboardComponents[i] = &runtimev1.DashboardComponent{
			Component: name,
			X:         c.X,
			Y:         c.Y,
			Width:     c.Width,
			Height:    c.Height,
			FontSize:  c.FontSize,
		}

		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindComponent, Name: name})
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
	r.DashboardSpec.Components = dashboardComponents

	// Track inline components
	for _, def := range inlineDefs {
		r, err := p.insertResource(ResourceKindComponent, def.name, node.Paths, def.refs...)
		if err != nil {
			// Normally we could return the error, but we can't do that here because we've already inserted the dashboard.
			// Since the component has been validated with insertDryRun in parseDashboardComponent, this error should never happen in practice.
			// So let's panic.
			panic(err)
		}
		r.ComponentSpec = def.spec
	}

	return nil
}

// parseDashboardComponent parses a dashboard component's "component" property.
// It may be a string (name of an externally defined component) or an inline component definition.
func (p *Parser) parseDashboardComponent(dashboardName string, idx int, n yaml.Node) (string, *componentDef, error) {
	if n.Kind == yaml.ScalarNode {
		var name string
		err := n.Decode(&name)
		if err != nil {
			return "", nil, err
		}
		return name, nil, nil
	}

	if n.Kind != yaml.MappingNode {
		return "", nil, errors.New("expected a component name or inline declaration")
	}

	tmp := &ComponentYAML{}
	err := n.Decode(tmp)
	if err != nil {
		return "", nil, err
	}

	spec, refs, err := p.parseComponentYAML(tmp)
	if err != nil {
		return "", nil, err
	}

	name := fmt.Sprintf("%s--component%d", dashboardName, idx)

	err = p.insertDryRun(ResourceKindComponent, name)
	if err != nil {
		return "", nil, err
	}

	def := &componentDef{
		name: name,
		refs: refs,
		spec: spec,
	}

	return name, def, nil
}

type componentDef struct {
	name string
	refs []ResourceName
	spec *runtimev1.ComponentSpec
}
