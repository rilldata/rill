package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	mcputil "github.com/mark3labs/mcp-go/util"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/middleware"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const mcpInstructions = `
# Rill MCP Server
This server exposes APIs for querying **metrics views**, which represent Rill's metrics layer.

## Tool Availability
The available tools depend on your context:

### Dashboard Context (Context-Aware Mode)
When you're in a dashboard context, you have access to:
- **get_metrics_view**: Get detailed specifications for the current metrics view (dimensions, measures, descriptions)
- **query_metrics_view_with_context**: Execute queries that automatically use current dashboard context (filters, time ranges)

### Generic Context (Discovery Mode)  
When you're in a generic context, you have access to discovery tools:
- **list_metrics_views**: Discover available metrics views in the project
- **get_metrics_view**: Get detailed specifications for a metrics view
- **query_metrics_view_summary**: Get time range and sample data for a metrics view
- **query_metrics_view**: Execute queries on a metrics view

## Usage Guidelines

### In Dashboard Context
1. **First, call "get_metrics_view"** to discover available dimensions and measures
2. **Then use "query_metrics_view_with_context"** for data queries - it automatically applies dashboard context

### In Generic Context
Follow the step-by-step workflow:
1. List metrics views to discover available data
2. Get metrics view specifications to understand dimensions and measures  
3. Query summary to understand time ranges and data types
4. Execute queries to get results

If a response contains an "ai_instructions" field, interpret it as additional guidance for subsequent interactions.
`

func (s *Server) newMCPServer() *server.MCPServer {
	version := s.runtime.Version().Number
	if version == "" {
		version = "0.0.1"
	}

	mcpServer := server.NewMCPServer("rill", version,
		server.WithToolHandlerMiddleware(observability.MCPToolHandlerMiddleware()),
		server.WithToolHandlerMiddleware(mcpErrorMappingMiddleware),
		server.WithToolHandlerMiddleware(middleware.TimeoutMCPToolHandlerMiddleware(func(tool string) time.Duration {
			switch tool {
			case "query_metrics_view_summary", "query_metrics_view":
				return 120 * time.Second
			default:
				return 20 * time.Second
			}
		})),
		server.WithRecovery(),
		server.WithToolCapabilities(true),
		server.WithInstructions(mcpInstructions),
	)

	// Rill capabilities
	mcpServer.AddTool(s.mcpListMetricsViews())
	mcpServer.AddTool(s.mcpGetMetricsView())
	mcpServer.AddTool(s.mcpQueryMetricsView())
	mcpServer.AddTool(s.mcpQueryMetricsViewSummary())
	mcpServer.AddTool(s.mcpQueryMetricsViewWithContext())

	return mcpServer
}

func (s *Server) newMCPHTTPHandler(mcpServer *server.MCPServer) http.Handler {
	httpServer := server.NewStreamableHTTPServer(
		mcpServer,
		server.WithHeartbeatInterval(30*time.Second),
		server.WithHTTPContextFunc(s.mcpHTTPContextFunc),
		server.WithStateLess(true), // NOTE: Need to change if we start using notifications.
		server.WithLogger(mcpLogger{s.logger}),
	)

	return httpServer
}

type mcpLogger struct {
	logger *zap.Logger
}

var _ mcputil.Logger = mcpLogger{}

func (l mcpLogger) Infof(msg string, args ...any) {
	l.logger.Info("mcp: info log", zap.String("msg", fmt.Sprintf(msg, args...)))
}

func (l mcpLogger) Errorf(msg string, args ...any) {
	l.logger.Warn("mcp: error log", zap.String("msg", fmt.Sprintf(msg, args...)))
}

func (s *Server) mcpListMetricsViews() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("list_metrics_views",
		mcp.WithDescription("List all metrics views in the current project"),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		instanceID := mcpInstanceIDFromContext(ctx)
		resp, err := s.ListResources(ctx, &runtimev1.ListResourcesRequest{
			InstanceId: instanceID,
			Kind:       runtime.ResourceKindMetricsView,
		})
		if err != nil {
			return nil, err
		}

		res := make(map[string]any)

		// Find instance-wide AI context and add it to the response.
		// NOTE: These arguably belong in the top-level instructions or other metadata, but that doesn't currently support dynamic values.
		instance, err := s.runtime.Instance(ctx, instanceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get instance %q: %w", instanceID, err)
		}
		if instance.AIInstructions != "" {
			res["ai_instructions"] = instance.AIInstructions
		}

		var metricsViews []map[string]any
		for _, r := range resp.Resources {
			mv := r.GetMetricsView()
			if mv == nil || mv.State.ValidSpec == nil {
				continue
			}

			metricsViews = append(metricsViews, map[string]any{
				"name":         r.Meta.Name.Name,
				"display_name": mv.State.ValidSpec.DisplayName,
				"description":  mv.State.ValidSpec.Description,
			})
		}
		res["metrics_views"] = metricsViews

		return mcpNewToolResultJSON(res)
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

func (s *Server) mcpQueryMetricsViewSummary() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("query_metrics_view_summary",
		mcp.WithDescription(`
			Retrieve summary statistics for a metrics view including:
			- Total time range available
			- Sample values and data types for each dimension
			Note: All subsequent queries of the metrics view should be constrained to this time range to ensure accurate results.
		`),
		mcp.WithString("metrics_view",
			mcp.Required(),
			mcp.Description("Name of the metrics view"),
		),
	)
	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		instanceID := mcpInstanceIDFromContext(ctx)
		metricsViewName, err := req.RequireString("metrics_view")
		if err != nil {
			return nil, err
		}

		claims := auth.GetClaims(ctx)
		if !claims.CanInstance(instanceID, auth.ReadMetrics) {
			return nil, ErrForbidden
		}

		res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID: instanceID,
			Resolver:   "metrics_summary",
			ResolverProperties: map[string]any{
				"metrics_view": metricsViewName,
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
Note: 'time_dimension' is an optional parameter under "time_range" that can be used to specify the time dimension to use for the time range. If not provided, the default time column of the metrics view will be used.

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

Example: Get the total revenue by country and month for order shipped in 2024:
    {
        "metrics_view": "ecommerce_financials",
        "dimensions": [
            {"name": "event_time", "compute": {"time_floor": {"dimension": "event_time", "grain": "month"}}},
            {"name": "country"}
        ],
        "measures": [{"name": "total_revenue"}],
        "time_range": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2025-01-01T00:00:00Z",
            "time_dimension"": "order_shipped_time",
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

		// Get the raw response data
		data, err := res.MarshalJSON()
		if err != nil {
			return nil, err
		}

		// Generate an open URL for the query
		openURL, err := s.generateOpenURL(ctx, instanceID, metricsProps)
		if err != nil {
			return nil, fmt.Errorf("failed to generate open URL: %w", err)
		}

		// Add both the open URL and metricsProps to the response
		response, err := s.addOpenURLAndMetricsPropsToResponse(data, openURL, metricsProps)
		if err != nil {
			return nil, fmt.Errorf("failed to add open URL and metricsProps to response: %w", err)
		}

		return mcp.NewToolResultText(string(response)), nil
	}

	return tool, handler
}

func (s *Server) mcpQueryMetricsViewWithContext() (mcp.Tool, server.ToolHandlerFunc) {
	description := `
