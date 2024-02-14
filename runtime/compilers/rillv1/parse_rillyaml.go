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
	Title       string            `yaml:"title"`
	Description string            `yaml:"description"`
	Vars        map[string]string `yaml:"vars"`
	Connectors  []struct {
		Type     string            `yaml:"type"`
		Name     string            `yaml:"name"`
		Defaults map[string]string `yaml:"defaults"`
	} `yaml:"connectors"`
	Env        map[string]yaml.Node `yaml:"env"`
	Sources    yaml.Node            `yaml:"sources"`
	Models     yaml.Node            `yaml:"models"`
	Dashboards yaml.Node            `yaml:"dashboards"`
	Migrations yaml.Node            `yaml:"migrations"`
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

	// Ugly backwards compatibility hack: we have renamed "env" to "vars", and now use "env" for environment-specific overrides.
	// For backwards compatibility, we still consider "env" entries with scalar values as variables.
	for k := range tmp.Env {
		v := tmp.Env[k]
		if v.Kind == yaml.ScalarNode {
			if tmp.Vars == nil {
				tmp.Vars = make(map[string]string)
			}
			tmp.Vars[k] = v.Value
			delete(tmp.Env, k)
		}
	}

	// Apply environment-specific overrides
	if envOverride := tmp.Env[p.Environment]; !envOverride.IsZero() {
		if err := envOverride.Decode(tmp); err != nil {
			return newYAMLError(err)
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
		Variables:   make([]*VariableDef, len(tmp.Vars)),
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
	for k, v := range tmp.Vars {
		res.Variables[i] = &VariableDef{
			Name:    k,
			Default: v,
		}
		i++
	}

	p.RillYAML = res
	return nil
}
