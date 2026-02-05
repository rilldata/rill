package executor_test

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestResolveQueryAttributesTemplate(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"m1.sql": `
SELECT 1
`,
			"mv1.yaml": `
type: metrics_view
model: m1
dimensions:
  - name: dim1
    expression: getSettingOrDefault('custom_attr', 'x')
measures:
  - expression: count(*)
query_attributes:
  custom_attr: '{{ .user.org_id }}_{{ .user.tenant_id }}'
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	res, err := rt.Resolve(t.Context(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics",
		ResolverProperties: map[string]any{
			"metrics_view": "mv1",
			"dimensions":   []map[string]any{{"name": "dim1"}},
		},
		Claims: &runtime.SecurityClaims{
			UserAttributes: map[string]any{
				"org_id":    "org123",
				"tenant_id": "tenant456",
			},
			Permissions: runtime.AllPermissions,
		},
	})
	require.NoError(t, err)
	defer res.Close()
	row, err := res.Next()
	require.NoError(t, err)
	require.Equal(t, "org123_tenant456", row["dim1"])
}
