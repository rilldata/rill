package parser

import (
	"context"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestExploreDuplicateFields(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "dimensions",
			yaml:    "dimensions: [foo, bar, foo]",
			wantErr: `duplicate field "foo" in dimensions`,
		},
		{
			name:    "measures",
			yaml:    "measures: [count, total, count]",
			wantErr: `duplicate field "count" in measures`,
		},
		{
			name:    "default dimensions",
			yaml:    "defaults:\n  dimensions: [foo, bar, foo]",
			wantErr: `duplicate field "foo" in defaults.dimensions`,
		},
		{
			name:    "default measures",
			yaml:    "defaults:\n  measures: [count, total, count]",
			wantErr: `duplicate field "count" in defaults.measures`,
		},
		{
			name: "unique fields can also appear in defaults",
			yaml: "dimensions: [foo, bar]\nmeasures: [count, total]\ndefaults:\n  dimensions: [bar, foo]\n  measures: [total, count]",
		},
	}
	for _, inline := range []bool{false, true} {
		kind := "standalone"
		if inline {
			kind = "inline"
		}
		for _, tt := range tests {
			t.Run(kind+"/"+tt.name, func(t *testing.T) {
				path := "explores/e1.yaml"
				content := "type: explore\nmetrics_view: mv1\n" + tt.yaml
				if inline {
					path = "metrics_views/mv1.yaml"
					content = `type: metrics_view
version: 1
model: m1
dimensions:
  - name: foo
    expression: foo
  - name: bar
    expression: bar
measures:
  - name: count
    expression: COUNT(*)
  - name: total
    expression: SUM(foo)
explore:
  ` + strings.ReplaceAll(tt.yaml, "\n", "\n  ")
				}
				repo := makeRepo(t, map[string]string{
					"rill.yaml": "",
					path:        content,
				})
				p, err := Parse(context.Background(), repo, "", "", "duckdb", true)
				require.NoError(t, err)
				if tt.wantErr != "" {
					require.Len(t, p.Errors, 1)
					require.Equal(t, "/"+path, p.Errors[0].FilePath)
					require.Contains(t, p.Errors[0].Message, tt.wantErr)
					for _, resource := range p.Resources {
						require.NotEqual(t, ResourceKindExplore, resource.Name.Kind)
					}
					return
				}
				require.Empty(t, p.Errors)
				var explore *runtimev1.ExploreSpec
				for _, resource := range p.Resources {
					if resource.Name.Kind == ResourceKindExplore {
						explore = resource.ExploreSpec
					}
				}
				require.NotNil(t, explore)
				require.Equal(t, []string{"foo", "bar"}, explore.Dimensions)
				require.Equal(t, []string{"count", "total"}, explore.Measures)
				require.Equal(t, []string{"bar", "foo"}, explore.DefaultPreset.Dimensions)
				require.Equal(t, []string{"total", "count"}, explore.DefaultPreset.Measures)
			})
		}
	}
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
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}
