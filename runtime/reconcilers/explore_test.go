package reconcilers_test

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestExplores(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"models/m1.sql": `SELECT 'foo' as foo, 'bar' as bar, 'int' as internal, 1 as x, 2 as y`,
		"metrics_views/mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
- column: bar
- column: internal
measures:
- name: x
  expression: sum(x)
- name: y
  expression: sum(y)
security:
  access: true
  row_filter: true
  exclude:
    - if: "{{ not .user.admin }}"
      names: ['internal']
`,
		"explores/e1.yaml": `
type: explore
display_name: Hello
metrics_view: mv1
dimensions:
  exclude: ['internal']
measures: '*'
time_zones: ['UTC', 'America/Los_Angeles']
defaults:
  measures: ['x']
  comparison_mode: time
`,
	})

	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindExplore, Name: "e1"},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "mv1"}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/explores/e1.yaml"},
		},
		Resource: &runtimev1.Resource_Explore{
			Explore: &runtimev1.Explore{
				Spec: &runtimev1.ExploreSpec{
					DisplayName: "Hello",
					MetricsView: "mv1",
					Dimensions:  nil,
					DimensionsSelector: &runtimev1.FieldSelector{
						Invert:   true,
						Selector: &runtimev1.FieldSelector_Fields{Fields: &runtimev1.StringListValue{Values: []string{"internal"}}},
					},
					Measures:             nil,
					MeasuresSelector:     &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
					TimeZones:            []string{"UTC", "America/Los_Angeles"},
					AllowCustomTimeRange: true,
					DefaultPreset: &runtimev1.ExplorePreset{
						DimensionsSelector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
						Measures:           []string{"x"},
						ComparisonMode:     runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_TIME,
					},
				},
				State: &runtimev1.ExploreState{
					ValidSpec: &runtimev1.ExploreSpec{
						DisplayName:          "Hello",
						MetricsView:          "mv1",
						Dimensions:           []string{"foo", "bar"},
						Measures:             []string{"x", "y"},
						TimeZones:            []string{"UTC", "America/Los_Angeles"},
						AllowCustomTimeRange: true,
						DefaultPreset: &runtimev1.ExplorePreset{
							Dimensions:     []string{"foo", "bar"},
							Measures:       []string{"x"},
							ComparisonMode: runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_TIME,
						},
						SecurityRules: []*runtimev1.SecurityRule{
							{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
								ConditionExpression: "true",
								Allow:               true,
							}}},
							{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
								Allow:     true,
								AllFields: true,
							}}},
							{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
								ConditionExpression: "{{ not .user.admin }}",
								Allow:               false,
								Fields:              []string{"internal"},
							}}},
						},
					},
				},
			},
		},
	})
}

func TestExploreTheme(t *testing.T) {
	// Create source and model
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/data/foo.csv": `a,b,c,d,e
1,2,3,4,5
1,2,3,4,5
1,2,3,4,5
`,
		"/sources/foo.yaml": `
type: source
connector: local_file
path: data/foo.csv
`,
		"/models/bar.sql": `SELECT * FROM foo`,
		"/metrics/m1.yaml": `
version: 1
type: metrics_view
model: bar
dimensions:
- column: b
measures:
- expression: count(*)
`,
		"explores/e1.yaml": `
type: explore
metrics_view: m1
display_name: Hello
theme: t1
`,
		"themes/t1.yaml": `
type: theme
colors:
  primary: red
  secondary: grey
`,
	})

	theme := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindTheme, Name: "t1"},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/themes/t1.yaml"},
		},
		Resource: &runtimev1.Resource_Theme{
			Theme: &runtimev1.Theme{
				Spec: &runtimev1.ThemeSpec{
					PrimaryColor: &runtimev1.Color{
						Red:   1,
						Green: 0,
						Blue:  0,
						Alpha: 1,
					},
					SecondaryColor: &runtimev1.Color{
						Red:   0.5019608,
						Green: 0.5019608,
						Blue:  0.5019608,
						Alpha: 1,
					},
					PrimaryColorRaw:   "red",
					SecondaryColorRaw: "grey",
				},
			},
		},
	}

	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 0, 0)
	testruntime.RequireResource(t, rt, id, theme)

	exp := testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "e1")
	require.Equal(t, exp.GetExplore().State.ValidSpec.Theme, "t1")
	require.ElementsMatch(t, exp.Meta.Refs, []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindTheme, Name: "t1"},
		{Kind: runtime.ResourceKindMetricsView, Name: "m1"},
	})

	// make the theme invalid
	testruntime.PutFiles(t, rt, id, map[string]string{
		`themes/t1.yaml`: `
type: theme
colors:
  primary: xxx
  secondary: xxx
`,
	})

	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 2, 1)

	exp = testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "e1")
	require.Nil(t, exp.GetExplore().State.ValidSpec)

	// make the theme valid
	testruntime.PutFiles(t, rt, id, map[string]string{
		`themes/t1.yaml`: `
type: theme
colors:
  primary: red
  secondary: grey
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 0, 0)
}

