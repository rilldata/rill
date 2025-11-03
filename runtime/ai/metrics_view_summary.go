package ai

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

type QueryMetricsViewSummary struct {
	Runtime *runtime.Runtime
}

var _ Tool[*QueryMetricsViewSummaryArgs, *QueryMetricsViewSummaryResult] = (*QueryMetricsViewSummary)(nil)

type QueryMetricsViewSummaryArgs struct {
	MetricsView string `json:"metrics_view" jsonschema:"Name of the metrics view"`
}

type QueryMetricsViewSummaryResult struct {
	Data map[string]any `json:"data"`
}

func (t *QueryMetricsViewSummary) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  "query_metrics_view_summary",
		Title: "Query Metrics View Summary",
		Description: `
			Retrieve summary statistics for a metrics view including:
			- Total time range available
			- Sample values and data types for each dimension
			Note: All subsequent queries of the metrics view should be constrained to this time range to ensure accurate results.
		`,
	}
}

func (t *QueryMetricsViewSummary) CheckAccess(ctx context.Context) bool {
	s := GetSession(ctx)
	return s.Claims().Can(runtime.ReadMetrics)
}

func (t *QueryMetricsViewSummary) Handler(ctx context.Context, args *QueryMetricsViewSummaryArgs) (*QueryMetricsViewSummaryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	session := GetSession(ctx)

	res, err := t.Runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID: session.InstanceID(),
		Resolver:   "metrics_summary",
		ResolverProperties: map[string]any{
			"metrics_view": args.MetricsView,
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

	return &QueryMetricsViewSummaryResult{
		Data: data,
	}, nil
}
