package server

import (
	"context"
	"errors"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
)

func (s *Server) Complete(ctx context.Context, req *adminv1.CompleteRequest) (*adminv1.CompleteResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.Int("args.messages_len", len(req.Messages)),
		attribute.Int("args.tools_len", len(req.Tools)),
	)

	// Handle backwards compatibility: migrate deprecated 'data' field to 'content'
	messages := make([]*aiv1.CompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = migrateCompletionMessage(msg)
	}

	// Pass messages and tools to the AI service
	msg, err := s.admin.AI.Complete(ctx, messages, req.Tools)
	if err != nil {
		return nil, err
	}

	if len(msg.Content) == 0 {
		return nil, errors.New("the AI responded with an empty message")
	}

	// Any tool use response will be passed to the client (the runtime server) for execution.
	return &adminv1.CompleteResponse{Message: msg}, nil
}

// migrateCompletionMessage handles backwards compatibility for CompletionMessage
// If the message has a non-empty 'data' field but empty 'content', it converts
// the 'data' field to a text content block for backwards compatibility.
func migrateCompletionMessage(msg *aiv1.CompletionMessage) *aiv1.CompletionMessage {
	// If content is already populated, use it as-is
	if len(msg.Content) > 0 {
		return msg
	}

	// If data field is populated but content is empty, migrate data to content
	if msg.Data != "" {
		return &aiv1.CompletionMessage{
			Role: msg.Role,
			Data: msg.Data, // Keep original for compatibility
			Content: []*aiv1.ContentBlock{
				{
					BlockType: &aiv1.ContentBlock_Text{
						Text: msg.Data,
					},
				},
			},
		}
	}

	// Return as-is if both are empty
	return msg
}
