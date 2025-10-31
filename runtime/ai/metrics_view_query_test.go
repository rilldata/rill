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
	_, err := s.CallTool(t.Context(), ai.RoleUser, "query_metrics_view", &res, ai.QueryMetricsViewArgs{
		"metrics_view": "test_metrics",
		"dimensions":   []map[string]any{{"name": "country"}},
		"measures":     []map[string]any{{"name": "total_revenue"}},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Data)
	require.Contains(t, res.OpenURL, "https://ui.rilldata.com/test-org/test-project")
	require.Contains(t, res.OpenURL, "/-/open-query?query=")
}
