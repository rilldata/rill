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

func TestReasoning(t *testing.T) {
	script := &scriptedAIService{turns: []turnFunc{
		reasoningTurn("sig-1", &aiv1.ContentBlock{BlockType: &aiv1.ContentBlock_ToolCall{ToolCall: &aiv1.ToolCall{
			Id:    "call-1",
			Name:  ai.ListMetricsViewsName,
			Input: &structpb.Struct{},
		}}}),
		// The final result may also contain reasoning.
		reasoningTurn("sig-2", &aiv1.ContentBlock{BlockType: &aiv1.ContentBlock_Text{Text: "first answer"}}),
		textTurn("second answer"),
	}}
	s := newSkillReferencesSession(t, script)

	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: "first question"})
	require.NoError(t, err)

	// The reasoning is passed back on the next iteration of the tool call loop, before the tool call it led to.
	require.Len(t, script.inputs, 2)
	blocks := slices.Concat(contents(script.inputs[1])...)
	i := slices.IndexFunc(blocks, func(b *aiv1.ContentBlock) bool { return b.GetReasoning() != nil })
	require.GreaterOrEqual(t, i, 0)
	require.Equal(t, "sig-1", blocks[i].GetReasoning().Signature)
	require.Equal(t, ai.ListMetricsViewsName, blocks[i+1].GetToolCall().GetName())

	// The reasoning isn't kept after the tool call loop completes.
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: "second question"})
	require.NoError(t, err)
	require.Len(t, script.inputs, 3)
	for _, block := range slices.Concat(contents(script.inputs[2])...) {
		require.Nil(t, block.GetReasoning())
	}
}

// reasoningTurn makes the model return its reasoning followed by the given block.
func reasoningTurn(signature string, block *aiv1.ContentBlock) turnFunc {
	return func(_ *drivers.CompleteOptions) *aiv1.CompletionMessage {
		return &aiv1.CompletionMessage{
			Role: "assistant",
			Content: []*aiv1.ContentBlock{
				{BlockType: &aiv1.ContentBlock_Reasoning{Reasoning: &aiv1.Reasoning{Signature: signature}}},
				block,
			},
		}
	}
}

func contents(msgs []*aiv1.CompletionMessage) [][]*aiv1.ContentBlock {
	res := make([][]*aiv1.ContentBlock, len(msgs))
	for i, msg := range msgs {
		res[i] = msg.Content
	}
	return res
}
