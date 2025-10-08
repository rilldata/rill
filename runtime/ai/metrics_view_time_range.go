package ai

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

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
	return claims.Can(runtime.ReadMetrics)
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
