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
	require.Equal(t, "foo", comp1Props["foo"])
	require.Equal(t, "bar", comp1Props["bar"])

	// Check referenced metrics views
	require.Len(t, res.ReferencedMetricsViews, 1)
	require.Equal(t, "m1", res.ReferencedMetricsViews["mv1"].GetMetricsView().State.ValidSpec.Model)
	require.NotContains(t, res.ReferencedMetricsViews, "mv2", "mv2 is only in metrics_sql and should not be extracted")
}

func TestResolveCanvas_MetricsSQLNotExtracted(t *testing.T) {
	// This test validates that metrics views referenced only in metrics_sql fields
	// are NOT extracted by ResolveCanvas. Components should call ResolveTemplatedString
	// RPC directly to resolve metrics_sql queries at render time.
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    `SELECT 'US' AS country, 100 AS revenue`,
			"m2.sql":    `SELECT 'CA' AS country, 200 AS revenue`,
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: country
measures:
- expression: SUM(revenue)
  name: total_revenue
`,
			"mv2.yaml": `
type: metrics_view
version: 1
model: m2
dimensions:
- column: country
measures:
- expression: SUM(revenue)
  name: total_revenue
`,
			"c_metrics_sql_only.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_view: mv1
  - kpi:
      metrics_sql: "SELECT total_revenue FROM mv2"
  - kpi:
      metrics_sql: "SELECT total_revenue FROM mv2 WHERE country = 'CA'"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 9, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c_metrics_sql_only",
	})
	require.NoError(t, err)

	// Should only include mv1 (referenced via metrics_view field)
	// mv2 is only referenced via metrics_sql and should NOT be extracted
	require.Len(t, res.ReferencedMetricsViews, 1)
	require.Contains(t, res.ReferencedMetricsViews, "mv1")
	require.NotContains(t, res.ReferencedMetricsViews, "mv2", "mv2 is only referenced via metrics_sql and should not be in ReferencedMetricsViews")

	// All components should still be resolved
	require.Len(t, res.ResolvedComponents, 3)

	// Verify the metrics_sql fields are preserved in the resolved components
	comp1Props := res.ResolvedComponents["c_metrics_sql_only--component-0-1"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "SELECT total_revenue FROM mv2", comp1Props["metrics_sql"])
	comp2Props := res.ResolvedComponents["c_metrics_sql_only--component-0-2"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "SELECT total_revenue FROM mv2 WHERE country = 'CA'", comp2Props["metrics_sql"])
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

	// Should still resolve all components
	// Note: Invalid SQL in metrics_sql fields doesn't prevent component resolution
	// because metrics_sql is not parsed by ResolveCanvas
	require.Len(t, res.ResolvedComponents, 3, "All components should be resolved even with invalid SQL")
	require.Len(t, res.ReferencedMetricsViews, 1, "Should only include mv1 (from metrics_view field, not from metrics_sql)")
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

	// Note: Metrics views in metrics_sql are NOT extracted by ResolveCanvas.
	// The templates are resolved, but the metrics view extraction from SQL is not performed.
	// Components should call ResolveTemplatedString RPC to execute these queries.
	require.Len(t, res.ReferencedMetricsViews, 0, "No metrics views should be extracted from metrics_sql fields")

	// Check that templates were resolved in the components
	require.Len(t, res.ResolvedComponents, 2)
	comp0Props := res.ResolvedComponents["c_templated--component-0-0"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "SELECT country FROM mv1", comp0Props["metrics_sql"])
	comp1Props := res.ResolvedComponents["c_templated--component-0-1"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "SELECT country FROM mv2", comp1Props["metrics_sql"])
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

	// Only the first two components use metrics_view: mv1
	// The third uses metrics_sql which is NOT extracted
	require.Len(t, res.ReferencedMetricsViews, 1, "Only mv1 from metrics_view fields should be extracted")
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

	// Note: metrics_sql fields are NOT parsed to extract metrics views.
	// Components should call ResolveTemplatedString RPC to execute these queries.
	require.Len(t, res.ReferencedMetricsViews, 0, "No metrics views from metrics_sql should be extracted")
	require.Len(t, res.ResolvedComponents, 3)
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

	require.Len(t, res.ReferencedMetricsViews, 0, "No metrics views from metrics_sql should be extracted")
	require.Len(t, res.ResolvedComponents, 1)

	// Check that the custom_chart component has the correct properties
	comp0Props := res.ResolvedComponents["c_custom_chart--component-0-0"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	require.Equal(t, "hsl(246, 66%, 50%)", comp0Props["color"])
	require.Contains(t, comp0Props, "metrics_sql")
}

func TestResolveCanvasWithMarkdown(t *testing.T) {
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
- expression: SUM(revenue)
  name: total_revenue
`,
			"c_markdown.yaml": `
type: canvas
rows:
- items:
  - markdown:
      content: |
        # Sales Report
        Welcome {{ .args.user_name }}!
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c_markdown",
		Args: must(structpb.NewStruct(map[string]any{
			"user_name": "Alice",
		})),
	})
	require.NoError(t, err)

	// Markdown components don't reference metrics views via metrics_view field
	require.Len(t, res.ReferencedMetricsViews, 0)
	require.Len(t, res.ResolvedComponents, 1)

	// Verify the markdown content had templates resolved (args/env variables)
	comp0Props := res.ResolvedComponents["c_markdown--component-0-0"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	content, ok := comp0Props["content"].(string)
	require.True(t, ok, "content should be a string")
	require.Contains(t, content, "Welcome Alice!")
	require.NotContains(t, content, "{{ .args.user_name }}", "template should be resolved")
}

func TestResolveCanvasWithMarkdownAndMixedReferences(t *testing.T) {
	// Note: This test shows that markdown components with metrics_sql templates
	// are NOT resolved by ResolveCanvas. They would need to be resolved separately
	// via ResolveTemplatedString RPC by the client at render time.
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
- expression: SUM(revenue)
  name: total_revenue
`,
			"c_mixed.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_view: mv1
  - markdown:
      content: |
        # Simple markdown without metrics_sql
        This component references mv1 indirectly through the canvas.
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 6, 0, 0)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.ResolveCanvas(testCtx(), &runtimev1.ResolveCanvasRequest{
		InstanceId: instanceID,
		Canvas:     "c_mixed",
	})
	require.NoError(t, err)

	// Should only extract mv1 from the kpi component's metrics_view field
	// Markdown component doesn't directly reference any metrics views
	require.Len(t, res.ReferencedMetricsViews, 1)
	require.Contains(t, res.ReferencedMetricsViews, "mv1")
	require.Len(t, res.ResolvedComponents, 2)

	// Verify markdown content is present
	markdownProps := res.ResolvedComponents["c_mixed--component-0-1"].GetComponent().State.ValidSpec.RendererProperties.AsMap()
	content, ok := markdownProps["content"].(string)
	require.True(t, ok)
	require.Contains(t, content, "Simple markdown")
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

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
