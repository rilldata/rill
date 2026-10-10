package canvas_test

import (
	"slices"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/canvas"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPruneRendererProperties(t *testing.T) {
	restricted := map[string]bool{"gross_margin": true, "customer_name": true}
	canAccess := func(field string) bool { return !restricted[field] }

	tests := []struct {
		name     string
		renderer string
		props    string
		keep     bool
		changed  bool
		want     string // expected props after pruning; empty if the component is removed or unchanged
	}{
		{
			name:     "unrestricted fields are unchanged",
			renderer: "kpi_grid",
			props:    `{metrics_view: mv, measures: [total_revenue, order_count]}`,
			keep:     true,
		},
		{
			name:     "pivot drops a restricted measure and its sort",
			renderer: "pivot",
			props: `
metrics_view: mv
measures: [total_revenue, {name: gross_margin, format_preset: percentage}]
row_dimensions: [order_date_rill_TIME_GRAIN_MONTH, region]
sort_by: gross_margin
sort_dir: desc
`,
			keep:    true,
			changed: true,
			want: `
metrics_view: mv
measures: [total_revenue]
row_dimensions: [order_date_rill_TIME_GRAIN_MONTH, region]
`,
		},
		{
			name:     "pivot with only restricted fields is removed",
			renderer: "pivot",
			props:    `{metrics_view: mv, measures: [gross_margin], row_dimensions: [customer_name]}`,
			keep:     false,
			changed:  true,
		},
		{
			name:     "table drops a restricted column",
			renderer: "table",
			props:    `{metrics_view: mv, columns: [region, {name: customer_name, width: 200}, total_revenue]}`,
			keep:     true,
			changed:  true,
			want:     `{metrics_view: mv, columns: [region, total_revenue]}`,
		},
		{
			name:     "table keeps a sort on an allowed field",
			renderer: "table",
			props:    `{metrics_view: mv, columns: [region, gross_margin], sort_by: region}`,
			keep:     true,
			changed:  true,
			want:     `{metrics_view: mv, columns: [region], sort_by: region}`,
		},
		{
			name:     "kpi grid drops a restricted measure",
			renderer: "kpi_grid",
			props:    `{metrics_view: mv, measures: [total_revenue, gross_margin]}`,
			keep:     true,
			changed:  true,
			want:     `{metrics_view: mv, measures: [total_revenue]}`,
		},
		{
			name:     "kpi grid with only restricted measures is removed",
			renderer: "kpi_grid",
			props:    `{metrics_view: mv, measures: [gross_margin]}`,
			keep:     false,
			changed:  true,
		},
		{
			name:     "kpi on a restricted measure is removed",
			renderer: "kpi",
			props:    `{metrics_view: mv, measure: gross_margin}`,
			keep:     false,
			changed:  true,
		},
		{
			name:     "chart with a restricted y field is removed",
			renderer: "line_chart",
			props:    `{metrics_view: mv, x: {field: order_date, type: temporal}, y: {field: gross_margin, type: quantitative}}`,
			keep:     false,
			changed:  true,
		},
		{
			name:     "chart drops a restricted color dimension",
			renderer: "bar_chart",
			props:    `{metrics_view: mv, x: {field: region, type: nominal}, y: {field: total_revenue, type: quantitative}, color: {field: customer_name, type: nominal}}`,
			keep:     true,
			changed:  true,
			want:     `{metrics_view: mv, x: {field: region, type: nominal}, y: {field: total_revenue, type: quantitative}}`,
		},
		{
			name:     "chart drops a restricted measure from y.fields",
			renderer: "line_chart",
			props:    `{metrics_view: mv, x: {field: order_date, type: temporal}, y: {field: total_revenue, fields: [total_revenue, gross_margin]}}`,
			keep:     true,
			changed:  true,
			want:     `{metrics_view: mv, x: {field: order_date, type: temporal}, y: {field: total_revenue, fields: [total_revenue]}}`,
		},
		{
			name:     "chart with allowed fields and a filter on a restricted field is removed",
			renderer: "bar_chart",
			props:    `{metrics_view: mv, x: {field: region, type: nominal}, y: {field: total_revenue, type: quantitative}, dimension_filters: "customer_name EQ 'acme'"}`,
			keep:     false,
			changed:  true,
		},
		{
			name:     "chart with a filter the frontend can't parse is kept",
			renderer: "bar_chart",
			props:    `{metrics_view: mv, x: {field: region, type: nominal}, y: {field: total_revenue, type: quantitative}, dimension_filters: "customer_name = 'acme'"}`,
			keep:     true,
		},
		{
			name:     "color literals and virtual colors are kept",
			renderer: "combo_chart",
			props:    `{metrics_view: mv, x: {field: region}, y1: {field: total_revenue}, y2: {field: order_count}, color: {field: rill_measures, type: value}}`,
			keep:     true,
		},
		{
			name:     "pie without its category dimension is removed",
			renderer: "pie_chart",
			props:    `{metrics_view: mv, measure: {field: total_revenue}, color: {field: customer_name}}`,
			keep:     false,
			changed:  true,
		},
		{
			name:     "heatmap with a restricted color measure is removed",
			renderer: "heatmap",
			props:    `{metrics_view: mv, x: {field: region}, y: {field: order_date}, color: {field: gross_margin}}`,
			keep:     false,
			changed:  true,
		},
		{
			name:     "leaderboard drops a restricted dimension",
			renderer: "leaderboard",
			props:    `{metrics_view: mv, dimensions: [region, customer_name], measures: [total_revenue]}`,
			keep:     true,
			changed:  true,
			want:     `{metrics_view: mv, dimensions: [region], measures: [total_revenue]}`,
		},
		{
			name:     "leaderboard without dimensions left is removed",
			renderer: "leaderboard",
			props:    `{metrics_view: mv, dimensions: [customer_name], measures: [total_revenue]}`,
			keep:     false,
			changed:  true,
		},
		{
			name:     "map drops a restricted tooltip dimension",
			renderer: "map",
			props:    `{metrics_view: mv, geo_dimension: {field: region}, color: {measure: total_revenue}, tooltip_dimension: {field: customer_name}}`,
			keep:     true,
			changed:  true,
			want:     `{metrics_view: mv, geo_dimension: {field: region}, color: {measure: total_revenue}}`,
		},
		{
			name:     "adhoc measure on a restricted measure is dropped with its definition",
			renderer: "table",
			props: `
metrics_view: mv
adhoc_measures:
- {name: margin_share, expression: gross_margin / total_revenue}
- {name: revenue_per_order, expression: total_revenue / order_count}
columns: [region, margin_share, revenue_per_order]
`,
			keep:    true,
			changed: true,
			want: `
metrics_view: mv
adhoc_measures:
- {name: revenue_per_order, expression: total_revenue / order_count}
columns: [region, revenue_per_order]
`,
		},
		{
			name:     "markdown is never pruned",
			renderer: "markdown",
			props:    `{content: "Total is {{ metrics_sql \"SELECT gross_margin FROM mv\" }}"}`,
			keep:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := mustYAMLMap(t, tt.props)
			keep, changed := canvas.PruneRendererProperties(tt.renderer, props, "order_date", canAccess)
			require.Equal(t, tt.keep, keep)
			require.Equal(t, tt.changed, changed)
			if tt.keep && tt.changed {
				require.Equal(t, mustYAMLMap(t, tt.want), props)
			}
			if tt.keep && !tt.changed {
				require.Equal(t, mustYAMLMap(t, tt.props), props)
			}
		})
	}
}

