package parser

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// ConnectorYAML is the raw structure of a Connector resource defined in YAML (does not include common fields)
type ConnectorYAML struct {
	commonYAML `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into Properties
	// Driver name
	Driver     string         `yaml:"driver"`
	Managed    yaml.Node      `yaml:"managed"` // Boolean or map of properties
	Properties map[string]any `yaml:",inline" mapstructure:",remain"`
}

// parseConnector parses a connector definition and adds the resulting resource to p.Resources.
func (p *Parser) parseConnector(node *Node) error {
	// Parse YAML
	tmp := &ConnectorYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
	}

	// "Managed" indicates that we should automatically provision the connector
	var provision bool
	var provisionArgsPB *structpb.Struct
	if !tmp.Managed.IsZero() {
		switch tmp.Managed.Kind {
		case yaml.ScalarNode:
			err := tmp.Managed.Decode(&provision)
			if err != nil {
				return fmt.Errorf("failed to decode 'managed': %w", err)
			}
		case yaml.MappingNode:
			provision = true
			var provisionArgs map[string]any
			err := tmp.Managed.Decode(&provisionArgs)
			if err != nil {
				return fmt.Errorf("failed to decode 'managed': %w", err)
			}
			provisionArgsPB, err = structpb.NewStruct(provisionArgs)
			if err != nil {
				return fmt.Errorf("failed to convert provision args to proto: %w", err)
			}
		default:
			return fmt.Errorf("invalid type for 'managed': expected boolean or map of args")
		}
	}

	// Find out if any properties are templated
	templatedProps, err := analyzeTemplatedProperties(tmp.Properties)
	if err != nil {
		return fmt.Errorf("failed to analyze templated properties: %w", err)
	}
	var propertiesPB *structpb.Struct
	if tmp.Properties != nil {
		propertiesPB, err = structpb.NewStruct(tmp.Properties)
		if err != nil {
			return fmt.Errorf("failed to convert connector properties to proto: %w", err)
		}
	}

	// Insert the connector
	r, err := p.insertResource(ResourceKindConnector, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.ConnectorSpec.Driver = tmp.Driver
	r.ConnectorSpec.Properties = propertiesPB
	r.ConnectorSpec.TemplatedProperties = templatedProps
	r.ConnectorSpec.Provision = provision
	r.ConnectorSpec.ProvisionArgs = provisionArgsPB

	return nil
}

// analyzeTemplatedProperties returns a slice of map keys that have a value which contains templating tags.
func analyzeTemplatedProperties(m map[string]any) ([]string, error) {
	var res []string
	for k, v := range m {
		hasTemplate, err := containsTemplating(v)
		if err != nil {
			return nil, err
		}
		if hasTemplate {
			res = append(res, k)
		}
	}
	return res, nil
}

// containsTemplating recursively checks whether a value (string/map/array) uses templating.
func containsTemplating(v any) (bool, error) {
	switch val := v.(type) {
	case string:
		meta, err := AnalyzeTemplate(val)
		if err != nil {
			return false, err
		}
		return meta.UsesTemplating, nil
	case map[string]any:
		for _, sub := range val {
			ok, err := containsTemplating(sub)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	case []any:
		for _, item := range val {
			ok, err := containsTemplating(item)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	}
	return false, nil
}
