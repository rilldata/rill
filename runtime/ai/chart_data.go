package ai

import (
	"context"
	"fmt"

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

var _ Tool[ChartDataArgs, *ChartDataResult] = (*ChartData)(nil)

var _ MCPAppTool = (*ChartData)(nil)

// ChartDataArgs are the arguments of chart_data, which are the same as those of create_chart.
type ChartDataArgs map[string]any

func (t *ChartData) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  ChartDataName,
		Title: "Chart data",
		Description: `Return the data of a chart configuration without creating a chart in the conversation.
It is called by the chart widget to refetch a chart's data when the user changes its time grain, limits or comparison,
and is not meant to be called directly: use create_chart to show a chart.
It takes the same arguments as create_chart and returns the chart-ready rows, field metadata and resolved time ranges.`,
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
		InputSchema: chartsInputSchema(),
	}
}

// MCPSpec is the definition advertised on the MCP server: the same Flint shape of the arguments as create_chart (see CreateChart.MCPSpec).
func (t *ChartData) MCPSpec() *mcp.Tool {
	spec := t.Spec()
	spec.Description = `Return the data of a chart without creating a chart in the conversation.
It is called by the chart widget to refetch a chart's data when the user changes its time grain, top-N limits or comparison,
and is not meant to be called directly: use create_chart to show a chart.
It takes the same arguments as create_chart (chart_spec and data) and returns the chart-ready rows, field metadata,
semantic types and resolved time ranges.`
	spec.InputSchema = flintChartsInputSchema()
	return spec
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

func (t *ChartData) Handler(ctx context.Context, args ChartDataArgs) (*ChartDataResult, error) {
	s := GetSession(ctx)
	req, err := parseChartRequest(ctx, t.Runtime, s, args)
	if err != nil {
		return nil, err
	}
	instance, err := t.Runtime.Instance(ctx, s.InstanceID())
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	res, _, err := queryChartData(ctx, t.Runtime, s, instance, req)
	return res, err
}

// MCPAppUI restricts chart_data to the widget: hosts hide tools with app visibility from the model.
// Being app-only also makes its calls plumbing rather than conversation turns (see MCPAppTool).
func (t *ChartData) MCPAppUI() *mcpapp.ToolUI {
	return &mcpapp.ToolUI{Visibility: []string{mcpapp.VisibilityApp}}
}
