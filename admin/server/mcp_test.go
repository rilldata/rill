package server

import (
	"encoding/json"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/stretchr/testify/require"
)

func TestMCPTools(t *testing.T) {
	tools, err := mcpTools()
	require.NoError(t, err)
	require.Len(t, tools, len(mcpForwardedTools)+1)

	for _, tool := range tools {
		schema, ok := tool.InputSchema.(*jsonschema.Schema)
		require.True(t, ok, "tool %q", tool.Name)
		if tool.Name == mcpListProjectsName {
			continue
		}
		require.Contains(t, schema.Properties, mcpProjectArg, "tool %q", tool.Name)
		require.Contains(t, schema.Required, mcpProjectArg, "tool %q", tool.Name)
	}

	// query_metrics_view has a hand-written schema, which must be extended and not replaced.
	for _, tool := range tools {
		if tool.Name != ai.QueryMetricsViewName {
			continue
		}
		schema := tool.InputSchema.(*jsonschema.Schema)
		require.Contains(t, schema.Properties, "metrics_view")
		require.NotEmpty(t, schema.Defs)
	}
}

func TestTakeMCPProjectArg(t *testing.T) {
	project, rest, err := takeMCPProjectArg(json.RawMessage(`{"project":"org/proj","metrics_view":"mv"}`))
	require.NoError(t, err)
	require.Equal(t, "org/proj", project)
	require.JSONEq(t, `{"metrics_view":"mv"}`, string(rest))

	// The project argument is required, and must be a non-empty string.
	for _, args := range []string{``, `{}`, `{"project":""}`, `{"project":123}`, `not json`} {
		_, _, err := takeMCPProjectArg(json.RawMessage(args))
		require.Error(t, err, "args %q", args)
	}
}
