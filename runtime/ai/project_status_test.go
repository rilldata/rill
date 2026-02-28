package ai_test

import (
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestProjectStatus(t *testing.T) {
	// Setup a basic project with various files
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Variables: map[string]string{
			"empty_var":     "",
			"non_empty_var": "hello_world",
		},
		Files: map[string]string{
			// Create some models with SQL content (self-contained, no external tables)
			"models/orders.yaml": `
type: model
sql: |
  SELECT
    1 AS order_id,
    101 AS customer_id,
    TIMESTAMP '2024-01-01' AS order_date,
    100.50 AS total_amount
  WHERE 1=1
`,
			"models/customers.yaml": `
type: model
sql: |
  SELECT
    101 AS customer_id,
    'John Doe' AS customer_name,
    'john@example.com' AS email,
    TIMESTAMP '2023-01-01' AS signup_date
`,
			// Create a metrics view
			"metrics/orders_metrics.yaml": `
type: metrics_view
model: orders
dimensions:
  - column: customer_id
measures:
  - name: total_revenue
    expression: SUM(total_amount)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	// Initialize test session
	s := newSession(t, rt, instanceID)

	t.Run("list all resources", func(t *testing.T) {
		var res *ai.ProjectStatusResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "duckdb", res.DefaultOLAPConnector)
		require.Len(t, res.Env, 1)
		require.Contains(t, res.Env, "non_empty_var")
		require.GreaterOrEqual(t, len(res.Resources), 3) // At least orders, customers, orders_metrics
		require.Empty(t, res.ParseErrors)

		// Verify resources have the expected fields
		for _, r := range res.Resources {
			require.NotEmpty(t, r["kind"])
			require.NotEmpty(t, r["name"])
			require.NotEmpty(t, r["reconcile_status"])
		}
	})

	t.Run("no logs by default", func(t *testing.T) {
		var res *ai.ProjectStatusResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Empty(t, res.Logs)
	})

	t.Run("tail logs", func(t *testing.T) {
		var res *ai.ProjectStatusResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{
			TailLogs: 100,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.Logs)

		// Verify log entries have expected fields
		for _, l := range res.Logs {
			require.NotEmpty(t, l["level"])
			require.NotEmpty(t, l["time"])
			require.NotEmpty(t, l["message"])
		}
	})

	t.Run("filter by kind", func(t *testing.T) {
		var res *ai.ProjectStatusResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{
			Kind: "rill.runtime.v1.Model",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Resources, 2) // orders and customers

		for _, r := range res.Resources {
			require.Equal(t, "rill.runtime.v1.Model", r["kind"])
		}
	})

	t.Run("filter by name", func(t *testing.T) {
		var res *ai.ProjectStatusResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{
			Name: "orders",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Resources, 1)
		require.Equal(t, "orders", res.Resources[0]["name"])
	})

	t.Run("filter by path", func(t *testing.T) {
		var res *ai.ProjectStatusResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{
			Path: "/models/orders.yaml",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Resources, 1)
		require.Equal(t, "orders", res.Resources[0]["name"])
		require.Equal(t, "/models/orders.yaml", res.Resources[0]["path"])
	})

	t.Run("filter where_error with no errors", func(t *testing.T) {
		var res *ai.ProjectStatusResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{
			WhereError: true,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Empty(t, res.Resources) // No errors in our test setup
	})

	t.Run("resources have refs", func(t *testing.T) {
		var res *ai.ProjectStatusResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{
			Kind: "rill.runtime.v1.MetricsView",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Resources, 1)

		// The metrics view should have refs to the orders model
		mv := res.Resources[0]
		require.Equal(t, "orders_metrics", mv["name"])
		refs, ok := mv["refs"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, refs)
	})
}

func TestProjectStatusWithParseErrors(t *testing.T) {
	// Setup a project with a parse error
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			// Valid model
			"models/valid.yaml": `
type: model
sql: SELECT 1 AS id
`,
			// Invalid YAML that will cause a parse error (tabs are not allowed in YAML)
			"models/invalid.yaml": "type: model\n\tsql: SELECT 1",
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 1, 1) // 2 resources (parser + valid), 1 reconcile error on ProjectParser, 1 parse error

	// Initialize test session
	s := newSession(t, rt, instanceID)

	t.Run("returns parse errors", func(t *testing.T) {
		var res *ai.ProjectStatusResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.ParseErrors)

		// Check parse error has expected fields
		pe := res.ParseErrors[0]
		require.NotEmpty(t, pe["path"])
		require.NotEmpty(t, pe["message"])
	})

	t.Run("filter parse errors by path", func(t *testing.T) {
		var res *ai.ProjectStatusResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{
			Path: "/models/valid.yaml",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Empty(t, res.ParseErrors) // No parse errors for the valid file
	})
}

func TestProjectStatusWaitUntilIdle(t *testing.T) {
	// Use ClickHouse so that materializing a model with sleep() takes real time,
	// eliminating the race where reconciliation finishes before we call WaitUntilIdle.
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"models/orders.yaml": `
type: model
sql: SELECT 1 AS order_id, 100.50 AS total_amount
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0) // parser + model

	s := newSession(t, rt, instanceID)

	// Add a materialized model with sleep(2) so reconciliation takes ~2 seconds.
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"/models/slow.yaml": `
type: model
materialize: true
sql: SELECT sleep(2), 1 AS id
`,
	})
	ctrl, err := rt.Controller(t.Context(), instanceID)
	require.NoError(t, err)
	err = ctrl.Reconcile(t.Context(), runtime.GlobalProjectParserName)
	require.NoError(t, err)

	// Sleep briefly to let reconciliation begin.
	time.Sleep(100 * time.Millisecond)

	// Call project_status with WaitUntilIdle; it should wait for reconciliation to finish.
	var res *ai.ProjectStatusResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.ProjectStatusName, &res, &ai.ProjectStatusArgs{
		WaitUntilIdle: true,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	// All resources should be idle after waiting.
	require.NotEmpty(t, res.Resources)
	for _, r := range res.Resources {
		require.Equal(t, "RECONCILE_STATUS_IDLE", r["reconcile_status"])
	}

	// The new materialized model should be present.
	var found bool
	for _, r := range res.Resources {
		if r["name"] == "slow" {
			found = true
			break
		}
	}
	require.True(t, found, "expected 'slow' model to be present after WaitUntilIdle")
}
