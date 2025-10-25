package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
)

type QueryMetricsView struct {
	Runtime *runtime.Runtime
}

var _ Tool[QueryMetricsViewArgs, *QueryMetricsViewResult] = (*QueryMetricsView)(nil)

type QueryMetricsViewArgs map[string]any

type QueryMetricsViewResult struct {
	Data    []map[string]any `json:"data"`
	OpenURL string           `json:"open_url,omitempty"`
}

func (t *QueryMetricsView) Spec() *mcp.Tool {
	description := `
Perform an arbitrary aggregation on a metrics view.

The JSON schema defines all available parameters. Key considerations:

Request:
• 'time_range' is inclusive of start time, exclusive of end time
• 'time_range.time_dimension' (optional) specifies which time column to filter; defaults to the metrics view's default time column
• Include 'sort' and 'limit' parameters to optimize query performance and avoid unbounded result sets
• For comparisons, 'time_range' and 'comparison_time_range' must be non-overlapping and similar in duration (~20% tolerance)

Response:
• Returns aggregated data matching your query parameters
• Includes 'open_url' field with a shareable link to view results in the Rill UI
• Always cite the source of quantitative claims by including 'open_url' as a markdown link
• When presenting insights from multiple queries, cite each query's 'open_url' inline; when presenting multiple insights from the same query, cite once at the end

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

	var inputSchema *jsonschema.Schema
	err := json.Unmarshal([]byte(metricsview.QueryJSONSchema), &inputSchema)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal input schema: %w", err))
	}

	return &mcp.Tool{
		Name:        "query_metrics_view",
		Title:       "Query Metrics View",
		Description: description,
		InputSchema: inputSchema,
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Querying metrics…",
			"openai/toolInvocation/invoked":  "Completed metrics query",
		},
	}
}

func (t *QueryMetricsView) CheckAccess(ctx context.Context) bool {
	s := GetSession(ctx)
	return s.Claims().Can(runtime.ReadMetrics)
}

func (t *QueryMetricsView) Handler(ctx context.Context, args QueryMetricsViewArgs) (*QueryMetricsViewResult, error) {
	session := GetSession(ctx)

	// Apply a timeout to prevent runaway queries
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Run the metrics query
	res, err := t.Runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         session.InstanceID(),
		Resolver:           "metrics",
		ResolverProperties: map[string]any(args),
		Claims:             session.Claims(),
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	// Gather the result rows
	var data []map[string]any
	for {
		row, err := res.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		data = append(data, row)
	}

	// Generate an open URL for the query
	openURL, err := t.generateOpenURL(ctx, session.InstanceID(), args)
	if err != nil {
		return nil, fmt.Errorf("failed to generate open URL: %w", err)
	}

	return &QueryMetricsViewResult{
		Data:    data,
		OpenURL: openURL,
	}, nil
}

// generateOpenURL generates an open URL for the given query parameters
func (t *QueryMetricsView) generateOpenURL(ctx context.Context, instanceID string, metricsQuery map[string]any) (string, error) {
	// Get instance to access the configured frontend URL
	instance, err := t.Runtime.Instance(ctx, instanceID)
	if err != nil {
		return "", fmt.Errorf("failed to get instance: %w", err)
	}

	// If there's no frontend URL (e.g. perhaps in test cases or during rollout), return an empty string
	if instance.FrontendURL == "" {
		return "", nil
	}

	// Build the complete URL for the query
	openURL, err := url.Parse(instance.FrontendURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse frontend URL %q: %w", instance.FrontendURL, err)
	}

	openURL.Path, err = url.JoinPath(openURL.Path, "-", "open-query")
	if err != nil {
		return "", fmt.Errorf("failed to join path: %w", err)
	}

	queryJSON, err := json.Marshal(metricsQuery)
	if err != nil {
		return "", fmt.Errorf("failed to marshal MCP query to JSON: %w", err)
	}
	values := make(url.Values)
	values.Set("query", string(queryJSON))
	openURL.RawQuery = values.Encode()

	return openURL.String(), nil
}
