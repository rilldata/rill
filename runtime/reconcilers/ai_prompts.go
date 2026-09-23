package reconcilers

import (
	"context"
	"errors"
	"maps"
	"slices"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

// reconcileSuggestedPrompts computes the AI-generated starter prompts for a dashboard.
// It is shared by the explore and canvas reconcilers and runs after the dashboard's valid spec has been written, so it never delays or fails the dashboard itself.
//
// Behavior:
//   - If the dashboard is invalid (input is nil) or the author configured `ai_prompts`, generated prompts are cleared.
//   - Otherwise prompts are generated once per distinct input hash. Data refreshes and controller restarts reuse the stored prompts.
//   - If the instance has no AI connector, nothing is stored (checking for a connector is cheap, so this is re-evaluated on every reconcile).
//   - If generation fails, the previous prompts are kept but the new hash is still stored so the call is not retried on every reconcile; the next spec change retries.
//
// It returns the prompts and hash to store, and whether they differ from the current state.
func reconcileSuggestedPrompts(ctx context.Context, c *runtime.Controller, name *runtimev1.ResourceName, configured, current []*runtimev1.AIPrompt, currentHash string, in *ai.SuggestedPromptsInput) ([]*runtimev1.AIPrompt, string, bool) {
	if in == nil || len(configured) > 0 {
		return nil, "", len(current) > 0 || currentHash != ""
	}

	hash := ai.SuggestedPromptsHash(in)
	if hash == currentHash {
		return current, currentHash, false
	}

	prompts, err := ai.GenerateSuggestedPrompts(ctx, c.Runtime, c.InstanceID, in)
	if err != nil {
		if errors.Is(err, runtime.ErrAINotConfigured) {
			return nil, "", len(current) > 0 || currentHash != ""
		}
		if ctx.Err() != nil {
			// The reconcile was cancelled (e.g. the resource changed again). Leave the state untouched so the next reconcile retries.
			return current, currentHash, false
		}
		c.Logger.Warn("Failed to generate suggested AI prompts", zap.String("name", name.Name), zap.String("kind", name.Kind), zap.Error(err), observability.ZapCtx(ctx))
		return current, hash, true
	}
	return prompts, hash, true
}

// suggestedPromptsInputForExplore builds the prompt generation input for a valid explore spec, limited to the explore's visible dimensions and measures.
func suggestedPromptsInputForExplore(ctx context.Context, c *runtime.Controller, name string, spec *runtimev1.ExploreSpec) (*ai.SuggestedPromptsInput, error) {
	mvr, err := c.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: spec.MetricsView}, false)
	if err != nil {
		return nil, err
	}
	mv := mvr.GetMetricsView().State.ValidSpec
	if mv == nil {
		return nil, nil
	}

	inst, err := c.Runtime.Instance(ctx, c.InstanceID)
	if err != nil {
		return nil, err
	}

	return &ai.SuggestedPromptsInput{
		DashboardKind:       "explore",
		DashboardName:       name,
		DisplayName:         spec.DisplayName,
		Description:         spec.Description,
		MetricsViews:        []*ai.SuggestedPromptsMetricsView{ai.NewSuggestedPromptsMetricsView(spec.MetricsView, mv, spec.Dimensions, spec.Measures)},
		ProjectInstructions: inst.AIInstructions,
	}, nil
}

// suggestedPromptsInputForCanvas builds the prompt generation input for a valid canvas spec from the metrics views its components reference.
func suggestedPromptsInputForCanvas(ctx context.Context, c *runtime.Controller, name string, spec *runtimev1.CanvasSpec, metricsViews map[string]*runtimev1.Resource) (*ai.SuggestedPromptsInput, error) {
	if len(metricsViews) == 0 {
		return nil, nil
	}

	inst, err := c.Runtime.Instance(ctx, c.InstanceID)
	if err != nil {
		return nil, err
	}

	in := &ai.SuggestedPromptsInput{
		DashboardKind:       "canvas",
		DashboardName:       name,
		DisplayName:         spec.DisplayName,
		ProjectInstructions: inst.AIInstructions,
	}
	// Iterate in name order so the hash is stable across reconciles.
	for _, mvName := range slices.Sorted(maps.Keys(metricsViews)) {
		mv := metricsViews[mvName].GetMetricsView().State.ValidSpec
		if mv == nil {
			continue
		}
		in.MetricsViews = append(in.MetricsViews, ai.NewSuggestedPromptsMetricsView(mvName, mv, nil, nil))
	}
	if len(in.MetricsViews) == 0 {
		return nil, nil
	}
	return in, nil
}
