package reconcilers_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// Opening a report through a recipient token analyzes the resolver to expand transitive access.
// Analysis must not initialize a resolver and re-enter security resolution.
func TestReportResolveTransitiveAccessNoRecursion(t *testing.T) {
	for _, tc := range []struct {
		name       string
		data       string
		wantD      bool
		wantFilter bool
	}{
		{name: "api", data: "  api: open_api"},
		{name: "metrics", data: "  metrics:\n    metrics_view: mv1\n    measures:\n      - name: count"},
		{name: "metrics_sql", data: "  metrics_sql: SELECT count FROM mv1 WHERE d = 'x'", wantD: true, wantFilter: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
				Files: map[string]string{
					"rill.yaml":    "",
					"models/m.sql": "SELECT 1 AS a, 'x' AS d",
					"metrics/mv1.yaml": `
type: metrics_view
model: m
dimensions:
  - column: d
measures:
  - name: count
    expression: count(*)
`,
					"metrics/mv2.yaml": `
type: metrics_view
model: m
dimensions:
  - column: d
measures:
  - name: count
    expression: count(*)
`,
					"apis/open_api.yaml": `
type: api
metrics_sql: SELECT count FROM mv1
`,
					"reports/r1.yaml": `
type: report
refresh:
  cron: 0 8 * * *
export:
  format: csv
data:
` + tc.data + `
notify:
  email:
    recipients:
      - somebody@example.com
`,
				},
			})
			testruntime.RequireReconcileState(t, rt, id, 8, 0, 0)

			claims := &runtime.SecurityClaims{
				UserAttributes: map[string]any{"email": "somebody@example.com"},
				AdditionalRules: []*runtimev1.SecurityRule{{
					Rule: &runtimev1.SecurityRule_TransitiveAccess{
						TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
							Resource: &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: "r1"},
						},
					},
				}},
			}
			ctx := t.Context()

			r1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindReport, "r1")
			sec, err := rt.ResolveSecurity(ctx, id, claims, r1)
			require.NoError(t, err)
			require.True(t, sec.CanAccess())

			mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
			sec, err = rt.ResolveSecurity(ctx, id, claims, mv1)
			require.NoError(t, err)
			require.True(t, sec.CanAccess())
			require.True(t, sec.CanAccessField("count"))
			require.Equal(t, tc.wantD, sec.CanAccessField("d"))
			require.Equal(t, tc.wantFilter, sec.QueryFilter() != nil)

			mv2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
			sec, err = rt.ResolveSecurity(ctx, id, claims, mv2)
			require.NoError(t, err)
			require.False(t, sec.CanAccess())

			// Real query execution still applies the inferred policy with the original claims.
			result, _, err := rt.Resolve(ctx, &runtime.ResolveOptions{
				InstanceID: id,
				Resolver:   "metrics",
				ResolverProperties: map[string]any{
					"metrics_view": "mv1",
					"measures":     []any{map[string]any{"name": "count"}},
				},
				Claims: claims,
			})
			require.NoError(t, err)
			row, err := result.Next()
			require.NoError(t, err)
			require.EqualValues(t, 1, row["count"])
			require.NoError(t, result.Close())

			denied := *claims
			denied.AdditionalRules = append([]*runtimev1.SecurityRule{}, claims.AdditionalRules...)
			denied.AdditionalRules = append(denied.AdditionalRules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
					Allow: false, ConditionResources: []*runtimev1.ResourceName{mv1.Meta.Name},
				}},
			})
			sec, err = rt.ResolveSecurity(ctx, id, &denied, mv1)
			require.NoError(t, err)
			require.False(t, sec.CanAccess(), "transitive analysis must not override an explicit denial")
		})
	}
}
