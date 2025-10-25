package ai

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
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
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Listing metrics…",
			"openai/toolInvocation/invoked":  "Listed metrics",
		},
	}
}

func (t *ListMetricsViews) CheckAccess(ctx context.Context) bool {
	s := GetSession(ctx)
	return s.Claims().Can(runtime.ReadObjects)
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
		r, access, err := t.Runtime.ApplySecurityPolicy(ctx, session.InstanceID(), session.Claims(), r)
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
