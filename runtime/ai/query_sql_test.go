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

func TestQuerySQLLimit(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/test_data.yaml": `
type: model
sql: |
  SELECT UNNEST(range(1, 11)) AS id
`,
		},
		Variables: map[string]string{
			"rill.ai.max_query_limit": "5",
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	s := newSession(t, rt, instanceID)

	tests := []struct {
		name        string
		sql         string
		wantRows    int
		wantWarning string
	}{
		{
			name:        "result under limit",
			sql:         "SELECT * FROM test_data WHERE id <= 2 ORDER BY id",
			wantRows:    2,
			wantWarning: "",
		},
		{
			name:        "result truncated at limit",
			sql:         "SELECT * FROM test_data ORDER BY id",
			wantRows:    5,
			wantWarning: "The system truncated the result to 5 rows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res *ai.QuerySQLResult
			_, err := s.CallTool(t.Context(), ai.RoleUser, ai.QuerySQLName, &res, &ai.QuerySQLArgs{
				SQL: tt.sql,
			})
			require.NoError(t, err)
			require.Len(t, res.Data, tt.wantRows)
			require.Equal(t, tt.wantWarning, res.TruncationWarning)
		})
	}
}
