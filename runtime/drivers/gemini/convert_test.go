package gemini

import (
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/stretchr/testify/require"
)

func TestConvertMessagesSkipsReasoning(t *testing.T) {
	// Reasoning blocks from other providers are skipped instead of failing the request.
	_, contents, err := convertMessages([]*aiv1.CompletionMessage{
		{Role: "user", Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_Text{Text: "hi"}}}},
		{Role: "assistant", Content: []*aiv1.ContentBlock{
			{BlockType: &aiv1.ContentBlock_Reasoning{Reasoning: &aiv1.Reasoning{Signature: "sig"}}},
			{BlockType: &aiv1.ContentBlock_Text{Text: "hello"}},
		}},
	})
	require.NoError(t, err)
	require.Len(t, contents, 2)
	require.Len(t, contents[1].Parts, 1)
	require.Equal(t, "hello", contents[1].Parts[0].Text)
}
