package ai

import (
	"fmt"
	"strings"
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/stretchr/testify/require"
)

func TestMaybeTruncateMessagesWithinBudget(t *testing.T) {
	messages := newTextMessages(500)
	result := maybeTruncateMessages(messages, sumEstimatedTokens(messages))
	require.Equal(t, messages, result)
}

func TestMaybeTruncateMessagesNoBudget(t *testing.T) {
	// A budget <= 0 disables truncation entirely.
	messages := newTextMessages(500)
	result := maybeTruncateMessages(messages, 0)
	require.Equal(t, messages, result)
}

func TestMaybeTruncateMessagesOverBudget(t *testing.T) {
	messages := newTextMessages(121)
	// A budget that fits all but one message forces a skip, which rounds up to truncateStep.
	result := maybeTruncateMessages(messages, sumEstimatedTokens(messages)-1)

	// Expect the first messages, a truncation indicator, then everything after the skipped range.
	require.Len(t, result, truncateKeepFirst+1+len(messages)-truncateKeepFirst-truncateStep)
	require.Equal(t, messages[:truncateKeepFirst], result[:truncateKeepFirst])
	require.Equal(t, fmt.Sprintf("... [%d messages omitted for brevity] ...", truncateStep), result[truncateKeepFirst].Content[0].GetText())
	require.Equal(t, messages[truncateKeepFirst+truncateStep:], result[truncateKeepFirst+1:])
	require.LessOrEqual(t, sumEstimatedTokens(result), sumEstimatedTokens(messages))
}

// TestMaybeTruncateMessagesStablePrefix verifies that as messages accumulate, each truncated
// result is an append-only extension of the previous one until the skipped count steps up.
// This property is what keeps LLM prompt caching effective across completion iterations.
func TestMaybeTruncateMessagesStablePrefix(t *testing.T) {
	// Uniform-sized messages so the token budget corresponds to an exact message count.
	const keepMessages = 100
	const maxN = 200
	budget := sumEstimatedTokens(newTextMessages(keepMessages))

	var steps int
	prev := maybeTruncateMessages(newTextMessages(keepMessages+1), budget)
	require.Less(t, len(prev), keepMessages+1, "expected truncation to trigger")
	for n := keepMessages + 2; n <= maxN; n++ {
		curr := maybeTruncateMessages(newTextMessages(n), budget)
		if len(curr) > len(prev) {
			// Within a step: the previous result must be a prefix of the current result.
			require.Equal(t, prev, curr[:len(prev)], "prefix changed at %d messages", n)
		} else {
			// The skipped count stepped up: the result shrinks by truncateStep-1 (truncateStep more skipped, 1 appended).
			require.Len(t, curr, len(prev)-(truncateStep-1), "unexpected step at %d messages", n)
			steps++
		}
		prev = curr
	}
	require.Equal(t, (maxN-keepMessages-1)/truncateStep, steps)
}

// TestMaybeTruncateMessagesBudgetTooSmall verifies the degenerate case where even heavy
// truncation can't fit the budget: everything between the first messages and the last message
// is skipped, and the result is still returned.
func TestMaybeTruncateMessagesBudgetTooSmall(t *testing.T) {
	messages := make([]*aiv1.CompletionMessage, 10)
	for i := range messages {
		messages[i] = NewTextCompletionMessage(RoleUser, strings.Repeat("x", 3000))
	}

	result := maybeTruncateMessages(messages, 1)
	require.Len(t, result, truncateKeepFirst+1+1)
	require.Equal(t, messages[:truncateKeepFirst], result[:truncateKeepFirst])
	require.Equal(t, messages[len(messages)-1], result[len(result)-1])
}

// TestMaybeTruncateMessagesUnbalancedToolCalls verifies that tool results whose calls were
// truncated away are also removed.
func TestMaybeTruncateMessagesUnbalancedToolCalls(t *testing.T) {
	messages := newTextMessages(121)
	// Make the last skipped message a tool call and the first kept message its result
	// (which must then be removed as unbalanced).
	messages[truncateKeepFirst+truncateStep-1] = &aiv1.CompletionMessage{
		Role: "assistant",
		Content: []*aiv1.ContentBlock{
			{BlockType: &aiv1.ContentBlock_ToolCall{ToolCall: &aiv1.ToolCall{Id: "call-1"}}},
		},
	}
	messages[truncateKeepFirst+truncateStep] = &aiv1.CompletionMessage{
		Role: "tool",
		Content: []*aiv1.ContentBlock{
			{BlockType: &aiv1.ContentBlock_ToolResult{ToolResult: &aiv1.ToolResult{Id: "call-1"}}},
		},
	}

	// A budget that exactly fits everything after the skipped range forces skipping truncateStep messages.
	budget := sumEstimatedTokens(messages) - sumEstimatedTokens(messages[truncateKeepFirst:truncateKeepFirst+truncateStep])
	result := maybeTruncateMessages(messages, budget)
	require.Len(t, result, truncateKeepFirst+1+len(messages)-truncateKeepFirst-truncateStep-1)
	for _, msg := range result {
		for _, block := range msg.Content {
			if res := block.GetToolResult(); res != nil {
				require.NotEqual(t, "call-1", res.Id)
			}
		}
	}
}

// newTextMessages generates n messages with identical estimated token counts.
func newTextMessages(n int) []*aiv1.CompletionMessage {
	messages := make([]*aiv1.CompletionMessage, n)
	for i := range messages {
		messages[i] = NewTextCompletionMessage(RoleUser, fmt.Sprintf("message %03d", i))
	}
	return messages
}

func sumEstimatedTokens(messages []*aiv1.CompletionMessage) int {
	var total int
	for _, m := range messages {
		total += estimateMessageTokens(m)
	}
	return total
}
