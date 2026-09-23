package ai

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestSuggestedPromptsHash(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		DisplayName: "Ads",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "campaign", DisplayName: "Campaign"},
			{Name: "publisher", DisplayName: "Publisher"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "impressions", DisplayName: "Impressions"},
			{Name: "clicks", DisplayName: "Clicks"},
		},
	}
	input := func() *SuggestedPromptsInput {
		return &SuggestedPromptsInput{
			DashboardKind: "explore",
			DashboardName: "ads",
			DisplayName:   "Ads dashboard",
			MetricsViews:  []*SuggestedPromptsMetricsView{NewSuggestedPromptsMetricsView("ads", mv, []string{"campaign"}, nil)},
		}
	}

	// Stable for equal inputs
	require.Equal(t, SuggestedPromptsHash(input()), SuggestedPromptsHash(input()))

	// The explore's field selection is respected
	require.Len(t, input().MetricsViews[0].Dimensions, 1)
	require.Len(t, input().MetricsViews[0].Measures, 2)

	// Changes to any relevant input change the hash
	changed := input()
	changed.DisplayName = "Other"
	require.NotEqual(t, SuggestedPromptsHash(input()), SuggestedPromptsHash(changed))

	changed = input()
	changed.MetricsViews[0].Measures[0].Description = "Number of impressions"
	require.NotEqual(t, SuggestedPromptsHash(input()), SuggestedPromptsHash(changed))

	changed = input()
	changed.ProjectInstructions = "Focus on clicks"
	require.NotEqual(t, SuggestedPromptsHash(input()), SuggestedPromptsHash(changed))

	// Field order matters: a moved field is a different prompt input, and ambiguous concatenations must not collide
	a := &SuggestedPromptsInput{MetricsViews: []*SuggestedPromptsMetricsView{{Name: "ab", DisplayName: "c"}}}
	b := &SuggestedPromptsInput{MetricsViews: []*SuggestedPromptsMetricsView{{Name: "a", DisplayName: "bc"}}}
	require.NotEqual(t, SuggestedPromptsHash(a), SuggestedPromptsHash(b))
}