func TestExploreValidState(t *testing.T) {
	// Create an instance with StageChanges==true
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:        map[string]string{"rill.yaml": ""},
		StageChanges: true,
	})

	// Create basic model + metrics_view + explore
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"e1.yaml": `
type: explore
metrics_view: mv1
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)
	e1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "e1")
	require.NotNil(t, e1.GetExplore().State.ValidSpec)

	// Change the model so it breaks the metrics view and explore, check valid spec is preserved
	testruntime.PutFiles(t, rt, id, map[string]string{"m1.sql": `SELECT 'bar' as bar, 2 as y`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 2, 0)
	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.NotNil(t, mv1.GetMetricsView().State.ValidSpec)
	e1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "e1")
	require.NotNil(t, e1.GetExplore().State.ValidSpec)

	// Fix everything
	testruntime.PutFiles(t, rt, id, map[string]string{"m1.sql": `SELECT 'foo' as foo, 1 as x`})
	testruntime.ReconcileParserAndWait(t, rt, id)

	// Break only the explore. Check valid spec is preserved.
	testruntime.PutFiles(t, rt, id, map[string]string{"e1.yaml": `
type: explore
metrics_view: mv1
dimensions: ['doesnt_exist']
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 1, 0)
	mv1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.NotNil(t, mv1.GetMetricsView().State.ValidSpec)
	e1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "e1")
	require.NotNil(t, e1.GetExplore().State.ValidSpec)
	require.NotEmpty(t, e1.Meta.ReconcileError)
}

