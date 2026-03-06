package server

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/templates"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GenerateTemplate generates a connector or model YAML file from structured form data.
// Deprecated: use GenerateFile instead. This handler delegates to the templates package.
func (s *Server) GenerateTemplate(ctx context.Context, req *runtimev1.GenerateTemplateRequest) (*runtimev1.GenerateTemplateResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.resource_type", req.ResourceType),
		attribute.String("args.driver", req.Driver),
		attribute.String("args.connector_name", req.ConnectorName),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	// Validate resource type
	if req.ResourceType != "connector" && req.ResourceType != "model" {
		return nil, status.Errorf(codes.InvalidArgument, "resource_type must be \"connector\" or \"model\"")
	}

	// Validate driver exists
	drv, ok := drivers.Connectors[req.Driver]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unknown driver %q", req.Driver)
	}
	spec := drv.Spec()

	// Convert properties
	var props map[string]any
	if req.Properties != nil {
		props = req.Properties.AsMap()
	} else {
		props = make(map[string]any)
	}

	// Validate properties against the driver spec (backward compat)
	if err := validateProperties(spec, req.ResourceType, props); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}

	// Look up template
	tmpl, ok := s.templateRegistry.LookupByDriver(req.Driver, req.ResourceType)
	if !ok {
		return nil, status.Errorf(codes.Internal, "no template for driver %q resource type %q", req.Driver, req.ResourceType)
	}

	// Read existing .env for env var conflict resolution
	existingEnv := make(map[string]bool)
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err == nil {
		existingEnv = templates.ReadEnvKeys(ctx, repo)
		release()
	} else {
		s.logger.Warn("failed to open repo for .env conflict resolution; env var conflicts may not be detected", zap.Error(err))
	}

	// Render using the templates package
	result, err := templates.Render(&templates.RenderInput{
		Template:      tmpl,
		Output:        req.ResourceType,
		Properties:    props,
		ConnectorName: req.ConnectorName,
		ExistingEnv:   existingEnv,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "rendering template: %s", err)
	}

	blob := ""
	if len(result.Files) > 0 {
		blob = result.Files[0].Blob
	}

	// Backward compat: the old API returned "duckdb" as the driver for
	// object store, file store, HTTPS, and SQLite models that get rewritten to DuckDB.
	responseDriver := req.Driver
	if req.ResourceType == "model" {
		if spec.ImplementsObjectStore || spec.ImplementsFileStore || req.Driver == "sqlite" || req.Driver == "https" {
			responseDriver = "duckdb"
		}
	}

	return &runtimev1.GenerateTemplateResponse{
		Blob:         blob,
		EnvVars:      result.EnvVars,
		ResourceType: req.ResourceType,
		Driver:       responseDriver,
	}, nil
}

// validateProperties rejects unknown property keys.
func validateProperties(spec drivers.Spec, resourceType string, properties map[string]any) error {
	allowed := make(map[string]bool)
	var props []*drivers.PropertySpec
	if resourceType == "connector" {
		props = spec.ConfigProperties
	} else {
		props = spec.SourceProperties
		// Universal model properties: sql and name are always allowed for models,
		// even if the driver's SourceProperties doesn't list them explicitly.
		allowed["sql"] = true
		allowed["name"] = true
	}
	for _, p := range props {
		allowed[p.Key] = true
	}
	for key := range properties {
		if !allowed[key] {
			return fmt.Errorf("unknown property %q for driver", key)
		}
	}
	return nil
}
