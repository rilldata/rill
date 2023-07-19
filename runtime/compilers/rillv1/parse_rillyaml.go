package rillv1

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
)

// RillYAML is the parsed contents of rill.yaml
type RillYAML struct {
	Title       string
	Description string
	Connectors  []*ConnectorDef
	Variables   []*VariableDef
}

// ConnectorDef is a subtype of RillYAML, defining connectors required by the project
type ConnectorDef struct {
	Type     string
	Name     string
	Defaults map[string]string
}

// VariableDef is a subtype of RillYAML, defining defaults for project variables
type VariableDef struct {
	Name    string
	Default string
}

// rillYAML is the raw YAML structure of rill.yaml
type rillYAML struct {
	Title       string            `yaml:"title"`
	Description string            `yaml:"description"`
	Env         map[string]string `yaml:"env"`
	Connectors  []struct {
		Type     string            `yaml:"type"`
		Name     string            `yaml:"name"`
		Defaults map[string]string `yaml:"defaults"`
	} `yaml:"connectors"`
}

// parseRillYAML parses rill.yaml
func (p *Parser) parseRillYAML(ctx context.Context, path string) error {
	data, err := p.Repo.Get(ctx, p.InstanceID, path)
	if err != nil {
		return fmt.Errorf("error loading %q: %w", path, err)
	}

	tmp := &rillYAML{}
	if err := yaml.Unmarshal([]byte(data), tmp); err != nil {
		return newYAMLError(err)
	}

	res := &RillYAML{
		Title:       tmp.Title,
		Description: tmp.Description,
		Connectors:  make([]*ConnectorDef, len(tmp.Connectors)),
		Variables:   make([]*VariableDef, len(tmp.Env)),
	}

	for i, c := range tmp.Connectors {
		res.Connectors[i] = &ConnectorDef{
			Type:     c.Type,
			Name:     c.Name,
			Defaults: c.Defaults,
		}
	}

	i := 0
	for k, v := range tmp.Env {
		res.Variables[i] = &VariableDef{
			Name:    k,
			Default: v,
		}
		i++
	}

	p.RillYAML = res
	return nil
}
