package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestQuerySQL(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/test_data.yaml": `
type: model
sql: |
  SELECT 1 AS id, 'Alice' AS name, 100 AS value
  UNION ALL
  SELECT 2 AS id, 'Bob' AS name, 200 AS value
  UNION ALL
  SELECT 3 AS id, 'Charlie' AS name, 300 AS value
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	s := newSession(t, rt, instanceID)

	t.Run("basic query", func(t *testing.T) {
		var res *ai.QuerySQLResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.QuerySQLName, &res, &ai.QuerySQLArgs{
			SQL: "SELECT * FROM test_data ORDER BY id",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Data, 3)
		require.EqualValues(t, 1, res.Data[0]["id"])
		require.Equal(t, "Alice", res.Data[0]["name"])
		require.EqualValues(t, 100, res.Data[0]["value"])
	})

	t.Run("explicit connector", func(t *testing.T) {
		var res *ai.QuerySQLResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.QuerySQLName, &res, &ai.QuerySQLArgs{
			Connector: "duckdb",
			SQL:       "SELECT 42 AS answer",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Data, 1)
		require.EqualValues(t, 42, res.Data[0]["answer"])
	})

	t.Run("aggregation query", func(t *testing.T) {
		var res *ai.QuerySQLResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.QuerySQLName, &res, &ai.QuerySQLArgs{
			SQL: "SELECT COUNT(*) AS count, SUM(value) AS total FROM test_data",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Data, 1)
		require.EqualValues(t, 3, res.Data[0]["count"])
		require.EqualValues(t, 600, res.Data[0]["total"])
	})

	t.Run("missing sql", func(t *testing.T) {
		var res *ai.QuerySQLResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.QuerySQLName, &res, &ai.QuerySQLArgs{})
		require.Error(t, err)
	})

	t.Run("invalid sql", func(t *testing.T) {
		var res *ai.QuerySQLResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.QuerySQLName, &res, &ai.QuerySQLArgs{
			SQL: "SELECT * FROM nonexistent_table",
		})
		require.Error(t, err)
	})
}
