package server

import (
	"context"
	"testing"

	"github.com/rilldata/rill/admin"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCompleteReasoning(t *testing.T) {
	// Older runtimes don't support reasoning blocks, so they're removed unless the client opted in.
	s := &Server{
		logger: zap.NewNop(),
		admin:  &admin.Service{AI: reasoningAIService{}},
	}
	for _, includeReasoning := range []bool{false, true} {
		res, err := s.Complete(t.Context(), &adminv1.CompleteRequest{
			Messages:         []*aiv1.CompletionMessage{{Role: "user", Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_Text{Text: "hi"}}}}},
			IncludeReasoning: includeReasoning,
		})
		require.NoError(t, err)
		if includeReasoning {
			require.Len(t, res.Message.Content, 2)
			require.Equal(t, "sig", res.Message.Content[0].GetReasoning().GetSignature())
		} else {
			require.Len(t, res.Message.Content, 1)
			require.Equal(t, "hello", res.Message.Content[0].GetText())
		}
	}
}

// reasoningAIService is a drivers.AIService that responds with reasoning followed by text.
type reasoningAIService struct{}

func (reasoningAIService) Complete(_ context.Context, _ *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	return &drivers.CompleteResult{
		Message: &aiv1.CompletionMessage{
			Role: "assistant",
			Content: []*aiv1.ContentBlock{
				{BlockType: &aiv1.ContentBlock_Reasoning{Reasoning: &aiv1.Reasoning{Signature: "sig"}}},
				{BlockType: &aiv1.ContentBlock_Text{Text: "hello"}},
			},
		},
		Provider: "claude",
	}, nil
}
