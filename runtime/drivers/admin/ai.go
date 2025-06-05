package admin

import (
	"context"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

func (h *Handle) Complete(ctx context.Context, msgs []*drivers.CompletionMessage, config *drivers.CompletionConfig) (*drivers.CompletionMessage, error) {
	// Convert input to admin format
	reqMsgs := make([]*adminv1.CompletionMessage, len(msgs))
	for i, msg := range msgs {
		reqMsgs[i] = convertRuntimeMessageToAdminMessage(msg)
	}

	var reqTools []*adminv1.Tool
	if config != nil && len(config.Tools) > 0 {
		reqTools = make([]*adminv1.Tool, len(config.Tools))
		for i, tool := range config.Tools {
			reqTools[i] = convertRuntimeToolToAdminTool(tool)
		}
	}

	// Make the admin service call
	res, err := h.admin.Complete(ctx, &adminv1.CompleteRequest{
		Messages: reqMsgs,
		Tools:    reqTools,
	})
	if err != nil {
		return nil, err
	}

	// Convert response back to runtime format
	return convertAdminMessageToRuntimeMessage(res.Message), nil
}

// convertRuntimeMessageToAdminMessage converts a single runtime CompletionMessage to admin CompletionMessage format.
func convertRuntimeMessageToAdminMessage(msg *drivers.CompletionMessage) *adminv1.CompletionMessage {
	adminContentBlocks := make([]*adminv1.ContentBlock, len(msg.Content))
	for i, block := range msg.Content {
		adminBlock := &adminv1.ContentBlock{}

		if text := block.GetText(); text != "" {
			adminBlock.BlockType = &adminv1.ContentBlock_Text{Text: text}
		} else if toolCall := block.GetToolCall(); toolCall != nil {
			adminBlock.BlockType = &adminv1.ContentBlock_ToolCall{
				ToolCall: &adminv1.ToolCall{
					Id:    toolCall.Id,
					Name:  toolCall.Name,
					Input: toolCall.Input,
				},
			}
		} else if toolResult := block.GetToolResult(); toolResult != nil {
			adminBlock.BlockType = &adminv1.ContentBlock_ToolResult{
				ToolResult: &adminv1.ToolResult{
					Id:      toolResult.Id,
					Content: toolResult.Content,
					IsError: toolResult.IsError,
				},
			}
		}

		adminContentBlocks[i] = adminBlock
	}

	return &adminv1.CompletionMessage{
		Role:    msg.Role,
		Content: adminContentBlocks,
	}
}

// convertRuntimeToolToAdminTool converts a single runtime Tool to admin Tool format.
func convertRuntimeToolToAdminTool(tool drivers.Tool) *adminv1.Tool {
	return &adminv1.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		InputSchema: tool.InputSchema,
	}
}

// convertAdminMessageToRuntimeMessage converts admin CompletionMessage to runtime CompletionMessage format.
func convertAdminMessageToRuntimeMessage(adminMsg *adminv1.CompletionMessage) *drivers.CompletionMessage {
	contentBlocks := make([]*runtimev1.ContentBlock, len(adminMsg.Content))
	for i, adminBlock := range adminMsg.Content {
		runtimeBlock := &runtimev1.ContentBlock{}

		if text := adminBlock.GetText(); text != "" {
			runtimeBlock.BlockType = &runtimev1.ContentBlock_Text{Text: text}
		} else if adminToolCall := adminBlock.GetToolCall(); adminToolCall != nil {
			runtimeBlock.BlockType = &runtimev1.ContentBlock_ToolCall{
				ToolCall: &runtimev1.ToolCall{
					Id:    adminToolCall.Id,
					Name:  adminToolCall.Name,
					Input: adminToolCall.Input,
				},
			}
		} else if adminToolResult := adminBlock.GetToolResult(); adminToolResult != nil {
			runtimeBlock.BlockType = &runtimev1.ContentBlock_ToolResult{
				ToolResult: &runtimev1.ToolResult{
					Id:      adminToolResult.Id,
					Content: adminToolResult.Content,
					IsError: adminToolResult.IsError,
				},
			}
		}

		contentBlocks[i] = runtimeBlock
	}

	return &drivers.CompletionMessage{
		Role:    adminMsg.Role,
		Content: contentBlocks,
	}
}
