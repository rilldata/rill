package parser

import (
	"context"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestModelOnSchemaChangeValidation(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		wantErr   string
		wantValid bool
	}{
		{
			name: "valid partition overwrite",
			model: `
type: model
incremental: true
partitions:
  sql: SELECT 1 AS partition_id
sql: SELECT 1
output:
  on_schema_change: append_new_columns
`,
			wantValid: true,
		},
		{
			name: "valid merge",
			model: `
type: model
incremental: true
partitions:
  sql: SELECT 1 AS partition_id
sql: SELECT 1 AS id
output:
  incremental_strategy: merge
  unique_key: [id]
  on_schema_change: fail
`,
			wantValid: true,
		},
		{
			name: "unknown mode",
			model: `
type: model
incremental: true
partitions:
  sql: SELECT 1 AS partition_id
sql: SELECT 1
output:
  on_schema_change: replace
`,
			wantErr: `invalid "output.on_schema_change" value "replace"`,
		},
		{
			name: "non string mode",
			model: `
type: model
incremental: true
partitions:
  sql: SELECT 1 AS partition_id
sql: SELECT 1
output:
  on_schema_change: true
`,
			wantErr: `"output.on_schema_change" must be a string`,
		},
		{
			name: "non duckdb output",
			model: `
type: model
incremental: true
partitions:
  sql: SELECT 1 AS partition_id
sql: SELECT 1
output:
  connector: clickhouse
  on_schema_change: ignore
`,
			wantErr: `only supported for DuckDB outputs`,
		},
		{
			name: "non incremental model",
			model: `
type: model
partitions:
  sql: SELECT 1 AS partition_id
sql: SELECT 1
output:
  on_schema_change: ignore
`,
			wantErr: `only supported for models with "incremental: true" and partitions`,
		},
		{
			name: "model without partitions",
			model: `
type: model
incremental: true
sql: SELECT 1
output:
  on_schema_change: ignore
`,
			wantErr: `only supported for models with "incremental: true" and partitions`,
		},
		{
			name: "append strategy",
			model: `
type: model
incremental: true
partitions:
  sql: SELECT 1 AS partition_id
sql: SELECT 1
output:
  incremental_strategy: append
  on_schema_change: ignore
`,
			wantErr: `only supported for the "merge" and "partition_overwrite" incremental strategies`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := makeRepo(t, map[string]string{"rill.yaml": "", "model.yaml": tt.model})
			p, err := Parse(context.Background(), repo, "", "", "duckdb", true)
			require.NoError(t, err)
			if tt.wantValid {
				require.Empty(t, p.Errors)
				require.Contains(t, p.Resources, ResourceName{Kind: ResourceKindModel, Name: "model"})
				return
			}
			require.NotEmpty(t, p.Errors)
			var messages []string
			for _, parseErr := range p.Errors {
				messages = append(messages, parseErr.Message)
			}
			require.Contains(t, fmt.Sprint(messages), tt.wantErr)
		})
	}
}

func TestModelOnSchemaChangeNamedOutputConnector(t *testing.T) {
	model := `
type: model
incremental: true
partitions:
  sql: SELECT 1 AS partition_id
sql: SELECT 1
output:
  connector: custom_olap
  on_schema_change: ignore
`
	tests := []struct {
		driver  string
		wantErr string
	}{
		{driver: "duckdb"},
		{driver: "clickhouse", wantErr: "only supported for DuckDB outputs"},
	}
	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			// The connector path intentionally sorts after the model path to verify validation is independent of parse order.
			repo := makeRepo(t, map[string]string{
				"rill.yaml":    "",
				"a_model.yaml": model,
				"z_connector.yaml": fmt.Sprintf(`
type: connector
name: custom_olap
driver: %s
`, tt.driver),
			})
			p, err := Parse(context.Background(), repo, "", "", "duckdb", true)
			require.NoError(t, err)
			if tt.wantErr == "" {
				require.Empty(t, p.Errors)
				return
			}
			require.Len(t, p.Errors, 1)
			require.ErrorContains(t, fmt.Errorf("%s", p.Errors[0].Message), tt.wantErr)
		})
	}
}

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
