package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestSearchFiles(t *testing.T) {
	// Setup a basic project with various files
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
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
			// Create a non-YAML file
			"README.md": `# My Rill Project

This project analyzes order data and customer information.
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	// Initialize test session
	s := newSession(t, rt, instanceID)

	t.Run("search for SQL keyword", func(t *testing.T) {
		var res *ai.SearchFilesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SearchFilesName, &res, &ai.SearchFilesArgs{
			Pattern: "SELECT",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Matches, 2) // Should find both model files

		// Verify we found the expected files
		foundPaths := make(map[string]bool)
		for _, match := range res.Matches {
			foundPaths[match.Path] = true
			require.NotEmpty(t, match.Lines)
			require.NotEmpty(t, match.Snippets)
		}
		require.True(t, foundPaths["/models/orders.yaml"])
		require.True(t, foundPaths["/models/customers.yaml"])
	})

	t.Run("search with case insensitive pattern", func(t *testing.T) {
		var res *ai.SearchFilesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SearchFilesName, &res, &ai.SearchFilesArgs{
			Pattern:       "customer",
			CaseSensitive: false,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.GreaterOrEqual(t, len(res.Matches), 2) // Should find customers.yaml and orders_metrics.yaml at minimum
	})

	t.Run("search with case sensitive pattern", func(t *testing.T) {
		var res *ai.SearchFilesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SearchFilesName, &res, &ai.SearchFilesArgs{
			Pattern:       "AS",
			CaseSensitive: true,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.GreaterOrEqual(t, len(res.Matches), 2) // Should find both model files with uppercase AS
	})

	t.Run("search with glob pattern filter", func(t *testing.T) {
		var res *ai.SearchFilesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SearchFilesName, &res, &ai.SearchFilesArgs{
			Pattern:     "type:",
			GlobPattern: "**models/**",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, res.Matches, 2) // Should only find files in models directory

		for _, match := range res.Matches {
			require.Contains(t, match.Path, "/models/")
		}
	})

	t.Run("search with regex pattern", func(t *testing.T) {
		var res *ai.SearchFilesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SearchFilesName, &res, &ai.SearchFilesArgs{
			Pattern: `customer_\w+`,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.Matches)

		// Verify snippets contain context lines
		for _, match := range res.Matches {
			for _, snippet := range match.Snippets {
				require.Contains(t, snippet, ">") // Should have the highlight marker
				require.Contains(t, snippet, ":") // Should have line numbers
			}
		}
	})

	t.Run("search with no matches", func(t *testing.T) {
		var res *ai.SearchFilesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SearchFilesName, &res, &ai.SearchFilesArgs{
			Pattern: "nonexistent_pattern_xyz123",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Empty(t, res.Matches)
	})

	t.Run("search with invalid regex", func(t *testing.T) {
		var res *ai.SearchFilesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SearchFilesName, &res, &ai.SearchFilesArgs{
			Pattern: "[invalid(regex",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid pattern")
	})

	t.Run("search returns line numbers", func(t *testing.T) {
		var res *ai.SearchFilesResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.SearchFilesName, &res, &ai.SearchFilesArgs{
			Pattern: "type: model",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.Matches)

		// Verify that line numbers are present and valid
		for _, match := range res.Matches {
			require.NotEmpty(t, match.Lines)
			for _, lineNum := range match.Lines {
				require.Greater(t, lineNum, 0)
			}
		}
	})
}
