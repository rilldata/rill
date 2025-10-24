package markdown

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
)

// Renderer renders markdown templates with embedded Metrics SQL queries
type Renderer struct {
	runtime *runtime.Runtime
}

// NewRenderer creates a new markdown renderer
func NewRenderer(rt *runtime.Runtime) *Renderer {
	return &Renderer{runtime: rt}
}

// RenderContext contains the context for rendering a markdown template
type RenderContext struct {
	InstanceID string
	Where      *runtimev1.Expression
	TimeRange  *runtimev1.TimeRange
	TimeZone   string
	Claims     *runtime.SecurityClaims
}

// RenderResult contains both formatted and raw markdown
type RenderResult struct {
	FormattedMarkdown string // With format tokens
	RawMarkdown       string // Without tokens
}

// Render renders a markdown template with embedded queries
// Returns both formatted (with tokens) and raw (without tokens) versions
func (r *Renderer) Render(ctx context.Context, tmplStr string, renderCtx RenderContext) (*RenderResult, error) {
	// Create Go template with Sprig functions + custom helpers
	tmpl := template.New("markdown").
		Funcs(sprig.FuncMap()).
		Funcs(r.customFunctions(ctx, renderCtx))

	tmpl, err := tmpl.Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template (functions return format tokens)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	formattedMarkdown := buf.String()

	// Strip format tokens to create raw version
	rawMarkdown := stripFormatTokens(formattedMarkdown)

	return &RenderResult{
		FormattedMarkdown: formattedMarkdown,
		RawMarkdown:       rawMarkdown,
	}, nil
}

// stripFormatTokens removes format tokens and returns just the raw values
func stripFormatTokens(markdown string) string {
	// Pattern: __RILL_FORMAT__metricsview::measure::value__END__
	// Replace with just the value (3rd capture group)
	re := regexp.MustCompile(`__RILL_FORMAT__[^:]+::[^:]+::(.+?)__END__`)
	return re.ReplaceAllString(markdown, "$1")
}

// executeQuery executes a single Metrics SQL query
func (r *Renderer) executeQuery(ctx context.Context, sql string, renderCtx RenderContext) ([]map[string]any, error) {
	// Build resolver properties
	props := map[string]any{"sql": sql}

	// Add optional properties using proper types
	if renderCtx.TimeZone != "" {
		props["time_zone"] = renderCtx.TimeZone
	}
	if renderCtx.Where != nil {
		// Convert proto Expression to metricsview.Expression
		props["additional_where"] = metricsview.NewExpressionFromProto(renderCtx.Where)
	}
	if renderCtx.TimeRange != nil && renderCtx.TimeRange.Start != nil && renderCtx.TimeRange.End != nil {
		props["additional_time_range"] = &metricsview.TimeRange{
			Start: renderCtx.TimeRange.Start.AsTime(),
			End:   renderCtx.TimeRange.End.AsTime(),
		}
	}

	// Initialize and execute the metrics_sql resolver
	initializer := runtime.ResolverInitializers["metrics_sql"]
	resolver, err := initializer(ctx, &runtime.ResolverOptions{
		Runtime:    r.runtime,
		InstanceID: renderCtx.InstanceID,
		Properties: props,
		Args:       make(map[string]any),
		Claims:     renderCtx.Claims,
		ForExport:  false,
	})
	if err != nil {
		return nil, err
	}
	defer resolver.Close()

	result, err := resolver.ResolveInteractive(ctx)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	// Read all rows
	rows := make([]map[string]any, 0)
	for {
		row, err := result.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		rowStruct, err := pbutil.ToStruct(row, result.Schema())
		if err != nil {
			return nil, err
		}
		rows = append(rows, rowStruct.AsMap())
	}

	return rows, nil
}
