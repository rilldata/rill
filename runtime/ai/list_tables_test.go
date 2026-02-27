package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestListTables(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/orders.yaml": `
type: model
sql: SELECT 1 AS order_id, 100 AS amount
`,
			"models/customers.yaml": `
type: model
sql: SELECT 1 AS customer_id, 'Alice' AS name
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	s := newSession(t, rt, instanceID)

	t.Run("list all tables", func(t *testing.T) {
		var res *ai.ListTablesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListTablesName, &res, &ai.ListTablesArgs{})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.GreaterOrEqual(t, len(res.Tables), 2)

		// Find our models in the results
		names := make(map[string]bool)
		for _, table := range res.Tables {
			names[table.Name] = true
		}
		require.True(t, names["orders"])
		require.True(t, names["customers"])
	})

	t.Run("with search pattern", func(t *testing.T) {
		var res *ai.ListTablesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListTablesName, &res, &ai.ListTablesArgs{
			SearchPattern: "orders",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.GreaterOrEqual(t, len(res.Tables), 1)

		found := false
		for _, table := range res.Tables {
			if table.Name == "orders" {
				found = true
				break
			}
		}
		require.True(t, found)
	})

	t.Run("explicit connector", func(t *testing.T) {
		var res *ai.ListTablesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListTablesName, &res, &ai.ListTablesArgs{
			Connector: "duckdb",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.GreaterOrEqual(t, len(res.Tables), 2)
	})
}
