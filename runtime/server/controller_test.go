package server_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateTriggerAll(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	// Create a table directly in the OLAP connector for testing metrics views without any refs.
	createTableAsSelect(t, rt, instanceID, "duckdb", "foo", "SELECT 'US' AS country")

	// Create test resources
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		// Model
		"m1.sql": `
SELECT 'US' AS country
`,
		// Metrics view with reference to the model
		"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: country
measures:
- expression: COUNT(*)
`,
		// Explore on mv1
		"e1.yaml": `
type: explore
metrics_view: mv1
`,
		// Metrics view on external table without any refs
		"mv2.yaml": `
type: metrics_view
version: 1
table: foo
dimensions:
- column: country
measures:
- expression: COUNT(*)
`,
		// Explore on mv2
		"e2.yaml": `
type: explore
metrics_view: mv2
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 6, 0, 0)

	// Verify that mv2 has no refs
	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "mv2"}, false)
	require.NoError(t, err)
	require.Len(t, r.Meta.Refs, 0)

	// Capture version numbers for all resources
	rs, err := ctrl.List(context.Background(), "", "", false)
	require.NoError(t, err)
	versions := make(map[string]int)
	for _, r := range rs {
		versions[r.Meta.Name.Name] = int(r.Meta.StateVersion)
	}

	// Create test server
	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Create all trigger
	_, err = server.CreateTrigger(testCtx(), &runtimev1.CreateTriggerRequest{
		InstanceId: instanceID,
		All:        true,
	})
	require.NoError(t, err)

	// Await all are idle
	err = ctrl.WaitUntilIdle(context.Background(), false)
	require.NoError(t, err)
	testruntime.RequireReconcileState(t, rt, instanceID, 6, 0, 0)

	// Verify that all were refreshed
	rs, err = ctrl.List(context.Background(), "", "", false)
	require.NoError(t, err)
	for _, r := range rs {
		oldVersion, ok := versions[r.Meta.Name.Name]
		require.True(t, ok)
		require.Greater(t, int(r.Meta.StateVersion), oldVersion, "resource %s was not refreshed", r.Meta.Name.Name)
	}
}

