package rillv1

import (
	"fmt"
	"regexp"
	"strings"
)

// envVarRegex matches a variable reference in the form {{   .vars.variable   }}
var envVarRegex = regexp.MustCompile(`^\{\{\s*\.vars\.\w+(?:\.\w+)*\s*\}\}$`)

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

	props, propsFromVariables, err := propertiesFromVariables(tmp.Defaults)
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
	r.ConnectorSpec.Properties = props
	r.ConnectorSpec.PropertiesFromVariables = propsFromVariables
	return nil
}

func propertiesFromVariables(in map[string]string) (map[string]string, map[string]string, error) {
	props := make(map[string]string)
	propsFromVars := make(map[string]string)
	for key, val := range in {
		meta, err := AnalyzeTemplate(val)
		if err != nil {
			return nil, nil, err
		}
		if len(meta.Variables) == 0 { // property does not use any variables
			props[key] = val
			continue
		}
		// property uses variables
		if len(meta.Variables) > 1 {
			return nil, nil, fmt.Errorf("connector property should contain atmost one variable. Property %q contains %q", key, len(meta.Variables))
		}
		if !envVarRegex.MatchString(val) {
			return nil, nil, fmt.Errorf("invalid property %q. When accessing a variable in a connector property, the value should match the form `{{ .vars.variable }}`", key)
		}

		after, _ := strings.CutPrefix(meta.Variables[0], "vars.")
		propsFromVars[key] = after
	}
	return props, propsFromVars, nil
}
