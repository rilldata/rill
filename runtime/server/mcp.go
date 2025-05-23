package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const mcpInstructions = `
## Rill MCP Server
This server exposes APIs for querying **metrics views** (Rill's analytical units).
### Workflow Overview
1. **List Metrics Views:** Use "list_metrics_views" to discover available metrics views in a project.
2. **Get Metrics View Spec:** Use "get_metrics_view" to fetch a metrics view's spec. This is important to understand all the dimensions and measures in the metrics view.
3. **Get Time Range:** Use "get_metrics_view_time_range_summary" to obtain the available time range for a metrics view. This is important to understand what time range the data spans.
4. **Query Aggregations:** Use "get_metrics_view_aggregation" to run queries.
In the workflow, do not proceed with the next step until the previous step has been completed. If the information from the previous step is already known (let's say for subsequent queries), you can skip it.
`

func (s *Server) newMCPHandler(basePath string) http.Handler {
	mcpServer := server.NewMCPServer("rill", "0.1.0", // TODO: Correct version
		server.WithRecovery(),
		// TODO: Observability: hooks, error logger
		server.WithToolCapabilities(true),
		server.WithInstructions(mcpInstructions),
	)

	mcpServer.AddTool(s.mcpEcho())
	mcpServer.AddTool(s.mcpListMetricsViews())
	mcpServer.AddTool(s.mcpGetMetricsView())
	mcpServer.AddTool(s.mcpGetMetricsViewTimeRangeSummary())
	mcpServer.AddTool(s.mcpGetMetricsViewAggregation())

	sseServer := server.NewSSEServer(
		mcpServer,
		server.WithStaticBasePath(basePath),
		server.WithUseFullURLForMessageEndpoint(false),
	)

	return sseServer
}

func (s *Server) mcpEcho() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("echo",
		mcp.WithDescription("Echoes the message back to the caller"),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("The message to echo"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		claims := auth.GetClaims(ctx)
		return mcp.NewToolResultText(fmt.Sprintf("Echo to subject %q: %v", claims.Subject(), req.GetArguments()["message"])), nil
	}

	return tool, handler
}

func (s *Server) mcpListMetricsViews() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("list_metrics_views",
		mcp.WithDescription("List all metrics views in the current project"),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := s.ListResources(ctx, &runtimev1.ListResourcesRequest{
			InstanceId: "default", // TODO: Dynamic
			Kind:       runtime.ResourceKindMetricsView,
		})
		if err != nil {
			return mcpErrorFromGRPC(err)
		}

		var names []string
		for _, r := range resp.Resources {
			mv := r.GetMetricsView()
			if mv == nil || mv.State.ValidSpec == nil {
				continue
			}

			names = append(names, r.Meta.Name.Name)
		}

		return mcpNewToolResultJSON(names)
	}

	return tool, handler
}

func (s *Server) mcpGetMetricsView() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_metrics_view",
		mcp.WithDescription("Retrieve the specification for a given metrics view, including available measures and dimensions"),
		mcp.WithString("metrics_view",
			mcp.Required(),
			mcp.Description("The name of the metrics view"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("metrics_view")
		if err != nil {
			return mcpNewToolError(err)
		}

		resp, err := s.GetResource(ctx, &runtimev1.GetResourceRequest{
			InstanceId: "default", // TODO: Dynamic
			Name: &runtimev1.ResourceName{
				Kind: runtime.ResourceKindMetricsView,
				Name: name,
			},
		})
		if err != nil {
			return mcpErrorFromGRPC(err)
		}

		mv := resp.Resource.GetMetricsView()
		if mv == nil || mv.State.ValidSpec == nil {
			return mcpNewToolErrorf("metrics view %q not valid", name)
		}

		return mcpNewToolResultJSON(mv.State.ValidSpec)
	}

	return tool, handler
}

