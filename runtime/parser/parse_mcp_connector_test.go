package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMCPConnector(t *testing.T) {
	ctx := context.Background()
	files := map[string]string{
		"rill.yaml": "",
		"connectors/linear.yaml": `
type: connector
driver: mcp
url: https://mcp.linear.app/mcp
display_name: Linear
ai_instructions: Use Linear to look up issues.
headers:
  Authorization: "Bearer {{ .env.LINEAR_TOKEN }}"
security:
  access: "{{ .user.admin }}"
  agents: [analyst]
  include:
    - if: true
      names: [search_issues, get_issue]
  exclude:
    - if: "NOT {{ .user.admin }}"
      names: [delete_issue]
`,
	}
	p, err := Parse(ctx, makeRepo(t, files), "", "", "duckdb", true)
	require.NoError(t, err)
	require.Empty(t, p.Errors)

	res := p.Resources[ResourceName{Kind: ResourceKindConnector, Name: "linear"}]
	require.NotNil(t, res)
	spec := res.ConnectorSpec
	require.Equal(t, "mcp", spec.Driver)
	require.Equal(t, []string{"analyst"}, spec.Agents)

	// Headers are resolved by the runtime like any other templated property.
	require.Equal(t, []string{"headers"}, spec.TemplatedProperties)

	props := spec.Properties.AsMap()
	require.Equal(t, "https://mcp.linear.app/mcp", props["url"])
	require.Equal(t, "Linear", props["display_name"])
	require.Equal(t, map[string]any{"Authorization": "Bearer {{ .env.LINEAR_TOKEN }}"}, props["headers"])
	require.NotContains(t, props, "security")

	// access + include + exclude
	require.Len(t, spec.SecurityRules, 3)
	access := spec.SecurityRules[0].GetAccess()
	require.NotNil(t, access)
	require.True(t, access.Allow)
	require.Equal(t, "{{ .user.admin }}", access.ConditionExpression)
	include := spec.SecurityRules[1].GetFieldAccess()
	require.NotNil(t, include)
	require.True(t, include.Allow)
	require.Equal(t, []string{"search_issues", "get_issue"}, include.Fields)
	exclude := spec.SecurityRules[2].GetFieldAccess()
	require.NotNil(t, exclude)
	require.False(t, exclude.Allow)
	require.Equal(t, []string{"delete_issue"}, exclude.Fields)
	require.Equal(t, "NOT {{ .user.admin }}", exclude.ConditionExpression)
}

func TestParseMCPConnectorDefaults(t *testing.T) {
	ctx := context.Background()

	// No security block: no rules, so the runtime only allows admins.
	p, err := Parse(ctx, makeRepo(t, map[string]string{
		"rill.yaml":            "",
		"connectors/mcp1.yaml": "type: connector\ndriver: mcp\nurl: https://example.com/mcp\n",
	}), "", "", "duckdb", true)
	require.NoError(t, err)
	require.Empty(t, p.Errors)
	spec := p.Resources[ResourceName{Kind: ResourceKindConnector, Name: "mcp1"}].ConnectorSpec
	require.Empty(t, spec.SecurityRules)
	require.Empty(t, spec.Agents)

	// A security block without access denies everyone, like metrics views.
	p, err = Parse(ctx, makeRepo(t, map[string]string{
		"rill.yaml":            "",
		"connectors/mcp2.yaml": "type: connector\ndriver: mcp\nurl: https://example.com/mcp\nsecurity:\n  exclude:\n    - if: true\n      names: [x]\n",
	}), "", "", "duckdb", true)
	require.NoError(t, err)
	require.Empty(t, p.Errors)
	spec = p.Resources[ResourceName{Kind: ResourceKindConnector, Name: "mcp2"}].ConnectorSpec
	require.NotEmpty(t, spec.SecurityRules)
	require.NotNil(t, spec.SecurityRules[0].GetAccess())
	require.False(t, spec.SecurityRules[0].GetAccess().Allow)
}

func TestParseMCPConnectorErrors(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name string
		yaml string
		err  string
	}{
		{"missing url", "type: connector\ndriver: mcp\n", "missing required property 'url'"},
		{"bad url", "type: connector\ndriver: mcp\nurl: not a url\n", "invalid 'url'"},
		{"non-string header", "type: connector\ndriver: mcp\nurl: https://x.com/mcp\nheaders:\n  X-Count: 1\n", "must be a string"},
		{"row filter", "type: connector\ndriver: mcp\nurl: https://x.com/mcp\nsecurity:\n  access: true\n  row_filter: \"1=1\"\n", "'row_filter' is not supported"},
		{"custom rules", "type: connector\ndriver: mcp\nurl: https://x.com/mcp\nsecurity:\n  access: true\n  rules:\n    - type: access\n      action: allow\n", "'rules' is not supported"},
		{"unknown agent", "type: connector\ndriver: mcp\nurl: https://x.com/mcp\nsecurity:\n  access: true\n  agents: [robot]\n", "unknown agent \"robot\""},
		{"unknown security key", "type: connector\ndriver: mcp\nurl: https://x.com/mcp\nsecurity:\n  acess: true\n", "invalid 'security'"},
		{"security on other driver", "type: connector\ndriver: slack\nbot_token: x\nsecurity:\n  access: true\n", "'security' is only supported for connectors with driver \"mcp\""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(ctx, makeRepo(t, map[string]string{"rill.yaml": "", "connectors/c.yaml": tc.yaml}), "", "", "duckdb", true)
			require.NoError(t, err)
			require.Len(t, p.Errors, 1)
			require.Contains(t, p.Errors[0].Message, tc.err)
			require.Nil(t, p.Resources[ResourceName{Kind: ResourceKindConnector, Name: "c"}])
		})
	}
}
