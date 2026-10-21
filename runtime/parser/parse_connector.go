package parser

import (
	"bytes"
	"fmt"
	"net/url"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// ConnectorYAML is the raw structure of a Connector resource defined in YAML (does not include common fields)
type ConnectorYAML struct {
	commonYAML `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into Properties
	// Driver name
	Driver  string    `yaml:"driver"`
	Managed yaml.Node `yaml:"managed"` // Boolean or map of properties
	// Security is an access policy for the connector. It is only supported by the "mcp" driver, where it controls
	// which users may use the server and which of its tools; see mcpSecurityYAML.
	Security   yaml.Node      `yaml:"security"`
	Properties map[string]any `yaml:",inline" mapstructure:",remain"`
}

// mcpDriverName is the driver name of connectors for external MCP servers.
const mcpDriverName = "mcp"

// mcpAgentNames are the AI agents an MCP connector's security policy can be scoped to.
var mcpAgentNames = []string{"analyst", "developer"}

// mcpSecurityYAML is the "security" block of an MCP connector.
// It reuses the metrics view security policy syntax, with tool names in place of field names in "include" and "exclude",
// plus an "agents" field that scopes the server to specific AI agents. Row filters and custom rules are not supported.
type mcpSecurityYAML struct {
	SecurityPolicyYAML `yaml:",inline"`
	Agents             []string `yaml:"agents"`
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

	// Validate driver-specific properties and parse the security policy (MCP connectors only)
	var securityRules []*runtimev1.SecurityRule
	var agents []string
	if tmp.Driver == mcpDriverName {
		securityRules, agents, err = parseMCPConnector(tmp)
		if err != nil {
			return err
		}
	} else if !tmp.Security.IsZero() {
		return fmt.Errorf("'security' is only supported for connectors with driver %q", mcpDriverName)
	}

	// Find out if any properties are templated
	templatedProps, err := analyzeTemplatedProperties(tmp.Properties)
	if err != nil {
		return fmt.Errorf("failed to analyze templated properties: %w", err)
	}
	if tmp.Driver == mcpDriverName {
		// MCP header templates may reference the requesting user ({{ .user.* }}), so they are resolved per request by the driver
		// rather than once by the runtime (which would fail on the missing user attributes).
		templatedProps = slices.DeleteFunc(templatedProps, func(k string) bool { return k == "headers" })
	}
	var propertiesPB *structpb.Struct
	if tmp.Properties != nil {
		propertiesPB, err = structpb.NewStruct(tmp.Properties)
		if err != nil {
			return fmt.Errorf("failed to convert connector properties to proto: %w", err)
		}
	}

	// Insert the connector
	r, err := p.insertResource(ResourceKindConnector, node.Name, node.Paths, node.Tags, node.Metadata, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.ConnectorSpec.Driver = tmp.Driver
	r.ConnectorSpec.Properties = propertiesPB
	r.ConnectorSpec.TemplatedProperties = templatedProps
	r.ConnectorSpec.Provision = provision
	r.ConnectorSpec.ProvisionArgs = provisionArgsPB
	r.ConnectorSpec.SecurityRules = securityRules
	r.ConnectorSpec.Agents = agents

	return nil
}

// parseMCPConnector validates the properties of an MCP connector and parses its security policy.
// It returns the security rules and the agents the server is scoped to (empty means all agents).
func parseMCPConnector(tmp *ConnectorYAML) ([]*runtimev1.SecurityRule, []string, error) {
	rawURL, _ := tmp.Properties["url"].(string)
	if rawURL == "" {
		return nil, nil, fmt.Errorf("missing required property 'url'")
	}
	if meta, err := AnalyzeTemplate(rawURL); err != nil {
		return nil, nil, fmt.Errorf("invalid 'url': %w", err)
	} else if !meta.UsesTemplating {
		u, err := url.Parse(rawURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return nil, nil, fmt.Errorf("invalid 'url': expected an http(s) URL, got %q", rawURL)
		}
	}

	auth, _ := tmp.Properties["auth"].(string)
	if auth != "" && auth != "oauth" {
		return nil, nil, fmt.Errorf("invalid 'auth': expected \"oauth\" or empty, got %q", auth)
	}

	if headers, ok := tmp.Properties["headers"]; ok {
		headersMap, ok := headers.(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("invalid 'headers': expected a map of header names to values")
		}
		for k, v := range headersMap {
			if _, ok := v.(string); !ok {
				return nil, nil, fmt.Errorf("invalid 'headers': value of %q must be a string", k)
			}
			if auth == "oauth" && strings.EqualFold(k, "authorization") {
				return nil, nil, fmt.Errorf("an 'Authorization' header cannot be combined with 'auth: oauth' (choose per-user OAuth or a project-scoped token, not both)")
			}
		}
	}

	if tmp.Security.IsZero() {
		return nil, nil, nil
	}

	// Re-encode the node so it can be decoded strictly, which reports unknown keys such as typos.
	secYAML, err := yaml.Marshal(&tmp.Security)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid 'security': %w", err)
	}
	sec := &mcpSecurityYAML{}
	dec := yaml.NewDecoder(bytes.NewReader(secYAML))
	dec.KnownFields(true)
	err = dec.Decode(sec)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid 'security': %w", err)
	}
	if sec.RowFilter != "" {
		return nil, nil, fmt.Errorf("invalid 'security': 'row_filter' is not supported for MCP connectors")
	}
	if len(sec.Rules) > 0 {
		return nil, nil, fmt.Errorf("invalid 'security': 'rules' is not supported for MCP connectors")
	}
	for _, a := range sec.Agents {
		if !slices.Contains(mcpAgentNames, a) {
			return nil, nil, fmt.Errorf("invalid 'security': unknown agent %q in 'agents' (expected one of %s)", a, strings.Join(mcpAgentNames, ", "))
		}
	}

	// Proto() applies the metrics view semantics: a security block without "access" denies everyone.
	rules, err := sec.SecurityPolicyYAML.Proto()
	if err != nil {
		return nil, nil, err
	}
	return rules, sec.Agents, nil
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