func TestModelChangeModes(t *testing.T) {
	// --- RESET MODE ---
	t.Run("reset", func(t *testing.T) {
		rt, instanceID := testruntime.NewInstance(t)
		createTableAsSelect(t, rt, instanceID, "duckdb", "foo", "SELECT 'US' AS country")

		// Create test resources m1, mv1
		testruntime.PutFiles(t, rt, instanceID, map[string]string{
			"m1.yaml": `
version: 1
type: model
sql: SELECT 'US' AS country
change_mode: reset
incremental: false
`,
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: country
measures:
- expression: COUNT(*)
`,
		})
		testruntime.ReconcileParserAndWait(t, rt, instanceID)
		testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0) // m1, mv1, project_settings

		ctrl, err := rt.Controller(context.Background(), instanceID)
		require.NoError(t, err)
		m1, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "m1"}, false)
		require.NoError(t, err)
		initialVersionM1 := int(m1.Meta.StateVersion)

		// Update m1.yaml with different SQL
		testruntime.PutFiles(t, rt, instanceID, map[string]string{
			"m1.yaml": `
version: 1
type: model
sql: SELECT 'CA' AS country
change_mode: reset
incremental: false
`,
		})
		testruntime.ReconcileParserAndWait(t, rt, instanceID)
		testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

		// Verify that the model m1 was refreshed
		m1, err = ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "m1"}, false)
		require.NoError(t, err)
		require.Greater(t, int(m1.Meta.StateVersion), initialVersionM1, "m1 was not refreshed in reset mode")
	})

	// --- MANUAL MODE ---
	t.Run("manual", func(t *testing.T) {
		rt, instanceID := testruntime.NewInstance(t)
		createTableAsSelect(t, rt, instanceID, "duckdb", "foo", "SELECT 'US' AS country")

		// Create test resources m2, mv2
		testruntime.PutFiles(t, rt, instanceID, map[string]string{
			"m2.yaml": `
version: 1
type: model
sql: SELECT 'DE' AS country
change_mode: manual
incremental: false
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
		})
		testruntime.ReconcileParserAndWait(t, rt, instanceID)
		testruntime.RequireReconcileState(t, rt, instanceID, 3, 2, 0) // m2, mv2, project_settings

		ctrl, err := rt.Controller(context.Background(), instanceID)
		require.NoError(t, err)
		m2, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "m2"}, false)
		require.NoError(t, err)
		initialVersionM2 := int(m2.Meta.StateVersion)

		// Update m2.yaml with different SQL, still in manual mode
		testruntime.PutFiles(t, rt, instanceID, map[string]string{
			"m2.yaml": `
version: 1
type: model
sql: SELECT 'FR' AS country
change_mode: manual
incremental: false
`,
		})
		testruntime.ReconcileParserAndWait(t, rt, instanceID)
		// Expects an error because changes in manual mode are not automatically applied
		testruntime.RequireReconcileState(t, rt, instanceID, 3, 2, 0) // m2, mv2, project_settings

		// Change the mode by removing change_mode (implicitly defaults, e.g., to reset)
		testruntime.PutFiles(t, rt, instanceID, map[string]string{
			"m2.yaml": `
version: 1
type: model
sql: SELECT 'FR' AS country # SQL is the new one
incremental: false
`,
		})
		testruntime.ReconcileParserAndWait(t, rt, instanceID)
		// Now it should reconcile without errors, and m2 should be updated
		testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0) // m2, mv2, project_settings

		// Verify that the model m2 was refreshed after changing mode
		m2, err = ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "m2"}, false)
		require.NoError(t, err)
		require.Greater(t, int(m2.Meta.StateVersion), initialVersionM2, "m2 was not refreshed after changing mode from manual")
	})

	// --- PATCH MODE ---
	t.Run("patch", func(t *testing.T) {
		rt, instanceID := testruntime.NewInstance(t)
		createTableAsSelect(t, rt, instanceID, "duckdb", "foo", "SELECT 'US' AS country")

		// Create test resources m3, mv3
		testruntime.PutFiles(t, rt, instanceID, map[string]string{
			"m3.yaml": `
version: 1
type: model
sql: SELECT 'GB' AS country
change_mode: patch
incremental: true
`,
			"mv3.yaml": `
type: metrics_view
version: 1
model: m3
dimensions:
- column: country
measures:
- expression: COUNT(*)
`,
		})
		testruntime.ReconcileParserAndWait(t, rt, instanceID)
		testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0) // m3, mv3, project_settings

		ctrl, err := rt.Controller(context.Background(), instanceID)
		require.NoError(t, err)
		m3, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "m3"}, false)
		require.NoError(t, err)
		initialVersionM3 := int(m3.Meta.StateVersion)

		// Update m3.yaml with different SQL
		testruntime.PutFiles(t, rt, instanceID, map[string]string{
			"m3.yaml": `
version: 1
type: model
sql: SELECT 'IT' AS country
change_mode: patch
incremental: true
`,
		})
		testruntime.ReconcileParserAndWait(t, rt, instanceID)
		testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0) // m3, mv3, project_settings

		// Verify that the model m3 was refreshed (patched)
		m3, err = ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "m3"}, false)
		require.NoError(t, err)
		require.Greater(t, int(m3.Meta.StateVersion), initialVersionM3, "m3 was not refreshed in patch mode")
	})
}

// createTableAsSelect is a test utility for creating a table directly in an OLAP connector.
// It invokes a model executor directly without actually creating a model resource.
// This is useful for testing resources against pre-existing/external tables.
func createTableAsSelect(t *testing.T, rt *runtime.Runtime, instanceID, connector, name, sql string) {
	h, release, err := rt.AcquireHandle(context.Background(), instanceID, connector)
	require.NoError(t, err)
	defer release()
	opts := &drivers.ModelExecutorOptions{
		Env:                         &drivers.ModelEnv{},
		ModelName:                   name,
		InputHandle:                 h,
		InputConnector:              connector,
		PreliminaryInputProperties:  map[string]any{"sql": sql},
		OutputHandle:                h,
		OutputConnector:             connector,
		PreliminaryOutputProperties: map[string]any{"table": name},
	}
	me, err := h.AsModelExecutor(instanceID, opts)
	require.NoError(t, err)
	_, err = me.Execute(context.Background(), &drivers.ModelExecuteOptions{
		ModelExecutorOptions: opts,
		InputProperties:      opts.PreliminaryInputProperties,
		OutputProperties:     opts.PreliminaryOutputProperties,
	})
	require.NoError(t, err)
}