func TestQueryReferencesRestrictedFields(t *testing.T) {
	restricted := map[string]bool{"gross_margin": true, "customer_name": true}
	canAccess := func(field string) bool { return !restricted[field] }

	tests := []struct {
		name  string
		query *metricsview.Query
		want  bool
	}{
		{
			name:  "allowed fields",
			query: &metricsview.Query{Dimensions: []metricsview.Dimension{{Name: "region"}}, Measures: []metricsview.Measure{{Name: "total_revenue"}}},
			want:  false,
		},
		{
			name:  "restricted measure",
			query: &metricsview.Query{Dimensions: []metricsview.Dimension{{Name: "region"}}, Measures: []metricsview.Measure{{Name: "gross_margin"}}},
			want:  true,
		},
		{
			name: "time floor of the time dimension",
			query: &metricsview.Query{Dimensions: []metricsview.Dimension{{
				Name:    "order_date_month",
				Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "order_date", Grain: metricsview.TimeGrainMonth}},
			}}},
			want: false,
		},
		{
			name: "restricted dimension in a nested where condition",
			query: &metricsview.Query{
				Measures: []metricsview.Measure{{Name: "total_revenue"}},
				Where: &metricsview.Expression{Condition: &metricsview.Condition{
					Operator: metricsview.OperatorAnd,
					Expressions: []*metricsview.Expression{
						{Condition: &metricsview.Condition{Operator: metricsview.OperatorEq, Expressions: []*metricsview.Expression{{Name: "region"}, {Value: "US"}}}},
						{Condition: &metricsview.Condition{Operator: metricsview.OperatorEq, Expressions: []*metricsview.Expression{{Name: "customer_name"}, {Value: "acme"}}}},
					},
				}},
			},
			want: true,
		},
		{
			name: "restricted measure in a having subquery",
			query: &metricsview.Query{
				Measures: []metricsview.Measure{{Name: "total_revenue"}},
				Having: &metricsview.Expression{Subquery: &metricsview.Subquery{
					Dimension: metricsview.Dimension{Name: "region"},
					Measures:  []metricsview.Measure{{Name: "gross_margin"}},
				}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, canvas.QueryReferencesRestrictedFields(tt.query, "order_date", canAccess))
		})
	}
}

