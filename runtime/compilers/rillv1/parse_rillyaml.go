package rillv1

import (
	"context"
	"fmt"
	"maps"

	"gopkg.in/yaml.v3"
)

var _reservedConnectorNames = map[string]bool{"admin": true, "repo": true, "metastore": true}

// RillYAML is the parsed contents of rill.yaml
type RillYAML struct {
	Title         string
	Description   string
	OLAPConnector string
	Connectors    []*ConnectorDef
	Variables     []*VariableDef
	Defaults      map[ResourceKind]yaml.Node
	FeatureFlags  map[string]bool

	// if adding any new field evaluate if equalRillYAML function needs changes
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
	// Title of the project
	Title string `yaml:"title"`
	// Description of the project
	Description string `yaml:"description"`
	// The project's default OLAP connector to use (can be overridden in the individual resources)
	OLAPConnector string `yaml:"olap_connector"`
	// Connectors required by the project
	Connectors []struct {
		Type     string            `yaml:"type"`
		Name     string            `yaml:"name"`
		Defaults map[string]string `yaml:"defaults"`
	} `yaml:"connectors"`
	// Variables required by the project and their default values
	Vars map[string]string `yaml:"vars"`
	// Environment-specific overrides for rill.yaml
	Env map[string]yaml.Node `yaml:"env"`
	// Shorthand for setting "env:dev:"
	Dev yaml.Node `yaml:"dev"`
	// Shorthand for setting "env:prod:"
	Prod yaml.Node `yaml:"prod"`
	// Default YAML values for sources
	Sources yaml.Node `yaml:"sources"`
	// Default YAML values for models
	Models yaml.Node `yaml:"models"`
	// Default YAML values for metric views
	Dashboards yaml.Node `yaml:"dashboards"`
	// Default YAML values for migrations
	Migrations yaml.Node `yaml:"migrations"`
	// Feature flags
	Features []string `yaml:"features"`
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

	// Handle "dev:" and "prod:" shorthands (copy to to tmp.Env)
	if !tmp.Dev.IsZero() {
		if tmp.Env == nil {
			tmp.Env = make(map[string]yaml.Node)
		}
		tmp.Env["dev"] = tmp.Dev
	}
	if !tmp.Prod.IsZero() {
		if tmp.Env == nil {
			tmp.Env = make(map[string]yaml.Node)
		}
		tmp.Env["prod"] = tmp.Prod
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

	featureFlags := map[string]bool{}
	for _, f := range tmp.Features {
		featureFlags[f] = true
	}

	res := &RillYAML{
		Title:         tmp.Title,
		Description:   tmp.Description,
		OLAPConnector: tmp.OLAPConnector,
		Connectors:    make([]*ConnectorDef, len(tmp.Connectors)),
		Variables:     make([]*VariableDef, len(tmp.Vars)),
		Defaults: map[ResourceKind]yaml.Node{
			ResourceKindSource:      tmp.Sources,
			ResourceKindModel:       tmp.Models,
			ResourceKindMetricsView: tmp.Dashboards,
			ResourceKindMigration:   tmp.Migrations,
		},
		FeatureFlags: featureFlags,
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

func equalRillYAML(r, other *RillYAML) bool {
	if r.OLAPConnector != other.OLAPConnector || len(r.Connectors) != len(other.Connectors) || len(r.Variables) != len(other.Variables) || len(r.Defaults) != len(other.Defaults) {
		return false
	}
	for _, connector := range r.Connectors {
		equal := false
		for _, otherConnector := range other.Connectors {
			if connector.Name == otherConnector.Name && connector.Type == otherConnector.Type && maps.Equal(connector.Defaults, otherConnector.Defaults) {
				equal = true
				break
			}
		}
		if !equal {
			return false
		}
	}

	for _, variable := range r.Variables {
		equal := false
		for _, otherVariable := range other.Variables {
			if variable.Name == otherVariable.Name && variable.Default == otherVariable.Default {
				equal = true
				break
			}
		}
		if !equal {
			return false
		}
	}

	if !maps.Equal(r.FeatureFlags, other.FeatureFlags) {
		return false
	}

	// keeping defaults comparison at last to avoid yaml marshalling
	for key1 := range r.Defaults {
		node1 := r.Defaults[key1]
		node2, ok := r.Defaults[key1]
		if !ok {
			return false
		}
		if !compareYAMLNodes(&node1, &node2) {
			return false
		}
	}
	return true
}

func compareYAMLNodes(node1, node2 *yaml.Node) bool {
	yamlStr1, err1 := marshalYAMLNode(node1)
	if err1 != nil {
		return false
	}

	yamlStr2, err2 := marshalYAMLNode(node2)
	if err2 != nil {
		return false
	}

	return yamlStr1 == yamlStr2
}

func marshalYAMLNode(node *yaml.Node) (string, error) {
	out, err := yaml.Marshal(node)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
