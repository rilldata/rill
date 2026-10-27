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
			name: "where_kind",
			files: map[string]string{
				"rill.yaml": ``,
				"m1.sql":    `SELECT 314`,
			},
			props: map[string]any{"where_kind": []string{"rill.runtime.v1.Model"}},
			expected: []map[string]any{
				{"type": "Model", "name": "m1", "status": "Idle", "error": ""},
			},
		},
		{
			name: "where_kind shorthand as string",
			files: map[string]string{
				"rill.yaml": ``,
				"m1.sql":    `SELECT 314`,
			},
			props: map[string]any{"where_kind": "project_parser"},
			expected: []map[string]any{
				{"type": "ProjectParser", "name": "parser", "status": "Idle", "error": ""},
			},
		},
		{
			name: "where_name is case-insensitive",
			files: map[string]string{
				"rill.yaml": ``,
				"m1.sql":    `SELECT 314`,
				"m2.sql":    `SELECT 159`,
			},
			props: map[string]any{"where_name": []string{"M2"}},
			expected: []map[string]any{
				{"type": "Model", "name": "m2", "status": "Idle", "error": ""},
			},
		},
		{
			name: "where_name with where_error",
			files: map[string]string{
				"rill.yaml": ``,
				"m1.sql":    `SELECT error("boom1")`,
				"m2.sql":    `SELECT error("boom2")`,
				"m3.sql":    `SELECT 314`,
			},
			props: map[string]any{"where_error": true, "where_kind": []string{"model"}, "where_name": []string{"m2", "m3"}},
			expected: []map[string]any{
				{"type": "Model", "name": "m2", "status": "Idle", "error": "boom2"},
			},
		},
		{
			name:  "unknown kind",
			files: map[string]string{`rill.yaml`: ``},
			props: map[string]any{"where_kind": []string{"dashboard"}},
			err:   `unknown resource kind "dashboard"`,
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
