package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestCreateChartDashboardFilters(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/orders.yaml": `
type: model
materialize: true
sql: |
  SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS event_time, 'United States' AS country, 'desktop' AS device, 100 AS revenue
  UNION ALL
  SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS event_time, 'Denmark' AS country, 'mobile' AS device, 10 AS revenue
`,
			"metrics/orders.yaml": `
type: metrics_view
model: orders
timeseries: event_time
dimensions:
- column: country
- column: device
measures:
- name: revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	s := newSession(t, rt, instanceID)

	// chartArgs builds a bar chart of revenue by country on the "orders" metrics view, with the given 'where' (omitted when nil).
	chartArgs := func(where map[string]any) ai.CreateChartArgs {
		spec := map[string]any{
			"metrics_view": "orders",
			"time_range":   map[string]any{"start": "2025-01-01T00:00:00Z", "end": "2025-01-02T00:00:00Z"},
			"x":            map[string]any{"field": "country", "type": "nominal"},
			"y":            map[string]any{"field": "revenue", "type": "quantitative"},
		}
		if where != nil {
			spec["where"] = where
		}
		return ai.CreateChartArgs{"chart_type": "bar_chart", "spec": spec}
	}
	// cond builds a chart 'where' leaf condition such as country = 'Denmark', in the JSON shape the model sends.
	cond := func(op, dimension string, value any) map[string]any {
		return map[string]any{"cond": map[string]any{
			"op":    op,
			"exprs": []any{map[string]any{"name": dimension}, map[string]any{"val": value}},
		}}
	}
	and := func(exprs ...any) map[string]any {
		return map[string]any{"cond": map[string]any{"op": "and", "exprs": exprs}}
	}
	// dashboardFilter builds a dashboard filter such as country IN ('Denmark'), in the shape the dashboards send.
	dashboardFilter := func(dimension string, value any) *metricsview.Expression {
		return &metricsview.Expression{Condition: &metricsview.Condition{
			Operator:    metricsview.OperatorIn,
			Expressions: []*metricsview.Expression{{Name: dimension}, {Value: value}},
		}}
	}

	// The user is viewing the "orders" dashboard filtered to Denmark.
	denmarkCtx := ai.WithDashboardFilters(t.Context(), map[string]*metricsview.Expression{"orders": dashboardFilter("country", "Denmark")})

	t.Run("rejects a chart without a filter", func(t *testing.T) {
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, nil, chartArgs(nil))
		require.ErrorContains(t, err, "the chart has no 'where' filter")
		require.ErrorContains(t, err, "country IN ('Denmark')")
	})

	t.Run("rejects a chart whose filter drops the dashboard's filter", func(t *testing.T) {
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, nil, chartArgs(cond("eq", "device", "mobile")))
		require.ErrorContains(t, err, `drops the dashboard's filter on "country"`)
	})

	t.Run("rejects a chart that changes the dashboard's filter values", func(t *testing.T) {
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, nil, chartArgs(cond("in", "country", []any{"Denmark", "United States"})))
		require.ErrorContains(t, err, `filters "country" to [Denmark United States], but the dashboard the user is viewing filters it to [Denmark]`)
	})

	t.Run("accepts a chart that restates the dashboard's filter", func(t *testing.T) {
		var res *ai.CreateChartResult
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, &res, chartArgs(cond("eq", "country", "Denmark")))
		require.NoError(t, err)
		require.NotNil(t, res.Spec["where"])
	})

	t.Run("accepts a chart that adds its own filter to the dashboard's", func(t *testing.T) {
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, nil, chartArgs(and(cond("in", "country", []any{"Denmark"}), cond("eq", "device", "mobile"))))
		require.NoError(t, err)
	})

	t.Run("accepts a chart that narrows the dashboard's filter", func(t *testing.T) {
		ctx := ai.WithDashboardFilters(t.Context(), map[string]*metricsview.Expression{"orders": dashboardFilter("country", []any{"Denmark", "United States"})})
		_, err := s.CallTool(ctx, ai.RoleUser, ai.CreateChartName, nil, chartArgs(cond("eq", "country", "Denmark")))
		require.NoError(t, err)
	})

	t.Run("accepts a chart that restates an excluding filter on the same dimension", func(t *testing.T) {
		excludeDenmark := &metricsview.Expression{Condition: &metricsview.Condition{
			Operator:    metricsview.OperatorNin,
			Expressions: []*metricsview.Expression{{Name: "country"}, {Value: []any{"Denmark"}}},
		}}
		ctx := ai.WithDashboardFilters(t.Context(), map[string]*metricsview.Expression{"orders": excludeDenmark})
		_, err := s.CallTool(ctx, ai.RoleUser, ai.CreateChartName, nil, chartArgs(cond("eq", "country", "United States")))
		require.NoError(t, err)
	})

	t.Run("accepts an explicit empty filter", func(t *testing.T) {
		var res *ai.CreateChartResult
		_, err := s.CallTool(denmarkCtx, ai.RoleUser, ai.CreateChartName, &res, chartArgs(and()))
		require.NoError(t, err)
		require.Equal(t, "bar_chart", res.ChartType)
	})

	t.Run("ignores filters on other metrics views", func(t *testing.T) {
		ctx := ai.WithDashboardFilters(t.Context(), map[string]*metricsview.Expression{"other": dashboardFilter("country", "Denmark")})
		_, err := s.CallTool(ctx, ai.RoleUser, ai.CreateChartName, nil, chartArgs(nil))
		require.NoError(t, err)
	})

	t.Run("ignores empty dashboard filters", func(t *testing.T) {
		emptyFilter := &metricsview.Expression{Condition: &metricsview.Condition{Operator: metricsview.OperatorAnd}}
		ctx := ai.WithDashboardFilters(t.Context(), map[string]*metricsview.Expression{"orders": emptyFilter})
		_, err := s.CallTool(ctx, ai.RoleUser, ai.CreateChartName, nil, chartArgs(nil))
		require.NoError(t, err)
	})

	t.Run("accepts a chart when the dashboard has no filters", func(t *testing.T) {
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.CreateChartName, nil, chartArgs(nil))
		require.NoError(t, err)
	})
}
