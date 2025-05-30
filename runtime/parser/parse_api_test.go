package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

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
