package runtime

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

// ===== CONSTANTS AND TYPES =====

// Constants for AI completion
const aiGenerateTimeout = 30 * time.Second
const maxToolCallIterations = 20

// ===== PUBLIC API =====

// CompletionRequest represents the input for AI completion
type CompletionRequest struct {
	InstanceID     string
	ConversationID string // Empty string means create new conversation
	Messages       []*runtimev1.Message
}

// CompletionResult represents the output of AI completion
type CompletionResult struct {
	ConversationID string
	Messages       []*runtimev1.Message
}

// Tool represents a tool that can be called by the AI
type Tool struct {
	Name        string
	Description string
	InputSchema string // TODO: Use a better type
}

// ToolService interface for managing and executing tools - will be implemented by server layer
type ToolService interface {
	ListTools(ctx context.Context) ([]Tool, error)
	ExecuteTool(ctx context.Context, toolName string, toolArgs map[string]any) (any, error)
}

// CompleteWithTools runs a conversational AI completion with tool calling support using the provided tool service
func (r *Runtime) CompleteWithTools(ctx context.Context, req *CompletionRequest, toolService ToolService) (*CompletionResult, error) {
	// Get instance-specific logger
	logger, err := r.InstanceLogger(ctx, req.InstanceID)
	if err != nil {
		r.Logger.Error("failed to get instance logger", zap.Error(err), zap.String("instance_id", req.InstanceID), observability.ZapCtx(ctx))
		logger = r.Logger.With(zap.String("instance_id", req.InstanceID))
	}

	logger.Info("starting AI completion",
		zap.String("conversation_id", req.ConversationID),
		zap.Int("message_count", len(req.Messages)),
		observability.ZapCtx(ctx))

	start := time.Now()
	defer func() {
		logger.Info("completed AI completion",
			zap.Duration("duration", time.Since(start)),
			observability.ZapCtx(ctx))
	}()

	// 1. Determine conversation ID (create if needed)
	conversationID, err := r.ensureConversation(ctx, req.InstanceID, req.ConversationID, req.Messages)
	if err != nil {
		logger.Error("failed to ensure conversation", zap.Error(err), observability.ZapCtx(ctx))
		return nil, err
	}

	// 2. Save user messages to database first
	addedMessageIDs, err := r.saveUserMessages(ctx, req.InstanceID, conversationID, req.Messages)
	if err != nil {
		return nil, err
	}

	// 3. Load complete conversation context from database (single DB call)
	allMessages, err := r.loadConversationContext(ctx, req.InstanceID, conversationID)
	if err != nil {
		return nil, err
	}

	// 4. Execute AI completion with database-backed context
	logger.Debug("executing AI completion", zap.Int("context_messages", len(allMessages)), observability.ZapCtx(ctx))
	contentBlocks, err := r.executeAICompletion(ctx, req.InstanceID, allMessages, toolService)
	if err != nil {
		logger.Error("AI completion failed", zap.Error(err), observability.ZapCtx(ctx))
		return nil, err
	}

	// 5. Save assistant message and build response
	return r.buildCompletionResult(ctx, req.InstanceID, conversationID, contentBlocks, addedMessageIDs)
}

// noOpToolService provides a no-op implementation for conversations without tool calling
type noOpToolService struct{}

func (n *noOpToolService) ListTools(ctx context.Context) ([]Tool, error) {
	return []Tool{}, nil
}

func (n *noOpToolService) ExecuteTool(ctx context.Context, toolName string, toolArgs map[string]any) (any, error) {
	return nil, fmt.Errorf("tool execution not implemented in runtime layer")
}

// Complete runs a conversational AI completion without tool calling support
func (r *Runtime) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResult, error) {
	// Use the no-op tool service for conversations that don't need tools
	return r.CompleteWithTools(ctx, req, &noOpToolService{})
}

// ===== BUSINESS LOGIC HELPERS =====

// ensureConversation determines the conversation ID - creates new if empty, returns existing if provided
func (r *Runtime) ensureConversation(ctx context.Context, instanceID, conversationID string, newMessages []*runtimev1.Message) (string, error) {
	if conversationID == "" {
		// Create new conversation using the first user message as title
		title := createConversationTitle(newMessages)
		conv, err := r.createConversation(ctx, instanceID, title)
		if err != nil {
			return "", err
		}
		return conv.Id, nil
	}

	// For existing conversations, just return the ID - validation happens in loadConversationContext
	return conversationID, nil
}

// saveUserMessages saves all user messages from the request to the conversation database
func (r *Runtime) saveUserMessages(ctx context.Context, instanceID, conversationID string, messages []*runtimev1.Message) ([]string, error) {
	var addedMessageIDs []string
	for _, msg := range messages {
		messageID, err := r.addMessage(ctx, instanceID, conversationID, msg.Role, msg.Content)
		if err != nil {
			return nil, err
		}
		addedMessageIDs = append(addedMessageIDs, messageID)
	}
	return addedMessageIDs, nil
}

