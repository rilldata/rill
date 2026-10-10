package server_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/canvas"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestResolveCanvas(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			// Model
			"m1.sql": `
SELECT 'US' AS country
`,
			"m2.sql": `
SELECT 'PA' AS state
`,
			// Metrics view
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
`,
			// Metrics view
			"mv2.yaml": `
type: metrics_view
version: 1
model: m2
dimensions:
- column: state
measures:
- name: count
  expression: COUNT(*)
`,
			// Canvas
			"c1.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_view: mv1
      measure: count
  - kpi:
      metrics_view: mv1
      measure: count
      foo: "{{ .args.foo }}"
      bar: "{{ .env.bar }}"
  - custom_chart:
      metrics_sql: "SELECT state FROM mv2 WHERE state = 'PA'"
`,
		},
		Variables: map[string]string{
			"bar": "bar",
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 9, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c1",
		Args: must(structpb.NewStruct(map[string]any{
			"foo": "foo",
		})),
	})
	require.NoError(t, err)

	// Check canvas is valid
	require.Equal(t, "c1", res.Canvas.Meta.Name.Name)
	require.NotNil(t, res.Canvas.GetCanvas().State.ValidSpec)

	require.Len(t, res.ResolvedComponents, 3)
	comp0Props := res.ResolvedComponents["c1--component-0-0"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Len(t, comp0Props, 2)
	require.Equal(t, "mv1", comp0Props["metrics_view"])
	require.Equal(t, "count", comp0Props["measure"])
	comp1Props := res.ResolvedComponents["c1--component-0-1"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Len(t, comp1Props, 4)
	require.Equal(t, "mv1", comp1Props["metrics_view"])
	// Templates should NOT be resolved - canvas returns raw templates
	require.Equal(t, "{{ .args.foo }}", comp1Props["foo"])
	require.Equal(t, "{{ .env.bar }}", comp1Props["bar"])

	// Check referenced metrics views
	require.Len(t, res.ReferencedMetricsViews, 2)
	require.Equal(t, "m1", res.ReferencedMetricsViews["mv1"].GetMetricsView().State.ValidSpec.Model)
	require.Equal(t, "m2", res.ReferencedMetricsViews["mv2"].GetMetricsView().State.ValidSpec.Model)
}

func TestResolveCanvasWithInvalidSQL(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    `SELECT 'US' AS country`,
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
`,
			"c_invalid.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_view: mv1
      measure: count
  - custom_chart:
      metrics_sql: "INVALID SQL SYNTAX HERE"
  - custom_chart:
      metrics_sql: "SELECT * FROM nonexistent_mv"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 7, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c_invalid",
	})

	// Should still resolve components and their metrics views
	require.Len(t, res.ResolvedComponents, 3, "All components should be resolved even with invalid SQL")
	require.Len(t, res.ReferencedMetricsViews, 1, "Should only include mv1 from valid component")
	require.Contains(t, res.ReferencedMetricsViews, "mv1")
}

func TestResolveCanvasWithTemplatedSQL(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    `SELECT 'US' AS country`,
			"m2.sql":    `SELECT 'CA' AS country`,
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
`,
			"mv2.yaml": `
type: metrics_view
version: 1
model: m2
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
`,
			"c_templated.yaml": `
type: canvas
rows:
- items:
  - custom_chart:
      metrics_sql: "SELECT country FROM {{ .args.metrics_view_name }}"
  - custom_chart:
      metrics_sql: "SELECT country FROM {{ .env.default_mv }}"
`,
		},
		Variables: map[string]string{
			"default_mv": "mv2",
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c_templated",
		Args: must(structpb.NewStruct(map[string]any{
			"metrics_view_name": "mv1",
		})),
	})
	require.NoError(t, err)

	// Templates are NOT resolved by ResolveCanvas anymore - that's the job of ResolveTemplatedString
	// Canvas should still track referenced metrics views by parsing the SQL
	// Note: Template parsing for metrics view tracking happens at a later stage,
	// so templates won't be detected until resolved
	require.Len(t, res.ResolvedComponents, 2)

	// Verify templates are preserved as-is (not resolved)
	comp0Props := res.ResolvedComponents["c_templated--component-0-0"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "SELECT country FROM {{ .args.metrics_view_name }}", comp0Props["metrics_sql"])
	comp1Props := res.ResolvedComponents["c_templated--component-0-1"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "SELECT country FROM {{ .env.default_mv }}", comp1Props["metrics_sql"])
}

