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
## Workflow Overview
1. **List metrics views:** Use "list_metrics_views" to discover available metrics views in the project.
2. **Get metrics view spec:** Use "get_metrics_view" to fetch a metrics view's specification. This is important to understand all the dimensions and measures in a metrics view.
3. **Query the summary:** Use "query_metrics_view_summary" to obtain the available time range for a metrics view and sample values with their data types for each dimension. This provides a richer context for understanding the data.
4. **Query the metrics:** Use "query_metrics_view" to run queries to get aggregated results.
5. **Create a chart:** After running "query_metrics_view" create a chart using "create_chart" unless:
   - The user explicitly requests a table-only response
   - The query returns only a single scalar value
   - The data structure doesn't lend itself to visualization (e.g., text-heavy data)
	 - There is no appropriate chart type which can be created for the underlying data

In the workflow, do not proceed with the next step until the previous step has been completed. If the information from the previous step is already known (let's say for subsequent queries), you can skip it.
If a response contains an "ai_instructions" field, you should interpret it as additional instructions for how to behave in subsequent responses that relate to that tool call.

## Visualization Best Practices
Choose the appropriate chart type based on your data:
- Time series data → line_chart or area_chart (better for cummalative trends)
- Category comparisons → bar_chart
- Part-to-whole relationships → donut_chart
- Multiple dimensions → Use color encoding or combo_chart
- Multiple measures (more that 2) -> Use stacked bar chart with multiple measure fields
- Distribution across two dimensions → heatmap
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
	mcpServer.AddTool(s.mcpCreateChart())

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

		claims := auth.GetClaims(ctx, instanceID)
		if !claims.Can(runtime.ReadMetrics) {
			return nil, ErrForbidden
		}

		res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID: instanceID,
			Resolver:   "metrics_summary",
			ResolverProperties: map[string]any{
				"metrics_view": metricsViewName,
			},
			Claims: claims,
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

Parameter notes:
• The 'time_range' parameter is inclusive of the start time and exclusive of the end time
• 'time_dimension' is optional under 'time_range' to specify which time column to use (defaults to the metrics view's default time column)

Best practices:
• Use 'sort' and 'limit' parameters for best results and to avoid large, unbounded result sets
• For comparison queries: ensure 'time_range' and 'comparison_time_range' are non-overlapping and similar in duration (~20% tolerance) to ensure valid period-over-period comparisons

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

		claims := auth.GetClaims(ctx, instanceID)
		if !claims.Can(runtime.ReadMetrics) {
			return nil, ErrForbidden
		}

		res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID:         instanceID,
			Resolver:           "metrics",
			ResolverProperties: metricsProps,
			Claims:             claims,
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

		// Add the open URL to the response
		response, err := s.addOpenURLToResponse(data, openURL)
		if err != nil {
			return nil, fmt.Errorf("failed to add open URL to response: %w", err)
		}

		return mcp.NewToolResultText(string(response)), nil
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

// addOpenURLToResponse adds the open URL to the response data
func (s *Server) addOpenURLToResponse(data []byte, openURL string) ([]byte, error) {
	// Parse the JSON response to understand its structure
	var response any
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Create a wrapper object with the response data and open URL
	wrappedResponse := map[string]any{
		"response": response,
		"open_url": openURL,
	}

	// Marshal back to JSON
	modifiedData, err := json.Marshal(wrappedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified response: %w", err)
	}

	return modifiedData, nil
}

func (s *Server) mcpCreateChart() (mcp.Tool, server.ToolHandlerFunc) {
	description := `# Chart Visualization Tool

Create visualization charts based on metrics views. This tool generates chart specifications that will be rendered in the chat interface.

## Required Parameters

All chart specifications must include:
- ` + "`chart_type`" + `: The type of visualization to create
- ` + "`spec`" + `: The chart specification object containing:
  - ` + "`metrics_view`" + `: The name of the metrics view to query
  - ` + "`time_range`" + `: **Required** - Time range for the data query
    - ` + "`start`" + `: ISO 8601 timestamp (inclusive)
    - ` + "`end`" + `: ISO 8601 timestamp (exclusive)
    - Example: ` + "`\"start\": \"2024-01-01T00:00:00Z\", \"end\": \"2024-12-31T23:59:59Z\"`" + `

## Supported Chart Types

### 1. Bar Chart (` + "`bar_chart`" + `)
**Use for:** Comparing values across different categories

Example Specification
` + "```json" + `
{
  "chart_type": "bar_chart",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": "primary",
    "x": {
      "field": "advertiser_name",
      "limit": 20,
      "showNull": true,
      "type": "nominal",
      "sort": "-y"
    },
    "y": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

### 2. Line Chart (` + "`line_chart`" + `)
**Use for:** Showing trends over time

Example Specification
` + "```json" + `
{
  "chart_type": "line_chart",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "device_os",
      "limit": 3,
      "type": "nominal"
    },
    "x": {
      "field": "__time",
      "limit": 20,
      "sort": "-y",
      "type": "temporal"
    },
    "y": {
      "field": "total_bids",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

### 3. Area Chart (` + "`area_chart`" + `)
**Use for:** Showing magnitude of change over time with filled areas

Example Specification
` + "```json" + `
{
  "chart_type": "area_chart",
  "spec": {
    "metrics_view": "auction_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "app_or_site",
      "type": "nominal"
    },
    "x": {
      "field": "__time",
      "limit": 20,
      "showNull": true,
      "type": "temporal"
    },
    "y": {
      "field": "requests",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

### 4. Stacked Bar Chart (` + "`stacked_bar`" + `)
**Use for:** Showing multiple data series stacked on top of each other

Example Specification
` + "```json" + `
{
  "chart_type": "stacked_bar",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "rill_measures",
      "legendOrientation": "top",
      "type": "value"
    },
    "x": {
      "field": "__time",
      "limit": 20,
      "type": "temporal"
    },
    "y": {
      "field": "clicks",
      "fields": [
        "video_starts",
        "video_completes",
        "ctr",
        "clicks",
        "ecpm",
        "impressions"
      ],
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

### 5. Normalized Stacked Bar Chart (` + "`stacked_bar_normalized`" + `)
**Use for:** Showing proportions instead of absolute values (100% stacked)

Example Specification
` + "```json" + `
{
  "chart_type": "stacked_bar_normalized",
  "spec": {
    "metrics_view": "rill_commits_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "username",
      "limit": 3,
      "type": "nominal"
    },
    "x": {
      "field": "date",
      "limit": 20,
      "type": "temporal"
    },
    "y": {
      "field": "number_of_commits",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

### 6. Donut Chart (` + "`donut_chart`" + `)
**Use for:** Displaying data as segments of a circle with a hollow center

Example Specification
` + "```json" + `
{
  "chart_type": "donut_chart",
  "spec": {
    "metrics_view": "rill_commits_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "username",
      "limit": 20,
      "type": "nominal"
    },
    "innerRadius": 50,
    "measure": {
      "field": "number_of_commits",
      "type": "quantitative"
    }
  }
}
` + "```" + `

### 7. Funnel Chart (` + "`funnel_chart`" + `)
**Use for:** Showing flow through a process with decreasing/increasing values at each stage

Example Specification
` + "```json" + `
{
  "chart_type": "funnel_chart",
  "spec": {
    "metrics_view": "Funnel_Dataset_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": "stage",
    "measure": {
      "field": "total_users_measure",
      "type": "quantitative"
    },
    "mode": "width",
    "stage": {
      "field": "stage",
      "limit": 15,
      "type": "nominal"
    }
  }
}
` + "```" + `

### 8. Heat Map (` + "`heatmap`" + `)
**Use for:** Visualizing data density using color intensity across two dimensions

Example Specification
` + "```json" + `
{
  "chart_type": "heatmap",
  "spec": {
    "metrics_view": "bids_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "total_bids",
      "type": "quantitative"
    },
    "x": {
      "field": "day",
      "limit": 10,
      "type": "nominal",
      "sort": [
        "Sunday",
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
        "Saturday"
      ]
    },
    "y": {
      "field": "hour",
      "limit": 24,
      "type": "nominal",
      "sort": "-color"
    }
  }
}
` + "```" + `

### 9. Combo Chart (` + "`combo_chart`" + `)
**Use for:** Combining different chart types (like bars and lines) in a single visualization

Example Specification
` + "```json" + `
{
  "chart_type": "combo_chart",
  "spec": {
    "metrics_view": "auction_metrics",
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-31T23:59:59Z"
    },
    "color": {
      "field": "measures",
      "legendOrientation": "top",
      "type": "value"
    },
    "x": {
      "field": "__time",
      "limit": 20,
      "type": "temporal"
    },
    "y1": {
      "field": "1d_qps",
      "mark": "bar",
      "type": "quantitative",
      "zeroBasedOrigin": true
    },
    "y2": {
      "field": "requests",
      "mark": "line",
      "type": "quantitative",
      "zeroBasedOrigin": true
    }
  }
}
` + "```" + `

## Field Type Definitions

### Data Types
- **nominal**: Categorical data (e.g., categories, names, labels)
- **temporal**: Time-based data (dates, timestamps)
- **quantitative**: Numerical data (counts, amounts, measurements)
- **value**: Special type for multiple measures (used in color field)

### Common Field Properties
- **field**: The field name from the metrics view
- **type**: Data type (nominal, temporal, quantitative, value)
- **limit**: Maximum number of values to display
- **showNull**: Include null values in the visualization (true/false)
- **sort**: Sorting order
  - ` + "`\"x\"`" + ` or ` + "`\"-x\"`" + `: Sort by x-axis values (ascending/descending)
  - ` + "`\"y\"`" + ` or ` + "`\"-y\"`" + `: Sort by y-axis values (ascending/descending)
	- ` + "`\"color\"`" + ` or ` + "`\"-color\"`" + `: Sort by color field values (ascending/descending) (Useful for heatmaps)

  - Array of values for custom sort order (e.g., weekday names)
- **zeroBasedOrigin**: Start y-axis from zero (true/false)

### Special Fields
- **__time**: Built-in time dimension field
- **rill_measures**: Special field for multiple measures in stacked charts

## Color Configuration
- Can be a string (e.g., ` + "`\"primary\"`" + `, ` + "`\"stage\"`" + `) for single color
- Can be an object for field-based coloring:
  ` + "```json" + `
  {
    "field": "dimension_name",
    "type": "nominal|temporal|value",
    "limit": 10,
    "legendOrientation": "top|bottom|left|right"
  }
  ` + "```" + `

## Usage Guidelines

1. **Time-based visualizations**: Use ` + "`line_chart`" + ` or ` + "`area_chart`" + ` with temporal x-axis
2. **Categorical comparisons**: Use ` + "`bar_chart`" + ` with nominal x-axis
3. **Part-to-whole relationships**: Use ` + "`donut_chart`" + ` or ` + "`stacked_bar_normalized`" + `
4. **Process flows**: Use ` + "`funnel_chart`" + ` for conversion or stage-based data
5. **Two-dimensional patterns**: Use ` + "`heatmap`" + ` for correlations or patterns across two dimensions
6. **Multiple metrics**: Use ` + "`combo_chart`" + ` to compare different scales or ` + "`stacked_bar`" + ` for cumulative values

## Important Notes

- The ` + "`time_range`" + ` parameter is **required** for all charts
- Time range ` + "`start`" + ` is inclusive, ` + "`end`" + ` is exclusive
- You do not always have to include color for different bar chart and line charts. Use when required or when more than 1 dimensions has to be visualized.
- Ensure the metrics_view name matches exactly with available views
- Field names must match the exact field names in the metrics view
- When using ` + "`__time`" + ` field, set type to ` + "`\"temporal\"`" + `
- For multiple measures, use the ` + "`fields`" + ` array in the y-axis configuration`

	tool := mcp.NewTool("create_chart",
		mcp.WithDescription(description),
		mcp.WithString("chart_type",
			mcp.Required(),
			mcp.Description("Type of chart: bar_chart, line_chart, area_chart, stacked_bar, stacked_bar_normalized, donut_chart, funnel_chart, heatmap, combo_chart"),
		),
		mcp.WithObject("spec",
			mcp.Required(),
			mcp.Description("Chart specification object containing metrics_view, x, y, color, and other chart properties"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		instanceID := mcpInstanceIDFromContext(ctx)

		// Validate access
		claims := auth.GetClaims(ctx)
		if !claims.CanInstance(instanceID, auth.ReadMetrics) {
			return nil, ErrForbidden
		}

		// Get arguments as a map
		args, ok := req.GetRawArguments().(map[string]any)
		if !ok {
			return nil, errors.New("invalid arguments: expected an object")
		}

		chartType, ok := args["chart_type"].(string)
		if !ok || chartType == "" {
			return nil, fmt.Errorf("chart_type is required and must be a string")
		}

		spec, ok := args["spec"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("spec is required and must be an object")
		}

		// Validate that metrics_view is specified
		metricsView, ok := spec["metrics_view"].(string)
		if !ok || metricsView == "" {
			return nil, fmt.Errorf("spec must contain a 'metrics_view' field")
		}

		// Validate that time_range is specified
		_, hasTimeRange := spec["time_range"]
		if !hasTimeRange {
			return nil, fmt.Errorf("spec must contain a 'time_range' field with 'start' and 'end' properties")
		}

		// Validate that the metrics view exists
		var err error
		_, err = s.GetResource(ctx, &runtimev1.GetResourceRequest{
			InstanceId: instanceID,
			Name: &runtimev1.ResourceName{
				Kind: runtime.ResourceKindMetricsView,
				Name: metricsView,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("metrics view %q not found: %w", metricsView, err)
		}

		// Return the chart specification in a structured format
		result := map[string]any{
			"chart_type": chartType,
			"spec":       spec,
			"message":    fmt.Sprintf("Chart created successfully: %s", chartType),
		}

		return mcpNewToolResultJSON(result)
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