Query the current dashboard's metrics view with automatic context awareness.

This tool automatically uses the current dashboard context (metrics view, filters, time ranges, visible measures/dimensions) to execute your query efficiently.

Simply provide the query parameters directly - do NOT include context information as it's automatically provided.

If you need to discover available dimensions and measures, use the "get_metrics_view" tool first to see what's available.

Parameters (same format as query_metrics_view):
- dimensions: Array of dimension objects to group by
- measures: Array of measure objects to aggregate  
- sort: Array of sort objects for ordering results
- limit: Maximum number of results to return
- where: Optional filter conditions (will be merged with dashboard filters)
- time_range: Optional time range (will use dashboard time range if not specified)

Example usage:
{
  "dimensions": [{"name": "product_category"}],
  "measures": [{"name": "total_sales"}], 
  "sort": [{"name": "total_sales", "desc": true}],
  "limit": 10
}

The tool automatically:
- Uses the current dashboard's metrics view
- Applies current dashboard filters and time ranges
- Merges your query with the dashboard context for accurate results
`

	// Use the same schema as query_metrics_view since context is automatically injected
	tool := mcp.NewToolWithRawSchema("query_metrics_view_with_context", description, json.RawMessage(metricsview.QueryJSONSchema))

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		instanceID := mcpInstanceIDFromContext(ctx)

		// Parse the request arguments as a direct query (context will be injected by mcpExecuteTool)
		queryObj, ok := req.GetRawArguments().(map[string]any)
		if !ok {
			return nil, errors.New("invalid arguments: expected a query object")
		}

		// For this handler, we expect the context to be already injected by mcpExecuteTool
		// Check if we have the injected context structure
		var contextObj map[string]any
		var metricsViewName string
		var autoGatherContext bool = true

		if injectedContext, hasContext := queryObj["context"]; hasContext {
			// Context was injected by mcpExecuteTool
			if ctx, ok := injectedContext.(map[string]any); ok {
				contextObj = ctx
				if mv, ok := ctx["metrics_view"].(string); ok {
					metricsViewName = mv
				}
			}
			// Extract the actual query from the injected structure
			if injectedQuery, hasQuery := queryObj["query"]; hasQuery {
				if q, ok := injectedQuery.(map[string]any); ok {
					queryObj = q
				}
			}
		}

		// If no context was injected, this is an error for the context-aware tool
		if metricsViewName == "" {
			return nil, errors.New("this tool requires dashboard context - metrics view not found")
		}

		claims := auth.GetClaims(ctx)
		if !claims.CanInstance(instanceID, auth.ReadMetrics) {
			return nil, ErrForbidden
		}

		// Prepare the response with gathered context
		response := map[string]any{
			"query_result": nil,
			"context_used": map[string]any{
				"metrics_view":  metricsViewName,
				"auto_gathered": autoGatherContext,
			},
		}

		// Auto-gather context if enabled and not provided
		if autoGatherContext {
			// Get metrics view specification
			mvResp, err := s.GetResource(ctx, &runtimev1.GetResourceRequest{
				InstanceId: instanceID,
				Name: &runtimev1.ResourceName{
					Kind: runtime.ResourceKindMetricsView,
					Name: metricsViewName,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get metrics view spec: %w", err)
			}

			mv := mvResp.Resource.GetMetricsView()
			if mv == nil || mv.State.ValidSpec == nil {
				return nil, fmt.Errorf("metrics view %q not valid", metricsViewName)
			}

			response["context_used"].(map[string]any)["metrics_view_spec"] = mv.State.ValidSpec

			// Get time range summary
			summaryRes, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
				InstanceID: instanceID,
				Resolver:   "metrics_summary",
				ResolverProperties: map[string]any{
					"metrics_view": metricsViewName,
				},
				Claims: claims.SecurityClaims(),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get time range summary: %w", err)
			}
			defer summaryRes.Close()

			summaryData, err := summaryRes.MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal summary: %w", err)
			}

			var summaryResult any
			if err := json.Unmarshal(summaryData, &summaryResult); err != nil {
				return nil, fmt.Errorf("failed to parse summary: %w", err)
			}

			response["context_used"].(map[string]any)["time_range_summary"] = summaryResult
		}

		// Merge current_state from context into the query if provided
		if currentState, exists := contextObj["current_state"].(map[string]any); exists {
			// Merge time_range if provided and not already in query
			if timeRange, exists := currentState["time_range"]; exists && queryObj["time_range"] == nil {
				queryObj["time_range"] = timeRange
			}

			// Merge filters if provided
			if filters, exists := currentState["filters"]; exists && filters != nil {
				fmt.Printf("Dashboard filters received: %+v\n", filters)

				// Convert the dashboard filters to a proper runtimev1.Expression
				dashboardExpr, err := convertToRuntimeExpression(filters)
				if err != nil {
					fmt.Printf("Failed to convert dashboard filters: %v\n", err)
				} else if dashboardExpr != nil {
					// Check if there's already a where clause in the query
					if existingWhere, hasWhere := queryObj["where"]; hasWhere {
						// Convert existing where to runtimev1.Expression
						existingExpr, err := convertToRuntimeExpression(existingWhere)
						if err != nil {
							fmt.Printf("Failed to convert existing where clause: %v\n", err)
						} else {
							// Combine both expressions with AND
							combinedExpr := &runtimev1.Expression{
								Expression: &runtimev1.Expression_Cond{
									Cond: &runtimev1.Condition{
										Op:    runtimev1.Operation_OPERATION_AND,
										Exprs: []*runtimev1.Expression{existingExpr, dashboardExpr},
									},
								},
							}

							// Convert back to metricsview expression and then to map for the resolver
							mvExpr := metricsview.NewExpressionFromProto(combinedExpr)
							queryObj["where"] = convertMetricsViewExpressionToMap(mvExpr)
						}
					} else {
						// Use dashboard filters as the only where clause
						mvExpr := metricsview.NewExpressionFromProto(dashboardExpr)
						queryObj["where"] = convertMetricsViewExpressionToMap(mvExpr)
					}
					fmt.Printf("Applied dashboard filters using proper expression conversion\n")
				}
			}

			response["context_used"].(map[string]any)["current_state"] = currentState
		}

		// Add the metrics_view to the query object
		queryObj["metrics_view"] = metricsViewName

		// Execute the actual metrics query
		res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID:         instanceID,
			Resolver:           "metrics",
			ResolverProperties: queryObj,
			Claims:             claims.SecurityClaims(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
		defer res.Close()

		// Get the raw response data
		data, err := res.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal query result: %w", err)
		}

		var queryResult any
		if err := json.Unmarshal(data, &queryResult); err != nil {
			return nil, fmt.Errorf("failed to parse query result: %w", err)
		}

		response["query_result"] = queryResult

		// Generate an open URL for the query
		openURL, err := s.generateOpenURL(ctx, instanceID, queryObj)
		if err != nil {
			return nil, fmt.Errorf("failed to generate open URL: %w", err)
		}

		response["open_url"] = openURL
		response["metricsProps"] = queryObj

		// Marshal the complete response
		responseData, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return mcp.NewToolResultText(string(responseData)), nil
	}

	return tool, handler
}

// generateOpenURL generates an open URL for the given query parameters
func (s *Server) generateOpenURL(ctx context.Context, instanceID string, metricsProps map[string]any) (string, error) {
	// Get instance to access the configured frontend URL
	instance, err := s.runtime.Instance(ctx, instanceID)
	if err != nil {
		return "", fmt.Errorf("failed to get instance: %w", err)
	}

	// If there's no frontend URL (e.g. perhaps in test cases or during rollout), return an empty string
	if instance.FrontendURL == "" {
		return "", nil
	}

	// Build the complete URL for the query
	jsonBytes, err := json.Marshal(metricsProps)
	if err != nil {
		return "", fmt.Errorf("failed to marshal MCP query to JSON: %w", err)
	}

	values := make(url.Values)
	values.Set("query", string(jsonBytes))

	return fmt.Sprintf("%s/-/open-query?%s", instance.FrontendURL, values.Encode()), nil
}

// addOpenURLAndMetricsPropsToResponse adds both the open URL and metricsProps to the response data
func (s *Server) addOpenURLAndMetricsPropsToResponse(data []byte, openURL string, metricsProps map[string]any) ([]byte, error) {
	// Parse the JSON response to understand its structure
	var response any
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Create a wrapper object with the response data, open URL, and metricsProps
	wrappedResponse := map[string]any{
		"response":     response,
		"open_url":     openURL,
		"metricsProps": metricsProps,
	}

	// Marshal back to JSON
	modifiedData, err := json.Marshal(wrappedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified response: %w", err)
	}

	return modifiedData, nil
}

// addMetricsPropsToResponse adds the metricsProps to the response data
func (s *Server) addMetricsPropsToResponse(data []byte, metricsProps map[string]any) ([]byte, error) {
	// Parse the JSON response to understand its structure
	var response any
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Create a wrapper object with the response data and metricsProps
	wrappedResponse := map[string]any{
		"response":     response,
		"metricsProps": metricsProps,
	}

	// Marshal back to JSON
	modifiedData, err := json.Marshal(wrappedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified response: %w", err)
	}

	return modifiedData, nil
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

// convertOperationEnumToString converts V1Operation enum constants to the string values expected by the query engine
func convertOperationEnumToString(op any) string {
	if opStr, ok := op.(string); ok {
		switch opStr {
		case "OPERATION_AND":
			return "and"
		case "OPERATION_OR":
			return "or"
		case "OPERATION_EQ":
			return "eq"
		case "OPERATION_NEQ":
			return "neq"
		case "OPERATION_LT":
			return "lt"
		case "OPERATION_LTE":
			return "lte"
		case "OPERATION_GT":
			return "gt"
		case "OPERATION_GTE":
			return "gte"
		case "OPERATION_IN":
			return "in"
		case "OPERATION_NIN":
			return "nin"
		case "OPERATION_LIKE":
			return "ilike"
		case "OPERATION_NLIKE":
			return "nilike"
		default:
			// If it's already a lowercase string, return as-is
			return opStr
		}
	}
	return ""
}

// normalizeFilterExpression recursively converts operation enum constants to strings in filter expressions
func normalizeFilterExpression(expr any) any {
	if exprMap, ok := expr.(map[string]any); ok {
		result := make(map[string]any)
		for key, value := range exprMap {
			if key == "cond" {
				if condMap, ok := value.(map[string]any); ok {
					normalizedCond := make(map[string]any)
					for condKey, condValue := range condMap {
						if condKey == "op" {
							normalizedCond[condKey] = convertOperationEnumToString(condValue)
						} else if condKey == "exprs" {
							if exprs, ok := condValue.([]any); ok {
								// Check if this is an IN or NIN operation that needs value consolidation
								op := convertOperationEnumToString(condMap["op"])
								if (op == "in" || op == "nin") && len(exprs) > 2 {
									fmt.Printf("Normalizing %s operation with %d expressions\n", op, len(exprs))
									// Consolidate multiple value expressions into a single array
									normalizedExprs := make([]any, 2)
									normalizedExprs[0] = normalizeFilterExpression(exprs[0]) // First expr (the field)

									// Collect all values from remaining expressions
									values := make([]any, 0, len(exprs)-1)
									for i := 1; i < len(exprs); i++ {
										if valExpr, ok := exprs[i].(map[string]any); ok {
											if val, hasVal := valExpr["val"]; hasVal {
												values = append(values, val)
											}
										}
									}

									// Create a single value expression with the array
									normalizedExprs[1] = map[string]any{"val": values}
									normalizedCond[condKey] = normalizedExprs
									fmt.Printf("Consolidated %s values: %+v\n", op, values)
								} else {
									// Normal case - just normalize each expression
									normalizedExprs := make([]any, len(exprs))
									for i, subExpr := range exprs {
										normalizedExprs[i] = normalizeFilterExpression(subExpr)
									}
									normalizedCond[condKey] = normalizedExprs
								}
							} else {
								normalizedCond[condKey] = condValue
							}
						} else {
							normalizedCond[condKey] = condValue
						}
					}
					result[key] = normalizedCond
				} else {
					result[key] = value
				}
			} else {
				result[key] = value
			}
		}
		return result
	}
	return expr
}

// convertToRuntimeExpression converts a map[string]any to *runtimev1.Expression
func convertToRuntimeExpression(exprMap any) (*runtimev1.Expression, error) {
	if exprMap == nil {
		return nil, nil
	}

	// First normalize the expression to fix enum constants and IN operations
	normalized := normalizeFilterExpression(exprMap)

	// Convert to JSON and then unmarshal to runtimev1.Expression
	jsonBytes, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal expression: %w", err)
	}

	var expr runtimev1.Expression
	err = protojson.Unmarshal(jsonBytes, &expr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal to runtimev1.Expression: %w", err)
	}

	return &expr, nil
}

// convertMetricsViewExpressionToMap converts a metricsview.Expression to map[string]any for the resolver
func convertMetricsViewExpressionToMap(expr *metricsview.Expression) map[string]any {
	if expr == nil {
		return nil
	}

	result := make(map[string]any)

	if expr.Name != "" {
		result["ident"] = expr.Name
	}

	if expr.Value != nil {
		result["val"] = expr.Value
	}

	if expr.Condition != nil {
		cond := make(map[string]any)

		// Convert operator enum to string
		switch expr.Condition.Operator {
		case metricsview.OperatorAnd:
			cond["op"] = "and"
		case metricsview.OperatorOr:
			cond["op"] = "or"
		case metricsview.OperatorEq:
			cond["op"] = "eq"
		case metricsview.OperatorNeq:
			cond["op"] = "neq"
		case metricsview.OperatorLt:
			cond["op"] = "lt"
		case metricsview.OperatorLte:
			cond["op"] = "lte"
		case metricsview.OperatorGt:
			cond["op"] = "gt"
		case metricsview.OperatorGte:
			cond["op"] = "gte"
		case metricsview.OperatorIn:
			cond["op"] = "in"
		case metricsview.OperatorNin:
			cond["op"] = "nin"
		case metricsview.OperatorIlike:
			cond["op"] = "ilike"
		case metricsview.OperatorNilike:
			cond["op"] = "nilike"
		}

		if len(expr.Condition.Expressions) > 0 {
			exprs := make([]any, len(expr.Condition.Expressions))
			for i, subExpr := range expr.Condition.Expressions {
				exprs[i] = convertMetricsViewExpressionToMap(subExpr)
			}
			cond["exprs"] = exprs
		}

		result["cond"] = cond
	}

	if expr.Subquery != nil {
		// Handle subquery if needed
		subquery := make(map[string]any)
		if expr.Subquery.Dimension.Name != "" {
			subquery["dimension"] = expr.Subquery.Dimension.Name
		}
		if len(expr.Subquery.Measures) > 0 {
			measures := make([]string, len(expr.Subquery.Measures))
			for i, m := range expr.Subquery.Measures {
				measures[i] = m.Name
			}
			subquery["measures"] = measures
		}
		if expr.Subquery.Where != nil {
			subquery["where"] = convertMetricsViewExpressionToMap(expr.Subquery.Where)
		}
		if expr.Subquery.Having != nil {
			subquery["having"] = convertMetricsViewExpressionToMap(expr.Subquery.Having)
		}
		result["subquery"] = subquery
	}

	return result
}
