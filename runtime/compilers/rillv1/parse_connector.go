package rillv1

import (
	"fmt"
)

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

	templatedProps, err := analyzeTemplatedProperties(tmp.Defaults)
	if err != nil {
		return fmt.Errorf("failed to analyze templated properties: %w", err)
	}

	// Insert the connector
	r, err := p.insertResource(ResourceKindConnector, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.ConnectorSpec.Driver = tmp.Driver
	r.ConnectorSpec.Properties = tmp.Defaults
	r.ConnectorSpec.TemplatedProperties = templatedProps
	return nil
}

// analyzeTemplatedProperties returns a slice of map keys that have a value which contains templating tags.
func analyzeTemplatedProperties(m map[string]string) ([]string, error) {
	var res []string
	for k, v := range m {
		meta, err := AnalyzeTemplate(v)
		if err != nil {
			return nil, err
		}
		if !meta.UsesTemplating {
			continue
		}
		res = append(res, k)
	}
	return res, nil
}
