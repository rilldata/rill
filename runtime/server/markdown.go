package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/markdown"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RenderMarkdown renders a markdown template with embedded Metrics SQL queries
func (s *Server) RenderMarkdown(ctx context.Context, req *runtimev1.RenderMarkdownRequest) (*runtimev1.RenderMarkdownResponse, error) {
	// Add observability
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.Int("args.template_length", len(req.Template)),
	)

	// Check permissions - require ReadAPI permission
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadAPI) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to render markdown")
	}

	// Create markdown renderer
	renderer := markdown.NewRenderer(s.runtime)

	// Build render context
	renderCtx := markdown.RenderContext{
		InstanceID: req.InstanceId,
		Where:      req.Where,
		TimeRange:  req.TimeRange,
		TimeZone:   req.TimeZone,
		Claims:     claims,
	}

	// Render the template
	result, err := renderer.Render(ctx, req.Template, renderCtx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to render markdown: %s", err.Error())
	}

	return &runtimev1.RenderMarkdownResponse{
		RenderedMarkdown: result.FormattedMarkdown,
		RawMarkdown:      result.RawMarkdown,
	}, nil
}
