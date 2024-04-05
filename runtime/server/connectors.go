package server

import (
	"context"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/auth"
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
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadInstance) {
		return nil, ErrForbidden
	}

	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	p, err := rillv1.Parse(ctx, repo, req.InstanceId, inst.Environment, inst.OLAPConnector)
	if err != nil {
		return nil, err
	}

	connectors, err := p.AnalyzeConnectors(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*runtimev1.AnalyzedConnector)

	for _, connector := range connectors {
		cfg, err := s.runtime.ConnectorConfig(ctx, req.InstanceId, connector.Name)
		if err != nil {
			return nil, err
		}

		c := &runtimev1.AnalyzedConnector{
			Name:               connector.Name,
			Driver:             driverSpecToPB(connector.Driver, connector.Spec),
			Config:             cfg.ResolveStrings(),
			PresetConfig:       cfg.Preset,
			ProjectConfig:      connector.DefaultConfig, // NOTE: Could also use cfg.Project, but connector.DefaultConfig might be slightly more up-to-date
			EnvConfig:          cfg.Env,
			HasAnonymousAccess: connector.AnonymousAccess,
			UsedBy:             nil,
		}

		for _, r := range connector.Resources {
			c.UsedBy = append(c.UsedBy, runtime.ResourceNameFromCompiler(r.Name))
		}

		res[connector.Name] = c
	}

	return &runtimev1.AnalyzeConnectorsResponse{
		Connectors: maps.Values(res),
	}, nil
}

func (s *Server) ListNotifierConnectors(ctx context.Context, req *runtimev1.ListNotifierConnectorsRequest) (*runtimev1.ListNotifierConnectorsResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadObjects) {
		return nil, ErrForbidden
	}

	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*runtimev1.Connector)

	for _, c := range inst.Connectors {
		if driverIsNotifier(c.Type) {
			res[c.Name] = &runtimev1.Connector{
				Type: c.Type,
				Name: c.Name,
			}
		}
	}

	for _, c := range inst.ProjectConnectors {
		if driverIsNotifier(c.Type) {
			res[c.Name] = &runtimev1.Connector{
				Type: c.Type,
				Name: c.Name,
			}
		}
	}

	// Connectors may be implicitly defined just by adding variables in the format "connector.<name>.<property>".
	// NOTE: We can remove this if we move to explicitly defined connectors.
	for k := range inst.ResolveVariables() {
		if !strings.HasPrefix(k, "connector.") {
			continue
		}

		parts := strings.Split(k, ".")
		if len(parts) <= 2 {
			continue
		}

		// Implicitly defined connectors always have the same name as the driver
		name := parts[1]
		if driverIsNotifier(name) {
			res[name] = &runtimev1.Connector{
				Type: name,
				Name: name,
			}
		}
	}

	return &runtimev1.ListNotifierConnectorsResponse{
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
		ImplementsNotifier:    spec.ImplementsNotifier,
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

func driverIsNotifier(driver string) bool {
	connector, ok := drivers.Connectors[driver]
	if !ok {
		return false
	}

	return connector.Spec().ImplementsNotifier
}
