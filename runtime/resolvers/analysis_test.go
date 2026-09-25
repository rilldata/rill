package resolvers

import (
	"context"
	"encoding/json"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestResolverAnalysis(t *testing.T) {
	rt, id := newAnalysisInstance(t)
	ctx := t.Context()
	for _, tc := range []struct {
		name     string
		resolver string
		props    map[string]any
		args     map[string]any
		fields   []string
		filter   string
	}{
		{
			name: "metrics", resolver: "metrics",
			props:  map[string]any{"metrics_view": "mv", "measures": []any{map[string]any{"name": "count"}}},
			fields: []string{"count"},
		},
		{
			name: "templated SQL", resolver: "metrics_sql",
			props:  map[string]any{"sql": "SELECT count FROM mv WHERE d = '{{ .user.country }}'"},
			fields: []string{"count", "d"}, filter: "Denmark",
		},
		{
			name: "additional filter", resolver: "metrics_sql",
			props: map[string]any{
				"sql": "SELECT count FROM mv",
				"additional_where_by_metrics_view": map[string]any{"mv": map[string]any{"cond": map[string]any{
					"op": "eq", "exprs": []any{map[string]any{"name": "d"}, map[string]any{"val": "Denmark"}},
				}}},
			},
			fields: []string{"count", "d"}, filter: "Denmark",
		},
		{
			name: "legacy", resolver: "legacy_metrics",
			props: map[string]any{
				"query_name":      "MetricsViewAggregation",
				"query_args_json": `{"metrics_view":"mv","measures":[{"name":"count"}]}`,
			},
			fields: []string{"count"},
		},
		{
			name: "nested API static args", resolver: "api",
			props:  map[string]any{"api": "outer", "args": map[string]any{"country": "France"}},
			args:   map[string]any{"country": "Germany"},
			fields: []string{"count", "d"}, filter: "Denmark",
		},
		{
			name: "builtin metrics", resolver: "api",
			props:  map[string]any{"api": "metrics"},
			args:   map[string]any{"metrics_view": "mv", "measures": []any{map[string]any{"name": "count"}}},
			fields: []string{"count"},
		},
		{
			name: "builtin metrics SQL", resolver: "api",
			props: map[string]any{"api": "metrics-sql"}, args: map[string]any{"sql": "SELECT count FROM mv"},
			fields: []string{"count"},
		},
		{
			name: "union repeated API", resolver: "union",
			props: map[string]any{"resolvers": []any{
				map[string]any{"name": "api", "properties": map[string]any{"api": "outer"}},
				map[string]any{"name": "api", "properties": map[string]any{"api": "outer"}},
			}},
			fields: []string{"count", "d", "count", "d"}, filter: "Denmark",
		},
		{
			name: "text references", resolver: "text",
			props: map[string]any{"text": `{{ metrics_sql "SELECT count FROM mv" }} {{ metrics_sql "SELECT count FROM mv" }}`},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := &runtime.ResolverAnalysisOptions{
				InstanceID: id, Properties: tc.props, Args: tc.args,
				UserAttributes: map[string]any{"country": "Denmark"},
			}
			before, err := json.Marshal(opts)
			require.NoError(t, err)
			analysis, err := rt.AnalyzeResolver(ctx, tc.resolver, opts)
			require.NoError(t, err)
			require.Len(t, analysis.Refs, 1)
			require.Equal(t, runtime.ResourceKindMetricsView, analysis.Refs[0].Kind)
			require.Equal(t, "mv", analysis.Refs[0].Name)
			var fields []string
			var filters string
			for _, rule := range analysis.RequiredSecurityRules {
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
			after, err := json.Marshal(opts)
			require.NoError(t, err)
			require.Equal(t, string(before), string(after), "analysis must not mutate its inputs")

			// Analysis and execution share parsing; their refs must agree.
			resolver, err := runtime.ResolverInitializers[tc.resolver](ctx, &runtime.ResolverOptions{
				Runtime:    rt,
				InstanceID: id,
				Properties: tc.props,
				Args:       tc.args,
				Claims:     &runtime.SecurityClaims{UserAttributes: opts.UserAttributes, SkipChecks: true},
			})
			require.NoError(t, err)
			defer resolver.Close()
			require.ElementsMatch(t, refNames(analysis.Refs), refNames(normalizeRefs(resolver.Refs())))
		})
	}

	for _, name := range []string{"self", "cycle_a"} {
		_, err := rt.AnalyzeResolver(ctx, "api", &runtime.ResolverAnalysisOptions{
			InstanceID: id, Properties: map[string]any{"api": name},
		})
		require.ErrorContains(t, err, "infinite recursion detected")
	}
	_, err := rt.AnalyzeResolver(ctx, "union", &runtime.ResolverAnalysisOptions{
		InstanceID: id, Properties: map[string]any{"resolvers": []any{
			map[string]any{"name": "sql", "properties": map[string]any{"sql": "SELECT 1"}},
		}},
	})
	require.ErrorContains(t, err, "security rule inference not implemented")
	analysis, err := rt.AnalyzeResolver(ctx, "ai", &runtime.ResolverAnalysisOptions{
		InstanceID: id, Properties: map[string]any{"is_report": true},
	})
	require.NoError(t, err)
	require.Empty(t, analysis.Refs)
	require.Empty(t, analysis.RequiredSecurityRules)
}

func TestResolverAnalysisDoesNotInitializeOrQuery(t *testing.T) {
	rt, id := newAnalysisInstance(t)
	// Removing the table leaves valid catalog metadata, but makes accidental data queries fail.
	olap, release, err := rt.OLAP(t.Context(), id, "duckdb")
	require.NoError(t, err)
	defer release()
	require.NoError(t, olap.Exec(t.Context(), &drivers.Statement{Query: "DROP VIEW m"}))
	for _, name := range []string{"api", "metrics", "metrics_sql", "text", "union", "legacy_metrics", "ai", "sql"} {
		initializer := runtime.ResolverInitializers[name]
		runtime.ResolverInitializers[name] = func(context.Context, *runtime.ResolverOptions) (runtime.Resolver, error) {
			t.Fatalf("analysis initialized executable resolver %q", name)
			return nil, nil
		}
		t.Cleanup(func() { runtime.ResolverInitializers[name] = initializer })
	}
	_, err = rt.AnalyzeResolver(t.Context(), "union", &runtime.ResolverAnalysisOptions{
		InstanceID: id, Properties: map[string]any{"resolvers": []any{
			map[string]any{"name": "api", "properties": map[string]any{"api": "outer"}},
			map[string]any{"name": "text", "properties": map[string]any{"text": `{{ metrics_sql "SELECT count FROM mv" }}`}},
			map[string]any{"name": "legacy_metrics", "properties": map[string]any{
				"query_name": "MetricsViewAggregation", "query_args_json": `{"metrics_view":"mv","measures":[{"name":"count"}]}`,
			}},
			map[string]any{"name": "ai", "properties": map[string]any{"is_report": true}},
		}},
	})
	require.NoError(t, err)
	_, err = rt.AnalyzeResolver(t.Context(), "sql", &runtime.ResolverAnalysisOptions{InstanceID: id})
	require.ErrorContains(t, err, "security rule inference not implemented")
}

func TestResolverAnalysisDynamicTime(t *testing.T) {
	rt, id := newAnalysisInstance(t)
	props := map[string]any{"sql": "SELECT count FROM mv WHERE ts >= time_range_start('P1D') AND ts < time_range_end('P1D')"}
	analysis, err := rt.AnalyzeResolver(t.Context(), "metrics_sql", &runtime.ResolverAnalysisOptions{InstanceID: id, Properties: props})
	require.NoError(t, err)
	require.Len(t, analysis.RequiredSecurityRules, 2)
	filter := analysis.RequiredSecurityRules[0].GetRowFilter().GetExpression()
	require.NotNil(t, filter)
	data, err := protojson.Marshal(filter)
	require.NoError(t, err)
	require.Contains(t, string(data), "2024-01-")
	require.ElementsMatch(t, []string{"count", "ts"}, analysis.RequiredSecurityRules[1].GetFieldAccess().Fields)

	// Invalid time dimensions propagate lookup errors, including on repeated calls after cleanup.
	for range 2 {
		_, err = rt.AnalyzeResolver(t.Context(), "metrics_sql", &runtime.ResolverAnalysisOptions{
			InstanceID: id, Properties: map[string]any{"sql": "SELECT count FROM mv WHERE d >= time_range_start('P1D')"},
		})
		require.Error(t, err)
	}
}

func TestResolverAnalysisDoesNotAuthorizeExecution(t *testing.T) {
	rt, id := newAnalysisInstance(t)
	for _, tc := range []struct {
		resolver string
		props    map[string]any
	}{
		{"api", map[string]any{"api": "denied"}},
		{"metrics_sql", map[string]any{"sql": "SELECT count FROM locked"}},
		{"metrics", map[string]any{"metrics_view": "locked", "measures": []any{map[string]any{"name": "count"}}}},
	} {
		_, err := rt.AnalyzeResolver(t.Context(), tc.resolver, &runtime.ResolverAnalysisOptions{InstanceID: id, Properties: tc.props})
		require.NoError(t, err)
		// First populate the result cache with an authorized execution.
		res, _, err := rt.Resolve(t.Context(), &runtime.ResolveOptions{
			InstanceID: id, Resolver: tc.resolver, ResolverProperties: tc.props,
			Claims: &runtime.SecurityClaims{SkipChecks: true},
		})
		require.NoError(t, err)
		require.NoError(t, res.Close())
		_, _, err = rt.Resolve(t.Context(), &runtime.ResolveOptions{
			InstanceID: id, Resolver: tc.resolver, ResolverProperties: tc.props, Claims: &runtime.SecurityClaims{},
		})
		require.ErrorIs(t, err, runtime.ErrForbidden)
	}
}

func newAnalysisInstance(t *testing.T) (*runtime.Runtime, string) {
	t.Helper()
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
		"apis/self.yaml":      "type: api\napi: self\n",
		"apis/cycle_a.yaml":   "type: api\napi: cycle_b\n",
		"apis/cycle_b.yaml":   "type: api\napi: cycle_a\n",
	}})
	testruntime.RequireParseErrors(t, rt, id, nil)
	return rt, id
}

func refNames(refs []*runtimev1.ResourceName) []string {
	names := make([]string, 0, len(refs))
	for _, ref := range refs {
		names = append(names, ref.Kind+"/"+ref.Name)
	}
	return names
}
