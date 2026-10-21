package ai_test

import (
	"slices"
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestReasoningIsReplayedWithinTheToolLoopOnly verifies that the model's reasoning is passed back on the next step of the
// same tool loop, which some providers require, and left out of later turns, where it is optional.
func TestReasoningIsReplayedWithinTheToolLoopOnly(t *testing.T) {
	script := &scriptedAIService{turns: []turnFunc{
		reasoningAndToolCallTurn("sig-1", ai.ListMetricsViewsName),
		textTurn("first"),
		textTurn("second"),
	}}
	s := newSkillReferencesSession(t, script)

	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: "first question"})
	require.NoError(t, err)

	// Passed back on the next step of the same loop, before the tool call it led to.
	require.Equal(t, []string{"sig-1"}, reasoningSignatures(script.inputs[1]))
	blocks := flatten(script.inputs[1])
	i := slices.IndexFunc(blocks, func(b *aiv1.ContentBlock) bool { return b.GetReasoning() != nil })
	require.Equal(t, ai.ListMetricsViewsName, blocks[i+1].GetToolCall().GetName())

	// Left out of the next turn.
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: "second question"})
	require.NoError(t, err)
	require.Empty(t, reasoningSignatures(script.inputs[2]))
}

// reasoningAndToolCallTurn makes the model return its reasoning followed by a call to the given tool.
func reasoningAndToolCallTurn(signature, tool string) turnFunc {
	return func(_ *drivers.CompleteOptions) *aiv1.CompletionMessage {
		input, _ := structpb.NewStruct(map[string]any{})
		return &aiv1.CompletionMessage{
			Role: "assistant",
			Content: []*aiv1.ContentBlock{
				{BlockType: &aiv1.ContentBlock_Reasoning{Reasoning: &aiv1.Reasoning{Signature: signature}}},
				{BlockType: &aiv1.ContentBlock_ToolCall{ToolCall: &aiv1.ToolCall{Id: "call-1", Name: tool, Input: input}}},
			},
		}
	}
}

func reasoningSignatures(messages []*aiv1.CompletionMessage) []string {
	var res []string
	for _, m := range messages {
		for _, b := range m.Content {
			if r := b.GetReasoning(); r != nil {
				res = append(res, r.Signature)
			}
		}
	}
	return res
}

func flatten(messages []*aiv1.CompletionMessage) []*aiv1.ContentBlock {
	var res []*aiv1.ContentBlock
	for _, m := range messages {
		res = append(res, m.Content...)
	}
	return res
}
