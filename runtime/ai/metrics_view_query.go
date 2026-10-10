package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

const QueryMetricsViewName = "query_metrics_view"

type QueryMetricsView struct {
	Runtime *runtime.Runtime
}

var _ Tool[QueryMetricsViewArgs, *QueryMetricsViewResult] = (*QueryMetricsView)(nil)

type QueryMetricsViewArgs map[string]any

type QueryMetricsViewResult struct {
	Schema []SchemaField `json:"schema"`
	Data   [][]any       `json:"data"`
	// ResolvedTimeRange & ResolvedComparisonTimeRange store the exact time ranges used for the query.
	// This helps when opening the citation url and get the exact time range for relative time ranges.
	ResolvedTimeRange           *metricsview.TimeRange `json:"resolved_time_range,omitempty"`
	ResolvedComparisonTimeRange *metricsview.TimeRange `json:"resolved_comparison_time_range,omitempty"`
	OpenURL                     string                 `json:"open_url,omitempty"`
	TruncationWarning           string                 `json:"truncation_warning,omitempty"`
}

func (t *QueryMetricsView) Spec() *mcp.Tool {
	description := `
Perform an arbitrary aggregation on a metrics view.

The JSON schema defines all available parameters. Key considerations:

Request:
- Include 'limit' and 'sort' parameters to optimize performance. Keep the limit as low as realistically possible for your task (ideally below 100 rows). Regardless of whether you include a limit, the server will truncate large results (and return a warning if it does).
- 'time_range' is inclusive of start time, exclusive of end time
- 'time_range.time_dimension' (optional) specifies which time column to filter; defaults to the metrics view's default time column
- Pass 'dimensions', 'measures', 'sort', 'where' and other structured parameters as JSON arrays and objects, not as JSON-encoded strings
- Only use dimension and measure names that exist in the metrics view, in their matching list (a dimension can't be used as a measure)
- Comparison measures ('comparison_value', 'comparison_delta', 'comparison_ratio') require 'comparison_time_range'. For a trend over consecutive periods (e.g. month over month), query a 'time_floor' dimension instead and compare the rows
- For comparisons, 'time_range' and 'comparison_time_range' must be non-overlapping and similar in duration (~20% tolerance)
- In 'where' and 'having', 'eq'/'neq' take a single value and 'in'/'nin' take a list. Each operand of 'and'/'or' must itself be a condition ({"cond": {...}})
- A derived measure computes arithmetic over existing measures: {"name": "cost_per_order", "compute": {"expression": {"expression": "total_cost / total_orders"}}}

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
            }
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
            "time_dimension": "order_shipped_time"
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

Example: Get the top 10 demographic segments (by country, gender, and age group) with the largest absolute revenue difference comparing last month as of latest day (base period) to previous month (comparison period):
	{
		"metrics_view": "ecommerce_financials",
		"measures": [
			{"name": "total_revenue"},
			{"name": "total_revenue__delta_abs", "compute": {"comparison_delta": {"measure": "total_revenue"}}},
			{"name": "total_revenue__delta_rel", "compute": {"comparison_ratio": {"measure": "total_revenue"}}}
		],
		"dimensions": [{"name": "country"}, {"name": "gender"}, {"name": "age_group"}],
		"time_range": {
			"expression": "1M as of latest/D"
		},
		"comparison_time_range": {
			"expression": "1M as of latest/D offset -1M"
		},
		"sort": [{"name": "total_revenue__delta_abs", "desc": true}],
		"limit": 10
	}

Example: Get the top 10 demographic segments (by country, gender, and age group) with the largest absolute revenue difference comparing last day as of latest minute (base period) to previous day (comparison period):
	{
		"metrics_view": "ecommerce_financials",
		"measures": [
			{"name": "total_revenue"},
			{"name": "total_revenue__delta_abs", "compute": {"comparison_delta": {"measure": "total_revenue"}}},
			{"name": "total_revenue__delta_rel", "compute": {"comparison_ratio": {"measure": "total_revenue"}}}
		],
		"dimensions": [{"name": "country"}, {"name": "gender"}, {"name": "age_group"}],
		"time_range": {
			"expression": "1D as of latest/m"
		},
		"comparison_time_range": {
			"expression": "1D as of latest/m offset -1D"
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
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
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
	normalizeQueryArgs(args)

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
	// ideally can be moved to executor.enforceQueryLimits, but then we cannot return the warning message in the result as easily
	var limit int64
	var isSystemLimit bool
	if v, ok := args["limit"]; ok { // Hackily extracting the query's 'limit' to avoid parsing the entire query outside of the resolver
		limit, err = strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid limit value: %w", err)
		}
		if limit > cfg.AIMaxQueryLimit {
			limit = cfg.AIMaxQueryLimit
			isSystemLimit = true
		}
	} else {
		limit = cfg.AIDefaultQueryLimit
		isSystemLimit = true
	}
	args["limit"] = limit
	args["query_limits"] = metricsview.QueryLimits{
		RequireTimeRange: cfg.AIRequireTimeRange,
		MaxTimeRangeDays: cfg.AIMaxTimeRangeDays,
	}

	// Apply a timeout to prevent runaway queries
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Run the metrics query
	res, _, err := t.Runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         session.InstanceID(),
		Resolver:           "metrics",
		ResolverProperties: map[string]any(args),
		Claims:             session.Claims(),
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	// Gather the result in tabular format
	schema, data, err := resolverResultToTabular(res)
	if err != nil {
		return nil, err
	}

	// Resolve time ranges and store them in the result to record the exact resolved time ranges for this tool call
	tr, ctr, err := resolveTimeRanges(res)
	if err != nil {
		return nil, err
	}

	// Generate an open URL for the query
	openURL, err := t.generateOpenURL(ctx, session.InstanceID(), session.ID(), session.ParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate open URL: %w", err)
	}

	// Build the result
	result := &QueryMetricsViewResult{
		Schema:                      schema,
		Data:                        data,
		OpenURL:                     openURL,
		ResolvedTimeRange:           tr,
		ResolvedComparisonTimeRange: ctr,
	}
	if isSystemLimit && int64(len(data)) >= limit { // Add a warning if we hit the system limit
		msg := fmt.Sprintf("The system truncated the result to %d rows", limit)
		if limit != cfg.AIMaxQueryLimit {
			msg += fmt.Sprintf("; to fetch more rows, explicitly set a limit (max allowed limit: %d)", cfg.AIMaxQueryLimit)
		}
		result.TruncationWarning = msg
	}
	return result, nil
}

// queryArgsJSONFields are the query fields that take an object or array, which LLMs sometimes pass as a JSON-encoded string.
var queryArgsJSONFields = []string{"dimensions", "measures", "pivot_on", "spine", "sort", "time_range", "comparison_time_range", "where", "having"}

// normalizeQueryArgs repairs common LLM formatting mistakes in query args where the intent is unambiguous.
func normalizeQueryArgs(args QueryMetricsViewArgs) {
	for _, k := range queryArgsJSONFields {
		s, ok := args[k].(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, "[") && !strings.HasPrefix(s, "{") {
			continue
		}
		var v any
		if err := json.Unmarshal([]byte(s), &v); err == nil {
			args[k] = v
		}
	}

	// Derived measures nest the expression in an object, but are often passed as {"compute": {"expression": "a / b"}}.
	measures, _ := args["measures"].([]any)
	for _, m := range measures {
		m, _ := m.(map[string]any)
		compute, _ := m["compute"].(map[string]any)
		expr, ok := compute["expression"].(string)
		if !ok {
			continue
		}
		obj := map[string]any{"expression": expr}
		if dn, ok := compute["display_name"]; ok {
			obj["display_name"] = dn
			delete(compute, "display_name")
		}
		compute["expression"] = obj
	}

	normalizeQueryExpression(args["where"])
	normalizeQueryExpression(args["having"])
}

// normalizeQueryExpression rewrites "eq"/"neq" conditions against a list value to "in"/"nin".
func normalizeQueryExpression(v any) {
	e, ok := v.(map[string]any)
	if !ok {
		return
	}

	if sub, ok := e["subquery"].(map[string]any); ok {
		normalizeQueryExpression(sub["where"])
		normalizeQueryExpression(sub["having"])
	}

	cond, ok := e["cond"].(map[string]any)
	if !ok {
		return
	}
	exprs, _ := cond["exprs"].([]any)
	for _, x := range exprs {
		normalizeQueryExpression(x)
	}

	if len(exprs) != 2 {
		return
	}
	rhs, _ := exprs[1].(map[string]any)
	if _, isList := rhs["val"].([]any); !isList {
		return
	}
	switch cond["op"] {
	case "eq":
		cond["op"] = "in"
	case "neq":
		cond["op"] = "nin"
	}
}

// generateOpenURL generates an open URL for the given query parameters
func (t *QueryMetricsView) generateOpenURL(ctx context.Context, instanceID, sessionID, callID string) (string, error) {
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

	openURL.Path, err = url.JoinPath(openURL.Path, "-", "ai", sessionID, "message", callID, "-", "open")
	if err != nil {
		return "", fmt.Errorf("failed to join path: %w", err)
	}

	return openURL.String(), nil
}

func resolveTimeRanges(res runtime.ResolverResult) (*metricsview.TimeRange, *metricsview.TimeRange, error) {
	meta := res.Meta()
	if meta == nil {
		return nil, nil, nil
	}

	var tr *metricsview.TimeRange
	rawTr, ok := meta["time_range"]
	if ok {
		tr = &metricsview.TimeRange{}
		if err := mapstructureutil.WeakDecode(rawTr, tr); err != nil {
			return nil, nil, err
		}
	}

	var ctr *metricsview.TimeRange
	rawCtr, ok := meta["comparison_time_range"]
	if ok {
		ctr = &metricsview.TimeRange{}
		if err := mapstructureutil.WeakDecode(rawCtr, ctr); err != nil {
			return nil, nil, err
		}
	}

	return tr, ctr, nil
}
