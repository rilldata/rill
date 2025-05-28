package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const mcpInstructions = `
## Rill MCP Server
This server exposes APIs for querying **metrics views**, which represent Rill's metrics layer.
### Workflow Overview
1. **List metrics views:** Use "list_metrics_views" to discover available metrics views in the project.
2. **Get metrics view spec:** Use "get_metrics_view" to fetch a metrics view's specification. This is important to understand all the dimensions and measures in a metrics view.
3. **Query the time range:** Use "query_metrics_view_time_range" to obtain the available time range for a metrics view. This is important to understand what time range the data spans.
4. **Query the metrics:** Use "query_metrics_view" to run queries to get aggregated results.
In the workflow, do not proceed with the next step until the previous step has been completed. If the information from the previous step is already known (let's say for subsequent queries), you can skip it.
`

func (s *Server) newMCPServer() *server.SSEServer {
	version := s.runtime.Version().Number
	if version == "" {
		version = "0.0.1"
	}

	mcpServer := server.NewMCPServer("rill", version,
		server.WithToolHandlerMiddleware(observability.MCPToolHandlerMiddleware()),
		server.WithToolHandlerMiddleware(mcpErrorMappingMiddleware),
		server.WithToolHandlerMiddleware(middleware.TimeoutMCPToolHandlerMiddleware(func(tool string) time.Duration {
			switch tool {
			case "query_metrics_view_time_range", "query_metrics_view":
				return 120 * time.Second
			default:
				return 20 * time.Second
			}
		})),
		server.WithRecovery(),
		server.WithToolCapabilities(true),
		server.WithInstructions(mcpInstructions),
	)

	mcpServer.AddTool(s.mcpListMetricsViews())
	mcpServer.AddTool(s.mcpGetMetricsView())
	mcpServer.AddTool(s.mcpQueryMetricsViewTimeRange())
	mcpServer.AddTool(s.mcpQueryMetricsView())

	sseServer := server.NewSSEServer(
		mcpServer,
		server.WithHTTPContextFunc(s.mcpHTTPContextFunc),
		server.WithUseFullURLForMessageEndpoint(false),
		server.WithDynamicBasePath(func(r *http.Request, sessionID string) string {
			// We don't know the base path because the MCP handler can be served from three base URLs:
			//   1. <runtime>/mcp
			//   2. <runtime>/v1/instances/{instance_id}/mcp
			//   3. <admin>/<runtime proxy path>/mcp

			// Get the base path.
			basePath := r.URL.Path

			// If the call was proxied from the admin runtime proxy, use the original base path.
			if originalURI := r.Header.Get("X-Original-URI"); originalURI != "" {
				parsedURL, err := url.Parse(originalURI)
				if err == nil {
					basePath = parsedURL.Path
				}
			}

			// We know the path ends with /mcp/sse or /mcp/message and we want to return the path up to /mcp.
			// So we keep cutting off the last segment until the path ends with "mcp".
			// Just to be extra safe, we limit the lookback to five iterations.
			basePath = path.Clean(basePath)
			for i := 0; i < 5; i++ {
				if path.Base(basePath) == "mcp" || len(basePath) <= 1 {
					break
				}
				basePath = path.Dir(basePath) // Cut off the last path segment
			}

			return basePath
		}),
	)

	return sseServer
}

