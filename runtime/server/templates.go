package server

import (
	"context"
	"os"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/templates"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
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
		var schemaPb *structpb.Struct
		if t.JSONSchema != nil {
			var err error
			schemaPb, err = structpb.NewStruct(t.JSONSchema)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "converting schema for %q: %s", t.Name, err)
			}
		}

		pbs[i] = &runtimev1.Template{
			Name:        t.Name,
			DisplayName: t.DisplayName,
			Description: t.Description,
			DocsUrl:     t.DocsURL,
			Driver:      t.Driver,
			Olap:        t.OLAP,
			Icon:        t.Icon,
			SmallIcon:   t.SmallIcon,
			Tags:        t.Tags,
			Files:       files,
			JsonSchema:  schemaPb,
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
	} else {
		s.logger.Warn("failed to open repo for .env conflict resolution; env var conflicts may not be detected", zap.Error(err))
	}

	// Render
	result, err := templates.Render(&templates.RenderInput{
		Template:      tmpl,
		Output:        req.Output,
		Properties:    props,
		ConnectorName: req.ConnectorName,
		ExistingEnv:   existingEnv,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "rendering template: %s", err)
	}

	// Write files if not preview mode
	if !req.Preview {
		if err := s.writeRenderedFiles(ctx, req.InstanceId, result); err != nil {
			return nil, err
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

// writeRenderedFiles writes all rendered output files and merges env vars into .env.
// Writes .env first so that YAML files referencing {{ .env.VAR }} have their secrets available.
func (s *Server) writeRenderedFiles(ctx context.Context, instanceID string, result *templates.RenderOutput) error {
	repo, release, err := s.runtime.Repo(ctx, instanceID)
	if err != nil {
		return status.Errorf(codes.Internal, "opening repo: %s", err)
	}
	defer release()

	// Write .env first so secrets are available when YAML files are parsed
	if len(result.EnvVars) > 0 {
		envContent, err := repo.Get(ctx, ".env")
		if err != nil && !os.IsNotExist(err) {
			s.logger.Warn("failed to read .env; existing env vars may be overwritten", zap.Error(err))
		}
		for key, val := range result.EnvVars {
			envContent = appendEnvVar(envContent, key, val)
		}
		if err := repo.Put(ctx, ".env", strings.NewReader(envContent)); err != nil {
			return status.Errorf(codes.Internal, "writing .env: %s", err)
		}
	}

	// Write rendered files
	for _, f := range result.Files {
		if err := repo.Put(ctx, f.Path, strings.NewReader(f.Blob)); err != nil {
			return status.Errorf(codes.Internal, "writing file %q: %s", f.Path, err)
		}
	}

	return nil
}

// appendEnvVar updates an existing env var or appends a new one.
// If the key already exists, its value is replaced in-place.
// Values are sanitized: newlines are stripped to prevent injection of extra
// env vars, and values containing spaces or special characters are quoted.
func appendEnvVar(content, key, value string) string {
	// Sanitize: strip newlines to prevent env var injection
	value = strings.ReplaceAll(value, "\n", "")
	value = strings.ReplaceAll(value, "\r", "")

	// Quote values that contain spaces, '=', or '#' (comment char).
	// Escape backslashes first, then double quotes, to avoid double-escaping.
	if strings.ContainsAny(value, ` =#'"\\`) {
		value = strings.ReplaceAll(value, `\`, `\\`)
		value = `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}

	prefix := key + "="
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = prefix + value
			return strings.Join(lines, "\n")
		}
	}
	// Key not found; append
	if content != "" && content[len(content)-1] != '\n' {
		content += "\n"
	}
	return content + prefix + value + "\n"
}
