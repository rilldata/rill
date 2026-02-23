package parser

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

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
	p, err := Parse(ctx, repo, "", "", "duckdb", zap.NewNop())
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}
