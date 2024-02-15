package rillv1

import (
	"context"
	"strings"
)

// MigrationYAML is the raw structure of a Migration resource defined in YAML (does not include common fields)
type MigrationYAML struct {
	Version uint `yaml:"version" mapstructure:"version"`
}

// parseMigration parses a migration definition and adds the resulting resource to p.Resources.
func (p *Parser) parseMigration(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &MigrationYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
	}

	// Add resource
	r, err := p.insertResource(ResourceKindMigration, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

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
