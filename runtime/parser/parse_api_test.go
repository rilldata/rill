package parser

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAPI(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 1`,
		// api a1
		`apis/a1.yaml`: `
type: api
sql: select * from m1
`,
		// api a2
		`apis/a2.yaml`: `
type: api
metrics_sql: select * from m1
`,
		// api a3 with security rules
		`apis/a3.yaml`: `
type: api
sql: select * from m1
security:
  access: true
`,
		// api a4
		`apis/a4.yaml`: `
type: api
metrics_sql: select * from m1
security:
  access: '{{ .user.admin }}'
`,
		// api a5
		`apis/a5.yaml`: `
type: api
metrics_sql: select * from m1
skip_nested_security: true
security:
  access: '{{ .user.admin }}'
`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindModel, Name: "m1"},
			Paths: []string{"/models/m1.sql"},
			ModelSpec: &runtimev1.ModelSpec{
				RefreshSchedule: &runtimev1.Schedule{RefUpdate: true},
				InputConnector:  "duckdb",
				InputProperties: must(structpb.NewStruct(map[string]any{"sql": `SELECT 1`})),
				OutputConnector: "duckdb",
				ChangeMode:      runtimev1.ModelChangeMode_MODEL_CHANGE_MODE_RESET,
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a1"},
			Paths: []string{"/apis/a1.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"connector": "duckdb", "sql": "select * from m1"})),
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a2"},
			Paths: []string{"/apis/a2.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "metrics_sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"sql": "select * from m1"})),
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a3"},
			Paths: []string{"/apis/a3.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"connector": "duckdb", "sql": "select * from m1"})),
				SecurityRules: []*runtimev1.SecurityRule{
					{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
						ConditionExpression: "true",
						Allow:               true,
					}}},
				},
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a4"},
			Paths: []string{"/apis/a4.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "metrics_sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"sql": "select * from m1"})),
				SecurityRules: []*runtimev1.SecurityRule{
					{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
						ConditionExpression: "{{ .user.admin }}",
						Allow:               true,
					}}},
				},
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindAPI, Name: "a5"},
			Paths: []string{"/apis/a5.yaml"},
			APISpec: &runtimev1.APISpec{
				Resolver:           "metrics_sql",
				ResolverProperties: must(structpb.NewStruct(map[string]any{"sql": "select * from m1"})),
				SecurityRules: []*runtimev1.SecurityRule{
					{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
						ConditionExpression: "{{ .user.admin }}",
						Allow:               true,
					}}},
				},
				SkipNestedSecurity: true,
			},
		},
	}
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, nil)
}

func TestAPIWithOpenAPI(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// model m1
		`models/m1.sql`: `SELECT 'a' AS foo, 'b' AS bar, 1 AS baz`,
		// api a1
		`apis/a1.yaml`: `
type: api
sql: SELECT * FROM m1 WHERE foo = '{{ .args.foo }}' AND bar = '{{ .args.bar }}'
openapi:
  summary: Test API
  request_schema:
    type: object
    required: [foo, bar]
    properties:
      foo:
        type: string
        description: "Foo"
      bar:
        type: string
        description: "Bar"
  response_schema:
    type: object
    required: [foo, bar, baz]
    properties:
      foo:
        type: string
        description: "Foo"
      bar:
        type: string
        description: "Bar"
      baz:
        type: integer
        description: "Baz"
`,
	})

	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	require.Len(t, p.Errors, 0)
	require.Len(t, p.Resources, 2)

	api := p.Resources[ResourceName{ResourceKindAPI, "a1"}]
	require.NotNil(t, api)
	require.NotNil(t, api.APISpec.OpenapiRequestSchemaJson)
	require.NotNil(t, api.APISpec.OpenapiResponseSchemaJson)
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single word", "hello", "Hello"},
		{"multiple words", "hello world", "HelloWorld"},
		{"with underscores", "hello_world", "HelloWorld"},
		{"with dashes", "hello-world", "HelloWorld"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := toPascalCase(test.input)
			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}
