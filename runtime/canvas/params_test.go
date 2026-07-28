package canvas_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/canvas"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func testParams(t *testing.T) []*runtimev1.ComponentParam {
	t.Helper()
	return []*runtimev1.ComponentParam{
		{Name: "metrics_view", Type: "metrics_view", Required: true},
		{Name: "measure", Type: "measure", Required: true, MetricsViewParam: "metrics_view"},
		{Name: "dim", Type: "dimension", MetricsViewParam: "metrics_view"},
		{Name: "time_dim", Type: "time_dimension", MetricsViewParam: "metrics_view"},
		{Name: "limit", Type: "number", Default: structpb.NewNumberValue(500)},
		{Name: "shape", Type: "string", Default: structpb.NewStringValue("line"), Options: []*structpb.Value{structpb.NewStringValue("line"), structpb.NewStringValue("area")}},
		{Name: "smooth", Type: "boolean"},
	}
}

func testMetricsViews(t *testing.T) map[string]*runtimev1.MetricsViewSpec {
	t.Helper()
	return map[string]*runtimev1.MetricsViewSpec{
		"mv1": {
			TimeDimension: "ts",
			Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
				{Name: "country", Type: runtimev1.MetricsViewSpec_DIMENSION_TYPE_CATEGORICAL},
				{Name: "updated_at", Type: runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME},
			},
			Measures: []*runtimev1.MetricsViewSpec_Measure{
				{Name: "total"},
			},
		},
	}
}

func TestEffectiveArgs(t *testing.T) {
	spec := &runtimev1.ComponentSpec{
		Params: testParams(t),
		Input: []*runtimev1.ComponentVariable{
			{Name: "legacy", DefaultValue: structpb.NewStringValue("x")},
			{Name: "limit", DefaultValue: structpb.NewNumberValue(1)}, // Overridden by the param default
		},
	}

	args := canvas.EffectiveArgs(spec, map[string]any{"metrics_view": "mv1", "shape": "area"})
	require.Equal(t, map[string]any{
		"metrics_view": "mv1",
		"limit":        float64(500),
		"shape":        "area",
		"legacy":       "x",
	}, args)
}

func TestValidateParamBindings(t *testing.T) {
	params := testParams(t)
	mvs := testMetricsViews(t)

	valid := map[string]any{
		"metrics_view": "mv1",
		"measure":      "total",
		"dim":          "country",
		"time_dim":     "ts",
		"limit":        100,
		"shape":        "area",
		"smooth":       true,
	}

	cases := []struct {
		name    string
		mutate  func(m map[string]any)
		mvs     map[string]*runtimev1.MetricsViewSpec
		wantErr string
	}{
		{name: "valid", mutate: func(m map[string]any) {}},
		{name: "valid without metrics views", mutate: func(m map[string]any) { m["measure"] = "unknown" }, mvs: nil},
		{name: "unknown param", mutate: func(m map[string]any) { m["nope"] = 1 }, wantErr: `unknown param "nope"`},
		{name: "missing required", mutate: func(m map[string]any) { delete(m, "measure") }, wantErr: `missing value for required param "measure"`},
		{name: "type mismatch number", mutate: func(m map[string]any) { m["limit"] = "abc" }, wantErr: `param "limit" must be a number`},
		{name: "type mismatch boolean", mutate: func(m map[string]any) { m["smooth"] = "yes" }, wantErr: `param "smooth" must be a boolean`},
		{name: "type mismatch string", mutate: func(m map[string]any) { m["measure"] = 1 }, wantErr: `param "measure" must be a string`},
		{name: "value not in options", mutate: func(m map[string]any) { m["shape"] = "donut" }, wantErr: "not one of its options"},
		{name: "invalid metrics view", mutate: func(m map[string]any) { m["metrics_view"] = "mv2" }, wantErr: `metrics view "mv2" bound to param "metrics_view" is invalid or does not exist`},
		{name: "invalid measure", mutate: func(m map[string]any) { m["measure"] = "country" }, wantErr: `not a measure in metrics view "mv1"`},
		{name: "invalid dimension", mutate: func(m map[string]any) { m["dim"] = "total" }, wantErr: `not a dimension in metrics view "mv1"`},
		{name: "time dimension accepts time-typed dimension", mutate: func(m map[string]any) { m["time_dim"] = "updated_at" }},
		{name: "invalid time dimension", mutate: func(m map[string]any) { m["time_dim"] = "country" }, wantErr: `not a time dimension in metrics view "mv1"`},
		{name: "templated value skips validation", mutate: func(m map[string]any) { m["measure"] = "{{ .env.measure }}" }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bound := make(map[string]any, len(valid))
			for k, v := range valid {
				bound[k] = v
			}
			tc.mutate(bound)

			effectiveMVs := mvs
			if tc.mvs == nil && tc.name == "valid without metrics views" {
				effectiveMVs = nil
			}

			err := canvas.ValidateParamBindings(params, bound, effectiveMVs)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErr)
			}
		})
	}

	// Optional params with no default can be omitted entirely.
	err := canvas.ValidateParamBindings(params, map[string]any{"metrics_view": "mv1", "measure": "total"}, mvs)
	require.NoError(t, err)

	// Defaults are validated too: a field param with a default referencing a missing field errors without a binding.
	withDefault := []*runtimev1.ComponentParam{
		{Name: "metrics_view", Type: "metrics_view", Default: structpb.NewStringValue("mv1")},
		{Name: "measure", Type: "measure", MetricsViewParam: "metrics_view", Default: structpb.NewStringValue("missing")},
	}
	err = canvas.ValidateParamBindings(withDefault, nil, mvs)
	require.ErrorContains(t, err, `not a measure in metrics view "mv1"`)
}