func TestResolveCanvasWithEmptyCanvas(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"c_empty.yaml": `
type: canvas
rows: []
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c_empty",
	})
	require.NoError(t, err)

	require.Equal(t, "c_empty", res.Canvas.Meta.Name.Name)
	require.Len(t, res.ResolvedComponents, 0)
	require.Len(t, res.ReferencedMetricsViews, 0)
}

func TestResolveCanvasWithMultipleMetricsViewsReferences(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    `SELECT 'US' AS country`,
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
`,
			"c_duplicate.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_view: mv1
      measure: count
  - kpi:
      metrics_view: mv1
      measure: count
  - custom_chart:
      metrics_sql: "SELECT country FROM mv1"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 7, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c_duplicate",
	})
	require.NoError(t, err)

	require.Len(t, res.ReferencedMetricsViews, 1)
	require.Contains(t, res.ReferencedMetricsViews, "mv1")
	require.Len(t, res.ResolvedComponents, 3)
}

func TestResolveCanvasWithMetricsSQL(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    `SELECT 'US' AS country, 100 AS revenue`,
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: country
measures:
- expression: COUNT(*)
  name: total_records
- expression: SUM(revenue)
  name: total_revenue
`,
			"c_complex.yaml": `
type: canvas
rows:
- items:
  - custom_chart:
      metrics_sql: "SELECT country, total_revenue FROM mv1 WHERE country = 'US'"
  - custom_chart:
      metrics_sql: "SELECT COUNT(*) as count FROM mv1 GROUP BY country HAVING count > 5"
  - custom_chart:
      metrics_sql: "SELECT country FROM mv1 ORDER BY total_revenue DESC LIMIT 10"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 7, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c_complex",
	})
	require.NoError(t, err)

	require.Len(t, res.ReferencedMetricsViews, 1)
	require.Contains(t, res.ReferencedMetricsViews, "mv1")
	require.Len(t, res.ResolvedComponents, 3)

	comp0Props := res.ResolvedComponents["c_complex--component-0-0"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "SELECT country, total_revenue FROM mv1 WHERE country = 'US'", comp0Props["metrics_sql"])
}

