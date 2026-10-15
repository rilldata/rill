package reconcilers_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// An alert or report that names no explore grants its link access to the only explore on its metrics view,
// which is the one the open page maps it to. With several explores it grants none: picking one would be a guess.
func TestAlertAndReportExploreWithoutAnnotation(t *testing.T) {
	alert := `
type: alert
refresh:
  cron: "0 8 * * *"
data:
  metrics_sql: SELECT foo, x FROM mv1
notify:
  email:
    recipients:
      - user_1@example.com
`
	report := `
type: report
refresh:
  cron: "0 8 * * *"
data:
  metrics_sql: SELECT foo, x FROM mv1
export:
  format: csv
notify:
  email:
    recipients:
      - user_1@example.com
`
	explore := func(name string) string {
		return `
type: explore
metrics_view: mv1
display_name: ` + name + `
`
	}

	cases := []struct {
		name     string
		files    map[string]string
		explores map[string]bool
	}{
		{
			name:     "the only explore",
			files:    map[string]string{"e1.yaml": explore("e1")},
			explores: map[string]bool{"e1": true},
		},
		{
			name:     "no guess among several",
			files:    map[string]string{"e1.yaml": explore("e1"), "e2.yaml": explore("e2")},
			explores: map[string]bool{"e1": false, "e2": false},
		},
		{
			name: "the one it names",
			files: map[string]string{
				"e1.yaml": explore("e1"),
				"e2.yaml": explore("e2"),
				"a1.yaml": alert + "annotations:\n  explore: e2\n",
				"r1.yaml": report + "annotations:\n  explore: e2\n",
			},
			explores: map[string]bool{"e1": false, "e2": true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
				Files: map[string]string{"rill.yaml": ""},
			})
			files := map[string]string{
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
				"a1.yaml": alert,
				"r1.yaml": report,
			}
			for k, v := range tc.files {
				files[k] = v
			}
			testruntime.PutFiles(t, rt, id, files)
			testruntime.ReconcileParserAndWait(t, rt, id)

			for _, owner := range []*runtimev1.ResourceName{
				{Kind: runtime.ResourceKindAlert, Name: "a1"},
				{Kind: runtime.ResourceKindReport, Name: "r1"},
			} {
				claims := &runtime.SecurityClaims{
					AdditionalRules: []*runtimev1.SecurityRule{{
						Rule: &runtimev1.SecurityRule_TransitiveAccess{
							TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{Resource: owner},
						},
					}},
				}
				for name, want := range tc.explores {
					e := testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, name)
					sec, err := rt.ResolveSecurity(t.Context(), id, claims, e)
					require.NoError(t, err)
					require.Equal(t, want, sec.CanAccess(), "%s %s: access to explore %s", owner.Kind, owner.Name, name)
				}
			}
		})
	}
}