// loadConversationContext loads the complete conversation context from database for AI processing
func (r *Runtime) loadConversationContext(ctx context.Context, instanceID, conversationID string) ([]*drivers.CompletionMessage, error) {
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	conversation, err := catalog.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Convert all conversation messages (including just-saved ones) to completion messages
	allMessages := make([]*drivers.CompletionMessage, len(conversation.Messages))
	for i, msg := range conversation.Messages {
		allMessages[i] = &drivers.CompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return allMessages, nil
}

// executeAICompletion runs the AI completion loop with tool calling support
func (r *Runtime) executeAICompletion(ctx context.Context, instanceID string, allMessages []*drivers.CompletionMessage, toolService ToolService) ([]*runtimev1.ContentBlock, error) {
	// Get instance-specific logger
	logger, err := r.InstanceLogger(ctx, instanceID)
	if err != nil {
		r.Logger.Error("failed to get instance logger in executeAICompletion", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		logger = r.Logger.With(zap.String("instance_id", instanceID))
	}

	// Connect to the AI service configured for the instance
	ai, release, err := r.AI(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, aiGenerateTimeout)
	defer cancel()

	// Get available tools and convert to drivers.Tool format (once per conversation)
	tools, err := toolService.ListTools(ctx)
	if err != nil {
		logger.Error("failed to list tools", zap.Error(err), observability.ZapCtx(ctx))
		return nil, err
	}

	logger.Debug("loaded tools for completion", zap.Int("tool_count", len(tools)), observability.ZapCtx(ctx))

	// Convert runtime.Tool to drivers.Tool
	driverTools := make([]drivers.Tool, len(tools))
	for i, tool := range tools {
		driverTools[i] = drivers.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		}
	}

	// Tool calling loop - accumulate all content blocks for a single assistant message
	var contentBlocks []*runtimev1.ContentBlock
	iteration := 0

	for range maxToolCallIterations {
		iteration++
		logger.Debug("starting tool calling iteration", zap.Int("iteration", iteration), observability.ZapCtx(ctx))

		// Call the AI service - returns structured ContentBlocks
		res, err := ai.Complete(ctx, allMessages, driverTools)
		if err != nil {
			logger.Error("AI completion call failed", zap.Error(err), zap.Int("iteration", iteration), observability.ZapCtx(ctx))
			return nil, err
		}

		logger.Debug("received AI response", zap.Int("content_blocks", len(res.Content)), zap.Int("iteration", iteration), observability.ZapCtx(ctx))

		// Process the response content blocks
		var toolCalls []*runtimev1.ToolCall
		var hasToolCalls bool

		for _, block := range res.Content {
			// Add all content blocks from the response
			contentBlocks = append(contentBlocks, block)

			// Check for tool calls
			if toolCall := block.GetToolCall(); toolCall != nil {
				toolCalls = append(toolCalls, toolCall)
				hasToolCalls = true
			}
		}

		// If no tool calls, this is the final response
		if !hasToolCalls {
			logger.Debug("AI completion finished - no tool calls requested", observability.ZapCtx(ctx))
			return contentBlocks, nil
		}

		logger.Info("executing tool calls", zap.Int("tool_call_count", len(toolCalls)), zap.Int("iteration", iteration), observability.ZapCtx(ctx))

		// Add the assistant's response with tool calls to the AI conversation context
		allMessages = append(allMessages, res)

		// Execute each tool call and add results as content blocks
		for _, toolCall := range toolCalls {
			logger.Debug("executing tool", zap.String("tool_name", toolCall.Name), zap.String("tool_id", toolCall.Id), observability.ZapCtx(ctx))

			// Convert protobuf Struct back to map for MCP tool execution
			inputMap := toolCall.Input.AsMap()

			// Call tool service
			resp, err := toolService.ExecuteTool(ctx, toolCall.Name, inputMap)

			var toolResult *runtimev1.ToolResult
			if err != nil {
				logger.Error("tool execution failed", zap.Error(err), zap.String("tool_name", toolCall.Name), zap.String("tool_id", toolCall.Id), observability.ZapCtx(ctx))
				toolResult = &runtimev1.ToolResult{
					Id:      toolCall.Id,
					Content: fmt.Sprintf("Error executing tool %s: %v", toolCall.Name, err),
					IsError: true,
				}
			} else {
				logger.Debug("tool executed successfully", zap.String("tool_name", toolCall.Name), zap.String("tool_id", toolCall.Id), observability.ZapCtx(ctx))
				// Convert response to ToolResult
				var result string
				if respContent, ok := resp.([]mcp.Content); ok && len(respContent) > 0 {
					if textContent, ok := respContent[0].(mcp.TextContent); ok {
						result = textContent.Text
					} else {
						result = fmt.Sprintf("%v", respContent[0])
					}
				} else {
					result = fmt.Sprintf("%v", resp)
				}

				toolResult = &runtimev1.ToolResult{
					Id:      toolCall.Id,
					Content: result,
					IsError: false,
				}
			}

			// Add tool result as content block
			contentBlocks = append(contentBlocks, &runtimev1.ContentBlock{
				BlockType: &runtimev1.ContentBlock_ToolResult{
					ToolResult: toolResult,
				},
			})
		}

		// Add tool results to AI conversation context
		var toolResultBlocks []*runtimev1.ContentBlock
		for _, toolCall := range toolCalls {
			// Find the corresponding result
			for _, block := range contentBlocks {
				if result := block.GetToolResult(); result != nil && result.Id == toolCall.Id {
					toolResultBlocks = append(toolResultBlocks, block)
					break
				}
			}
		}

		// Add tool results as a user message to continue the conversation
		allMessages = append(allMessages, &drivers.CompletionMessage{
			Role:    "user",
			Content: toolResultBlocks,
		})

		// Continue the loop to get the next response
	}

	// If we reach here, we've exceeded the maximum tool call iterations
	logger.Warn("maximum tool call iterations reached, requesting final response",
		zap.Int("max_iterations", maxToolCallIterations),
		observability.ZapCtx(ctx))

	// Instead of erroring, inform the AI and get a final response
	limitMessage := &drivers.CompletionMessage{
		Role: "user",
		Content: []*runtimev1.ContentBlock{
			{
				BlockType: &runtimev1.ContentBlock_Text{
					Text: fmt.Sprintf("Tool call limit reached (%d iterations). Please provide a final response without additional tool calls.", maxToolCallIterations),
				},
			},
		},
	}
	allMessages = append(allMessages, limitMessage)

	// Get final response from AI without tools
	res, err := ai.Complete(ctx, allMessages, []drivers.Tool{}) // No tools provided
	if err != nil {
		logger.Error("final AI completion call failed after tool limit", zap.Error(err), observability.ZapCtx(ctx))
		return nil, err
	}

	logger.Info("received final response after tool call limit", observability.ZapCtx(ctx))

	// Add the final response content blocks
	contentBlocks = append(contentBlocks, res.Content...)

	return contentBlocks, nil
}

// buildCompletionResult saves the assistant message and builds the final result
func (r *Runtime) buildCompletionResult(ctx context.Context, instanceID, conversationID string, contentBlocks []*runtimev1.ContentBlock, addedMessageIDs []string) (*CompletionResult, error) {
	// Save the complete assistant message with all content blocks
	messageID, err := r.addMessage(ctx, instanceID, conversationID, "assistant", contentBlocks)
	if err != nil {
		return nil, err
	}
	addedMessageIDs = append(addedMessageIDs, messageID)

	// Get the conversation to retrieve all saved messages with timestamps
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	conversation, err := catalog.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Collect all messages that were added during this Complete call
	var addedMessages []*runtimev1.Message
	for _, messageID := range addedMessageIDs {
		for _, msg := range conversation.Messages {
			if msg.Id == messageID {
				addedMessages = append(addedMessages, msg)
				break
			}
		}
	}

	// Return final result with all messages added during this Complete call
	return &CompletionResult{
		ConversationID: conversationID,
		Messages:       addedMessages,
	}, nil
}

// ===== UTILITY FUNCTIONS =====

// createConversationTitle creates a conversation title from the first user message.
// It truncates the message to a reasonable length and cleans it up.
func createConversationTitle(messages []*runtimev1.Message) string {
	// Find the first user message with text content
	for _, msg := range messages {
		if msg.Role == "user" {
			for _, block := range msg.Content {
				if block.GetText() != "" {
					title := strings.TrimSpace(block.GetText())

					// Truncate to 50 characters and add ellipsis if needed
					if len(title) > 50 {
						title = title[:50] + "..."
					}

					// Replace newlines with spaces
					title = strings.ReplaceAll(title, "\n", " ")
					title = strings.ReplaceAll(title, "\r", " ")

					// Collapse multiple spaces
					for strings.Contains(title, "  ") {
						title = strings.ReplaceAll(title, "  ", " ")
					}

					return title
				}
			}
		}
	}

	// Fallback if no user message found
	return "New Conversation"
}

// createConversation creates a new conversation with the given title
func (r *Runtime) createConversation(ctx context.Context, instanceID, title string) (*runtimev1.Conversation, error) {
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	conversationID, err := catalog.CreateConversation(ctx, title)
	if err != nil {
		return nil, err
	}

	// Return the created conversation
	return catalog.GetConversation(ctx, conversationID)
}

// addMessage adds a message to a conversation
func (r *Runtime) addMessage(ctx context.Context, instanceID, conversationID, role string, content []*runtimev1.ContentBlock) (string, error) {
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return "", err
	}
	defer release()

	return catalog.AddMessage(ctx, conversationID, role, content, nil)
}