func TestResolveCanvasWithSecurity(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			// Model
			"m1.sql": `SELECT 'US' AS country, 1 AS value`,
			// Metrics view
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
- name: sum
  expression: SUM(value)

security:
  access: "'{{ .user.domain }}' = 'rilldata.com'"
  exclude:
  - if: true
    names: [sum]
`,
			// Canvas
			"c1.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_view: mv1
      measure: count

security:
  access: '{{ .user.admin }}'
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Check with open access.
	ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{SkipChecks: true})
	res, err := server.ResolveCanvas(ctx, &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c1",
	})
	require.NoError(t, err)
	require.NotNil(t, res.Canvas)
	require.Len(t, res.ResolvedComponents, 1)
	require.Len(t, res.ReferencedMetricsViews, 1)
	require.Len(t, res.ReferencedMetricsViews["mv1"].GetMetricsView().State.ValidSpec.Measures, 2)

	// Check when doesn't have access to the canvas.
	claims := &runtime.SecurityClaims{
		UserAttributes: map[string]any{"admin": false, "domain": "rilldata.com"},
		Permissions:    []runtime.Permission{runtime.ReadAPI},
	}
	ctx = auth.WithClaims(context.Background(), claims)
	res, err = server.ResolveCanvas(ctx, &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c1",
	})
	require.ErrorIs(t, err, runtime.ErrForbidden)

	// Check metrics view column-level security.
	// The 'sum' measure should be excluded.
	claims = &runtime.SecurityClaims{
		UserAttributes: map[string]any{"admin": true, "domain": "rilldata.com"},
		Permissions:    []runtime.Permission{runtime.ReadAPI},
	}
	ctx = auth.WithClaims(context.Background(), claims)
	res, err = server.ResolveCanvas(ctx, &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c1",
	})
	require.NoError(t, err)
	require.NotNil(t, res.Canvas)
	require.Len(t, res.ResolvedComponents, 1)
	require.Len(t, res.ReferencedMetricsViews, 1)
	require.Len(t, res.ReferencedMetricsViews["mv1"].GetMetricsView().State.ValidSpec.Measures, 1)
	require.Equal(t, res.ReferencedMetricsViews["mv1"].GetMetricsView().State.ValidSpec.Measures[0].Name, "count")

	// Check metrics view access security.
	// Should have access to the canvas, but not the metrics view, so its component and the row it was in are removed.
	claims = &runtime.SecurityClaims{
		UserAttributes: map[string]any{"admin": true, "domain": "notrilldata.com"},
		Permissions:    []runtime.Permission{runtime.ReadAPI},
	}
	ctx = auth.WithClaims(context.Background(), claims)
	res, err = server.ResolveCanvas(ctx, &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c1",
	})
	require.NoError(t, err)
	require.NotNil(t, res.Canvas)
	require.Len(t, res.ResolvedComponents, 0)
	require.Len(t, res.ReferencedMetricsViews, 0)
	require.Len(t, res.Canvas.GetCanvas().State.ValidSpec.Rows, 0)
}

func TestResolveCanvasPrunesInaccessibleContent(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    `SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS order_date, 'US' AS region, 'acme' AS customer_name, 10 AS revenue, 3 AS cost`,
			"m2.sql":    `SELECT 'ops' AS cost_center, 3 AS cost`,
			"m3.sql":    `SELECT 'won' AS stage, 100 AS value`,
			"orders.yaml": `
type: metrics_view
version: 1
model: m1
timeseries: order_date
dimensions:
- column: region
- column: customer_name
measures:
- name: total_revenue
  expression: SUM(revenue)
- name: order_count
  expression: COUNT(*)
- name: gross_margin
  expression: (SUM(revenue) - SUM(cost)) / SUM(revenue)
security:
  access: true
  exclude:
  - if: 'NOT {{ has "finance" .user.groups }}'
    names: [gross_margin, customer_name]
`,
			"costs.yaml": `
type: metrics_view
version: 1
model: m2
dimensions:
- column: cost_center
measures:
- name: total_cost
  expression: SUM(cost)
security:
  access: '{{ .user.admin }}'
`,
			"pipeline.yaml": `
type: metrics_view
version: 1
model: m3
dimensions:
- column: stage
measures:
- name: pipeline_value
  expression: SUM(value)
security:
  access: '{{ has "sales" .user.groups }}'
`,
			"c1.yaml": `
type: canvas
rows:
- items:
  - width: 8
    line_chart:
      metrics_view: orders
      x: {field: order_date, type: temporal}
      y: {field: gross_margin, type: quantitative}
  - width: 4
    table:
      metrics_view: costs
      columns: [cost_center, total_cost]
- items:
  - pivot:
      metrics_view: orders
      row_dimensions: [region]
      measures: [total_revenue, order_count, gross_margin]
      sort_by: gross_margin
  - kpi_grid:
      metrics_view: orders
      measures: [total_revenue]
- name: deep_dives
  tabs:
  - label: Pipeline
    rows:
    - items:
      - kpi:
          metrics_view: pipeline
          measure: pipeline_value
  - label: Overview
    rows:
    - items:
      - kpi_grid:
          metrics_view: orders
          measures: [order_count]
- items:
  - custom_chart:
      metrics_sql: "SELECT stage, pipeline_value FROM pipeline"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 15, 0, 0)

	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	viewer := []runtime.Permission{runtime.ReadAPI}
	editor := []runtime.Permission{runtime.ReadAPI, runtime.EditRepo}
	resolveAs := func(attrs map[string]any, perms []runtime.Permission, req *runtimev1.ResolveCanvasRequest) (*runtimev1.ResolveCanvasResponse, error) {
		ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
			UserAttributes: attrs,
			Permissions:    perms,
		})
		req.InstanceId = instanceID
		req.Canvas = "c1"
		return srv.ResolveCanvas(ctx, req)
	}
	resolve := func(attrs map[string]any, unsafe bool) *runtimev1.ResolveCanvasResponse {
		perms := viewer
		if unsafe {
			perms = editor
		}
		res, err := resolveAs(attrs, perms, &runtimev1.ResolveCanvasRequest{Unsafe: unsafe})
		require.NoError(t, err)
		return res
	}

	// A viewer with access to everything sees the whole canvas.
	res := resolve(map[string]any{"admin": true, "groups": []any{"finance", "sales"}}, false)
	require.Len(t, res.ResolvedComponents, 7)
	require.Len(t, res.ReferencedMetricsViews, 3)
	require.Len(t, res.Canvas.GetCanvas().State.ValidSpec.Rows, 4)

	// A viewer outside finance, sales and admins only sees what they can query.
	res = resolve(map[string]any{"admin": false, "groups": []any{}}, false)
	rows := res.Canvas.GetCanvas().State.ValidSpec.Rows
	require.Len(t, rows, 2)

	// The first row lost the margin chart (restricted y field) and the cost table (inaccessible metrics view), so it's gone.
	// The pivot lost its restricted measure and the sort on it.
	pivot := rows[0].Items[0].Component
	pivotProps := res.ResolvedComponents[pivot].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, []any{"total_revenue", "order_count"}, pivotProps["measures"])
	require.NotContains(t, pivotProps, "sort_by")

	// The tab group lost its Pipeline tab (inaccessible metrics view); the custom chart's metrics_sql targets a forbidden metrics view.
	group := rows[1].GetTabGroup()
	require.Len(t, group.Tabs, 1)
	require.Equal(t, "Overview", group.Tabs[0].DisplayName)

	require.Len(t, res.ResolvedComponents, 3)
	require.Len(t, res.ReferencedMetricsViews, 1)
	require.Contains(t, res.ReferencedMetricsViews, "orders")
	for _, ref := range res.Canvas.Meta.Refs {
		if ref.Kind == runtime.ResourceKindComponent {
			require.Contains(t, res.ResolvedComponents, ref.Name)
		}
	}

	// A token without an 'admin' attribute (e.g. an embed) can't evaluate the costs policy.
	// The cost table is removed instead of the whole canvas failing, and the margin chart widens to fill the row.
	res = resolve(map[string]any{"groups": []any{"finance"}}, false)
	rows = res.Canvas.GetCanvas().State.ValidSpec.Rows
	require.Len(t, rows[0].Items, 1)
	require.Equal(t, uint32(12), rows[0].Items[0].GetWidth())
	require.NotContains(t, res.ReferencedMetricsViews, "costs")

	// The visual editor gets the full canvas, because it maps components back to their YAML positions.
	// Older editors only send unsafe, which includes hidden content for callers who may edit the project.
	restricted := map[string]any{"admin": false, "groups": []any{}}
	res = resolve(restricted, true)
	require.Len(t, res.ResolvedComponents, 7)
	require.Len(t, res.Canvas.GetCanvas().State.ValidSpec.Rows, 4)

	// A viewer who sends unsafe still gets the pruned canvas.
	res, err = resolveAs(restricted, viewer, &runtimev1.ResolveCanvasRequest{Unsafe: true})
	require.NoError(t, err)
	require.Len(t, res.ResolvedComponents, 3)

	// include_hidden requires permission to edit the canvas.
	_, err = resolveAs(restricted, viewer, &runtimev1.ResolveCanvasRequest{IncludeHidden: true})
	require.ErrorIs(t, err, runtime.ErrForbidden)

	// With permission, it returns the full canvas and the paths of the content the caller can't query.
	res, err = resolveAs(restricted, editor, &runtimev1.ResolveCanvasRequest{IncludeHidden: true})
	require.NoError(t, err)
	require.Len(t, res.ResolvedComponents, 7)
	require.Len(t, res.Canvas.GetCanvas().State.ValidSpec.Rows, 4)
	require.ElementsMatch(t, []string{
		"rows.0.items.0",               // margin chart: restricted y field
		"rows.0.items.1",               // cost table: inaccessible metrics view
		"rows.2.tabs.0.rows.0.items.0", // pipeline KPI: inaccessible metrics view
		"rows.3.items.0",               // custom chart: metrics_sql on a forbidden metrics view
	}, res.HiddenPaths)
}

