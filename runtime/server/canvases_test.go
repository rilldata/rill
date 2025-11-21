package server_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
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
- expression: COUNT(*)
`,
			// Metrics view
			"mv2.yaml": `
type: metrics_view
version: 1
model: m2
dimensions:
- column: state
measures:
- expression: COUNT(*)
`,
			// Canvas
			"c1.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_view: mv1
  - kpi:
      metrics_view: mv1
      foo: "{{ .args.foo }}"
      bar: "{{ .env.bar }}"
  - kpi:
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
	require.Len(t, comp0Props, 1)
	require.Equal(t, "mv1", comp0Props["metrics_view"])
	comp1Props := res.ResolvedComponents["c1--component-0-1"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Len(t, comp1Props, 3)
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
- expression: COUNT(*)
`,
			"c_invalid.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_view: mv1
  - kpi:
      metrics_sql: "INVALID SQL SYNTAX HERE"
  - kpi:
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
- expression: COUNT(*)
`,
			"mv2.yaml": `
type: metrics_view
version: 1
model: m2
dimensions:
- column: country
measures:
- expression: COUNT(*)
`,
			"c_templated.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_sql: "SELECT country FROM {{ .args.metrics_view_name }}"
  - kpi:
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
- expression: COUNT(*)
`,
			"c_duplicate.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_view: mv1
  - kpi:
      metrics_view: mv1
  - kpi:
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
  - kpi:
      metrics_sql: "SELECT country, total_revenue FROM mv1 WHERE country = 'US'"
  - kpi:
      metrics_sql: "SELECT COUNT(*) as count FROM mv1 GROUP BY country HAVING count > 5"
  - kpi:
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

func TestResolveCanvasWithCustomChart(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    `SELECT 'Advertiser A' AS advertiser_name, 1.25 AS avg_bid_price UNION ALL SELECT 'Advertiser B', 2.50`,
			"bids.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: advertiser_name
measures:
- expression: AVG(avg_bid_price)
  name: avg_bid_price
`,
			"c_custom_chart.yaml": `
type: canvas
rows:
- items:
  - custom_chart:
      color: hsl(246, 66%, 50%)
      metrics_sql:
        - select advertiser_name, avg_bid_price from bids order by advertiser_name limit 10
        - select avg_bid_price from bids
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c_custom_chart",
	})
	require.NoError(t, err)

	// Should reference the bids metrics view from both SQL queries
	require.Len(t, res.ReferencedMetricsViews, 1)
	require.Contains(t, res.ReferencedMetricsViews, "bids")
	require.Len(t, res.ResolvedComponents, 1)

	comp0Props := res.ResolvedComponents["c_custom_chart--component-0-0"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "hsl(246, 66%, 50%)", comp0Props["color"])
	require.Contains(t, comp0Props, "metrics_sql")

	metricsSQL := comp0Props["metrics_sql"].([]any)
	require.Len(t, metricsSQL, 2)
	require.Equal(t, "select advertiser_name, avg_bid_price from bids order by advertiser_name limit 10", metricsSQL[0])
	require.Equal(t, "select avg_bid_price from bids", metricsSQL[1])
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
	require.Error(t, err)
	require.ErrorContains(t, err, "does not have access")

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
	// Should have access to the canvas, but not the metrics view.
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
	require.Len(t, res.ResolvedComponents, 1)
	require.Len(t, res.ReferencedMetricsViews, 0)
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
  - kpi:
      title: "Revenue for {{ .user.country }}"
      metrics_sql: "SELECT total_revenue FROM mv WHERE country = '{{ .user.country }}'"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	ctx := auth.WithClaims(context.Background(), &runtime.SecurityClaims{
		SkipChecks: true,
		UserAttributes: map[string]any{
			"country": "US",
		},
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
	require.Equal(t, "Revenue for {{ .user.country }}", props["title"])
	require.Equal(t, "SELECT total_revenue FROM mv WHERE country = '{{ .user.country }}'", props["metrics_sql"])

	// Step 2: Use ResolveTemplatedString to resolve the title
	titleRes, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       props["title"].(string),
	})
	require.NoError(t, err)
	require.Equal(t, "Revenue for US", titleRes.Body)

	// Step 3: Use ResolveTemplatedString with metrics_sql to get the actual value
	// First resolve the template in the SQL, then execute metrics_sql
	valueRes, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId: instanceID,
		Body:       `The total is {{ metrics_sql "SELECT total_revenue FROM mv WHERE country = 'US'" }}`,
	})
	require.NoError(t, err)
	require.Equal(t, "The total is 100", valueRes.Body)

	// Step 4: Get formatted value using format tokens
	formatRes, err := server.ResolveTemplatedString(ctx, &runtimev1.ResolveTemplatedStringRequest{
		InstanceId:      instanceID,
		Body:            `{{ metrics_sql "SELECT total_revenue FROM mv WHERE country = 'US'" }}`,
		UseFormatTokens: true,
	})
	require.NoError(t, err)
	require.Contains(t, formatRes.Body, "__RILL__FORMAT__")
	require.Contains(t, formatRes.Body, "mv")
	require.Contains(t, formatRes.Body, "total_revenue")
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
