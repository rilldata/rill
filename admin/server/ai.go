package server

import (
	"context"
	"errors"
	"strings"

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
	needsBackwardsCompatibleResponse := false
	for i, msg := range req.Messages {
		if msg.Data != "" && len(msg.Content) == 0 {
			// Convert deprecated 'data' field to 'content'
			messages[i] = convertDataToContent(msg)
			needsBackwardsCompatibleResponse = true
		} else {
			// Use message as-is (content exists or both fields empty)
			messages[i] = msg
		}
	}

	// Pass messages and tools to the AI service
	msg, err := s.admin.AI.Complete(ctx, messages, req.Tools)
	if err != nil {
		return nil, err
	}

	if len(msg.Content) == 0 {
		return nil, errors.New("the AI responded with an empty message")
	}

	// Handle response backwards compatibility: if request used old format,
	// populate both data and content fields for old runtime compatibility
	responseMessage := msg
	if needsBackwardsCompatibleResponse {
		responseMessage = convertContentToData(msg)
	}

	// Any tool use response will be passed to the client (the runtime server) for execution.
	return &adminv1.CompleteResponse{Message: responseMessage}, nil
}

// convertDataToContent converts a message's deprecated 'data' field to 'content' blocks
// This function assumes the message has a non-empty data field and empty content
func convertDataToContent(msg *aiv1.CompletionMessage) *aiv1.CompletionMessage {
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

// convertContentToData creates a response message with both content and data fields
// for compatibility with old runtimes that expect the deprecated data field
func convertContentToData(msg *aiv1.CompletionMessage) *aiv1.CompletionMessage {
	// Extract text content from content blocks for the data field
	var textParts []string
	for _, block := range msg.Content {
		if text := block.GetText(); text != "" {
			textParts = append(textParts, text)
		}
	}

	// Create new message with both content (new format) and data (old format)
	return &aiv1.CompletionMessage{
		Role:    msg.Role,
		Data:    strings.Join(textParts, ""), // Populate deprecated field for old runtimes
		Content: msg.Content,                 // Keep new format for modern runtimes
	}
}
