package ai

import (
	"fmt"
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/stretchr/testify/require"
)

func TestMaybeTruncateMessagesReasoning(t *testing.T) {
	newMessages := func(n int) []*aiv1.CompletionMessage {
		var msgs []*aiv1.CompletionMessage
		for i := range n {
			block := &aiv1.ContentBlock{BlockType: &aiv1.ContentBlock_Text{Text: fmt.Sprintf("message %d", i)}}
			if i%2 == 1 {
				block = &aiv1.ContentBlock{BlockType: &aiv1.ContentBlock_Reasoning{Reasoning: &aiv1.Reasoning{Signature: fmt.Sprintf("sig-%d", i)}}}
			}
			msgs = append(msgs, &aiv1.CompletionMessage{Role: "assistant", Content: []*aiv1.ContentBlock{block}})
		}
		return msgs
	}
	countReasoning := func(msgs []*aiv1.CompletionMessage) int {
		var n int
		for _, msg := range msgs {
			for _, block := range msg.Content {
				if block.GetReasoning() != nil {
					n++
				}
			}
		}
		return n
	}

	// Without truncation, reasoning is kept.
	msgs := newMessages(50)
	require.Equal(t, msgs, maybeTruncateMessages(msgs))

	// With truncation, the history changes, so all reasoning is removed.
	truncated := maybeTruncateMessages(newMessages(150))
	require.Less(t, len(truncated), 150)
	require.Zero(t, countReasoning(truncated))
}
