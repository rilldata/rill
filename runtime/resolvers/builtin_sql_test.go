package resolvers

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestBuiltinSQL(t *testing.T) {
	ctx := context.Background()
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			`rill.yaml`:      ``,
			`models/foo.sql`: `SELECT 10 AS a`,
		},
	})

	api, err := rt.APIForName(ctx, instanceID, "sql")
	require.NoError(t, err)

	tt := []struct {
		args       map[string]any
		attrs      map[string]any
		skipChecks bool
		want       string
		wantErr    string
	}{
		{
			args:       map[string]any{"sql": "SELECT a FROM foo"},
			skipChecks: true,
			want:       `[{"a":10}]`,
		},
		{
			args:  map[string]any{"sql": "SELECT a FROM foo"},
			attrs: map[string]any{"admin": true},
			want:  `[{"a":10}]`,
		},
		{
			args:    map[string]any{"sql": "SELECT a FROM foo"},
			attrs:   map[string]any{"admin": false},
			wantErr: "must be an admin to run arbitrary SQL queries",
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
}
