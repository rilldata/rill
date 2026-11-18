package reconcilers_test

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

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

func TestExploreFilters(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	mv := `
version: 1
type: metrics_view
model: m1
timeseries: t1
dimensions:
- column: t2
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
  exclude:
    - if: "{{ not .user.admin }}"
      names: ['internal', 'y']
`
	testruntime.PutFiles(t, rt, id, map[string]string{
		"models/m1.sql": `SELECT '2025-11-20T00:00:00Z'::TIMESTAMP as t1, '2025-11-20T00:00:00Z'::TIMESTAMP as t2, 'foo' as foo, 'bar' as bar, 'int' as internal, 1 as x, 2 as y`,
		// The inline explore takes the same filter properties as a standalone explore.
		"metrics_views/mv1.yaml": mv + `
explore:
  name: inline
  filters:
    pinned: [bar]
    required: [foo]
  defaults:
    filters: foo = 'a'
`,
		"explores/e1.yaml": `
type: explore
metrics_view: mv1
filters:
  pinned: [bar, internal, x]
  required: [foo, internal]
defaults:
  filters: >
    foo IN ('a', 'b') AND internal = 'c'
    AND bar IN (SELECT bar FROM mv1 HAVING x > 1)
    AND bar IN (SELECT bar FROM mv1 HAVING y > 1)
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 0, 0)

	fooEqualsA, err := metricssql.ParseFilter(`foo = 'a'`)
	require.NoError(t, err)
	inline := testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "inline").GetExplore().State.ValidSpec
	require.Equal(t, []string{"bar"}, inline.PinnedFilters)
	require.Equal(t, []string{"foo"}, inline.RequiredFilters)
	require.True(t, proto.Equal(metricsview.ExpressionToProto(fooEqualsA), inline.DefaultPreset.Where))

	e1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "e1")
	require.Equal(t, []string{"bar", "internal", "x"}, e1.GetExplore().State.ValidSpec.PinnedFilters)
	require.Equal(t, []string{"foo", "internal"}, e1.GetExplore().State.ValidSpec.RequiredFilters)
	where := metricsview.NewExpressionFromProto(e1.GetExplore().State.ValidSpec.DefaultPreset.Where)
	require.Equal(t, []string{"bar", "foo", "internal", "x", "y"}, metricsview.AnalyzeExpressionFields(where))

	// A viewer without access to `internal` and `y` gets the explore without the filters on those fields.
	// Otherwise the required filter could never be set, and the default filter would fail every query.
	viewer := &runtime.SecurityClaims{UserAttributes: map[string]any{"admin": false}}
	res, access, err := rt.ApplySecurityPolicy(t.Context(), id, viewer, e1)
	require.NoError(t, err)
	require.True(t, access)
	require.Equal(t, []string{"bar", "x"}, res.GetExplore().State.ValidSpec.PinnedFilters)
	require.Equal(t, []string{"foo"}, res.GetExplore().State.ValidSpec.RequiredFilters)
	where = metricsview.NewExpressionFromProto(res.GetExplore().State.ValidSpec.DefaultPreset.Where)
	sql, err := metricsview.ExpressionToSQL(where)
	require.NoError(t, err)
	require.Equal(t, `(((foo IN ('a', 'b'))) AND (bar IN (SELECT bar FROM metrics_view HAVING x > 1)))`, sql)

	// The stored resource is left as is.
	e1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "e1")
	require.Equal(t, []string{"foo", "internal"}, e1.GetExplore().State.ValidSpec.RequiredFilters)

	// Filters that reference fields the explore doesn't show, or that can't be filtered, are reconcile errors.
	invalid := []struct {
		yaml    string
		wantErr string
	}{
		{
			yaml:    "filters:\n  pinned: [nope]",
			wantErr: `pinned filter "nope" is not a dimension or measure in this explore`,
		},
		{
			yaml:    "dimensions: [bar]\nfilters:\n  required: [foo]",
			wantErr: `required filter "foo" is not a dimension or measure in this explore`,
		},
		{
			yaml:    "defaults:\n  filters: nope = 'a'",
			wantErr: `invalid default filter: "nope" is not a dimension in this explore`,
		},
		{
			yaml:    "defaults:\n  filters: foo = 'a' AND x > 1",
			wantErr: `invalid default filter: measure "x" can only be filtered in the HAVING clause of a subquery`,
		},
		{
			yaml:    "defaults:\n  filters: foo IN (SELECT foo FROM mv1 HAVING bar > 1)",
			wantErr: `invalid default filter: "bar" is not a measure in this explore`,
		},
		{
			yaml:    "filters:\n  required: [t2]",
			wantErr: `required filter "t2" is a time dimension, which can't be filtered`,
		},
		{
			yaml:    "defaults:\n  filters: t2 = '2025-11-20'",
			wantErr: `invalid default filter: "t2" is a time dimension, which can't be filtered`,
		},
	}
	for _, tt := range invalid {
		testruntime.PutFiles(t, rt, id, map[string]string{"explores/e1.yaml": "type: explore\nmetrics_view: mv1\n" + tt.yaml})
		testruntime.ReconcileParserAndWait(t, rt, id)
		testruntime.RequireReconcileState(t, rt, id, 5, 1, 0)
		testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindExplore, "e1", tt.wantErr)
	}

	// A default filter that isn't a valid expression is a parse error, in a standalone and in an inline explore.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"explores/e1.yaml":       "type: explore\nmetrics_view: mv1\ndefaults:\n  filters: foo ==",
		"metrics_views/mv1.yaml": mv + "explore:\n  defaults:\n    filters:\n      mv1: foo = 'a'\n",
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireParseErrors(t, rt, id, map[string]string{
		"/explores/e1.yaml":       "invalid filter expression in defaults",
		"/metrics_views/mv1.yaml": "invalid defaults.filters: must be a single Metrics SQL filter expression",
	})
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
						Dimensions:           []string{"t1", "t2", "foo"},
						Measures:             []string{"x"},
						AllowCustomTimeRange: true,
						DefaultPreset: &runtimev1.ExplorePreset{
							Dimensions:     []string{"t2", "foo"},
							Measures:       []string{"x"},
							ComparisonMode: runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE,
						},
					},
				},
			},
		},
	})
}

func TestExploreAIPrompts(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"models/m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"metrics_views/mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"explores/e1.yaml": `
type: explore
metrics_view: mv1
ai_prompts:
  - What is the total x?
  - label: Foo breakdown
    prompt: Break down x by foo.
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	e1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "e1").GetExplore()
	require.NotNil(t, e1.State.ValidSpec)
	require.Len(t, e1.State.ValidSpec.AiPrompts, 2)
	require.Equal(t, "What is the total x", e1.State.ValidSpec.AiPrompts[0].Label)
	require.Equal(t, "What is the total x?", e1.State.ValidSpec.AiPrompts[0].Prompt)
	require.Equal(t, "Foo breakdown", e1.State.ValidSpec.AiPrompts[1].Label)
	require.Equal(t, "Break down x by foo.", e1.State.ValidSpec.AiPrompts[1].Prompt)

	// An invalid prompt list is a parse error and the explore is not created.
	testruntime.PutFiles(t, rt, id, map[string]string{"explores/e1.yaml": `
type: explore
metrics_view: mv1
ai_prompts:
  - label: Missing prompt
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireParseErrors(t, rt, id, map[string]string{"/explores/e1.yaml": "non-empty prompt"})
}
