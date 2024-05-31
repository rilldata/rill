package rillv1

import "strings"

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

	r.ConnectorSpec.Driver = tmp.Driver
	r.ConnectorSpec.Properties = tmp.Defaults
	r.ConnectorSpec.PropertiesFromVariables, err = propertiesFromVariables(tmp.Defaults)
	if err != nil {
		return err
	}
	return nil
}

func propertiesFromVariables(props map[string]string) (map[string]string, error) {
	res := make(map[string]string)
	for key, val := range props {
		meta, err := AnalyzeTemplate(val)
		if err != nil {
			return nil, err
		}
		// assume that only one variable will be set
		for _, k := range meta.Variables {
			if after, found := strings.CutPrefix(k, "vars."); found {
				res[key] = after
			}
		}
	}
	return res, nil
}