func (s *Server) mcpListMetricsViews() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("list_metrics_views",
		mcp.WithDescription("List all metrics views in the current project"),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := s.ListResources(ctx, &runtimev1.ListResourcesRequest{
			InstanceId: mcpInstanceIDFromContext(ctx),
			Kind:       runtime.ResourceKindMetricsView,
		})
		if err != nil {
			return nil, err
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
		mcp.WithDescription("Get the specification for a given metrics view, including available measures and dimensions"),
		mcp.WithString("metrics_view",
			mcp.Required(),
			mcp.Description("Name of the metrics view"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("metrics_view")
		if err != nil {
			return nil, err
		}

		resp, err := s.GetResource(ctx, &runtimev1.GetResourceRequest{
			InstanceId: mcpInstanceIDFromContext(ctx),
			Name: &runtimev1.ResourceName{
				Kind: runtime.ResourceKindMetricsView,
				Name: name,
			},
		})
		if err != nil {
			return nil, err
		}

		mv := resp.Resource.GetMetricsView()
		if mv == nil || mv.State.ValidSpec == nil {
			return nil, fmt.Errorf("metrics view %q not valid", name)
		}

		return mcpNewToolResultJSON(mv.State.ValidSpec)
	}

	return tool, handler
}

func (s *Server) mcpQueryMetricsViewTimeRange() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("query_metrics_view_time_range",
		mcp.WithDescription(`
            Retrieve the total time range available for a given metrics view.
            Note: All subsequent queries of the metrics view should be constrained to this time range to ensure accurate results.
        `),
		mcp.WithString("metrics_view",
			mcp.Required(),
			mcp.Description("Name of the metrics view"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		instanceID := mcpInstanceIDFromContext(ctx)
		name, err := req.RequireString("metrics_view")
		if err != nil {
			return nil, err
		}

		claims := auth.GetClaims(ctx)
		if !claims.CanInstance(instanceID, auth.ReadMetrics) {
			return nil, ErrForbidden
		}

		res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID: instanceID,
			Resolver:   "metrics_time_range",
			ResolverProperties: map[string]any{
				"metrics_view": name,
			},
			Claims: claims.SecurityClaims(),
		})
		if err != nil {
			return nil, err
		}
		defer res.Close()

		data, err := res.MarshalJSON()
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(string(data)), nil
	}

	return tool, handler
}

func (s *Server) mcpQueryMetricsView() (mcp.Tool, server.ToolHandlerFunc) {
	description := `
Perform an arbitrary aggregation on a metrics view.
Tip: Use the 'sort' and 'limit' parameters for best results and to avoid large, unbounded result sets.
Important note: The 'time_range' parameter is inclusive of the start time and exclusive of the end time.

Example: Get the total revenue by country and product category for 2024:
    {
        "metrics_view": "ecommerce_financials",
        "dimensions": [{"name": "country"}, {"name": "product_category"}],
        "measures": [{"name": "total_revenue"}, {"name": "total_orders"}],
        "time_range": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2025-01-01T00:00:00Z"
        },
        "where": {
            "cond": {
                "op": "and",
                "exprs": [
                    {
                        "cond": {
                            "op": "in",
                            "exprs": [
                                {"name": "country"},
                                {"val": ["US", "CA", "GB"]}
                            ]
                        }
                    },
                    {
                        "cond": {
                            "op": "eq",
                            "exprs": [
                                {"name": "product_category"},
                                {"val": "Electronics"}
                            ]
                        }
                    }
                ]
            },
        },
        "sort": [{"name": "total_revenue", "desc": true}],
        "limit": 10
    }
    
Example: Get the total revenue by country and month for 2024:
    {
        "metrics_view": "ecommerce_financials",
        "dimensions": [
            {"name": "event_time", "compute": {"time_floor": {"dimension": "event_time", "grain": "month"}}},
            {"name": "country"},
        ],
        "measures": [{"name": "total_revenue"}],
        "time_range": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2025-01-01T00:00:00Z"
        },
        "sort": [
            {"name": "event_time"},
            {"name": "total_revenue", "desc": true},
        ],
    }

Example: Get the top 10 demographic segments (by country, gender, and age group) with the largest absolute revenue difference comparing May 2025 (base period) to April 2025 (comparison period):
		{
			"metrics_view": "ecommerce_financials",
			"measures": [
				{"name": "total_revenue"},
				{"name": "total_revenue__delta_abs", "compute": {"comparison_delta": {"measure": "total_revenue"}}},
				{"name": "total_revenue__delta_rel", "compute": {"comparison_ratio": {"measure": "total_revenue"}}},
			],
			"dimensions": [{"name": "country"}, {"name": "gender"}, {"name": "age_group"}],
			"time_range": {
				"start": "2025-05-01T00:00:00Z",
				"end": "2025-05-31T23:59:59Z"
			},
			"comparison_time_range": {
				"start": "2025-04-01T00:00:00Z",
				"end": "2025-04-30T23:59:59Z"
			},
			"sort": [{"name": "total_revenue__delta_abs", "desc": true}],
			"limit": 10
		}
`

	tool := mcp.NewToolWithRawSchema("query_metrics_view", description, json.RawMessage(metricsview.QueryJSONSchema))

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		instanceID := mcpInstanceIDFromContext(ctx)
		metricsProps, ok := req.GetRawArguments().(map[string]any)
		if !ok {
			return nil, errors.New("invalid arguments: expected an object")
		}

		claims := auth.GetClaims(ctx)
		if !claims.CanInstance(instanceID, auth.ReadMetrics) {
			return nil, ErrForbidden
		}

		res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID:         instanceID,
			Resolver:           "metrics",
			ResolverProperties: metricsProps,
			Claims:             claims.SecurityClaims(),
		})
		if err != nil {
			return nil, err
		}
		defer res.Close()

		data, err := res.MarshalJSON()
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(string(data)), nil
	}

	return tool, handler
}

