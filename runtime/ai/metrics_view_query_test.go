package ai_test

import (
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
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.QueryMetricsViewName, &res, ai.QueryMetricsViewArgs{
		"metrics_view": "test_metrics",
		"dimensions":   []map[string]any{{"name": "country"}},
		"measures":     []map[string]any{{"name": "total_revenue"}},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Data)
	require.Contains(t, res.OpenURL, "https://ui.rilldata.com/test-org/test-project")
	require.Contains(t, res.OpenURL, "/-/open-query?query=")
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
	require.Len(t, res.Data, 1)
	row := res.Data[0]
	require.Equal(t, nil, row["inf"])
	require.Equal(t, nil, row["nan"])
}

func TestMetricsViewQueryRowLimit(t *testing.T) {
	// Setup a project with a metrics view that can return many rows
	// Configure a low AI query row limit to test the cap
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			// Create a model with 20 rows
			"test_data.sql": `SELECT UNNEST(RANGE(1, 21)) AS id, 100 AS revenue`,
			// Create a metrics view
			"test_metrics.yaml": `
type: metrics_view
model: test_data
dimensions:
- column: id
measures:
- expression: SUM(revenue)
  name: total_revenue
explore:
  skip: true
cache:
  enabled: false
`,
		},
		// Set a low AI query row limit for testing
		Variables: map[string]string{
			"rill.ai.query_row_limit": "10",
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	// Initialize test session
	s := newSession(t, rt, instanceID)

	// Test 1: Query without a limit should be capped at 10
	var res *ai.QueryMetricsViewResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.QueryMetricsViewName, &res, ai.QueryMetricsViewArgs{
		"metrics_view": "test_metrics",
		"dimensions":   []map[string]any{{"name": "id"}},
		"measures":     []map[string]any{{"name": "total_revenue"}},
	})
	require.NoError(t, err)
	require.Len(t, res.Data, 10, "expected results to be capped at 10 rows")

	// Test 2: Query with a higher limit should still be capped at 10
	var res2 *ai.QueryMetricsViewResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.QueryMetricsViewName, &res2, ai.QueryMetricsViewArgs{
		"metrics_view": "test_metrics",
		"dimensions":   []map[string]any{{"name": "id"}},
		"measures":     []map[string]any{{"name": "total_revenue"}},
		"limit":        100, // Request 100 rows, but should be capped at 10
	})
	require.NoError(t, err)
	require.Len(t, res2.Data, 10, "expected results to be capped at 10 rows even with higher limit")

	// Test 3: Query with a lower limit should respect the lower limit
	var res3 *ai.QueryMetricsViewResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.QueryMetricsViewName, &res3, ai.QueryMetricsViewArgs{
		"metrics_view": "test_metrics",
		"dimensions":   []map[string]any{{"name": "id"}},
		"measures":     []map[string]any{{"name": "total_revenue"}},
		"limit":        5, // Request 5 rows, should return 5
	})
	require.NoError(t, err)
	require.Len(t, res3.Data, 5, "expected results to respect lower limit of 5 rows")
}