func TestFilterReferencesRestrictedFields(t *testing.T) {
	// Like ResolvedSecurity.CanAccessField, fields that aren't explicitly allowed are restricted.
	allowed := map[string]bool{"region": true, "country name": true, "revenue": true, "orders": true}

	tests := []struct {
		name       string
		filter     string
		fields     []string // the fields the query engine checks; none if the frontend can't parse the filter, since it then doesn't apply it
		restricted bool
	}{
		{name: "no filter", filter: ""},
		{name: "dimension filters", filter: "region IN ('US','CA') AND region NEQ 'MX'", fields: []string{"region"}},
		{name: "keywords in any case", filter: "region not in list ('US') and region nlike '%x%'", fields: []string{"region"}},
		{name: "quoted field name", filter: `"country name" EQ 'Canada'`, fields: []string{"country name"}},
		{name: "restricted quoted field name", filter: `"customer\"name" EQ 'acme'`, fields: []string{`customer"name`}, restricted: true},
		{name: "quoted string on the left of a comparison is a field", filter: "'customer_name' EQ 'acme'", fields: []string{"customer_name"}, restricted: true},
		{name: "field named like a keyword", filter: "list EQ 'private'", fields: []string{"list"}, restricted: true},
		{name: "field named like a comparison value", filter: "region_delta EQ 'private'", fields: []string{"region_delta"}, restricted: true},
		{name: "values are not fields", filter: `region IN ('customer_name', 'O\'Brien') AND region LIKE '%margin%'`, fields: []string{"region"}},
		{name: "time dimension", filter: "order_date GTE '2024-01-01'"},
		{name: "measure filter", filter: "region HAVING (revenue GT 100 AND orders LT 5)", fields: []string{"region", "revenue", "orders"}},
		{name: "measure filter on a comparison value", filter: "region HAVING (revenue_delta GT 0)", fields: []string{"region", "revenue_delta"}, restricted: true},
		{name: "measure filter grouped by a restricted dimension", filter: "customer_name HAVING (revenue GT 100)", fields: []string{"customer_name", "revenue"}, restricted: true},
		{name: "nested condition", filter: "(region EQ 'x' OR customer_name EQ 'y') AND region IN ([1, 2], {'k': 'v'})", fields: []string{"region", "customer_name"}, restricted: true},
		{name: "one switch from AND to OR", filter: "region EQ 'x' AND region EQ 'y' OR customer_name EQ -1.5", fields: []string{"region", "customer_name"}, restricted: true},
		{name: "SQL operators", filter: "customer_name = 'acme'"},
		{name: "two switches between AND and OR", filter: "region EQ 'x' AND region EQ 'y' OR customer_name EQ 'z' AND region EQ 'w'"},
		{name: "double-quoted value", filter: `customer_name EQ "acme"`},
		{name: "incomplete filter", filter: "customer_name EQ 'acme' AND"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fields []string
			canvas.FilterReferencesRestrictedFields(tt.filter, "order_date", func(field string) bool {
				if !slices.Contains(fields, field) {
					fields = append(fields, field)
				}
				return true
			})
			require.ElementsMatch(t, tt.fields, fields)
			require.Equal(t, tt.restricted, canvas.FilterReferencesRestrictedFields(tt.filter, "order_date", func(field string) bool { return allowed[field] }))
		})
	}
}

