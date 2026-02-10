package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/jsonval"
)

const QueryMetricsViewName = "query_metrics_view"

type QueryMetricsView struct {
	Runtime *runtime.Runtime
}

var _ Tool[QueryMetricsViewArgs, *QueryMetricsViewResult] = (*QueryMetricsView)(nil)

type QueryMetricsViewArgs map[string]any

type QueryMetricsViewResult struct {
	Data              []map[string]any `json:"data"`
	OpenURL           string           `json:"open_url,omitempty"`
	TruncationWarning string           `json:"truncation_warning,omitempty"`
}

func (t *QueryMetricsView) Spec() *mcp.Tool {
	description := `
Perform an arbitrary aggregation on a metrics view.

The JSON schema defines all available parameters. Key considerations:

Request:
- Include 'limit' and 'sort' parameters to optimize performance. Keep the limit as low as realistically possible for your task (ideally below 100 rows). Regardless of whether you include a limit, the server will truncate large results (and return a warning if it does)."
- 'time_range' is inclusive of start time, exclusive of end time
- 'time_range.time_dimension' (optional) specifies which time column to filter; defaults to the metrics view's default time column
- For comparisons, 'time_range' and 'comparison_time_range' must be non-overlapping and similar in duration (~20% tolerance)

Response:
- Returns aggregated data matching your query parameters
- Includes 'open_url' field with a shareable link to view results in the Rill UI
- Always cite the source of quantitative claims by including 'open_url' as a markdown link
- When presenting insights from multiple queries, cite each query's 'open_url' inline; when presenting multiple insights from the same query, cite once at the end

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
            {"name": "country"}
        ],
        "measures": [{"name": "total_revenue"}],
        "time_range": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2025-01-01T00:00:00Z"
        },
        "sort": [
            {"name": "event_time"},
            {"name": "total_revenue", "desc": true}
        ],
        "limit": 120
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
            "time_dimension"": "order_shipped_time"
        },
        "sort": [
            {"name": "event_time"},
            {"name": "total_revenue", "desc": true}
        ],
        "limit": 120
    }

Example: Get the top 10 demographic segments (by country, gender, and age group) with the largest absolute revenue difference comparing May 2025 (base period) to April 2025 (comparison period):
	{
		"metrics_view": "ecommerce_financials",
		"measures": [
			{"name": "total_revenue"},
			{"name": "total_revenue__delta_abs", "compute": {"comparison_delta": {"measure": "total_revenue"}}},
			{"name": "total_revenue__delta_rel", "compute": {"comparison_ratio": {"measure": "total_revenue"}}}
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
		Name:        QueryMetricsViewName,
		Title:       "Query Metrics View",
		Description: description,
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Querying metrics...",
			"openai/toolInvocation/invoked":  "Queried metrics",
		},
		InputSchema: inputSchema,
	}
}

func (t *QueryMetricsView) CheckAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	return s.Claims().Can(runtime.ReadMetrics), nil
}

func (t *QueryMetricsView) Handler(ctx context.Context, args QueryMetricsViewArgs) (*QueryMetricsViewResult, error) {
	session := GetSession(ctx)

	// Load instance config
	instance, err := t.Runtime.Instance(ctx, session.InstanceID())
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	cfg, err := instance.Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get instance config: %w", err)
	}

	// Compute a hard limit to prevent large results that bloat the context
	var limit int64
	var isSystemLimit bool
	if v, ok := args["limit"]; ok {
		limit, err = strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid limit value: %w", err)
		}
		if limit > cfg.AIMaxQueryLimit {
			return nil, fmt.Errorf("requested limit %d exceeds maximum allowed query limit of %d", limit, cfg.AIMaxQueryLimit)
		}
	} else {
		limit = cfg.AIDefaultQueryLimit
		isSystemLimit = true
		if args != nil {
			args["limit"] = limit
		}
	}

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
	schema := &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, StructType: res.Schema()}
	for {
		row, err := res.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		// Cast non-JSON types to JSON-compatible types
		v, err := jsonval.ToValue(row, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert row to value: %w", err)
		}
		var ok bool
		row, ok = v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected row to be map[string]any, got %T", v)
		}

		data = append(data, row)
	}

	// Generate an open URL for the query
	openURL, err := t.generateOpenURL(ctx, session.InstanceID(), args)
	if err != nil {
		return nil, fmt.Errorf("failed to generate open URL: %w", err)
	}

	// Build the result
	result := &QueryMetricsViewResult{
		Data:    data,
		OpenURL: openURL,
	}
	if isSystemLimit && len(data) >= int(limit) { // Add a warning if we hit the hard limit
		result.TruncationWarning = fmt.Sprintf("The result was truncated to %d rows (max allowed limit: %d)", limit, cfg.AIMaxQueryLimit)
	}
	return result, nil
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
