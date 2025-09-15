package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"google.golang.org/protobuf/encoding/protojson"
)

type ListMetricsViews struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListMetricsViewsArgs, *ListMetricsViewsResult] = (*ListMetricsViews)(nil)

type ListMetricsViewsArgs struct{}

type ListMetricsViewsResult struct {
	MetricsViews []map[string]any `json:"metrics_views"`
}

func (t *ListMetricsViews) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_metrics_views",
		Title:       "List Metrics Views",
		Description: "List all metrics views in the current project",
	}
}

func (t *ListMetricsViews) CheckAccess(claims *runtime.SecurityClaims) bool {
	return true
}

func (t *ListMetricsViews) Handler(ctx context.Context, args *ListMetricsViewsArgs) (*ListMetricsViewsResult, error) {
	session := GetSession(ctx)

	ctrl, err := t.Runtime.Controller(ctx, session.InstanceID())
	if err != nil {
		return nil, err
	}

	rs, err := ctrl.List(ctx, runtime.ResourceKindMetricsView, "", false)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(rs, func(a, b *runtimev1.Resource) int {
		an := a.Meta.Name
		bn := b.Meta.Name
		if an.Kind < bn.Kind {
			return -1
		}
		if an.Kind > bn.Kind {
			return 1
		}
		return strings.Compare(an.Name, bn.Name)
	})

	i := 0
	for i < len(rs) {
		r := rs[i]
		r, access, err := t.Runtime.ApplySecurityPolicy(session.InstanceID(), session.Claims(), r)
		if err != nil {
			return nil, err
		}
		if !access {
			// Remove from the slice
			rs[i] = rs[len(rs)-1]
			rs[len(rs)-1] = nil
			rs = rs[:len(rs)-1]
			continue
		}
		rs[i] = r
		i++
	}

	res := make(map[string]any)

	// Find instance-wide AI context and add it to the response.
	// NOTE: These arguably belong in the top-level instructions or other metadata, but that doesn't currently support dynamic values.
	instance, err := t.Runtime.Instance(ctx, session.InstanceID())
	if err != nil {
		return nil, fmt.Errorf("failed to get instance %q: %w", session.InstanceID(), err)
	}
	if instance.AIInstructions != "" {
		res["ai_instructions"] = instance.AIInstructions
	}

	var metricsViews []map[string]any
	for _, r := range rs {
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

	return &ListMetricsViewsResult{
		MetricsViews: metricsViews,
	}, nil
}

type GetMetricsView struct {
	Runtime *runtime.Runtime
}

var _ Tool[*GetMetricsViewArgs, *GetMetricsViewResult] = (*GetMetricsView)(nil)

type GetMetricsViewArgs struct {
	MetricsView string `json:"metrics_view" jsonschema:"Name of the metrics view"`
}

type GetMetricsViewResult struct {
	Spec map[string]any `json:"spec"`
}

func (t *GetMetricsView) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_metrics_view",
		Title:       "Get Metrics View",
		Description: "Get the specification for a given metrics view, including available measures and dimensions",
	}
}

func (t *GetMetricsView) CheckAccess(claims *runtime.SecurityClaims) bool {
	return true
}

func (t *GetMetricsView) Handler(ctx context.Context, args *GetMetricsViewArgs) (*GetMetricsViewResult, error) {
	session := GetSession(ctx)

	ctrl, err := t.Runtime.Controller(ctx, session.InstanceID())
	if err != nil {
		return nil, err
	}

	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: args.MetricsView}, false)
	if err != nil {
		return nil, err
	}

	r, access, err := t.Runtime.ApplySecurityPolicy(session.InstanceID(), session.Claims(), r)
	if err != nil {
		return nil, err
	}
	if !access {
		return nil, fmt.Errorf("resource not found")
	}

	specJSON, err := protojson.Marshal(r.GetMetricsView().State.ValidSpec)
	if err != nil {
		return nil, err
	}
	var specMap map[string]any
	err = json.Unmarshal(specJSON, &specMap)
	if err != nil {
		return nil, err
	}

	return &GetMetricsViewResult{
		Spec: specMap,
	}, nil
}

type QueryMetricsViewTimeRange struct {
	Runtime *runtime.Runtime
}

var _ Tool[*QueryMetricsViewTimeRangeArgs, *QueryMetricsViewTimeRangeResult] = (*QueryMetricsViewTimeRange)(nil)

type QueryMetricsViewTimeRangeArgs struct {
	MetricsView   string `json:"metrics_view" jsonschema:"Name of the metrics view"`
	TimeDimension string `json:"time_dimension,omitempty" jsonschema:"Optional time dimension to use for resolving the time range. If not provided, the default time dimension defined under timeseries field of the metrics view will be used."`
}

type QueryMetricsViewTimeRangeResult struct {
	Data map[string]any `json:"data"`
}

func (t *QueryMetricsViewTimeRange) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  "query_metrics_view_time_range",
		Title: "Query Metrics View Time Range",
		Description: `
			Retrieve the total time range available for a given metrics view.
			Note: All subsequent queries of the metrics view should be constrained to this time range to ensure accurate results.
		`,
	}
}

func (t *QueryMetricsViewTimeRange) CheckAccess(claims *runtime.SecurityClaims) bool {
	return true
}

func (t *QueryMetricsViewTimeRange) Handler(ctx context.Context, args *QueryMetricsViewTimeRangeArgs) (*QueryMetricsViewTimeRangeResult, error) {
	session := GetSession(ctx)

	res, err := t.Runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID: session.InstanceID(),
		Resolver:   "metrics_time_range",
		ResolverProperties: map[string]any{
			"metrics_view": args.MetricsView,
		},
		Args: map[string]any{
			"time_dimension": args.TimeDimension,
		},
		Claims: session.Claims(),
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	data, err := res.Next()
	if err != nil {
		return nil, err
	}

	return &QueryMetricsViewTimeRangeResult{
		Data: data,
	}, nil
}

type QueryMetricsView struct {
	Runtime *runtime.Runtime
}

var _ Tool[QueryMetricsViewArgs, *QueryMetricsViewResult] = (*QueryMetricsView)(nil)

type QueryMetricsViewArgs map[string]any

type QueryMetricsViewResult struct {
	Data []map[string]any `json:"data"`
}

func (t *QueryMetricsView) Spec() *mcp.Tool {
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
	}
}

func (t *QueryMetricsView) CheckAccess(claims *runtime.SecurityClaims) bool {
	return true
}

func (t *QueryMetricsView) Handler(ctx context.Context, args QueryMetricsViewArgs) (*QueryMetricsViewResult, error) {
	session := GetSession(ctx)

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

	return &QueryMetricsViewResult{Data: data}, nil
}
