package parser

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestModelOutput(t *testing.T) {
	files := map[string]string{
		`rill.yaml`: ``,
		`m1.sql`: `
SELECT 1
`,
		`m2.yaml`: `
type: model
sql: SELECT 1
`,
		`m3.yaml`: `
type: model
connector: bigquery
sql: SELECT 1
`,
		`m4.yaml`: `
type: model
connector: bigquery
sql: SELECT 1
output:
  table: foobar
`,
		`m5.yaml`: `
type: model
connector: bigquery
sql: SELECT 1
output: clickhouse
`,
		`m6.yaml`: `
type: model
connector: bigquery
sql: SELECT 1
output:
  connector: clickhouse
`,
	}
	resources := []*Resource{
		// model m1
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 1"})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// model m2
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m2"},
			Paths: []string{"/m2.yaml"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 1"})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// model m3
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m3"},
			Paths: []string{"/m3.yaml"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "bigquery",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 1"})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// model m4
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m4"},
			Paths: []string{"/m4.yaml"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "bigquery",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 1"})),
				OutputConnector: "duckdb",
				OutputProperties: must(structpb.NewStruct(map[string]any{
					"table": "foobar",
				})),
				ChangeMode: runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// model m5
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m5"},
			Paths: []string{"/m5.yaml"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "bigquery",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 1"})),
				OutputConnector: "clickhouse",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		// model m6
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m6"},
			Paths: []string{"/m6.yaml"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "bigquery",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": "SELECT 1"})),
				OutputConnector: "clickhouse",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestModelWithExtraResolverFields(t *testing.T) {
	files := map[string]string{
		`rill.yaml`: ``,
		`m1.yaml`: `
type: model
sql: SELECT 1
partitions:
  sql: SELECT range::DATE AS day FROM range('2024-01-01T00:00:00Z'::TIMESTAMPTZ, now(), INTERVAL '1 DAY')
  concurrency: 10
`,
	}
	resources := []*Resource{
		// model m1 with an unexpected field in the resolver properties
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/m1.yaml"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule:    &runtimev1.Schedule{RefUpdate: true},
				InputConnector:     "duckdb",
				InputProperties:    must(structpb.NewStruct(map[string]any{"sql": "SELECT 1"})),
				PartitionsResolver: "sql",
				PartitionsResolverProperties: must(structpb.NewStruct(map[string]any{
					"connector": "duckdb",
					"sql":       "SELECT range::DATE AS day FROM range('2024-01-01T00:00:00Z'::TIMESTAMPTZ, now(), INTERVAL '1 DAY')",
				})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
	}

	ctx := context.Background()
	repo := makeRepo(t, files)
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, []*runtimev1.ParseError{
		{
			Message:  "undefined fields in resolver properties: [\"concurrency\"], will be ignored",
			FilePath: "/m1.yaml",
			Warning:  true,
		},
	})
}
