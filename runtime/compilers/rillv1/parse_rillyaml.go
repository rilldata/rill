package rillv1

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
)

var _reservedConnectorNames = map[string]bool{"admin": true, "repo": true, "metastore": true}

// RillYAML is the parsed contents of rill.yaml
type RillYAML struct {
	Title       string
	Description string
	Connectors  []*ConnectorDef
	Variables   []*VariableDef
	Defaults    map[ResourceKind]yaml.Node
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
	Title               string            `yaml:"title"`
	Description         string            `yaml:"description"`
	Variables           map[string]string `yaml:"variables"`
	VariablesDeprecated map[string]string `yaml:"env"`
	Connectors          []struct {
		Type     string            `yaml:"type"`
		Name     string            `yaml:"name"`
		Defaults map[string]string `yaml:"defaults"`
	} `yaml:"connectors"`
	Sources     yaml.Node            `yaml:"sources"`
	Models      yaml.Node            `yaml:"models"`
	Dashboards  yaml.Node            `yaml:"dashboards"`
	Migrations  yaml.Node            `yaml:"migrations"`
	Environment map[string]yaml.Node `yaml:"environment"`
}

// parseRillYAML parses rill.yaml
func (p *Parser) parseRillYAML(ctx context.Context, path string) error {
	data, err := p.Repo.Get(ctx, path)
	if err != nil {
		return fmt.Errorf("error loading %q: %w", path, err)
	}

	tmp := &rillYAML{}
	if err := yaml.Unmarshal([]byte(data), tmp); err != nil {
		return newYAMLError(err)
	}

	// Apply environment-specific overrides
	envNode := tmp.Environment[p.Environment]
	if !envNode.IsZero() {
		if err := envNode.Decode(tmp); err != nil {
			return newYAMLError(err)
		}
	}

	// Backwards compatibility for "env" -> "variables"
	if len(tmp.VariablesDeprecated) > 0 {
		if tmp.Variables == nil {
			tmp.Variables = make(map[string]string, len(tmp.VariablesDeprecated))
		}
		for k, v := range tmp.VariablesDeprecated {
			tmp.Variables[k] = v
		}
	}

	// Validate resource defaults
	if !tmp.Sources.IsZero() {
		if err := tmp.Sources.Decode(&SourceYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.Models.IsZero() {
		if err := tmp.Models.Decode(&ModelYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.Dashboards.IsZero() {
		if err := tmp.Dashboards.Decode(&MetricsViewYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.Migrations.IsZero() {
		if err := tmp.Migrations.Decode(&MigrationYAML{}); err != nil {
			return newYAMLError(err)
		}
	}

	res := &RillYAML{
		Title:       tmp.Title,
		Description: tmp.Description,
		Connectors:  make([]*ConnectorDef, len(tmp.Connectors)),
		Variables:   make([]*VariableDef, len(tmp.Variables)),
		Defaults: map[ResourceKind]yaml.Node{
			ResourceKindSource:      tmp.Sources,
			ResourceKindModel:       tmp.Models,
			ResourceKindMetricsView: tmp.Dashboards,
			ResourceKindMigration:   tmp.Migrations,
		},
	}

	for i, c := range tmp.Connectors {
		if _reservedConnectorNames[c.Name] {
			return fmt.Errorf("connector name %q is reserved", c.Name)
		}
		res.Connectors[i] = &ConnectorDef{
			Type:     c.Type,
			Name:     c.Name,
			Defaults: c.Defaults,
		}
	}

	i := 0
	for k, v := range tmp.Variables {
		res.Variables[i] = &VariableDef{
			Name:    k,
			Default: v,
		}
		i++
	}

	p.RillYAML = res
	return nil
}
