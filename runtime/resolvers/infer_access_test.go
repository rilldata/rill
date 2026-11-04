package resolvers_test

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestInferResolverAccess(t *testing.T) {
	rt, id := newInferAccessInstance(t)
	userAttrs := map[string]any{"country": "Denmark"}
	for _, tc := range []struct {
		name     string
		resolver string
		props    map[string]any
		fields   []string
		filter   string
	}{
		{
			name: "metrics", resolver: "metrics",
			props:  map[string]any{"metrics_view": "mv", "measures": []any{map[string]any{"name": "count"}}},
			fields: []string{"count"},
		},
		{
			name: "templated metrics SQL", resolver: "metrics_sql",
			props:  map[string]any{"sql": "SELECT count FROM mv WHERE d = '{{ .user.country }}'"},
			fields: []string{"count", "d"}, filter: "Denmark",
		},
		{
			name: "legacy metrics", resolver: "legacy_metrics",
			props: map[string]any{
				"query_name":      "MetricsViewAggregation",
				"query_args_json": `{"metrics_view":"mv","measures":[{"name":"count"}]}`,
			},
			fields: []string{"count"},
		},
		{
			name: "nested API", resolver: "api",
			props:  map[string]any{"api": "outer"},
			fields: []string{"count", "d"}, filter: "Denmark",
		},
		{
			name: "union", resolver: "union",
			props: map[string]any{"resolvers": []any{
				map[string]any{"name": "api", "properties": map[string]any{"api": "outer"}},
				map[string]any{"name": "metrics_sql", "properties": map[string]any{"sql": "SELECT count FROM mv"}},
			}},
			fields: []string{"count", "d", "count"}, filter: "Denmark",
		},
		{
			name: "text", resolver: "text",
			props: map[string]any{"text": `{{ metrics_sql "SELECT count FROM mv" }}`},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			refs, rules, err := rt.InferResolverAccess(t.Context(), id, tc.resolver, tc.props, userAttrs)
			require.NoError(t, err)
			require.Len(t, refs, 1)
			require.Equal(t, runtime.ResourceKindMetricsView, refs[0].Kind)
			require.Equal(t, "mv", refs[0].Name)

			var fields []string
			var filters string
			for _, rule := range rules {
				fields = append(fields, rule.GetFieldAccess().GetFields()...)
				if filter := rule.GetRowFilter(); filter != nil {
					data, err := protojson.Marshal(filter)
					require.NoError(t, err)
					filters += string(data)
				}
			}
			require.ElementsMatch(t, tc.fields, fields)
			if tc.filter != "" {
				require.Contains(t, filters, tc.filter)
			}
		})
	}

	_, _, err := rt.InferResolverAccess(t.Context(), id, "sql", map[string]any{"sql": "SELECT 1"}, nil)
	require.ErrorContains(t, err, "security rule inference not implemented")
}

// Inference skips security checks, but must not grant access to resources the caller cannot otherwise access.
func TestInferResolverAccessDoesNotAuthorizeExecution(t *testing.T) {
	rt, id := newInferAccessInstance(t)
	for _, tc := range []struct {
		resolver string
		props    map[string]any
	}{
		{"api", map[string]any{"api": "denied"}},
		{"metrics_sql", map[string]any{"sql": "SELECT count FROM locked"}},
		{"metrics", map[string]any{"metrics_view": "locked", "measures": []any{map[string]any{"name": "count"}}}},
	} {
		t.Run(tc.resolver, func(t *testing.T) {
			_, _, err := rt.InferResolverAccess(t.Context(), id, tc.resolver, tc.props, nil)
			require.NoError(t, err)

			// Populate the result cache with an authorized execution.
			res, _, err := rt.Resolve(t.Context(), &runtime.ResolveOptions{
				InstanceID: id, Resolver: tc.resolver, ResolverProperties: tc.props,
				Claims: &runtime.SecurityClaims{SkipChecks: true},
			})
			require.NoError(t, err)
			require.NoError(t, res.Close())

			_, _, err = rt.Resolve(t.Context(), &runtime.ResolveOptions{
				InstanceID: id, Resolver: tc.resolver, ResolverProperties: tc.props,
				Claims: &runtime.SecurityClaims{},
			})
			require.ErrorIs(t, err, runtime.ErrForbidden)
		})
	}
}

func newInferAccessInstance(t *testing.T) (*runtime.Runtime, string) {
	const mv = `type: metrics_view
model: m
timeseries: ts
dimensions:
  - column: d
measures:
  - name: count
    expression: count(*)
`
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: map[string]string{
		"models/m.sql":        "SELECT 'Denmark' AS d, TIMESTAMP '2024-01-02 00:00:00' AS ts",
		"metrics/mv.yaml":     mv,
		"metrics/locked.yaml": mv + "security:\n  access: false\n",
		"apis/inner.yaml":     "type: api\nmetrics_sql: SELECT count FROM mv WHERE d = '{{ .args.country }}'\n",
		"apis/outer.yaml":     "type: api\napi: inner\nargs:\n  country: Denmark\n",
		"apis/denied.yaml":    "type: api\nmetrics_sql: SELECT count FROM mv\nsecurity:\n  access: false\n",
	}})
	testruntime.RequireParseErrors(t, rt, id, nil)
	return rt, id
}
