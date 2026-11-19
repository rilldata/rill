package resolvers

import (
	"context"
	"io"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestResourceStatus(t *testing.T) {
	// A model, a metrics view, and its inline explore all named "orders".
	sameNameFiles := map[string]string{
		"rill.yaml":  ``,
		"orders.sql": `SELECT 1 AS x`,
		"metrics/orders.yaml": `
type: metrics_view
model: orders
dimensions:
- column: x
measures:
- name: count
  expression: count(*)
explore:
  display_name: Orders
`,
	}

	cases := []struct {
		name     string
		files    map[string]string
		props    map[string]any
		expected []map[string]any
		err      string
	}{
		{
			name:  "no resources",
			files: map[string]string{`rill.yaml`: ``},
			expected: []map[string]any{
				{"type": "ProjectParser", "name": "parser", "status": "Idle", "error": ""},
			},
		},
		{
			name:     "no resources with where_error",
			files:    map[string]string{`rill.yaml`: ``},
			props:    map[string]any{"where_error": true},
			expected: []map[string]any{},
		},
		{
			name: "multiple resources",
			files: map[string]string{
				"rill.yaml": ``,
				"m1.sql":    `SELECT 314`,
				"m2.sql":    `SELECT 159`,
			},
			expected: []map[string]any{
				{"type": "Model", "name": "m1", "status": "Idle", "error": ""},
				{"type": "Model", "name": "m2", "status": "Idle", "error": ""},
				{"type": "ProjectParser", "name": "parser", "status": "Idle", "error": ""},
			},
		},
		{
			name: "multiple resources with where_error",
			files: map[string]string{
				"rill.yaml": ``,
				"m1.sql":    `SELECT 314`,
				"m2.sql":    `SELECT 159`,
			},
			props:    map[string]any{"where_error": true},
			expected: []map[string]any{},
		},
		{
			name: "resource in error state",
			files: map[string]string{
				"rill.yaml": ``,
				"m1.sql":    `SELECT 314`,
				"m2.sql":    `SELECT error("booom!")`,
			},
			expected: []map[string]any{
				{"type": "Model", "name": "m1", "status": "Idle", "error": ""},
				{"type": "Model", "name": "m2", "status": "Idle", "error": "booom!"},
				{"type": "ProjectParser", "name": "parser", "status": "Idle", "error": ""},
			},
		},
		{
			name: "resource in error state with where_error",
			files: map[string]string{
				"rill.yaml": ``,
				"m1.sql":    `SELECT 314`,
				"m2.sql":    `SELECT error("booom!")`,
			},
			props: map[string]any{"where_error": true},
			expected: []map[string]any{
				{"type": "Model", "name": "m2", "status": "Idle", "error": "booom!"},
			},
		},
		{
			name: "parse errors",
			files: map[string]string{
				"rill.yaml": ``,
				"m1.yaml":   "type model\nsql SELECT 314", // Invalid YAML because it's missing colons
			},
			expected: []map[string]any{
				{"type": "ProjectParser", "name": "parser", "status": "Idle", "error": "encountered parse errors"},
			},
		},
		{
			name:  "where_resource exact match",
			files: sameNameFiles,
			props: map[string]any{"where_resource": []string{"metrics_view/orders"}},
			expected: []map[string]any{
				{"type": "MetricsView", "name": "orders", "status": "Idle", "error": ""},
			},
		},
		{
			name:  "where_resource does not cross kinds and names",
			files: sameNameFiles,
			props: map[string]any{"where_resource": []string{"model/orders", "explore/sales"}},
			expected: []map[string]any{
				{"type": "Model", "name": "orders", "status": "Idle", "error": ""},
			},
		},
		{
			name:  "where_resource wildcard name",
			files: sameNameFiles,
			props: map[string]any{"where_resource": "explore/*"},
			expected: []map[string]any{
				{"type": "Explore", "name": "orders", "status": "Idle", "error": ""},
			},
		},
		{
			name:  "where_resource wildcard kind",
			files: sameNameFiles,
			props: map[string]any{"where_resource": []string{"*/ORDERS"}},
			expected: []map[string]any{
				{"type": "Explore", "name": "orders", "status": "Idle", "error": ""},
				{"type": "MetricsView", "name": "orders", "status": "Idle", "error": ""},
				{"type": "Model", "name": "orders", "status": "Idle", "error": ""},
			},
		},
		{
			name:  "where_resource bare name",
			files: sameNameFiles,
			props: map[string]any{"where_resource": []string{"orders"}},
			expected: []map[string]any{
				{"type": "Explore", "name": "orders", "status": "Idle", "error": ""},
				{"type": "MetricsView", "name": "orders", "status": "Idle", "error": ""},
				{"type": "Model", "name": "orders", "status": "Idle", "error": ""},
			},
		},
		{
			name: "where_resource with where_error",
			files: map[string]string{
				"rill.yaml": ``,
				"m1.sql":    `SELECT error("boom1")`,
				"m2.sql":    `SELECT error("boom2")`,
				"m3.sql":    `SELECT 314`,
			},
			props: map[string]any{"where_error": true, "where_resource": []string{"model/m2", "model/m3"}},
			expected: []map[string]any{
				{"type": "Model", "name": "m2", "status": "Idle", "error": "boom2"},
			},
		},
		{
			name:  "where_resource unknown kind",
			files: map[string]string{`rill.yaml`: ``},
			props: map[string]any{"where_resource": []string{"dashboard/orders"}},
			err:   `invalid where_resource "dashboard/orders": unknown resource kind`,
		},
		{
			name:  "where_resource empty segment",
			files: map[string]string{`rill.yaml`: ``},
			props: map[string]any{"where_resource": []string{"model/"}},
			err:   `invalid where_resource "model/"`,
		},
		{
			name:  "where_resource too many segments",
			files: map[string]string{`rill.yaml`: ``},
			props: map[string]any{"where_resource": []string{"model/a/b"}},
			err:   `invalid where_resource "model/a/b"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
				Files: tc.files,
			})

			res, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
				InstanceID:         instanceID,
				Resolver:           "resource_status",
				ResolverProperties: tc.props,
				Args:               nil,
				Claims:             &runtime.SecurityClaims{},
			})
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			defer res.Close()

			for idx, exp := range tc.expected {
				nxt, err := res.Next()
				require.NoError(t, err, "unexpected error at row %d", idx)

				if exp["error"] != "" { // If we expect an error, compare using regexp instead of direct equality.
					require.Regexp(t, exp["error"], nxt["error"], "unexpected error message at index %d", idx)
					delete(exp, "error")
					delete(nxt, "error")
				}

				require.Equal(t, exp, nxt, "unexpected row at index %d", idx)
			}

			_, err = res.Next()
			require.Equal(t, io.EOF, err)
		})
	}
}
