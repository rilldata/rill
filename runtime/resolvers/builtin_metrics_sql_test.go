package resolvers

import (
	"context"
	"fmt"
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
			`models/foo.sql`: `SELECT 10 AS a, '2024-01-01T00:00:00Z'::TIMESTAMP as time`,
			`metrics/bar.yaml`: `
version: 1
type: metrics_view
model: foo
timeseries: time
dimensions:
- column: a
- name: time_7d
  expression: time + INTERVAL 7 DAYS
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
			args:  map[string]any{"sql": "SELECT count FROM bar where time >= '2024-01-01T00:00:00Z' and time < '2024-01-05T00:00:00Z'"},
			attrs: map[string]any{"admin": true},
			want:  `[{"count":1}]`,
		},
		{
			args:  map[string]any{"sql": "SELECT count FROM bar where time_7d >= '2024-01-10T00:00:00Z' OR time_7d < '2024-01-10T00:00:00Z'"},
			attrs: map[string]any{"admin": true},
			want:  `[{"count":1}]`,
		},
		{
			args:  map[string]any{"sql": "SELECT count FROM bar where time_7d >= time_range_start('P1D') AND time_7d < time_range_start('P3D')"},
			attrs: map[string]any{"admin": true},
			want:  `[{"count":0}]`,
		},
		{
			args:  map[string]any{"sql": "SELECT count FROM bar where time_7d BETWEEN time_range_start('P14D') AND '2024-01-05T00:00:00Z'"},
			attrs: map[string]any{"admin": true},
			want:  `[{"count":0}]`, // this time range falls in range of time but not time_7d
		},
		{
			args:  map[string]any{"sql": "SELECT count FROM bar where time >= '2024-01-01T00:00:00Z' and time_7d < '2024-01-05T00:00:00Z'"},
			attrs: map[string]any{"admin": true},
			want:  `[{"count":0}]`,
		},
		{
			args:    map[string]any{"sql": "SELECT count FROM bar"},
			attrs:   map[string]any{"admin": false},
			wantErr: `must be an admin to run arbitrary SQL queries`,
		},
	}

	for idx, tc := range tt {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			res, err := rt.Resolve(ctx, &runtime.ResolveOptions{
				InstanceID:         instanceID,
				Resolver:           api.Spec.Resolver,
				ResolverProperties: api.Spec.ResolverProperties.AsMap(),
				Args:               tc.args,
				Claims:             &runtime.SecurityClaims{UserAttributes: tc.attrs, SkipChecks: tc.skipChecks},
			})
			if tc.wantErr != "" {
				require.Equal(t, tc.wantErr, err.Error())
				return
			}
			require.NoError(t, err)
			defer res.Close()
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
		})
	}
}
