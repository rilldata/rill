package runtime_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestHealth(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
			"m1.sql":    `SELECT now() AS time, 'a' AS name, 1 AS value`,
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
timeseries: time
dimensions:
  - name: name
    column: name
measures:
  - name: count
    expression: count(*)
  - name: value
    expression: sum(value)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	health, err := rt.Health(t.Context(), true)
	require.NoError(t, err)

	require.Empty(t, health.HangingConn)
	require.Empty(t, health.Registry)

	require.Len(t, health.InstancesHealth, 1)
	ih, ok := health.InstancesHealth[instanceID]
	require.True(t, ok)

	require.Empty(t, ih.Controller)
	require.NotEmpty(t, ih.ControllerVersion)
	require.Empty(t, ih.OLAP)
	require.Empty(t, ih.Repo)
	require.Empty(t, ih.ParseErrCount)
	require.Empty(t, ih.ReconcileErrCount)

	require.Len(t, ih.MetricsViews, 1)
	mvErr, ok := ih.MetricsViews["mv1"]
	require.True(t, ok)
	require.Empty(t, mvErr.Err)
	require.NotEmpty(t, mvErr.Version)
}
