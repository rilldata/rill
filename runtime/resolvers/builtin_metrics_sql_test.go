package resolvers

import (
	"context"
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
			continue
		}
		defer res.Close()
		require.NoError(t, err)
		require.Equal(t, []byte(tc.want), must(res.MarshalJSON()))

		meta := res.Meta()
		require.NotNil(t, meta)

		schemaFields := map[string]bool{}
		for _, f := range res.Schema().Fields {
			schemaFields[f.Name] = true
		}

		for _, m := range meta["fields"].([]map[string]any) {
			name, ok := m["name"].(string)
			require.True(t, ok)
			_, exists := schemaFields[name]
			require.True(t, exists, "meta contains field not in schema: %s", name)
			delete(schemaFields, name)
		}
		require.Empty(t, schemaFields, "schema fields missing in meta: %v", schemaFields)
	}
}
