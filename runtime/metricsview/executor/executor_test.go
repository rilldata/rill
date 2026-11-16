package executor_test

import (
	"os"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
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

	res, _, err := rt.Resolve(t.Context(), &runtime.ResolveOptions{
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

	// Exports run through a model executor instead of a direct OLAP query; check that they also get the query attributes.
	mv := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindMetricsView, "mv1").GetMetricsView().State.ValidSpec
	e, err := executor.New(t.Context(), rt, instanceID, mv, false, runtime.ResolvedSecurityOpen, 0, map[string]any{
		"org_id":    "org123",
		"tenant_id": "tenant456",
	})
	require.NoError(t, err)
	defer e.Close()
	path, err := e.Export(t.Context(), &metricsview.Query{
		MetricsView: "mv1",
		Dimensions:  []metricsview.Dimension{{Name: "dim1"}},
	}, nil, drivers.FileFormatCSV, nil)
	require.NoError(t, err)
	defer os.Remove(path)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), "org123_tenant456")
}
