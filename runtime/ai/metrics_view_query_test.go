package ai_test

import (
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestMetricsViewQueryOpenURL(t *testing.T) {
	// Setup a basic project with a metrics view
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			// Create a simple model
			"test_data.sql": `SELECT 'US' AS country, 100 AS revenue, NOW() AS timestamp`,
			// Create a metrics view
			"test_metrics.yaml": `
type: metrics_view
version: 1
model: test_data
dimensions:
- column: country
measures:
- expression: SUM(revenue)
  name: total_revenue
explore:
  skip: true
`,
		},
		FrontendURL: "https://ui.rilldata.com/test-org/test-project",
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	// Initialize test session
	s := newSession(t, rt, instanceID)

	// Query the metrics view and check it returns a valid OpenURL
	var res *ai.QueryMetricsViewResult
	toolRes, err := s.CallTool(t.Context(), ai.RoleUser, ai.QueryMetricsViewName, &res, ai.QueryMetricsViewArgs{
		"metrics_view": "test_metrics",
		"dimensions":   []map[string]any{{"name": "country"}},
		"measures":     []map[string]any{{"name": "total_revenue"}},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Schema)
	require.NotEmpty(t, res.Data)
	require.Equal(t, res.OpenURL, fmt.Sprintf("https://ui.rilldata.com/test-org/test-project/-/ai/%s/message/%s/-/open", s.ID(), toolRes.Call.ID))
}

func TestMetricsViewQueryLimit(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"test_data.sql": `SELECT UNNEST(range(1, 11)) AS id`,
			"test_metrics.yaml": `
type: metrics_view
model: test_data
dimensions:
- column: id
measures:
- name: row_count
  expression: COUNT(*)
explore:
  skip: true
`,
		},
		Variables: map[string]string{
			"rill.ai.default_query_limit": "3",
			"rill.ai.max_query_limit":     "5",
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	s := newSession(t, rt, instanceID)

	tests := []struct {
		name        string
		limit       any // nil means omit limit from args
		wantRows    int
		wantWarning string
	}{
		{
			name:        "no limit applies default",
			limit:       nil,
			wantRows:    3,
			wantWarning: "The system truncated the result to 3 rows; to fetch more rows, explicitly set a limit (max allowed limit: 5)",
		},
		{
			name:        "user limit below max",
			limit:       2,
			wantRows:    2,
			wantWarning: "",
		},
		{
			name:        "user limit at max",
			limit:       5,
			wantRows:    5,
			wantWarning: "",
		},
		{
			name:        "user limit above max is capped",
			limit:       100,
			wantRows:    5,
			wantWarning: "The system truncated the result to 5 rows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := ai.QueryMetricsViewArgs{
				"metrics_view": "test_metrics",
				"dimensions":   []map[string]any{{"name": "id"}},
				"measures":     []map[string]any{{"name": "row_count"}},
				"sort":         []map[string]any{{"name": "id"}},
			}
			if tt.limit != nil {
				args["limit"] = tt.limit
			}

			var res *ai.QueryMetricsViewResult
			_, err := s.CallTool(t.Context(), ai.RoleUser, ai.QueryMetricsViewName, &res, args)
			require.NoError(t, err)
			require.Len(t, res.Data, tt.wantRows)
			require.Equal(t, tt.wantWarning, res.TruncationWarning)
		})
	}
}

func TestMetricsViewQueryNaN(t *testing.T) {
	// Setup a basic project with a metrics view that can produce NaN
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			// Create a simple model
			"test_data.sql": `SELECT 1`,
			// Create a metrics view
			"test_metrics.yaml": `
type: metrics_view
model: test_data
measures:
- name: inf
  expression: 1.0/0.0
- name: nan
  expression: ANY_VALUE('nan'::FLOAT)
explore:
  skip: true
cache:
  enabled: false
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	// Initialize test session
	s := newSession(t, rt, instanceID)

	// Query the metrics view and check it returns NaN values correctly
	var res *ai.QueryMetricsViewResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.QueryMetricsViewName, &res, ai.QueryMetricsViewArgs{
		"metrics_view": "test_metrics",
		"measures":     []map[string]any{{"name": "inf"}, {"name": "nan"}},
	})
	require.NoError(t, err)
	require.Len(t, res.Schema, 2)
	require.Equal(t, "inf", res.Schema[0].Name)
	require.Equal(t, "nan", res.Schema[1].Name)
	require.Len(t, res.Data, 1)
	require.Equal(t, nil, res.Data[0][0])
	require.Equal(t, nil, res.Data[0][1])
}
