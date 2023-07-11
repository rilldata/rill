package rillv1

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/connectors"
	"golang.org/x/exp/slices"
)

// Connector contains metadata about a connector used in a Rill project
type Connector struct {
	Driver          string
	Name            string
	Spec            connectors.Spec
	Resources       []*Resource
	AnonymousAccess bool
}

// AnalyzeConnectors extracts connector metadata from a Rill project
func (p *Parser) AnalyzeConnectors(ctx context.Context) ([]*Connector, error) {
	// Group resources by connector
	connectorResources := make(map[string][]*Resource)
	for _, r := range p.Resources {
		if r.SourceSpec != nil {
			name := r.SourceSpec.SourceConnector
			connectorResources[name] = append(connectorResources[name], r)
			if r.SourceSpec.SourceConnector != r.SourceSpec.SinkConnector {
				name = r.SourceSpec.SinkConnector
				connectorResources[name] = append(connectorResources[name], r)
			}
		} else if r.ModelSpec != nil {
			name := r.ModelSpec.Connector
			connectorResources[name] = append(connectorResources[name], r)
		} else if r.MigrationSpec != nil {
			name := r.MigrationSpec.Connector
			connectorResources[name] = append(connectorResources[name], r)
		}
		// NOTE: If we add more resource kinds that use connectors, add connector extraction here
	}

	// Build output
	res := make([]*Connector, 0, len(connectorResources))
	for name, resources := range connectorResources {
		// Skip default connector
		if name == "" {
			continue
		}

		// Get connector
		driver, connector, err := p.connectorForName(name)
		if err != nil {
			return nil, err
		}

		// Check if all resources have anon access
		anonAccess := true
		for _, r := range resources {
			// Only sources can have anon access (not sinks)
			if r.SourceSpec == nil || r.SourceSpec.SourceConnector != name {
				anonAccess = false
				break
			}
			// Poll for anon access
			res, _ := connector.HasAnonymousAccess(ctx, &connectors.Env{}, &connectors.Source{
				Name:       r.Name.Name,
				Connector:  name,
				Properties: r.SourceSpec.Properties.AsMap(),
			})
			if !res {
				anonAccess = false
				break
			}
		}

		// Add connector info to output
		res = append(res, &Connector{
			Driver:          driver,
			Name:            name,
			Spec:            connector.Spec(),
			Resources:       resources,
			AnonymousAccess: anonAccess,
		})
	}

	// Sort output to ensure deterministic ordering
	slices.SortFunc(res, func(a, b *Connector) bool {
		return a.Name < b.Name
	})

	return res, nil
}

// connectorForName resolves a connector name to a connector driver
func (p *Parser) connectorForName(name string) (string, connectors.Connector, error) {
	// Unless overridden in rill.yaml, the connector name is the driver name
	driver := name
	for _, c := range p.RillYAML.Connectors {
		if c.Name == name {
			driver = c.Type
			break
		}
	}

	connector, ok := connectors.Connectors[driver]
	if !ok {
		return "", nil, fmt.Errorf("unknown connector type %q", driver)
	}
	return driver, connector, nil
}