// mcpHTTPContextFunc is an MCP server middleware that adds the current instance ID to the context.
func (s *Server) mcpHTTPContextFunc(ctx context.Context, r *http.Request) context.Context {
	// Extract instance ID from the request path
	instanceID := r.PathValue("instance_id")
	if instanceID == "" {
		// We also mount the MCP server on <root>/mcp to make it easier to use in Rill Developer (on localhost).
		// In those settings, we pick the default instance ID.
		// This is safe because if there is no default instance, it'll just be the empty string and requests will error with "not found".
		instanceID, _ = s.runtime.DefaultInstanceID()
	}

	// Store instance ID in context for later use
	return context.WithValue(ctx, mcpInstanceIDKey{}, instanceID)
}

// mcpInstanceIDKey is a context key used to store the instance ID for the current MCP server request.
type mcpInstanceIDKey struct{}

// mcpInstanceIDFromContext retrieves the instance ID from the context.
// Only works for MCP server contexts (i.e. requests wrapped with mcpHTTPContextFunc).
func mcpInstanceIDFromContext(ctx context.Context) string {
	instanceID, _ := ctx.Value(mcpInstanceIDKey{}).(string)
	return instanceID
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
		return nil, mcpNewInternalError(fmt.Errorf("internal: failed to marshal metrics view names: %w", err))
	}
	return mcp.NewToolResultText(string(data)), nil
}

func mcpErrorMappingMiddleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := next(ctx, req)
		if err == nil {
			return res, nil
		}

		// Handle internal MCP errors.
		var internalErr mcpInternalError
		if errors.As(err, &internalErr) {
			return nil, internalErr.err
		}

		// Leverage our gRPC error mapper to avoid duplicating mapping of common errors.
		err = mapGRPCError(err)
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.Internal {
				return nil, fmt.Errorf("internal: %s", s.Message())
			}
			msg := fmt.Sprintf("%s: %s", s.Code(), s.Message())
			return mcp.NewToolResultError(msg), nil
		}

		// Default to returning as a user error.
		msg := err.Error()
		return mcp.NewToolResultError(msg), nil
	}
}

type mcpInternalError struct {
	err error
}

func mcpNewInternalError(err error) error {
	if err == nil {
		err = fmt.Errorf("internal: nil error")
	}
	return mcpInternalError{err: err}
}

func (e mcpInternalError) Error() string {
	return fmt.Sprintf("internal: %s", e.err.Error())
}
