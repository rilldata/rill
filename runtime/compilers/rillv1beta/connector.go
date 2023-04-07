package rillv1beta

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-yaml/yaml"
	"github.com/rilldata/rill/runtime/connectors"
)

// TODO :: return this to build support for all kind of variables
type Variables struct {
	ProjectVariables []connectors.VariableSchema
	Connectors       []*Connector
}

type Connector struct {
	Name      string
	Type      string
	Variables []connectors.VariableSchema
	Help      string
}

func ExtractConnectors(projectPath string) ([]*Connector, error) {
	sourcesPath := filepath.Join(projectPath, "sources")
	sources, err := doublestar.Glob(os.DirFS(sourcesPath), "*.{yaml,yml}", doublestar.WithFailOnPatternNotExist())
	if err != nil {
		return nil, err
	}

	// keeping a map to dedup connectors
	connectorMap := make(map[key]bool)
	for _, fileName := range sources {
		content, err := os.ReadFile(filepath.Join(sourcesPath, fileName))
		if err != nil {
			return nil, fmt.Errorf("error in reading file %v : %w", fileName, err)
		}

		// todo :: check for commented sources
		src := Source{}
		if err := yaml.Unmarshal(content, &src); err != nil {
			return nil, fmt.Errorf("error in unmarshalling yaml file %v : %w", fileName, err)
		}

		c := key{Name: src.Type, Type: src.Type}
		connectorMap[c] = true
	}

	result := make([]*Connector, 0)
	for k := range connectorMap {
		connector, ok := connectors.Connectors[k.Type]
		if !ok {
			return nil, fmt.Errorf("no source connector defined for type %q", k.Type)
		}

		spec := connector.Spec()
		result = append(result, &Connector{Name: k.Name, Type: k.Type, Variables: spec.ConnectorVariables, Help: spec.Help})
	}
	return result, nil
}

type key struct {
	Name string
	Type string
}
