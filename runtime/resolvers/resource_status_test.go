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
	cases := []struct {
		name       string
		files      map[string]string
		whereError bool
		expected   []map[string]any
	}{
		{
			name:  "no resources",
			files: map[string]string{`rill.yaml`: ``},
			expected: []map[string]any{
				{"type": "ProjectParser", "name": "parser", "status": "Idle", "error": ""},
			},
		},
		{
			name:       "no resources with where_error",
			files:      map[string]string{`rill.yaml`: ``},
			whereError: true,
			expected:   []map[string]any{},
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
			whereError: true,
			expected:   []map[string]any{},
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
			whereError: true,
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
				Files: tc.files,
			})

			res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
				InstanceID:         instanceID,
				Resolver:           "resource_status",
				ResolverProperties: map[string]any{"where_error": tc.whereError},
				Args:               nil,
				Claims:             &runtime.SecurityClaims{},
			})
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
