package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/exp/maps"
)

func (s *Server) ListConnectorDrivers(ctx context.Context, req *runtimev1.ListConnectorDriversRequest) (*runtimev1.ListConnectorDriversResponse, error) {
	var pbs []*runtimev1.ConnectorDriver
	for name, driver := range drivers.Connectors {
		pbs = append(pbs, driverSpecToPB(name, driver.Spec()))
	}
	return &runtimev1.ListConnectorDriversResponse{Connectors: pbs}, nil
}

func (s *Server) AnalyzeConnectors(ctx context.Context, req *runtimev1.AnalyzeConnectorsRequest) (*runtimev1.AnalyzeConnectorsResponse, error) {
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	p, err := rillv1.Parse(ctx, repo, req.InstanceId, inst.Environment, inst.ResolveOLAPConnector())
	if err != nil {
		return nil, err
	}

	connectors, err := p.AnalyzeConnectors(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*runtimev1.AnalyzedConnector)

	for _, connector := range connectors {
		c := &runtimev1.AnalyzedConnector{
			Name:               connector.Name,
			Driver:             driverSpecToPB(connector.Driver, connector.Spec),
			HasAnonymousAccess: connector.AnonymousAccess,
			UsedBy:             nil,
		}

		for _, r := range connector.Resources {
			c.UsedBy = append(c.UsedBy, runtime.ResourceNameFromCompiler(r.Name))
		}

		res[connector.Name] = c
	}

	// TODO: Incorporate logic from runtime/connections.go:connectorConfig to populate:
	// 1. Add connectors defined outside of the project files (e.g. in variables).
	// 2. Correctly populate the Config, DefaultConfig, EnvConfig properties.

	return &runtimev1.AnalyzeConnectorsResponse{
		Connectors: maps.Values(res),
	}, nil
}

func driverSpecToPB(name string, spec drivers.Spec) *runtimev1.ConnectorDriver {
	pb := &runtimev1.ConnectorDriver{
		Name:                  name,
		ConfigProperties:      nil,
		SourceProperties:      nil,
		DisplayName:           spec.DisplayName,
		Description:           spec.Description,
		ImplementsRegistry:    spec.ImplementsRegistry,
		ImplementsCatalog:     spec.ImplementsCatalog,
		ImplementsRepo:        spec.ImplementsRepo,
		ImplementsAdmin:       spec.ImplementsAdmin,
		ImplementsAi:          spec.ImplementsAI,
		ImplementsSqlStore:    spec.ImplementsSQLStore,
		ImplementsOlap:        spec.ImplementsOLAP,
		ImplementsObjectStore: spec.ImplementsObjectStore,
		ImplementsFileStore:   spec.ImplementsFileStore,
	}

	for _, prop := range spec.ConfigProperties {
		pb.ConfigProperties = append(pb.ConfigProperties, driverPropertySpecToPB(prop))
	}

	for _, prop := range spec.SourceProperties {
		pb.SourceProperties = append(pb.SourceProperties, driverPropertySpecToPB(prop))
	}

	return pb
}

func driverPropertySpecToPB(spec *drivers.PropertySpec) *runtimev1.ConnectorDriver_Property {
	var t runtimev1.ConnectorDriver_Property_Type
	switch spec.Type {
	case drivers.NumberPropertyType:
		t = runtimev1.ConnectorDriver_Property_TYPE_NUMBER
	case drivers.BooleanPropertyType:
		t = runtimev1.ConnectorDriver_Property_TYPE_BOOLEAN
	case drivers.StringPropertyType:
		t = runtimev1.ConnectorDriver_Property_TYPE_STRING
	case drivers.FilePropertyType:
		t = runtimev1.ConnectorDriver_Property_TYPE_FILE
	case drivers.InformationalPropertyType:
		t = runtimev1.ConnectorDriver_Property_TYPE_INFORMATIONAL
	}

	return &runtimev1.ConnectorDriver_Property{
		Key:         spec.Key,
		Type:        t,
		Required:    spec.Required,
		DisplayName: spec.DisplayName,
		Description: spec.Description,
		DocsUrl:     spec.DocsURL,
		Hint:        spec.Hint,
		Default:     spec.Default,
		Placeholder: spec.Placeholder,
		Secret:      spec.Secret,
	}
}
