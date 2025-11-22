package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/pkg/ai"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
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

	// Parse schema if given
	var outputSchema *jsonschema.Schema
	if req.OutputJsonSchema != "" {
		err := json.Unmarshal([]byte(req.OutputJsonSchema), &outputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to parse output JSON schema: %w", err)
		}
	}

	// Pass messages and tools to the AI service
	res, err := s.admin.AI.Complete(ctx, &ai.CompleteOptions{
		Messages:     messages,
		Tools:        req.Tools,
		OutputSchema: outputSchema,
	})
	if err != nil {
		return nil, err
	}
	if len(res.Message.Content) == 0 {
		return nil, errors.New("the AI responded with an empty message")
	}

	// Log token usage
	s.logger.Info("llm completion successful",
		zap.Int("input_messages", len(messages)),
		zap.Int("output_messages", len(res.Message.Content)),
		zap.Int("input_tokens", res.InputTokens),
		zap.Int("output_tokens", res.OutputTokens),
		observability.ZapCtx(ctx),
	)

	// Handle response backwards compatibility: if request used old format,
	// populate both data and content fields for old runtime compatibility
	responseMessage := res.Message
	if needsBackwardsCompatibleResponse {
		responseMessage = convertContentToData(res.Message)
	}

	// Any tool use response will be passed to the client (the runtime server) for execution.
	return &adminv1.CompleteResponse{
		Message:      responseMessage,
		InputTokens:  uint32(res.InputTokens),
		OutputTokens: uint32(res.OutputTokens),
	}, nil
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
