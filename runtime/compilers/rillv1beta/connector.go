package rillv1beta

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
)

// TODO :: return this to build support for all kind of variables
type Variables struct {
	ProjectVariables []connectors.VariableSchema
	Connectors       []*Connector
}

type Connector struct {
	Name            string
	Type            string
	Spec            connectors.Spec
	AnonymousAccess bool
}

func ExtractConnectors(ctx context.Context, projectPath string) ([]*Connector, error) {
	sourcesPath := filepath.Join(projectPath, "sources")
	sources, err := doublestar.Glob(os.DirFS(sourcesPath), "*.{yaml,yml}", doublestar.WithFailOnPatternNotExist())
	if err != nil {
		return nil, err
	}

	// keeping a map to dedup connectors
	connectorMap := make(map[key]bool)
	for _, fileName := range sources {
		src, err := readSource(ctx, filepath.Join(sourcesPath, fileName))
		if err != nil {
			return nil, fmt.Errorf("error in reading source file %v : %w", fileName, err)
		}

		connector, ok := connectors.Connectors[src.Connector]
		if !ok {
			return nil, fmt.Errorf("no source connector defined for type %q", src.Connector)
		}

		// ignoring error since failure to resolve this should not break the deployment flow
		// this can fail under cases such as full or host/bucket of URI is a variable
		access, _ := connector.HasAnonymousAccess(ctx, &connectors.Env{}, src)
		c := key{Name: src.Connector, Type: src.Connector, AnonymousAccess: access}
		connectorMap[c] = true
	}

	result := make([]*Connector, 0)
	for k := range connectorMap {
		connector := connectors.Connectors[k.Type]
		result = append(result, &Connector{Name: k.Name, Type: k.Type, Spec: connector.Spec(), AnonymousAccess: k.AnonymousAccess})
	}
	return result, nil
}

func readSource(ctx context.Context, path string) (*connectors.Source, error) {
	catalog, err := read(ctx, path)
	if err != nil {
		return nil, err
	}

	apiSource := catalog.GetSource()
	source := &connectors.Source{
		Name:          apiSource.Name,
		Connector:     apiSource.Connector,
		Properties:    apiSource.Properties.AsMap(),
		ExtractPolicy: apiSource.GetPolicy(),
		Timeout:       apiSource.GetTimeoutSeconds(),
	}

	return source, nil
}

func read(ctx context.Context, path string) (*drivers.CatalogEntry, error) {
	artifact, ok := artifacts.Artifacts[fileutil.FullExt(path)]
	if !ok {
		return nil, fmt.Errorf("no artifact found for %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error in reading file %v : %w", path, err)
	}

	catalog, err := artifact.DeSerialise(ctx, path, string(content))
	if err != nil {
		return nil, err
	}

	catalog.Path = path
	return catalog, nil
}

type key struct {
	Name            string
	Type            string
	AnonymousAccess bool
}
