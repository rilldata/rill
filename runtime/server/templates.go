package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/templates"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListTemplates returns available template definitions, optionally filtered by tags.
func (s *Server) ListTemplates(ctx context.Context, req *runtimev1.ListTemplatesRequest) (*runtimev1.ListTemplatesResponse, error) {
	tmplList := s.templateRegistry.ListByTags(req.Tags)

	pbs := make([]*runtimev1.Template, len(tmplList))
	for i, t := range tmplList {
		files := make([]*runtimev1.TemplateFile, len(t.Files))
		for j, f := range t.Files {
			files[j] = &runtimev1.TemplateFile{
				Name:        f.Name,
				PathPattern: f.PathTemplate,
			}
		}
		pbs[i] = &runtimev1.Template{
			Name:        t.Name,
			DisplayName: t.DisplayName,
			Driver:      t.Driver,
			Olap:        t.OLAP,
			Tags:        t.Tags,
			Files:       files,
		}
	}

	return &runtimev1.ListTemplatesResponse{Templates: pbs}, nil
}

// GenerateFile renders a template and optionally writes the resulting files.
func (s *Server) GenerateFile(ctx context.Context, req *runtimev1.GenerateFileRequest) (*runtimev1.GenerateFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.template_name", req.TemplateName),
		attribute.String("args.output", req.Output),
		attribute.String("args.connector_name", req.ConnectorName),
		attribute.Bool("args.preview", req.Preview),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	// Look up template
	tmpl, ok := s.templateRegistry.Get(req.TemplateName)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unknown template %q", req.TemplateName)
	}

	// Look up driver spec (nil for driverless templates like iceberg-duckdb)
	var driverSpec *drivers.Spec
	if tmpl.Driver != "" {
		drv, ok := drivers.Connectors[tmpl.Driver]
		if !ok {
			return nil, status.Errorf(codes.Internal, "template %q references unknown driver %q", req.TemplateName, tmpl.Driver)
		}
		spec := drv.Spec()
		driverSpec = &spec
	}

	// Convert properties
	var props map[string]any
	if req.Properties != nil {
		props = req.Properties.AsMap()
	} else {
		props = make(map[string]any)
	}

	// Read existing .env for conflict resolution
	existingEnv := make(map[string]bool)
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err == nil {
		existingEnv = templates.ReadEnvKeys(ctx, repo)
		release()
	}

	// Render
	result, err := templates.Render(&templates.RenderInput{
		Template:      tmpl,
		Output:        req.Output,
		DriverSpec:    driverSpec,
		Properties:    props,
		ConnectorName: req.ConnectorName,
		ExistingEnv:   existingEnv,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "rendering template: %s", err)
	}

	// Write files if not preview mode
	if !req.Preview {
		repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "opening repo: %s", err)
		}
		defer release()

		// Write rendered files
		for _, f := range result.Files {
			if err := repo.Put(ctx, f.Path, strings.NewReader(f.Blob)); err != nil {
				return nil, status.Errorf(codes.Internal, "writing file %q: %s", f.Path, err)
			}
		}

		// Merge env vars into .env
		if len(result.EnvVars) > 0 {
			envContent, _ := repo.Get(ctx, ".env")
			for key, val := range result.EnvVars {
				envContent = appendEnvVar(envContent, key, val)
			}
			if err := repo.Put(ctx, ".env", strings.NewReader(envContent)); err != nil {
				return nil, status.Errorf(codes.Internal, "writing .env: %s", err)
			}
		}
	}

	// Build response
	pbFiles := make([]*runtimev1.GeneratedFile, len(result.Files))
	for i, f := range result.Files {
		pbFiles[i] = &runtimev1.GeneratedFile{
			Path: f.Path,
			Blob: f.Blob,
		}
	}

	return &runtimev1.GenerateFileResponse{
		Files:   pbFiles,
		EnvVars: result.EnvVars,
	}, nil
}

// appendEnvVar appends or updates an env var in .env content.
func appendEnvVar(content, key, value string) string {
	// Simple append; the frontend already handles deduplication
	if content != "" && content[len(content)-1] != '\n' {
		content += "\n"
	}
	return fmt.Sprintf("%s%s=%s\n", content, key, value)
}