func (s *Server) mcpGetMetricsViewTimeRangeSummary() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("get_metrics_view_time_range_summary",
		mcp.WithDescription(`
            Retrieve the total time range available for a given metrics view.
            Note: All subsequent queries of the metrics view should be constrained to this time range to ensure accurate results.
        `),
		mcp.WithString("metrics_view",
			mcp.Required(),
			mcp.Description("The name of the metrics view"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("metrics_view")
		if err != nil {
			return mcpNewToolError(err)
		}

		resp, err := s.MetricsViewTimeRange(ctx, &runtimev1.MetricsViewTimeRangeRequest{
			InstanceId:      "default", // TODO: Dynamic
			MetricsViewName: name,
		})
		if err != nil {
			return mcpErrorFromGRPC(err)
		}

		return mcpNewToolResultJSON(resp.TimeRangeSummary)
	}

	return tool, handler
}

func (s *Server) mcpGetMetricsViewAggregation() (mcp.Tool, server.ToolHandlerFunc) {
	description := `
Perform an arbitrary aggregation on a metrics view.
Tip: Use the 'sort' and 'limit' parameters for best results and to avoid large, unbounded result sets.
Example: Get the total revenue by country and product category:
    {
        "metrics_view": "ecommerce_financials",
        "measures": [{"name": "total_revenue"}, {"name": "total_orders"}],
        "dimensions": [{"name": "country"}, {"name": "product_category"}],
        "time_range": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2024-12-31T23:59:59Z"
        },
        "where": {
            "cond": {
                "op": "OPERATION_AND",
                "exprs": [
                    {
                        "cond": {
                            "op": "OPERATION_IN",
                            "exprs": [
                                {"ident": "country"},
                                {"val": ["US", "CA", "GB"]}
                            ]
                        }
                    },
                    {
                        "cond": {
                            "op": "OPERATION_EQ",
                            "exprs": [
                                {"ident": "product_category"},
                                {"val": "Electronics"}
                            ]
                        }
                    }
                ]
            },
        },
        "sort": [{"name": "total_revenue", "desc": true}],
        "limit": "10"
    }
    
Example: Get the total revenue by country, grouped by month:
    {
        "metrics_view": "ecommerce_financials",
        "measures": [{"name": "total_revenue"}],
        "dimensions": [
            {"name": "transaction_timestamp", "time_grain": "TIME_GRAIN_MONTH"}
            {"name": "country"},
        ],
        "time_range": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2024-12-31T23:59:59Z"
        },
        "sort": [
            {"name": "transaction_timestamp"},
            {"name": "total_revenue", "desc": true},
        ],
    }
`

	tool := mcp.NewToolWithRawSchema("get_metrics_view_aggregation", description, json.RawMessage(metricsview.QueryJSONSchema))

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		instanceID := "default" // TODO: Dynamic
		metricsProps, ok := req.GetRawArguments().(map[string]any)
		if !ok {
			return mcpNewToolErrorf("invalid arguments: expected an object")
		}

		claims := auth.GetClaims(ctx)
		if !claims.CanInstance(instanceID, auth.ReadMetrics) {
			return mcpErrorFromGRPC(ErrForbidden)
		}

		res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID:         instanceID,
			Resolver:           "metrics",
			ResolverProperties: metricsProps,
			Claims:             claims.SecurityClaims(),
		})
		if err != nil {
			return mcpErrorFromGRPC(err)
		}
		defer res.Close()

		data, err := res.MarshalJSON()
		if err != nil {
			return mcpNewToolError(err)
		}

		return mcp.NewToolResultText(string(data)), nil
	}

	return tool, handler
}

func mcpNewToolResultJSON(val any) (*mcp.CallToolResult, error) {
	var data []byte
	var err error
	if msg, ok := val.(proto.Message); ok {
		data, err = protojson.Marshal(msg)
	} else {
		data, err = json.Marshal(val)
	}
	if err != nil {
		return nil, fmt.Errorf("internal: failed to marshal metrics view names: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

func mcpNewToolError(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(err.Error()), nil
}

func mcpNewToolErrorf(format string, args ...any) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(fmt.Sprintf(format, args...)), nil
}

func mcpErrorFromGRPC(err error) (*mcp.CallToolResult, error) {
	if err == nil {
		return nil, fmt.Errorf("mcpErrorFromGRPC: nil error")
	}

	err = mapGRPCError(err)

	if s, ok := status.FromError(err); ok {
		if s.Code() == codes.Internal {
			return nil, fmt.Errorf("internal: %s", s.Message())
		}
		msg := fmt.Sprintf("%s: %s", s.Code(), s.Message())
		return mcp.NewToolResultError(msg), nil
	}

	msg := err.Error()
	return mcp.NewToolResultError(msg), nil
}
