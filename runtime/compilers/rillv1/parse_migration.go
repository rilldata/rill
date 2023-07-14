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
func (p *Parser) parseMigration(ctx context.Context, stem *Stem) error {
	// Parse YAML
	tmp := &migrationYAML{}
	if stem.YAML != nil {
		if err := stem.YAML.Decode(tmp); err != nil {
			return pathError{path: stem.YAMLPath, err: newYAMLError(err)}
		}
	}

	// Override YAML config with SQL annotations
	err := mapstructureUnmarshal(stem.SQLAnnotations, tmp)
	if err != nil {
		return pathError{path: stem.SQLPath, err: fmt.Errorf("invalid SQL annotations: %w", err)}
	}

	// Upsert resource (in practice, this will always be an insert)
	r := p.upsertResource(ResourceKindMigration, stem.Name, stem.Paths, stem.Refs...)
	if stem.Connector != "" {
		r.MigrationSpec.Connector = stem.Connector
	}
	if stem.SQL != "" {
		r.MigrationSpec.Sql = strings.TrimSpace(stem.SQL)
	}
	if tmp.Version > 0 {
		r.MigrationSpec.Version = uint32(tmp.Version)
	}

	return nil
}
