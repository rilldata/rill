package parser

import (
	"context"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func parseExplorePreset(t *testing.T, exploreYAML string) (*runtimev1.ExplorePreset, error) {
	t.Helper()
	files := map[string]string{
		`rill.yaml`:          ``,
		`explores/test.yaml`: exploreYAML,
	}
	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	if err != nil {
		return nil, err
	}
	for _, parseErr := range p.Errors {
		return nil, fmt.Errorf("parse error: %s", parseErr.Message)
	}
	r, ok := p.Resources[ResourceName{Kind: ResourceKindExplore, Name: "test"}]
	require.True(t, ok, "explore 'test' not found in parsed resources")
	return r.ExploreSpec.DefaultPreset, nil
}

func TestExploreDefaults(t *testing.T) {
	t.Run("comparison_mode time (backwards compat)", func(t *testing.T) {
		preset, err := parseExplorePreset(t, `
type: explore
metrics_view: mv1
defaults:
  comparison_mode: time
`)
		require.NoError(t, err)
		require.NotNil(t, preset)
		require.Equal(t, runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_TIME, preset.ComparisonMode)
		require.Nil(t, preset.CompareTimeRange) // "time" does not set a specific compare_time_range
	})

	t.Run("comparison_mode rill-PM", func(t *testing.T) {
		preset, err := parseExplorePreset(t, `
type: explore
metrics_view: mv1
defaults:
  comparison_mode: rill-PM
`)
		require.NoError(t, err)
		require.NotNil(t, preset)
		require.Equal(t, runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_TIME, preset.ComparisonMode)
		require.NotNil(t, preset.CompareTimeRange)
		require.Equal(t, "rill-PM", *preset.CompareTimeRange)
	})

	t.Run("comparison_mode rill-PY", func(t *testing.T) {
		preset, err := parseExplorePreset(t, `
type: explore
metrics_view: mv1
defaults:
  comparison_mode: rill-PY
`)
		require.NoError(t, err)
		require.NotNil(t, preset)
		require.Equal(t, runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_TIME, preset.ComparisonMode)
		require.NotNil(t, preset.CompareTimeRange)
		require.Equal(t, "rill-PY", *preset.CompareTimeRange)
	})

	t.Run("comparison_mode dimension", func(t *testing.T) {
		preset, err := parseExplorePreset(t, `
type: explore
metrics_view: mv1
defaults:
  comparison_mode: dimension
  comparison_dimension: country
`)
		require.NoError(t, err)
		require.NotNil(t, preset)
		require.Equal(t, runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_DIMENSION, preset.ComparisonMode)
		require.Nil(t, preset.CompareTimeRange)
		require.NotNil(t, preset.ComparisonDimension)
		require.Equal(t, "country", *preset.ComparisonDimension)
	})

	t.Run("comparison_mode invalid", func(t *testing.T) {
		_, err := parseExplorePreset(t, `
type: explore
metrics_view: mv1
defaults:
  comparison_mode: bogus
`)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid comparison mode")
	})

	t.Run("filter", func(t *testing.T) {
		preset, err := parseExplorePreset(t, `
type: explore
metrics_view: mv1
defaults:
  filter: "country IN ('US', 'CA')"
`)
		require.NoError(t, err)
		require.NotNil(t, preset)
		require.NotNil(t, preset.Filter)
		require.NotNil(t, preset.Filter.Expression)
	})

	t.Run("filter invalid", func(t *testing.T) {
		_, err := parseExplorePreset(t, `
type: explore
metrics_view: mv1
defaults:
  filter: "not a valid expression %%"
`)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid filter expression")
	})

	t.Run("pinned", func(t *testing.T) {
		preset, err := parseExplorePreset(t, `
type: explore
metrics_view: mv1
defaults:
  pinned:
    - country
    - city
`)
		require.NoError(t, err)
		require.NotNil(t, preset)
		require.Equal(t, []string{"country", "city"}, preset.Pinned)
	})

	t.Run("combined defaults", func(t *testing.T) {
		preset, err := parseExplorePreset(t, `
type: explore
metrics_view: mv1
defaults:
  time_range: P7D
  comparison_mode: rill-PM
  filter: "country IN ('US')"
  pinned:
    - country
  measures:
    - impressions
  dimensions:
    - publisher
`)
		require.NoError(t, err)
		require.NotNil(t, preset)
		require.NotNil(t, preset.TimeRange)
		require.Equal(t, "P7D", *preset.TimeRange)
		require.Equal(t, runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_TIME, preset.ComparisonMode)
		require.NotNil(t, preset.CompareTimeRange)
		require.Equal(t, "rill-PM", *preset.CompareTimeRange)
		require.NotNil(t, preset.Filter)
		require.Equal(t, []string{"country"}, preset.Pinned)
		require.Equal(t, []string{"impressions"}, preset.Measures)
		require.Equal(t, []string{"publisher"}, preset.Dimensions)
	})
}

func TestExploreFieldSelector(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// explore e1
		`explores/e1.yaml`: `
type: explore
metrics_view: mv1
`,
		// explore e2
		`explores/e2.yaml`: `
type: explore
metrics_view: mv1
dimensions: '*'
measures:
  exclude: '*'
`,
		// explore e3
		`explores/e3.yaml`: `
type: explore
metrics_view: mv1
dimensions: foo
measures:
  exclude: bar
`,
		// explore e4
		`explores/e4.yaml`: `
type: explore
metrics_view: mv1
dimensions: [bar, baz]
measures:
  exclude: [foo, qux]
`,
		// explore e5
		`explores/e5.yaml`: `
type: explore
metrics_view: mv1
dimensions:
  regex: 'foo.*'
measures:
  exclude:
    regex: 'bar.*'
`,
	}

	resources := []*Resource{
		// explore e1
		{
			Name:  ResourceName{Kind: ResourceKindExplore, Name: "e1"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "mv1"}},
			Paths: []string{"/explores/e1.yaml"},
			ExploreSpec: &runtimev1.ExploreSpec{
				DisplayName:          "E1",
				MetricsView:          "mv1",
				DimensionsSelector:   &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
				MeasuresSelector:     &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
				AllowCustomTimeRange: true,
			},
		},
		// explore e2
		{
			Name:  ResourceName{Kind: ResourceKindExplore, Name: "e2"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "mv1"}},
			Paths: []string{"/explores/e2.yaml"},
			ExploreSpec: &runtimev1.ExploreSpec{
				DisplayName:          "E2",
				MetricsView:          "mv1",
				DimensionsSelector:   &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
				MeasuresSelector:     &runtimev1.FieldSelector{Invert: true, Selector: &runtimev1.FieldSelector_All{All: true}},
				AllowCustomTimeRange: true,
			},
		},
		// explore e3
		{
			Name:  ResourceName{Kind: ResourceKindExplore, Name: "e3"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "mv1"}},
			Paths: []string{"/explores/e3.yaml"},
			ExploreSpec: &runtimev1.ExploreSpec{
				DisplayName:          "E3",
				MetricsView:          "mv1",
				Dimensions:           []string{"foo"},
				MeasuresSelector:     &runtimev1.FieldSelector{Invert: true, Selector: &runtimev1.FieldSelector_Fields{Fields: &runtimev1.StringListValue{Values: []string{"bar"}}}},
				AllowCustomTimeRange: true,
			},
		},
		// explore e4
		{
			Name:  ResourceName{Kind: ResourceKindExplore, Name: "e4"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "mv1"}},
			Paths: []string{"/explores/e4.yaml"},
			ExploreSpec: &runtimev1.ExploreSpec{
				DisplayName:          "E4",
				MetricsView:          "mv1",
				Dimensions:           []string{"bar", "baz"},
				MeasuresSelector:     &runtimev1.FieldSelector{Invert: true, Selector: &runtimev1.FieldSelector_Fields{Fields: &runtimev1.StringListValue{Values: []string{"foo", "qux"}}}},
				AllowCustomTimeRange: true,
			},
		},
		// explore e5
		{
			Name:  ResourceName{Kind: ResourceKindExplore, Name: "e5"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "mv1"}},
			Paths: []string{"/explores/e5.yaml"},
			ExploreSpec: &runtimev1.ExploreSpec{
				DisplayName:          "E5",
				MetricsView:          "mv1",
				DimensionsSelector:   &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_Regex{Regex: "foo.*"}},
				MeasuresSelector:     &runtimev1.FieldSelector{Invert: true, Selector: &runtimev1.FieldSelector_Regex{Regex: "bar.*"}},
				AllowCustomTimeRange: true,
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}
