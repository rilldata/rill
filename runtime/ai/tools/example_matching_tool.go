package tools

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
	"github.com/rilldata/rill/runtime/ai/tools/examples"
	"gopkg.in/yaml.v3"
)

type exampleMatchingInput struct {
	Query string `mapstructure:"query"`
}

func newExampleMatchingInput(in map[string]any) (*exampleMatchingInput, error) {
	var input exampleMatchingInput
	if err := mapstructure.Decode(in, &input); err != nil {
		return nil, fmt.Errorf("failed to decode input: %w", err)
	}
	if input.Query == "" {
		return nil, fmt.Errorf("`query` parameter is required and must be a string")
	}
	return &input, nil
}

func FetchTopNExamples() *tool.FunctionTool {
	tool := tool.NewFunctionTool(
		"fetch_top_n_examples",
		"Fetches the top N examples based on a query",
		func(ctx context.Context, params map[string]any) (any, error) {
			input, err := newExampleMatchingInput(params)
			if err != nil {
				return nil, err
			}

			examples := examples.Top3Fuzzy(input.Query)
			if len(examples) > 0 {
				return yaml.Marshal(examples)
			}
			return "No examples found matching the query", nil
		},
	)
	tool.WithSchema(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The query to match against examples",
			},
		},
		"required": []any{"query"},
	})
	return tool
}
