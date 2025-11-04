package parser

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/rilldata/rill/runtime/pkg/env"
	"gopkg.in/yaml.v3"
)

var _reservedConnectorNames = map[string]bool{"admin": true, "repo": true, "metastore": true}

var ErrRillYAMLNotFound = errors.New("rill.yaml not found")

// RillYAML is the parsed contents of rill.yaml
type RillYAML struct {
	DisplayName    string
	Description    string
	AIInstructions string
	OLAPConnector  string
	AIConnector    string
	Connectors     []*ConnectorDef
	Variables      []*VariableDef
	Defaults       map[ResourceKind]yaml.Node
	FeatureFlags   map[string]string
	PublicPaths    []string
}

// ConnectorDef is a subtype of RillYAML, defining connectors required by the project
type ConnectorDef struct {
	Type     string
	Name     string
	Defaults map[string]any
}

// VariableDef is a subtype of RillYAML, defining defaults for project variables
type VariableDef struct {
	Name    string
	Default string
}

// rillYAML is the raw YAML structure of rill.yaml
type rillYAML struct {
	// Compiler is the parser version to use. It is not consumed here because at this point a parser has already been chosen.
	Compiler string `yaml:"compiler"`
	// RillVersion is deprecated and not used anymore.
	RillVersion any `yaml:"rill_version"`
	// Title of the project
	DisplayName string `yaml:"display_name"`
	// Title of the project
	Title string `yaml:"title"` // Deprecated: use display_name
	// Title of the project
	Name string `yaml:"name"` // Deprecated: use display_name
	// Description of the project
	Description string `yaml:"description"`
	// User-provided context for LLM/AI features
	AIInstructions string `yaml:"ai_instructions"`
	// Connector to use for the AI service
	AIConnector string `yaml:"ai_connector"`
	// The project's default OLAP connector to use (can be overridden in the individual resources)
	OLAPConnector string `yaml:"olap_connector"`
	// Connectors required by the project
	Connectors []struct {
		Type     string         `yaml:"type"`
		Name     string         `yaml:"name"`
		Defaults map[string]any `yaml:"defaults"`
	} `yaml:"connectors"`
	// Default values for project variables.
	// For backwards compatibility, "dev" and "prod" keys with nested values will populate "environment_overrides".
	Env map[string]yaml.Node `yaml:"env"`
	// Deprecated: Use "env" instead.
	Vars map[string]string `yaml:"vars"`
	// Environment-specific overrides for rill.yaml
	EnvironmentOverrides map[string]yaml.Node `yaml:"environment_overrides"`
	// Shorthand for setting "environment:dev:"
	Dev yaml.Node `yaml:"dev"`
	// Shorthand for setting "environment:prod:"
	Prod yaml.Node `yaml:"prod"`
	// Default YAML values for sources
	Sources yaml.Node `yaml:"sources"`
	// Default YAML values for models
	Models yaml.Node `yaml:"models"`
	// Default YAML values for metric views
	MetricsViews yaml.Node `yaml:"metrics_views"`
	// Default YAML values for metric views.
	// Deprecated: Use "metrics_views" instead
	MetricsViewsLegacy yaml.Node `yaml:"dashboards"`
	// Default YAML values for explores
	Explores yaml.Node `yaml:"explores"`
	// Default YAML values for canvases
	Canvases yaml.Node `yaml:"canvases"`
	// Default YAML values for custom apis
	APIs yaml.Node `yaml:"apis"`
	// Default YAML values for migrations
	Migrations yaml.Node `yaml:"migrations"`
	// Feature flags (preferably a map[string]bool, but can also be a []string for backwards compatibility)
	Features yaml.Node `yaml:"features"`
	// Paths to expose over HTTP (defaults to ./public)
	PublicPaths []string `yaml:"public_paths"`
	// Paths to ignore when watching for changes.
	// This is ignored in this parser because it's consumed directly by the repo driver.
	IgnorePaths []string `yaml:"ignore_paths"`
	// A list of mock users to test against dashboard security policies.
	// This is ignored in this parser because it's consumed directly by the local application.
	MockUsers []struct {
		Email      string         `yaml:"email"`
		Name       string         `yaml:"name"`
		Admin      bool           `yaml:"admin"`
		Groups     []string       `yaml:"groups"`
		Attributes map[string]any `yaml:",inline"`
	} `yaml:"mock_users"`
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

	dec := yaml.NewDecoder(strings.NewReader(data))
	dec.KnownFields(true)

	err = dec.Decode(tmp)
	if err != nil && !errors.Is(err, io.EOF) {
		return newYAMLError(err)
	}

	// Look for environment-specific overrides
	for k, v := range tmp.Env { // nolint: gocritic // Using a pointer changes parser behavior
		if v.Kind == yaml.ScalarNode {
			continue
		}
		// Backwards compatibility hack: we renamed "env" to "environment_overrides".
		// The only environments supported at the rename time were "dev" and "prod".
		if k == "dev" || k == "prod" {
			if tmp.EnvironmentOverrides == nil {
				tmp.EnvironmentOverrides = make(map[string]yaml.Node)
			}
			tmp.EnvironmentOverrides[k] = v
			continue
		}

		return fmt.Errorf(`invalid property "env": must be a map of strings`)
	}

	// Handle "dev:" and "prod:" shorthands (copy to to tmp.Env)
	if !tmp.Dev.IsZero() {
		if tmp.EnvironmentOverrides == nil {
			tmp.EnvironmentOverrides = make(map[string]yaml.Node)
		}
		tmp.EnvironmentOverrides["dev"] = tmp.Dev
	}
	if !tmp.Prod.IsZero() {
		if tmp.EnvironmentOverrides == nil {
			tmp.EnvironmentOverrides = make(map[string]yaml.Node)
		}
		tmp.EnvironmentOverrides["prod"] = tmp.Prod
	}

	// Apply environment-specific overrides
	if envOverride := tmp.EnvironmentOverrides[p.Environment]; !envOverride.IsZero() {
		if err := envOverride.Decode(tmp); err != nil {
			return newYAMLError(err)
		}
	}

	// Display name backwards compatibility
	if tmp.Title != "" && tmp.DisplayName == "" {
		tmp.DisplayName = tmp.Title
	}
	if tmp.Name != "" && tmp.DisplayName == "" {
		tmp.DisplayName = tmp.Name
	}

	// Parse environment variables from the "env:" (current) and "vars:" (deprecated) keys.
	vars := make(map[string]string)
	for k, v := range tmp.Vars { // Backwards compatibility
		if err := env.ValidateName(k); err != nil {
			return err
		}
		vars[k] = v
	}

	for k, v := range tmp.Env { // nolint: gocritic // Using a pointer changes parser behavior
		if v.Kind == yaml.ScalarNode {
			if err := env.ValidateName(k); err != nil {
				return err
			}
			vars[k] = v.Value
		}
	}

	// Validate resource defaults
	if !tmp.Sources.IsZero() {
		// NOTE: Validating against ModelYAML since SourceYAML is deprecated.
		if err := tmp.Sources.Decode(&ModelYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.Models.IsZero() {
		if err := tmp.Models.Decode(&ModelYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.MetricsViews.IsZero() {
		if err := tmp.MetricsViews.Decode(&MetricsViewYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.MetricsViewsLegacy.IsZero() {
		if err := tmp.MetricsViewsLegacy.Decode(&MetricsViewYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.Explores.IsZero() {
		if err := tmp.Explores.Decode(&ExploreYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.Canvases.IsZero() {
		if err := tmp.Canvases.Decode(&CanvasYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.APIs.IsZero() {
		if err := tmp.APIs.Decode(&APIYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.Migrations.IsZero() {
		if err := tmp.Migrations.Decode(&MigrationYAML{}); err != nil {
			return newYAMLError(err)
		}
	}

	// For backwards compatibility, we allow "features" to be either a map of strings (preferred) or a sequence of strings.
	var featureFlags map[string]string
	if !tmp.Features.IsZero() {
		switch tmp.Features.Kind {
		case yaml.MappingNode:
			if err := tmp.Features.Decode(&featureFlags); err != nil {
				return newYAMLError(err)
			}
		case yaml.SequenceNode:
			var fs []string
			if err := tmp.Features.Decode(&fs); err != nil {
				return newYAMLError(err)
			}

			featureFlags = map[string]string{}
			for _, f := range fs {
				featureFlags[f] = "true"
			}
		default:
			return fmt.Errorf(`invalid property "features": must be a map or a sequence`)
		}
	}
	// Validate the feature flags to give upfront error.
	for f, v := range featureFlags {
		_, err := ResolveTemplate(v, validationTemplateData, false)
		if err != nil {
			return fmt.Errorf(`invalid property "features": invalid value %q for %q: %w`, v, f, err)
		}
	}

	// We used to have camelCase for feature flags before. Convert it to snake_case to maintain backwards compatibility.
	snakeCaseFeatureFlags := make(map[string]string, len(featureFlags))
	for f, v := range featureFlags {
		sf := strcase.ToSnake(f)
		snakeCaseFeatureFlags[sf] = v
	}

	if len(tmp.PublicPaths) == 0 {
		tmp.PublicPaths = []string{"public"}
	}

	defaults := map[ResourceKind]yaml.Node{
		ResourceKindSource:      tmp.Sources,
		ResourceKindModel:       tmp.Models,
		ResourceKindMetricsView: tmp.MetricsViews,
		ResourceKindExplore:     tmp.Explores,
		ResourceKindMigration:   tmp.Migrations,
		ResourceKindCanvas:      tmp.Canvases,
		ResourceKindAPI:         tmp.APIs,
	}
	if !tmp.MetricsViewsLegacy.IsZero() {
		defaults[ResourceKindMetricsView] = tmp.MetricsViewsLegacy
	}

	res := &RillYAML{
		DisplayName:    tmp.DisplayName,
		Description:    tmp.Description,
		AIInstructions: tmp.AIInstructions,
		AIConnector:    tmp.AIConnector,
		OLAPConnector:  tmp.OLAPConnector,
		Connectors:     make([]*ConnectorDef, len(tmp.Connectors)),
		Variables:      make([]*VariableDef, len(vars)),
		Defaults:       defaults,
		FeatureFlags:   snakeCaseFeatureFlags,
		PublicPaths:    tmp.PublicPaths,
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
	for k, v := range vars {
		res.Variables[i] = &VariableDef{
			Name:    k,
			Default: v,
		}
		i++
	}

	p.RillYAML = res
	return nil
}
