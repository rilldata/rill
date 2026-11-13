package server_test

import (
	"context"
	"fmt"
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
)

// TestEmbedOpenReportAccess tests what an embed that opens a report can access when it is limited to a single dashboard.
// The embed user writes the report's query, and a report's transitive access is derived from its query,
// so the admin service limits the embed to the dashboard that the report is limited to, and not to the report itself (see GetIFrame).
// The creator can still read the report itself, which is all the open page needs from it.
func TestEmbedOpenReportAccess(t *testing.T) {
	metricsView := `
version: 1
type: metrics_view
model: bar
timeseries: __time
dimensions:
- column: country
measures:
- name: total
  expression: count(*)
`
	// Reports as the admin service creates them from an embed: they open the explore e1 and are owned by an embed user.
	report := func(ownerID, metricsView string) string {
		return fmt.Sprintf(`
type: report
display_name: Report
query:
  name: MetricsViewAggregation
  args:
    metrics_view: %s
    dimensions:
    - name: country
    measures:
    - name: total
export:
  format: csv
  attach: true
notify:
  email:
    recipients:
    - embed@example.com
annotations:
  admin_managed: "true"
  admin_embed_owner_id: %s
  web_open_mode: creator
  explore: e1
`, metricsView, ownerID)
	}

	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml":        "",
			"models/bar.sql":   `SELECT '2024-01-01T00:00:00Z'::TIMESTAMP AS __time, 'Denmark' AS country`,
			"metrics/mv1.yaml": metricsView,
			"metrics/mv2.yaml": metricsView,
			"explores/e1.yaml": "type: explore\nmetrics_view: mv1\n",
			"explores/e2.yaml": "type: explore\nmetrics_view: mv2\n",
			// Queries the metrics view of the explore it opens
			"reports/r1.yaml": report("ext_owner", "mv1"),
			// Queries a metrics view outside the explore it opens
			"reports/r2.yaml": report("ext_owner", "mv2"),
			// Created by another embed user
			"reports/r3.yaml": report("ext_other", "mv1"),
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 9, 0, 0)

	srv, err := server.NewServer(t.Context(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// embedCtx returns a context with the claims of an embed's runtime JWT that is limited to the given resources.
	// It mirrors the JWTs that the admin service issues for iframes (see issueRuntimeToken).
	embedCtx := func(userID string, resources ...*runtimev1.ResourceName) context.Context {
		return auth.WithClaims(t.Context(), &runtime.SecurityClaims{
			UserID:          userID,
			UserAttributes:  map[string]any{"embed": true},
			Permissions:     []runtime.Permission{runtime.ReadAPI, runtime.ReadMetrics, runtime.ReadObjects, runtime.UseAI},
			AdditionalRules: runtime.SecurityRulesForResources(true, resources),
		})
	}
	e1 := &runtimev1.ResourceName{Kind: runtime.ResourceKindExplore, Name: "e1"}
	r2 := &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r2"}

	// The calls below are the runtime calls of the embed's open page:
	// it gets the report, gets its explore, and resolves the report's time range and filters with queries on the explore's metrics view.
	getResource := func(ctx context.Context, kind, name string) error {
		_, err := srv.GetResource(ctx, &runtimev1.GetResourceRequest{InstanceId: instanceID, Name: &runtimev1.ResourceName{Kind: kind, Name: name}})
		return err
	}
	getExplore := func(ctx context.Context, name string) error {
		_, err := srv.GetExplore(ctx, &runtimev1.GetExploreRequest{InstanceId: instanceID, Name: name})
		return err
	}
	timeRange := func(ctx context.Context, metricsView string) error {
		_, err := srv.MetricsViewTimeRange(ctx, &runtimev1.MetricsViewTimeRangeRequest{InstanceId: instanceID, MetricsViewName: metricsView})
		return err
	}
	timeRanges := func(ctx context.Context, metricsView string) error {
		_, err := srv.MetricsViewTimeRanges(ctx, &runtimev1.MetricsViewTimeRangesRequest{InstanceId: instanceID, MetricsViewName: metricsView, Expressions: []string{"P7D"}})
		return err
	}
	aggregate := func(ctx context.Context, metricsView string) (int, error) {
		resp, err := srv.MetricsViewAggregation(ctx, &runtimev1.MetricsViewAggregationRequest{
			InstanceId:  instanceID,
			MetricsView: metricsView,
			Dimensions:  []*runtimev1.MetricsViewAggregationDimension{{Name: "country"}},
			Measures:    []*runtimev1.MetricsViewAggregationMeasure{{Name: "total"}},
		})
		if err != nil {
			return 0, err
		}
		return len(resp.Data), nil
	}
	// The handlers are called directly, so errors from the runtime are not mapped to gRPC codes by the server's middleware.
	// The forbidden errors of the server, the runtime and the query packages share the same message.
	requireForbidden := func(t *testing.T, err error) {
		t.Helper()
		require.ErrorContains(t, err, "action not allowed")
	}

	t.Run("claims with the report reach the metrics view of its query", func(t *testing.T) {
		// This is why the embed's claims must not include the report: r2 opens e1, but its query extends the claims to mv2.
		ctx := embedCtx("ext_owner", r2, e1)
		rows, err := aggregate(ctx, "mv2")
		require.NoError(t, err)
		require.Equal(t, 1, rows)
	})

	t.Run("claims limited to the explore can open the creator's reports", func(t *testing.T) {
		ctx := embedCtx("ext_owner", e1)
		for _, name := range []string{"r1", "r2"} {
			require.NoError(t, getResource(ctx, runtime.ResourceKindReport, name))
		}
		require.NoError(t, getExplore(ctx, "e1"))
		require.NoError(t, timeRange(ctx, "mv1"))
		require.NoError(t, timeRanges(ctx, "mv1"))
		rows, err := aggregate(ctx, "mv1")
		require.NoError(t, err)
		require.Equal(t, 1, rows)
	})

	t.Run("claims limited to the explore can't reach the metrics view of the report's query", func(t *testing.T) {
		ctx := embedCtx("ext_owner", e1)
		_, err := aggregate(ctx, "mv2")
		requireForbidden(t, err)
		requireForbidden(t, timeRange(ctx, "mv2"))
		requireForbidden(t, timeRanges(ctx, "mv2"))
		requireForbidden(t, getResource(ctx, runtime.ResourceKindMetricsView, "mv2"))
		requireForbidden(t, getExplore(ctx, "e2"))
	})

	t.Run("only the creator can read a report", func(t *testing.T) {
		requireForbidden(t, getResource(embedCtx("ext_owner", e1), runtime.ResourceKindReport, "r3"))
		requireForbidden(t, getResource(embedCtx("ext_other", e1), runtime.ResourceKindReport, "r1"))
		requireForbidden(t, getResource(embedCtx("", e1), runtime.ResourceKindReport, "r1"))
	})
}