func TestCanvasAndTemplatedString(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"model.sql": `
SELECT 'US' AS country, 100 AS revenue
UNION ALL
SELECT 'UK' AS country, 200 AS revenue
`,
			"mv.yaml": `
type: metrics_view
version: 1
model: model
dimensions:
- column: country
measures:
- name: total_revenue
  expression: SUM(revenue)
`,
			"canvas.yaml": `
type: canvas
rows:
- items:
  - markdown:
      content: The total is {{ metrics_sql "SELECT total_revenue FROM mv WHERE country = 'US'" }}
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		SkipChecks: true,
	})

	// Step 1: Get canvas with unresolved templates
	canvasRes, err := server.ResolveCanvas(ctx, &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "canvas",
	})
	require.NoError(t, err)
	require.Len(t, canvasRes.ResolvedComponents, 1)

	// Verify component has unresolved templates
	comp := canvasRes.ResolvedComponents["canvas--component-0-0"]
	props := comp.GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Contains(t, props["content"], "{{ metrics_sql ")

	// Use ResolveTemplatedString to resolve the content
	titleRes, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       props["content"].(string),
	})
	require.NoError(t, err)
	require.Equal(t, "The total is 100", titleRes.Body)

	// Step 4: Get formatted value using format tokens
	formatRes, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId:      instanceID,
		Body:            props["content"].(string),
		UseFormatTokens: true,
	})
	require.NoError(t, err)
	require.Contains(t, formatRes.Body, "__RILL__FORMAT__")
	require.Contains(t, formatRes.Body, "mv")
	require.Contains(t, formatRes.Body, "total_revenue")
}

func TestCanvasWithKPIGridAndMarkdown(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"bids.sql": `
SELECT 
  DATE '2025-11-04' AS timestamp,
  100 AS total_bids,
  50 AS winning_bids
UNION ALL
SELECT 
  DATE '2025-11-05' AS timestamp,
  150 AS total_bids,
  75 AS winning_bids
UNION ALL
SELECT 
  DATE '2025-11-06' AS timestamp,
  200 AS total_bids,
  100 AS winning_bids
`,
			"bids_metrics.yaml": `
type: metrics_view
version: 1
model: bids
timeseries: timestamp
dimensions:
- column: timestamp
  name: timestamp
measures:
- name: total_bids
  expression: SUM(total_bids)
- name: winning_bids
  expression: SUM(winning_bids)
`,
			"canvas.yaml": `
type: canvas
display_name: "Canvas Dashboard"
defaults:
  time_range: PT24H
  comparison_mode: time
rows:
  - items:
      - kpi_grid:
          comparison:
            - delta
            - percent_change
          metrics_view: bids_metrics
          measures:
            - total_bids
        width: 12
    height: 40px
  - items:
      - markdown:
          alignment:
            horizontal: left
            vertical: middle
          apply_formatting: true
          content: 'This is a cool metric: {{ metrics_sql "select total_bids from bids_metrics" }}'
          description: ""
          title: ""
        width: 12
    height: 40px
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 6, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	ctx := testCtx()

	// Step 1: Resolve the canvas
	canvasRes, err := server.ResolveCanvas(ctx, &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "canvas",
	})
	require.NoError(t, err)
	require.NotNil(t, canvasRes.Canvas)
	require.Equal(t, "canvas", canvasRes.Canvas.Meta.Name.Name)
	require.NotNil(t, canvasRes.Canvas.GetCanvas().State.ValidSpec)

	// Verify we have both components
	require.Len(t, canvasRes.ResolvedComponents, 2)
	require.Contains(t, canvasRes.ResolvedComponents, "canvas--component-0-0") // KPI grid
	require.Contains(t, canvasRes.ResolvedComponents, "canvas--component-1-0") // Markdown

	// Verify the KPI grid component
	kpiGrid := canvasRes.ResolvedComponents["canvas--component-0-0"]
	kpiGridProps := kpiGrid.GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "bids_metrics", kpiGridProps["metrics_view"])
	require.Equal(t, []any{"total_bids"}, kpiGridProps["measures"])
	require.Equal(t, []any{"delta", "percent_change"}, kpiGridProps["comparison"])

	// Verify the markdown component (templates should NOT be resolved by ResolveCanvas)
	markdown := canvasRes.ResolvedComponents["canvas--component-1-0"]
	markdownProps := markdown.GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "This is a cool metric: {{ metrics_sql \"select total_bids from bids_metrics\" }}", markdownProps["content"])

	// Verify referenced metrics views
	require.Len(t, canvasRes.ReferencedMetricsViews, 1)
	require.Contains(t, canvasRes.ReferencedMetricsViews, "bids_metrics")
	require.Equal(t, "bids", canvasRes.ReferencedMetricsViews["bids_metrics"].GetMetricsView().State.ValidSpec.Model)

	// Step 2: Resolve markdown template WITHOUT additional time range
	// This should return the total across all data
	markdownResNoFilter, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       markdownProps["content"].(string),
	})
	require.NoError(t, err)
	require.Equal(t, "This is a cool metric: 450", markdownResNoFilter.Body)

	// Step 3: Test with a WHERE clause to verify additional time range works
	// Note: Simple SELECT without WHERE doesn't auto-apply time filtering
	// The metrics SQL needs explicit time filtering for time ranges to work
	bodyWithTimeFilter := `Total for period: {{ metrics_sql "select total_bids from bids_metrics where timestamp >= '2025-11-04' and timestamp < '2025-11-06'" }}`

	markdownResWithExplicitTime, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       bodyWithTimeFilter,
	})
	require.NoError(t, err)
	// Should include data from 2025-11-04 and 2025-11-05 (100 + 150 = 250 bids)
	require.Equal(t, "Total for period: 250", markdownResWithExplicitTime.Body)

	// Step 4: Resolve with format tokens enabled
	markdownResFormatted, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId:      instanceID,
		Body:            markdownProps["content"].(string),
		UseFormatTokens: true,
	})
	require.NoError(t, err)
	require.Contains(t, markdownResFormatted.Body, "__RILL__FORMAT__")
	require.Contains(t, markdownResFormatted.Body, "bids_metrics")
	require.Contains(t, markdownResFormatted.Body, "total_bids")
	require.Contains(t, markdownResFormatted.Body, "450")

	// Step 5: Verify multiple metrics_sql calls in one template
	multiMetricBody := `Total: {{ metrics_sql "select total_bids from bids_metrics" }}, Winning: {{ metrics_sql "select winning_bids from bids_metrics" }}`
	multiMetricRes, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       multiMetricBody,
	})
	require.NoError(t, err)
	require.Equal(t, "Total: 450, Winning: 225", multiMetricRes.Body)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func TestResolveCanvasConditions(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    `SELECT 'US' AS region, 10 AS revenue`,
			"m2.sql":    `SELECT 'ops' AS cost_center, 3 AS cost`,
			"orders.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: region
measures:
- name: total_revenue
  expression: SUM(revenue)
- name: order_count
  expression: COUNT(*)
`,
			"costs.yaml": `
type: metrics_view
version: 1
model: m2
dimensions:
- column: cost_center
measures:
- name: total_cost
  expression: SUM(cost)
`,
			"c1.yaml": `
type: canvas
rows:
- items:
  - kpi_grid:
      metrics_view: orders
      measures: [total_revenue]
- if: '{{ has "finance" .user.groups }}'
  items:
  - width: 8
    kpi:
      metrics_view: orders
      measure: total_revenue
  - width: 4
    if: '{{ .user.admin }}'
    table:
      metrics_view: costs
      columns: [cost_center, total_cost]
- items:
  - width: 6
    kpi:
      metrics_view: orders
      measure: order_count
  - width: 6
    if: "'{{ .user.domain }}' = 'acme.com'"
    leaderboard:
      metrics_view: orders
      dimensions: [region]
      measures: [total_revenue]
- name: deep_dives
  tabs:
  - label: Pipeline
    if: '{{ has "sales" .user.groups }}'
    rows:
    - items:
      - kpi:
          metrics_view: orders
          measure: total_revenue
  - label: Cohorts
    if: '{{ eq .user.plan "premium" }}'
    rows:
    - items:
      - kpi:
          metrics_view: orders
          measure: order_count
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 13, 0, 0)

	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	priya := map[string]any{"domain": "acme.com", "groups": []any{"finance"}, "admin": false}
	sam := map[string]any{"domain": "acme.com", "groups": []any{"sales"}, "admin": false}
	tenant := map[string]any{"plan": "premium", "tenant_id": "t_42", "embed": true}
	financeAdmin := map[string]any{"domain": "acme.com", "groups": []any{"finance"}, "admin": true}

	viewerCtx := func(attrs map[string]any, rules ...*runtimev1.SecurityRule) context.Context {
		return auth.WithClaims(context.Background(), &runtime.SecurityClaims{
			UserAttributes:  attrs,
			Permissions:     []runtime.Permission{runtime.ReadAPI, runtime.ReadObjects},
			AdditionalRules: rules,
		})
	}

	// layout lists the components on each row (or "tabs:" and the visible tab names for a tab group), with their widths.
	layout := func(rows []*runtimev1.CanvasRow) []string {
		var res []string
		for _, row := range rows {
			if tg := row.GetTabGroup(); tg != nil {
				line := "tabs:"
				for _, tab := range tg.Tabs {
					line += " " + tab.Name
				}
				res = append(res, line)
				continue
			}
			line := ""
			for _, item := range row.Items {
				line += fmt.Sprintf("%s(%d) ", strings.TrimPrefix(item.Component, "c1--component-"), item.GetWidth())
			}
			res = append(res, strings.TrimSpace(line))
		}
		return res
	}
	resolve := func(ctx context.Context) *runtimev1.ResolveCanvasResponse {
		res, err := srv.ResolveCanvas(ctx, &runtimev1.ResolveCanvasRequest{InstanceId: instanceID, Canvas: "c1"})
		require.NoError(t, err)
		require.Empty(t, canvas.Conditions(res.Canvas.GetCanvas().State.ValidSpec.Rows), "conditions must not be sent to viewers")
		for _, ref := range res.Canvas.Meta.Refs {
			if ref.Kind == runtime.ResourceKindComponent {
				require.Contains(t, res.ResolvedComponents, ref.Name, "refs must not name hidden components")
			}
		}
		return res
	}

	// Priya (finance) doesn't see the admin-only cost table, so the margin KPI fills its row.
	res := resolve(viewerCtx(priya))
	require.Equal(t, []string{"0-0(0)", "1-0(12)", "2-0(6) 2-1(6)"}, layout(res.Canvas.GetCanvas().State.ValidSpec.Rows))
	require.Len(t, res.ResolvedComponents, 4)
	require.NotContains(t, res.ReferencedMetricsViews, "costs")

	// Sam (sales) sees the Pipeline tab but not the finance row.
	res = resolve(viewerCtx(sam))
	require.Equal(t, []string{"0-0(0)", "2-0(6) 2-1(6)", "tabs: pipeline"}, layout(res.Canvas.GetCanvas().State.ValidSpec.Rows))

	// The embed tenant has no domain or groups; built-in attributes default, so their conditions are false rather than errors.
	res = resolve(viewerCtx(tenant))
	require.Equal(t, []string{"0-0(0)", "2-0(12)", "tabs: cohorts"}, layout(res.Canvas.GetCanvas().State.ValidSpec.Rows))

	// A finance admin sees the cost table, but has no plan, so the strict custom attribute hides Cohorts.
	res = resolve(viewerCtx(financeAdmin))
	require.Equal(t, []string{"0-0(0)", "1-0(8) 1-1(4)", "2-0(6) 2-1(6)"}, layout(res.Canvas.GetCanvas().State.ValidSpec.Rows))

	// Local development (open access) shows everything, conditions included.
	res, err = srv.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{InstanceId: instanceID, Canvas: "c1"})
	require.NoError(t, err)
	require.Len(t, res.ResolvedComponents, 7)
	require.Len(t, canvas.Conditions(res.Canvas.GetCanvas().State.ValidSpec.Rows), 5)

	// The visual editor gets the full canvas with its conditions, plus what's hidden from the caller.
	editorCtx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		UserAttributes: priya,
		Permissions:    []runtime.Permission{runtime.ReadAPI, runtime.EditRepo},
	})
	res, err = srv.ResolveCanvas(editorCtx, &runtimev1.ResolveCanvasRequest{InstanceId: instanceID, Canvas: "c1", IncludeHidden: true})
	require.NoError(t, err)
	require.Len(t, res.ResolvedComponents, 7)
	require.Len(t, canvas.Conditions(res.Canvas.GetCanvas().State.ValidSpec.Rows), 5)
	require.Equal(t, []string{"rows.1.items.1", "rows.3.tabs.0", "rows.3.tabs.1"}, res.HiddenPaths)

	// Other reads of the canvas resource are pruned the same way.
	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	canvasRes, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindCanvas, Name: "c1"}, false)
	require.NoError(t, err)
	claims := auth.GetClaims(viewerCtx(sam), instanceID)
	pruned, access, err := rt.ApplySecurityPolicy(context.Background(), instanceID, claims, canvasRes)
	require.NoError(t, err)
	require.True(t, access)
	require.Equal(t, []string{"0-0(0)", "2-0(6) 2-1(6)", "tabs: pipeline"}, layout(pruned.GetCanvas().State.ValidSpec.Rows))

	// A hidden inline component can't be read directly.
	_, err = srv.ResolveComponent(viewerCtx(priya), &runtimev1.ResolveComponentRequest{InstanceId: instanceID, Component: "c1--component-1-1"})
	require.ErrorContains(t, err, "does not have access")
	_, err = srv.ResolveComponent(viewerCtx(financeAdmin), &runtimev1.ResolveComponentRequest{InstanceId: instanceID, Component: "c1--component-1-1"})
	require.NoError(t, err)

	// A public link created by Priya only grants what Priya can see on the canvas.
	linkRules := []*runtimev1.SecurityRule{{Rule: &runtimev1.SecurityRule_TransitiveAccess{TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
		Resource: &runtimev1.ResourceName{Kind: runtime.ResourceKindCanvas, Name: "c1"},
	}}}}
	link := auth.GetClaims(viewerCtx(priya, linkRules...), instanceID)
	for name, want := range map[string]bool{"c1--component-1-0": true, "c1--component-1-1": false, "c1--component-g3-t0-0-0": false} {
		cmp, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: name}, false)
		require.NoError(t, err)
		_, access, err := rt.ApplySecurityPolicy(context.Background(), instanceID, link, cmp)
		require.NoError(t, err)
		require.Equal(t, want, access, name)
	}
	costs, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "costs"}, false)
	require.NoError(t, err)
	_, access, err = rt.ApplySecurityPolicy(context.Background(), instanceID, link, costs)
	require.NoError(t, err)
	require.False(t, access, "the link must not grant the metrics view of hidden content")
}

