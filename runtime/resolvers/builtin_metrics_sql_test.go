package resolvers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestBuiltinMetricsSQL(t *testing.T) {
	ctx := context.Background()
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			`rill.yaml`:      ``,
			`models/foo.sql`: `SELECT 10 AS a`,
			`metrics/bar.yaml`: `
version: 1
type: metrics_view
model: foo
dimensions:
- column: a
measures:
- name: count
  expression: count(*)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	api, err := rt.APIForName(ctx, instanceID, "metrics-sql")
	require.NoError(t, err)

	tt := []struct {
		args       map[string]any
		attrs      map[string]any
		skipChecks bool
		want       string
		wantErr    string
	}{
		{
			args:       map[string]any{"sql": "SELECT count FROM bar"},
			skipChecks: true,
			want:       `[{"count":1}]`,
		},
		{
			args:  map[string]any{"sql": "SELECT count FROM bar"},
			attrs: map[string]any{"admin": true},
			want:  `[{"count":1}]`,
		},
		{
			args:    map[string]any{"sql": "SELECT count FROM bar"},
			attrs:   map[string]any{"admin": false},
			wantErr: `must be an admin to run arbitrary SQL queries`,
		},
	}

	for _, tc := range tt {
		res, err := rt.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID:         instanceID,
			Resolver:           api.Spec.Resolver,
			ResolverProperties: api.Spec.ResolverProperties.AsMap(),
			Args:               tc.args,
			Claims:             &runtime.SecurityClaims{UserAttributes: tc.attrs, SkipChecks: tc.skipChecks},
		})
		if tc.wantErr != "" {
			require.Equal(t, tc.wantErr, err.Error())
		} else {
			require.NoError(t, err)
			defer res.Close()
			require.Equal(t, []byte(tc.want), must(res.MarshalJSON()))
		}
	}

	t.Run("additional_where_structured_expression", func(t *testing.T) {
		props := api.Spec.ResolverProperties.AsMap()
		props["additional_where"] = map[string]any{
			"operator": "OPERATOR_EQ",
			"expressions": []any{
				map[string]any{"name": "a"},
				map[string]any{"literal": 10},
			},
		}
		res, err := rt.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID:         instanceID,
			Resolver:           api.Spec.Resolver,
			ResolverProperties: props,
			Args:               map[string]any{"sql": "SELECT a, count FROM bar"},
			Claims:             &runtime.SecurityClaims{SkipChecks: true},
		})
		require.NoError(t, err)
		defer res.Close()
		var rows []map[string]interface{}
		require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
		require.Len(t, rows, 1)
		require.Equal(t, float64(10), rows[0]["a"])
	})

	t.Run("additional_where_empty_expression", func(t *testing.T) {
		props := api.Spec.ResolverProperties.AsMap()
		props["additional_where"] = map[string]any{} // Empty expression should not filter anything
		res, err := rt.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID:         instanceID,
			Resolver:           api.Spec.Resolver,
			ResolverProperties: props,
			Args:               map[string]any{"sql": "SELECT a, count FROM bar"},
			Claims:             &runtime.SecurityClaims{SkipChecks: true},
		})
		require.NoError(t, err)
		defer res.Close()
		var rows []map[string]interface{}
		require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
		require.Len(t, rows, 1)
		// Should still return the row with a=10
		require.Equal(t, float64(10), rows[0]["a"])
	})
}
