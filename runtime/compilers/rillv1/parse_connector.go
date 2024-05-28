package rillv1

// ConnectorYAML is the raw structure of a Connector resource defined in YAML (does not include common fields)
type ConnectorYAML struct {
	commonYAML `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into Properties
	// Driver name
	Driver   string            `yaml:"driver"`
	Defaults map[string]string `yaml:",inline" mapstructure:",remain"`
}

// parseConnector parses a connector definition and adds the resulting resource to p.Resources.
func (p *Parser) parseConnector(node *Node) error {
	// Parse YAML
	tmp := &ConnectorYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
	}

	// Insert the connector
	r, err := p.insertResource(ResourceKindConnector, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	name := tmp.Name
	if name == "" {
		name = tmp.Driver
	}

	r.ConnectorSpec.Driver = tmp.Driver
	r.ConnectorSpec.Name = name
	r.ConnectorSpec.Properties = tmp.Defaults
	return nil
}
