package parser

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	// Load IANA time zone data
	_ "time/tzdata"
)

// parseSource parses a source definition and adds the resulting resource to p.Resources.
func (p *Parser) parseSource(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := make(map[string]any)
	if node.YAML == nil {
		node.YAML = &yaml.Node{}
	}
	err := node.YAML.Decode(tmp)
	if err != nil {
		return err
	}

	// Backwards compatibility: "type:" was previously used instead of "connector:".
	// So if "type:" is not a valid resource kind, we treat it as a connector.
	if typ, ok := tmp["type"].(string); ok {
		if _, err := ParseResourceKind(typ); err != nil {
			node.Connector = typ
			node.ConnectorInferred = false
		}
	}

	tmp["type"] = "model"
	tmp["defined_as_source"] = true
	tmp["materialize"] = true
	if _, ok := tmp["output"]; !ok {
		tmp["output"] = map[string]any{"connector": p.defaultOLAPConnector()}
	}

	// Backward compatibility: when the default connector is "olap", and it's a DuckDB connector, a source with connector "duckdb" should run on it
	if p.DefaultOLAPConnector == "olap" && node.Connector == "duckdb" {
		node.Connector = "olap"
	}

	// Validate the source has a connector
	if node.ConnectorInferred {
		return fmt.Errorf("must explicitly specify a connector for sources")
	}

	// Convert back to YAML
	err = node.YAML.Encode(tmp)
	if err != nil {
		return err
	}
	bytes, err := yaml.Marshal(node.YAML)
	if err != nil {
		return err
	}
	node.YAMLRaw = string(bytes)

	// NOTE: Not changing node.Kind such that the call to decodeNodeYAML in parseModel still applies the correct project-wide defaults.

	// We allowed a special resource type (source) to ingest data from external connector.
	// After the unification of sources and models everything is a model.
	return p.parseModel(ctx, node)
}
