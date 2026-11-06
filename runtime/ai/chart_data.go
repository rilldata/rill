package ai

import (
	"context"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/mcpapp"
)

const ChartDataName = "chart_data"

// ChartData returns the data of a chart without creating a chart in the conversation.
// It is an app-only tool: the MCP Apps widget calls it (through the host) to refetch data when the user changes
// the time grain, limits or comparison of a chart created with create_chart. Its _meta.ui marks it app-only,
// so hosts that support MCP Apps hide it from the model; its description tells other clients not to call it directly.
type ChartData struct {
	Runtime *runtime.Runtime
}

var _ Tool[ChartDataArgs, *ChartDataOutput] = (*ChartData)(nil)

var _ MCPResultDecorator[*ChartDataOutput] = (*ChartData)(nil)

var _ MCPAppTool = (*ChartData)(nil)

// ChartDataArgs are the arguments of chart_data, which are the same as those of create_chart.
type ChartDataArgs map[string]any

// ChartDataOutput is the result of chart_data: the data section of a create_chart result and the effective input (see ChartInput),
// so that the widget shows the limits and the grain that were applied.
type ChartDataOutput struct {
	Input *ChartInput `json:"input,omitempty"`
	*ChartDataResult
}

// Spec defines the tool for MCP clients only: Rill's own clients fetch chart data themselves (see CheckAccess).
// The input schema is deliberately minimal, since hosts without MCP Apps list the tool to the model, which should not call it,
// and every byte of a tool list is sent with each request. The arguments are validated by the handler like those of create_chart.
func (t *ChartData) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  ChartDataName,
		Title: "Chart data",
		Description: `Return the data of a chart without creating a chart in the conversation.
It is called by the chart widget to refetch a chart's data when the user changes its time grain, top-N limits or comparison,
and is not meant to be called directly: use create_chart to show a chart.
It takes the arguments of a create_chart call and returns the chart-ready rows, field metadata, semantic types and resolved time ranges.`,
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
		InputSchema: &jsonschema.Schema{
			Type:        "object",
			Description: "The arguments of a create_chart call: chart_spec, data and optionally semantic_types, plus project on a server that serves several projects.",
			Properties: map[string]*jsonschema.Schema{
				"chart_spec":     {Type: "object", Description: "As in create_chart."},
				"data":           {Type: "object", Description: "As in create_chart."},
				"semantic_types": {Type: "object", Description: "As in create_chart."},
			},
		},
	}
}

func (t *ChartData) CheckAccess(ctx context.Context) (bool, error) {
	// Like create_chart, it requires ReadMetrics and the chat_charts feature flag.
	// Only external MCP clients get it: Rill's own clients fetch chart data themselves.
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.ReadMetrics) || isRillUserAgent(s) {
		return false, nil
	}
	ff, err := s.FeatureFlags(ctx)
	if err != nil {
		return false, err
	}
	return ff["chat_charts"], nil
}

func (t *ChartData) Handler(ctx context.Context, args ChartDataArgs) (*ChartDataOutput, error) {
	s := GetSession(ctx)
	req, err := parseFlintChartRequest(ctx, t.Runtime, s, args)
	if err != nil {
		return nil, err
	}
	instance, err := t.Runtime.Instance(ctx, s.InstanceID())
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	data, limits, err := queryChartData(ctx, t.Runtime, s, instance, req)
	if err != nil {
		return nil, err
	}
	return &ChartDataOutput{
		Input:           flintInputWithLimits(req.input, limits, data.TimeGrain),
		ChartDataResult: data,
	}, nil
}

// MCPAppUI restricts chart_data to the widget: hosts hide tools with app visibility from the model.
// Being app-only also makes its calls plumbing rather than conversation turns (see MCPAppTool).
func (t *ChartData) MCPAppUI() *mcpapp.ToolUI {
	return &mcpapp.ToolUI{Visibility: []string{mcpapp.VisibilityApp}}
}

// DecorateMCPResult replaces the text content the SDK would derive from the result, which repeats all of structuredContent,
// with a one-line summary: the widget renders from structuredContent, and the model is not meant to see the result.
func (t *ChartData) DecorateMCPResult(ctx context.Context, out *ChartDataOutput, res *mcp.CallToolResult) {
	text := fmt.Sprintf("Chart data: %d rows.", len(out.Data))
	if out.TruncationWarning != "" {
		text += " " + out.TruncationWarning
	}
	res.Content = []mcp.Content{&mcp.TextContent{Text: text}}
}
