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
			`dashboards/bar.yaml`: `
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
		args    map[string]any
		attrs   map[string]any
		want    string
		wantErr string
	}{
		{
			args:  map[string]any{"sql": "SELECT count FROM bar"},
			attrs: map[string]any{},
			want:  `[{"count":1}]`,
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
			UserAttributes:     tc.attrs,
		})
		if tc.wantErr != "" {
			require.Equal(t, tc.wantErr, err.Error())
		} else {
			require.Equal(t, []byte(tc.want), res.Data)
		}
	}
}
