package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestShowTable(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/test_table.yaml": `
type: model
sql: |
  SELECT
    1 AS id,
    'Alice' AS name,
    100.50 AS amount,
    TIMESTAMP '2024-01-01' AS created_at
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	s := newSession(t, rt, instanceID)

	t.Run("show table schema", func(t *testing.T) {
		var res *ai.ShowTableResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ShowTableName, &res, &ai.ShowTableArgs{
			Table: "test_table",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "test_table", res.Name)
		require.GreaterOrEqual(t, len(res.Columns), 4)

		// Verify column names
		colNames := make(map[string]bool)
		for _, col := range res.Columns {
			colNames[col.Name] = true
			require.NotEmpty(t, col.Type)
		}
		require.True(t, colNames["id"])
		require.True(t, colNames["name"])
		require.True(t, colNames["amount"])
		require.True(t, colNames["created_at"])
	})

	t.Run("with explicit connector", func(t *testing.T) {
		var res *ai.ShowTableResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ShowTableName, &res, &ai.ShowTableArgs{
			Connector: "duckdb",
			Table:     "test_table",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "test_table", res.Name)
	})

	t.Run("table not found", func(t *testing.T) {
		var res *ai.ShowTableResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ShowTableName, &res, &ai.ShowTableArgs{
			Table: "nonexistent_table",
		})
		require.Error(t, err)
	})

	t.Run("missing table name", func(t *testing.T) {
		var res *ai.ShowTableResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ShowTableName, &res, &ai.ShowTableArgs{})
		require.Error(t, err)
	})
}
