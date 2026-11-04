package canvas_test

import (
	"slices"
	"testing"

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
			name:     "chart whose y.fields are all restricted and that has no y.field is removed",
			renderer: "bar_chart",
			props:    `{metrics_view: mv, x: {field: region, type: nominal}, y: {fields: [gross_margin]}}`,
			keep:     false,
			changed:  true,
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
			name:     "map drops a regions layer with a restricted measure",
			renderer: "map",
			props:    `{metrics_view: mv, geo_dimension: {field: region}, color: {measure: total_revenue}, regions: {dimension: country, boundaries: countries, color: {measure: gross_margin}}}`,
			keep:     true,
			changed:  true,
			want:     `{metrics_view: mv, geo_dimension: {field: region}, color: {measure: total_revenue}}`,
		},
		{
			name:     "map drops a points layer with a restricted measure but keeps regions",
			renderer: "map",
			props:    `{metrics_view: mv, geo_dimension: {field: region}, color: {measure: gross_margin}, size_measure: {field: order_count}, regions: {dimension: country, boundaries: countries, color: {measure: total_revenue}}}`,
			keep:     true,
			changed:  true,
			want:     `{metrics_view: mv, regions: {dimension: country, boundaries: countries, color: {measure: total_revenue}}}`,
		},
		{
			name:     "map with only a restricted regions layer is removed",
			renderer: "map",
			props:    `{metrics_view: mv, regions: {dimension: customer_name, boundaries: countries, color: {measure: total_revenue}}}`,
			keep:     false,
			changed:  true,
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
		fields     []string // the fields checked against canAccess
		restricted bool
	}{
		{name: "no filter", filter: ""},
		{name: "dimension filters", filter: "region IN ('US','CA') AND region NEQ 'MX'", fields: []string{"region"}},
		{name: "keywords in any case", filter: "region not in list ('US') and region nlike '%x%'", fields: []string{"region"}},
		{name: "quoted field name", filter: `"country name" EQ 'Canada'`, fields: []string{"country name"}},
		// Escapes in quoted field names aren't decoded, so the name doesn't match its field and counts as restricted.
		{name: "escaped quote in a quoted field name", filter: `"customer\"name" EQ 'acme'`, fields: []string{`customer\"name`}, restricted: true},
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
		{name: "NOT IN LIST and NOT LIKE are operators, not fields", filter: "region NOT IN LIST ('x') AND region NOT LIKE '%y%'", fields: []string{"region"}},
		// Fields named like an operator are assumed not to exist, so they're missed.
		{name: "field named like an operator", filter: "in IN ('x') AND not EQ 'y'"},
		// The frontend can't parse these, so it doesn't apply them; the fields before an operator still count.
		{name: "SQL operators", filter: "customer_name = 'acme'"},
		{name: "two switches between AND and OR", filter: "region EQ 'x' AND region EQ 'y' OR customer_name EQ 'z' AND region EQ 'w'", fields: []string{"region", "customer_name"}, restricted: true},
		{name: "double-quoted value", filter: `customer_name EQ "acme"`, fields: []string{"customer_name"}, restricted: true},
		{name: "incomplete filter", filter: "customer_name EQ 'acme' AND", fields: []string{"customer_name"}, restricted: true},
		{name: "unterminated string", filter: "region EQ 'acme", fields: []string{"region"}},
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
