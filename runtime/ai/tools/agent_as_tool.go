package tools

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

type runAgentInput struct {
	Input string `mapstructure:"input"`
}

func newRunAgentInput(in map[string]any) (*runAgentInput, error) {
	var input runAgentInput
	if err := mapstructure.Decode(in, &input); err != nil {
		return nil, fmt.Errorf("failed to decode input: %w", err)
	}
	if err := input.validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}
	return &input, nil
}

func (i *runAgentInput) validate() error {
	if i.Input == "" {
		return fmt.Errorf("expected 'input' parameter to be a string")
	}
	return nil
}

func RunAgent(a *agent.Agent, r *runner.Runner, name, description string) (*tool.FunctionTool, error) {
	// Create a tool that runs the agent
	tool := tool.NewFunctionTool(
		name,
		description,
		func(ctx context.Context, params map[string]any) (any, error) {
			userInput, err := newRunAgentInput(params)
			if err != nil {
				return nil, err
			}
			result, err := r.Run(ctx, a, &runner.RunOptions{
				Input:    userInput.Input,
				MaxTurns: 10,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to run agent: %w", err)
			}
			return newToolResult(result.FinalOutput, nil), nil
		},
	)

	tool.WithSchema(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"input": map[string]any{
				"type":        "string",
				"description": "input to the tool",
			},
		},
		"required": []string{"input"},
	})
	return tool, nil
}
