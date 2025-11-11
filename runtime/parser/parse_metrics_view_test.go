package parser

import (
	"context"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMetricsViewDimensionLookup(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 1 AS id`,
		// model m2
		`models/m2.sql`: `SELECT 1 AS id, 2 AS value`,
		// metrics view
		`metrics_views/mv1.yaml`: `
type: metrics_view
version: 1
model: m1
dimensions:
- name: foo
  expression: id
  lookup_table: m2 # Expect ref since it is a model in the same project
  lookup_key_column: id
  lookup_value_column: value
- name: bar
  expression: id
  lookup_table: ext # Expect no ref since it is not a known model (assumed to be a pre-existing table)
  lookup_key_column: id
  lookup_value_column: value
measures:
- name: count
  expression: COUNT(*)
`,
	}

	resources := []*Resource{
		// model m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m1.sql"])})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// model m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Paths: []string{"/models/m2.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m2.sql"])})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// metrics view
		{
			Name: ResourceName{Kind: ResourceKindMetricsView, Name: "mv1"},
			// Note: Expecting a ref to m2 since it's used as a lookup table and exists as a model in the same project.
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}, {Kind: ResourceKindModel, Name: "m2"}},
			Paths: []string{"/metrics_views/mv1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:   "duckdb",
				Model:       "m1",
				DisplayName: "Mv1",
				Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
					{
						Name:              "foo",
						DisplayName:       "Foo",
						Expression:        "id",
						LookupTable:       "m2",
						LookupKeyColumn:   "id",
						LookupValueColumn: "value",
					},
					{
						Name:              "bar",
						DisplayName:       "Bar",
						Expression:        "id",
						LookupTable:       "ext",
						LookupKeyColumn:   "id",
						LookupValueColumn: "value",
					},
				},
				Measures: []*runtimev1.MetricsViewSpec_Measure{
					{
						Name:        "count",
						DisplayName: "Count",
						Expression:  "COUNT(*)",
						Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
					},
				},
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestMetricsViewDimensionSmallestTimeGrain(t *testing.T) {
	files := map[string]string{
		// rill.yaml
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 1 AS id, '2025-01-01T00:00:00Z'::TIMESTAMP AS ts1, '2025-01-01'::DATE AS ts2`,
		// metrics view
		`metrics_views/mv1.yaml`: `
type: metrics_view
version: 1
model: m1
timeseries: ts1
smallest_time_grain: hour
dimensions:
- column: id
- column: ts2
  smallest_time_grain: day
measures:
- name: count
  expression: COUNT(*)
`,
	}

	resources := []*Resource{
		// model m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": strings.TrimSpace(files["models/m1.sql"])})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// metrics view
		{
			Name: ResourceName{Kind: ResourceKindMetricsView, Name: "mv1"},
			// Note: Expecting a ref to m2 since it's used as a lookup table and exists as a model in the same project.
			Refs:  []ResourceName{{Kind: ResourceKindModel, Name: "m1"}},
			Paths: []string{"/metrics_views/mv1.yaml"},
			MetricsViewSpec: &runtimev1.MetricsViewSpec{
				Connector:         "duckdb",
				Model:             "m1",
				DisplayName:       "Mv1",
				TimeDimension:     "ts1",
				SmallestTimeGrain: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
				Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
					{
						Name:        "ts1",
						DisplayName: "Ts1",
						Column:      "ts1",
					},
					{
						Name:        "id",
						DisplayName: "Id",
						Column:      "id",
					},
					{
						Name:              "ts2",
						DisplayName:       "Ts2",
						Column:            "ts2",
						SmallestTimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
					},
				},
				Measures: []*runtimev1.MetricsViewSpec_Measure{
					{
						Name:        "count",
						DisplayName: "Count",
						Expression:  "COUNT(*)",
						Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
					},
				},
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}
