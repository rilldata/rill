package rillv1

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gopkg.in/yaml.v3"
)

type DashboardYAML struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string           `yaml:"title"`
	Columns    uint32           `yaml:"columns"`
	Gap        uint32           `yaml:"gap"`
	Items      []*struct {
		Component yaml.Node `yaml:"component"` // Can be a name (string) or inline component definition (map)
		X         *uint32   `yaml:"x"`
		Y         *uint32   `yaml:"y"`
		Width     *uint32   `yaml:"width"`
		Height    *uint32   `yaml:"height"`
		FontSize  uint32    `yaml:"font_size"`
	} `yaml:"items"`
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

	// Ensure there's at least one item
	if len(tmp.Items) == 0 {
		return errors.New(`at least one item must be configured`)
	}

	// Parse items.
	// Each item can either reference an externally defined component by name or define a component inline.
	items := make([]*runtimev1.DashboardItem, len(tmp.Items))
	var inlineComponentDefs []*componentDef
	for i, item := range tmp.Items {
		if item == nil {
			return fmt.Errorf("item at index %d is nil", i)
		}

		component, inlineComponentDef, err := p.parseDashboardItemComponent(node.Name, i, item.Component)
		if err != nil {
			return fmt.Errorf("invalid component at index %d: %w", i, err)
		}

		// Track inline component definitions so we can insert them after we have validated all components
		if inlineComponentDef != nil {
			inlineComponentDefs = append(inlineComponentDefs, inlineComponentDef)
		}

		items[i] = &runtimev1.DashboardItem{
			Component: component,
			X:         item.X,
			Y:         item.Y,
			Width:     item.Width,
			Height:    item.Height,
			FontSize:  item.FontSize,
		}

		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindComponent, Name: component})
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
	r.DashboardSpec.Items = items

	// Track inline components
	for _, def := range inlineComponentDefs {
		r, err := p.insertResource(ResourceKindComponent, def.name, node.Paths, def.refs...)
		if err != nil {
			// Normally we could return the error, but we can't do that here because we've already inserted the dashboard.
			// Since the component has been validated with insertDryRun in parseDashboardItemComponent, this error should never happen in practice.
			// So let's panic.
			panic(err)
		}
		r.ComponentSpec = def.spec
	}

	return nil
}

// parseDashboardItemComponent parses a dashboard item's "component" property.
// It may be a string (name of an externally defined component) or an inline component definition.
func (p *Parser) parseDashboardItemComponent(dashboardName string, idx int, n yaml.Node) (string, *componentDef, error) {
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

	spec.DefinedInDashboard = true

	name := fmt.Sprintf("%s--component-%d", dashboardName, idx)

	err = p.insertDryRun(ResourceKindComponent, name)
	if err != nil {
		name = fmt.Sprintf("%s--component-%d-%s", dashboardName, idx, uuid.New())
		err = p.insertDryRun(ResourceKindComponent, name)
		if err != nil {
			return "", nil, err
		}
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