func TestExploreDerivedSecurity(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"models/m1.sql": `SELECT 'foo' as foo, 'bar' as bar, 'int' as internal, 1 as x, 2 as y`,
		"metrics_views/mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
- column: bar
- column: internal
measures:
- name: x
  expression: sum(x)
- name: y
  expression: sum(y)
security:
  access: true
  row_filter: true
  exclude:
    - if: "{{ not .user.admin }}"
      names: ['internal']
`,
		"explores/e1.yaml": `
type: explore
display_name: Hello
metrics_view: mv1
dimensions:
  exclude: ['internal']
measures: '*'
time_zones: ['UTC', 'America/Los_Angeles']
defaults:
  measures: ['x']
  comparison_mode: time
security:
  access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'rilldata.com'"
`,
	})

	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindExplore, Name: "e1"},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "mv1"}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/explores/e1.yaml"},
		},
		Resource: &runtimev1.Resource_Explore{
			Explore: &runtimev1.Explore{
				Spec: &runtimev1.ExploreSpec{
					DisplayName: "Hello",
					MetricsView: "mv1",
					Dimensions:  nil,
					DimensionsSelector: &runtimev1.FieldSelector{
						Invert:   true,
						Selector: &runtimev1.FieldSelector_Fields{Fields: &runtimev1.StringListValue{Values: []string{"internal"}}},
					},
					Measures:             nil,
					MeasuresSelector:     &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
					TimeZones:            []string{"UTC", "America/Los_Angeles"},
					AllowCustomTimeRange: true,
					DefaultPreset: &runtimev1.ExplorePreset{
						DimensionsSelector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
						Measures:           []string{"x"},
						ComparisonMode:     runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_TIME,
					},
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
							ConditionExpression: "{{ .user.admin }} OR '{{ .user.domain }}' == 'rilldata.com'",
							Allow:               true,
						}}},
					},
				},
				State: &runtimev1.ExploreState{
					ValidSpec: &runtimev1.ExploreSpec{
						DisplayName:          "Hello",
						MetricsView:          "mv1",
						Dimensions:           []string{"foo", "bar"},
						Measures:             []string{"x", "y"},
						TimeZones:            []string{"UTC", "America/Los_Angeles"},
						AllowCustomTimeRange: true,
						DefaultPreset: &runtimev1.ExplorePreset{
							Dimensions:     []string{"foo", "bar"},
							Measures:       []string{"x"},
							ComparisonMode: runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_TIME,
						},
						SecurityRules: []*runtimev1.SecurityRule{
							// Derived from metrics_view and explore
							{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
								ConditionExpression: "(true) AND ({{ .user.admin }} OR '{{ .user.domain }}' == 'rilldata.com')",
								Allow:               true,
							}}},
							// Inherited from metrics_view
							{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
								Allow:     true,
								AllFields: true,
							}}},
							{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
								ConditionExpression: "{{ not .user.admin }}",
								Fields:              []string{"internal"},
							}}},
						},
					},
				},
			},
		},
	})
}

func TestExploreDataRefreshedOn(t *testing.T) {
	// Create an instance with StageChanges==true
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:        map[string]string{"rill.yaml": ""},
		StageChanges: true,
	})

	// Create basic model + metrics_view + canvas
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `
-- @materialize: true
SELECT 'foo' as foo, 1 as x
`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"e1.yaml": `
type: explore
metrics_view: mv1
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	getAndCheckRefreshedOn := func() time.Time {
		e1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "e1")
		require.NotNil(t, e1.GetExplore().State.DataRefreshedOn)

		mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
		require.NotNil(t, mv1.GetMetricsView().State.DataRefreshedOn)

		require.Equal(t, e1.GetExplore().State.DataRefreshedOn, mv1.GetMetricsView().State.DataRefreshedOn)
		return e1.GetExplore().State.DataRefreshedOn.AsTime()
	}

	refreshedOn1 := getAndCheckRefreshedOn()
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "m1"})
	refreshedOn2 := getAndCheckRefreshedOn()
	require.Greater(t, refreshedOn2, refreshedOn1)
}

func TestExploreTimeDimensions(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"models/m1.sql": `SELECT '2025-11-20T00:00:00Z'::TIMESTAMP as t1, '2025-11-20T00:00:00Z'::TIMESTAMP as t2, 'foo' as foo, 1 as x`,
		"metrics_views/mv1.yaml": `
version: 1
type: metrics_view
model: m1
timeseries: t1
dimensions:
- column: t1
- column: t2
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"explores/e1.yaml": `
type: explore
display_name: Hello
metrics_view: mv1
dimensions: '*'
measures: '*'
defaults:
  dimensions:
    - t2
    - foo
`,
	})

	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindExplore, Name: "e1"},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: "mv1"}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/explores/e1.yaml"},
		},
		Resource: &runtimev1.Resource_Explore{
			Explore: &runtimev1.Explore{
				Spec: &runtimev1.ExploreSpec{
					DisplayName:          "Hello",
					MetricsView:          "mv1",
					Dimensions:           nil,
					DimensionsSelector:   &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
					Measures:             nil,
					MeasuresSelector:     &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
					AllowCustomTimeRange: true,
					DefaultPreset: &runtimev1.ExplorePreset{
						Dimensions:       []string{"t2", "foo"},
						MeasuresSelector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
						ComparisonMode:   runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE,
					},
				},
				State: &runtimev1.ExploreState{
					ValidSpec: &runtimev1.ExploreSpec{
						DisplayName:          "Hello",
						MetricsView:          "mv1",
						Dimensions:           []string{"foo"}, // NOTE: filtered out the time dimensions
						Measures:             []string{"x"},
						AllowCustomTimeRange: true,
						DefaultPreset: &runtimev1.ExplorePreset{
							Dimensions:     []string{"foo"}, // NOTE: filtered out the time dimensions
							Measures:       []string{"x"},
							ComparisonMode: runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE,
						},
					},
				},
			},
		},
	})
}
