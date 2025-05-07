package server_test

import (
	"context"
	"fmt"
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
	testCases := []struct {
		name            string
		changeMode      string
		initialSQL      string
		updatedSQL      string
		expectedCountry string
	}{
		{
			name:            "reset mode",
			changeMode:      "reset",
			initialSQL:      "SELECT 'US' AS country",
			updatedSQL:      "SELECT 'UK' AS country",
			expectedCountry: "UK",
		},
		{
			name:            "manual mode",
			changeMode:      "manual",
			initialSQL:      "SELECT 'US' AS country",
			updatedSQL:      "SELECT 'UK' AS country",
			expectedCountry: "US",
		},
		{
			name:            "patch mode",
			changeMode:      "patch",
			initialSQL:      "SELECT 'US' AS country, 1 as id",
			updatedSQL:      "SELECT 'US-updated' AS country, 1 as id",
			expectedCountry: "US",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new runtime instance for each test case
			rt, instanceID := testruntime.NewInstance(t)

			// Create initial model with specified change mode
			testruntime.PutFiles(t, rt, instanceID, map[string]string{
				"mv1.yaml": `
type: metrics_view
version: 1
model: my_model
dimensions:
- column: country
measures:
- expression: COUNT(*)
`,
				"model.yaml": fmt.Sprintf(`
type: model
name: my_model
change_mode: %s
sql: |
  %s
`, tc.changeMode, tc.initialSQL),
			})
			// Reconcile to create the initial model
			testruntime.ReconcileParserAndWait(t, rt, instanceID)
			testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

			// Get controller
			ctrl, err := rt.Controller(context.Background(), instanceID)
			require.NoError(t, err)

			// Wait until model is idle
			err = ctrl.WaitUntilIdle(context.Background(), false)
			require.NoError(t, err)

			// Verify initial data was loaded
			h, release, err := rt.AcquireHandle(context.Background(), instanceID, "duckdb")
			require.NoError(t, err)
			defer release()
			q := h.Driver()
			require.NotNil(t, q)

			// Capture initial state version
			r, err := ctrl.Get(testCtx(), &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "my_model"}, false)
			require.NoError(t, err)
			initialStateVersion := int64(r.Meta.StateVersion)

			// Update the model with new SQL
			testruntime.PutFiles(t, rt, instanceID, map[string]string{
				"model.yaml": fmt.Sprintf(`
type: model
name: my_model
change_mode: %s
sql: |
  %s
`, tc.changeMode, tc.updatedSQL),
			})
			// Reconcile to pick up changes
			testruntime.ReconcileParserAndWait(t, rt, instanceID)

			// Wait until model is idle again
			err = ctrl.WaitUntilIdle(testCtx(), false)
			require.NoError(t, err)

			// Check behavior based on change mode
			r, err = ctrl.Get(testCtx(), &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "my_model"}, false)
			require.NoError(t, err)

			olap, release, err := rt.OLAP(testCtx(), instanceID, "duckdb")
			require.NoError(t, err)
			defer release()

			switch tc.changeMode {
			case "reset":
				require.Greater(t, int64(r.Meta.StateVersion), initialStateVersion)
				verifyModelData(t, olap, tc.expectedCountry)

			case "manual":
				require.Equal(t, initialStateVersion, int64(r.Meta.StateVersion))
				verifyModelData(t, olap, tc.expectedCountry)

			case "patch":
				require.Equal(t, initialStateVersion, int64(r.Meta.StateVersion))
				verifyModelData(t, olap, tc.expectedCountry)

			default:
				t.Fatalf("unexpected change mode: %s", tc.changeMode)
			}
		})
	}
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

	me, ok := h.AsModelExecutor(instanceID, opts)
	require.True(t, ok)
	_, err = me.Execute(context.Background(), &drivers.ModelExecuteOptions{
		ModelExecutorOptions: opts,
		InputProperties:      opts.PreliminaryInputProperties,
		OutputProperties:     opts.PreliminaryOutputProperties,
	})
	require.NoError(t, err)
}

// Add this helper function after the TestModelChangeModes function
func verifyModelData(t *testing.T, olap drivers.OLAPStore, expectedCountry string) {
	t.Helper() // Marks this as a helper function for better test failure reporting
	res, err := olap.Query(testCtx(), &drivers.Statement{Query: "SELECT country FROM my_model"})
	require.NoError(t, err)
	defer res.Close()

	var country string
	for res.Next() {
		err = res.Scan(&country)
		require.NoError(t, err)
		require.Equal(t, expectedCountry, country, "country value should match expected value")
	}
	err = res.Err()
	require.NoError(t, err)
}
