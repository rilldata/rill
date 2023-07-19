package rillv1

import (
	"context"
	"fmt"
	"strings"
)

// migrationYAML is the raw structure of a Migration resource defined in YAML (does not include common fields)
type migrationYAML struct {
	Version uint `yaml:"version" mapstructure:"version"`
}

// parseMigration parses a migration definition and adds the resulting resource to p.Resources.
func (p *Parser) parseMigration(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &migrationYAML{}
	if node.YAML != nil {
		if err := node.YAML.Decode(tmp); err != nil {
			return pathError{path: node.YAMLPath, err: newYAMLError(err)}
		}
	}

	// Override YAML config with SQL annotations
	err := mapstructureUnmarshal(node.SQLAnnotations, tmp)
	if err != nil {
		return pathError{path: node.SQLPath, err: fmt.Errorf("invalid SQL annotations: %w", err)}
	}

	// Upsert resource (in practice, this will always be an insert)
	r := p.upsertResource(ResourceKindMigration, node.Name, node.Paths, node.Refs...)
	if node.Connector != "" {
		r.MigrationSpec.Connector = node.Connector
	}
	if node.SQL != "" {
		r.MigrationSpec.Sql = strings.TrimSpace(node.SQL)
	}
	if tmp.Version > 0 {
		r.MigrationSpec.Version = uint32(tmp.Version)
	}

	return nil
}