func TestResolveCanvasHidesRestrictedComponents(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    `SELECT 'US' AS region, 'acme' AS customer_name, 10 AS revenue, 3 AS cost`,
			"m2.sql":    `SELECT 'ops' AS cost_center, 3 AS cost`,
			"orders.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: region
- column: customer_name
measures:
- name: total_revenue
  expression: SUM(revenue)
- name: gross_margin
  expression: (SUM(revenue) - SUM(cost)) / SUM(revenue)
security:
  access: true
  exclude:
  - if: 'NOT {{ has "finance" .user.groups }}'
    names: [gross_margin, customer_name]
`,
			"costs.yaml": `
type: metrics_view
version: 1
model: m2
dimensions:
- column: cost_center
measures:
- name: total_cost
  expression: SUM(cost)
`,
			// Canvas a shows its cost table to admins only.
			"a.yaml": `
type: canvas
rows:
- items:
  - if: '{{ .user.admin }}'
    table:
      metrics_view: costs
      columns: [cost_center, total_cost]
`,
			// Canvas b references a's admin-only component, and has components that select or filter on restricted fields.
			"b.yaml": `
type: canvas
rows:
- items:
  - component: a--component-0-0
  - kpi:
      metrics_view: orders
      measure: total_revenue
- items:
  - custom_chart:
      metrics_sql: "SELECT region, gross_margin FROM orders"
- items:
  - custom_chart:
      metrics_sql: "SELECT region, total_revenue FROM orders WHERE customer_name = 'acme'"
- items:
  - custom_chart:
      metrics_sql: "SELECT region, total_revenue FROM orders"
- items:
  - bar_chart:
      metrics_view: orders
      x: {field: region, type: nominal}
      y: {field: total_revenue, type: quantitative}
      dimension_filters: "customer_name EQ 'acme'"
- items:
  - custom_chart:
      metrics_sql: "SELECT region, total_revenue FROM orders"
      dimension_filters: "customer_name IN ('acme')"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 14, 0, 0)

	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	viewer := map[string]any{"admin": false, "groups": []any{}}
	financeAdmin := map[string]any{"admin": true, "groups": []any{"finance"}}
	resolveAs := func(attrs map[string]any, perms []runtime.Permission, includeHidden bool) *runtimev1.ResolveCanvasResponse {
		ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{UserAttributes: attrs, Permissions: perms})
		res, err := srv.ResolveCanvas(ctx, &runtimev1.ResolveCanvasRequest{InstanceId: instanceID, Canvas: "b", IncludeHidden: includeHidden})
		require.NoError(t, err)
		return res
	}
	components := func(res *runtimev1.ResolveCanvasResponse) []string {
		var names []string
		for _, row := range res.Canvas.GetCanvas().State.ValidSpec.Rows {
			for _, item := range row.Items {
				names = append(names, item.Component)
			}
		}
		return names
	}

	// The viewer doesn't get a's admin-only component, nor the components that select or filter on restricted fields.
	res := resolveAs(viewer, []runtime.Permission{runtime.ReadAPI}, false)
	require.Equal(t, []string{"b--component-0-1", "b--component-3-0"}, components(res))
	require.Len(t, res.ResolvedComponents, 2)
	require.NotContains(t, res.ReferencedMetricsViews, "costs")

	// A finance admin gets everything.
	res = resolveAs(financeAdmin, []runtime.Permission{runtime.ReadAPI}, false)
	require.Equal(t, []string{"a--component-0-0", "b--component-0-1", "b--component-1-0", "b--component-2-0", "b--component-3-0", "b--component-4-0", "b--component-5-0"}, components(res))
	require.Contains(t, res.ReferencedMetricsViews, "costs")

	// The editor gets the full canvas, with the paths of the components hidden from the viewer.
	res = resolveAs(viewer, []runtime.Permission{runtime.ReadAPI, runtime.EditRepo}, true)
	require.Len(t, res.ResolvedComponents, 7)
	require.ElementsMatch(t, []string{"rows.0.items.0", "rows.1.items.0", "rows.2.items.0", "rows.4.items.0", "rows.5.items.0"}, res.HiddenPaths)

	// A public link to b only grants a's admin-only component, and the metrics view it uses, if its creator is an admin.
	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	linkRules := []*runtimev1.SecurityRule{{Rule: &runtimev1.SecurityRule_TransitiveAccess{TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
		Resource: &runtimev1.ResourceName{Kind: runtime.ResourceKindCanvas, Name: "b"},
	}}}}
	for _, tc := range []struct {
		creator map[string]any
		want    bool
	}{{viewer, false}, {financeAdmin, true}} {
		link := &runtime.SecurityClaims{UserAttributes: tc.creator, Permissions: []runtime.Permission{runtime.ReadAPI}, AdditionalRules: linkRules}
		for _, name := range []*runtimev1.ResourceName{
			{Kind: runtime.ResourceKindComponent, Name: "a--component-0-0"},
			{Kind: runtime.ResourceKindMetricsView, Name: "costs"},
		} {
			r, err := ctrl.Get(context.Background(), name, false)
			require.NoError(t, err)
			_, access, err := rt.ApplySecurityPolicy(context.Background(), instanceID, link, r)
			require.NoError(t, err)
			require.Equal(t, tc.want, access, name.Name)
		}
	}
}