func TestPruneRows(t *testing.T) {
	removed := map[string]bool{"cost_table": true, "pipeline_chart": true, "cohorts_chart": true, "leaderboard": true}
	keep := canvas.Keep{Item: func(item *runtimev1.CanvasItem) bool { return !removed[item.Component] }}

	t.Run("nothing removed returns the input", func(t *testing.T) {
		rows := []*runtimev1.CanvasRow{row(item("kpis", 12))}
		res, paths := canvas.PruneRows(rows, keep)
		require.Empty(t, paths)
		require.Same(t, rows[0], res[0])
	})

	t.Run("remaining items keep their proportions", func(t *testing.T) {
		rows := []*runtimev1.CanvasRow{
			row(item("margin_chart", 8), item("cost_table", 4)),
			row(item("a", 6), item("b", 3), item("leaderboard", 3)),
			row(item("c", 5), item("leaderboard", 4), item("d", 3)),
		}
		res, paths := canvas.PruneRows(rows, keep)
		require.Equal(t, []string{"rows.0.items.1", "rows.1.items.2", "rows.2.items.1"}, paths)
		require.Equal(t, []uint32{12}, widths(res[0]))
		require.Equal(t, []uint32{8, 4}, widths(res[1]))
		require.Equal(t, []uint32{8, 4}, widths(res[2]))

		// The input is not modified
		require.Equal(t, []uint32{8, 4}, widths(rows[0]))
	})

	t.Run("rows without widths are left for the frontend to split", func(t *testing.T) {
		rows := []*runtimev1.CanvasRow{row(&runtimev1.CanvasItem{Component: "a"}, &runtimev1.CanvasItem{Component: "leaderboard"})}
		res, _ := canvas.PruneRows(rows, keep)
		require.Len(t, res[0].Items, 1)
		require.Nil(t, res[0].Items[0].Width)
	})

	t.Run("rows, tabs and groups that lose all their content are removed", func(t *testing.T) {
		rows := []*runtimev1.CanvasRow{
			row(item("kpis", 12)),
			row(item("cost_table", 12)),
			tabGroup("deep_dives",
				tab("pipeline", row(item("pipeline_chart", 12))),
				tab("cohorts", row(item("cohorts_chart", 12))),
			),
			tabGroup("more",
				tab("pipeline", row(item("pipeline_chart", 12))),
				tab("regions", row(item("region_chart", 12))),
				tab("notes"),
			),
		}
		res, paths := canvas.PruneRows(rows, keep)
		require.Equal(t, []string{
			"rows.1.items.0",
			"rows.2.tabs.0.rows.0.items.0",
			"rows.2.tabs.1.rows.0.items.0",
			"rows.3.tabs.0.rows.0.items.0",
		}, paths)
		require.Len(t, res, 2)
		require.Equal(t, "kpis", res[0].Items[0].Component)

		// The group keeps its name, its remaining tab and the tab that was authored empty
		group := res[1].GetTabGroup()
		require.Equal(t, "more", group.Name)
		require.Len(t, group.Tabs, 2)
		require.Equal(t, "regions", group.Tabs[0].Name)
		require.Equal(t, "notes", group.Tabs[1].Name)

		// The input is not modified
		require.Len(t, rows[3].GetTabGroup().Tabs, 3)
	})

	t.Run("rows and tabs rejected by their own condition", func(t *testing.T) {
		hidden := func(condition string) bool { return condition == "hide" }
		conditional := canvas.Keep{
			Row:  func(r *runtimev1.CanvasRow) bool { return !hidden(r.ConditionExpression) },
			Tab:  func(t *runtimev1.CanvasTab) bool { return !hidden(t.ConditionExpression) },
			Item: func(i *runtimev1.CanvasItem) bool { return !hidden(i.ConditionExpression) },
		}
		hiddenRow := row(item("margin_chart", 12))
		hiddenRow.ConditionExpression = "hide"
		hiddenTab := tab("pipeline", row(item("pipeline_chart", 12)))
		hiddenTab.ConditionExpression = "hide"
		rows := []*runtimev1.CanvasRow{
			row(item("kpis", 12)),
			hiddenRow,
			tabGroup("deep_dives", hiddenTab, tab("cohorts", row(item("cohorts_chart", 12)))),
		}

		res, paths := canvas.PruneRows(rows, conditional)
		require.Equal(t, []string{"rows.1", "rows.2.tabs.0"}, paths)
		require.Len(t, res, 2)
		require.Equal(t, []string{"cohorts"}, []string{res[1].GetTabGroup().Tabs[0].Name})
		require.Equal(t, []string{"hide"}, canvas.Conditions(rows))

		canvas.StripConditions(res)
		require.Empty(t, canvas.Conditions(res))
	})
}

func mustYAMLMap(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(s), &m))
	// Normalize to the shapes structpb.Struct.AsMap produces: []any and map[string]any with float64 numbers.
	return normalize(m).(map[string]any)
}

func normalize(v any) any {
	switch v := v.(type) {
	case map[string]any:
		for k, e := range v {
			v[k] = normalize(e)
		}
		return v
	case []any:
		for i, e := range v {
			v[i] = normalize(e)
		}
		return v
	case int:
		return float64(v)
	default:
		return v
	}
}

func item(component string, width uint32) *runtimev1.CanvasItem {
	return &runtimev1.CanvasItem{Component: component, Width: &width}
}

func row(items ...*runtimev1.CanvasItem) *runtimev1.CanvasRow {
	return &runtimev1.CanvasRow{Items: items}
}

func tab(name string, rows ...*runtimev1.CanvasRow) *runtimev1.CanvasTab {
	return &runtimev1.CanvasTab{Name: name, DisplayName: name, Rows: rows}
}

func tabGroup(name string, tabs ...*runtimev1.CanvasTab) *runtimev1.CanvasRow {
	return &runtimev1.CanvasRow{TabGroup: &runtimev1.CanvasTabGroup{Name: name, Tabs: tabs}}
}

func widths(r *runtimev1.CanvasRow) []uint32 {
	var res []uint32
	for _, it := range r.Items {
		res = append(res, *it.Width)
	}
	return res
}
