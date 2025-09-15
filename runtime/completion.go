package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ===== CONSTANTS AND TYPES =====

// Constants for AI completion
const (
	completionTimeout     = 120 * time.Second // Overall timeout for entire AI completion
	aiRequestTimeout      = 30 * time.Second  // Timeout for individual AI API calls
	maxToolCallIterations = 20
)

// ===== PUBLIC API =====

// ToolService interface for managing and executing tools - will be implemented by server layer
type ToolService interface {
	ListTools(ctx context.Context) ([]*aiv1.Tool, error)
	ExecuteTool(ctx context.Context, toolName string, toolArgs map[string]any) (any, error)
}

// CompleteMessageCallback is called when a new message is added
type CompleteMessageCallback func(conversationID string, msg *runtimev1.Message) error

// CompleteWithToolsOptions represents the input for AI completion
type CompleteWithToolsOptions struct {
	OwnerID        string
	InstanceID     string
	ConversationID string                // Empty string means create new conversation
	AppContext     *runtimev1.AppContext // Used to seed new conversations with context
	Messages       []*runtimev1.Message
	ToolService    ToolService
	OnMessage      CompleteMessageCallback
}

// CompleteWithToolsResult represents the output of AI completion
type CompleteWithToolsResult struct {
	ConversationID string
	Messages       []*runtimev1.Message
}

// CompleteWithTools runs a conversational AI completion with tool calling support using the provided tool service
func (r *Runtime) CompleteWithTools(ctx context.Context, opts *CompleteWithToolsOptions) (result *CompleteWithToolsResult, err error) {
	// Get instance-specific logger
	logger, err := r.InstanceLogger(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	logger.Info("starting AI completion",
		zap.String("conversation_id", opts.ConversationID),
		zap.Int("message_count", len(opts.Messages)),
		observability.ZapCtx(ctx))

	start := time.Now()
	defer func() {
		if err != nil {
			logger.Info("failed AI completion",
				zap.Error(err),
				zap.Duration("duration", time.Since(start)),
				observability.ZapCtx(ctx))
		} else {
			logger.Info("completed AI completion",
				zap.Duration("duration", time.Since(start)),
				observability.ZapCtx(ctx))
		}
	}()

	// Determine conversation ID (create if needed)
	conversationID, err := r.ensureConversation(ctx, opts.InstanceID, opts.OwnerID, opts.ConversationID, opts.AppContext, opts.Messages)
	if err != nil {
		return nil, err
	}

	// If this is a new conversation, process app context and save system messages first
	var addedMessageIDs []string
	if opts.ConversationID == "" {
		// This was a new conversation, so process app context and save system messages
		var contextMessages []*runtimev1.Message
		contextMessages, err = r.processAppContext(ctx, opts.InstanceID, opts.AppContext, opts.ToolService)
		if err != nil {
			return nil, err
		}

		// Save system messages to database first
		for _, msg := range contextMessages {
			messageID, err := r.addMessage(ctx, opts.InstanceID, conversationID, msg.Role, msg.Content)
			if err != nil {
				return nil, err
			}
			addedMessageIDs = append(addedMessageIDs, messageID)
		}
	}

	// Save user messages to database
	var userMessageIDs []string
	userMessageIDs, err = r.saveUserMessages(ctx, opts.InstanceID, conversationID, opts.Messages)
	if err != nil {
		return nil, err
	}
	addedMessageIDs = append(addedMessageIDs, userMessageIDs...)

	// Emit preliminary user messages from the request.
	// NOTE: The messages are split up such that each block is emitted separately. This conforms with the new contract that CompleteStreaming only returns one block per message.
	// NOTE: Since the messages may be split up, it generates new non-persistent IDs for the streamed messages.
	if opts.OnMessage != nil {
		for _, msg := range opts.Messages {
			for _, block := range msg.Content {
				err := opts.OnMessage(conversationID, &runtimev1.Message{
					Id:        uuid.NewString(),
					Role:      msg.Role,
					Content:   []*aiv1.ContentBlock{block},
					CreatedOn: timestamppb.Now(),
					UpdatedOn: timestamppb.Now(),
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// Load complete conversation context from database (includes any saved system messages)
	var allMessages []*runtimev1.Message
	allMessages, err = r.loadConversationContext(ctx, opts.InstanceID, conversationID)
	if err != nil {
		return nil, err
	}

	// Execute AI completion with database-backed context
	var contentBlocks []*aiv1.ContentBlock
	contentBlocks, err = r.executeAICompletion(ctx, opts.InstanceID, conversationID, allMessages, opts.ToolService, opts.OnMessage)
	if err != nil {
		return nil, err
	}

	// Save assistant message and build response
	result, err = r.buildCompletionResult(ctx, opts.InstanceID, conversationID, contentBlocks, addedMessageIDs)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ===== BUSINESS LOGIC HELPERS =====

// ensureConversation determines the conversation ID - creates new if empty, validates existing if provided
func (r *Runtime) ensureConversation(ctx context.Context, instanceID, ownerID, conversationID string, appContext *runtimev1.AppContext, newMessages []*runtimev1.Message) (string, error) {
	if conversationID == "" {
		// Create new conversation using the first user message as title
		title := createConversationTitle(newMessages)
		conv, err := r.createConversation(ctx, instanceID, ownerID, title, appContext)
		if err != nil {
			return "", err
		}
		return conv.Id, nil
	}

	// For existing conversations, validate that it exists and belongs to the user
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return "", err
	}
	defer release()

	conversation, err := catalog.FindConversation(ctx, conversationID)
	if err != nil {
		return "", err
	}

	// Verify ownership
	if conversation.OwnerID != ownerID {
		return "", fmt.Errorf("conversation not found or access denied")
	}

	return conversationID, nil
}

// processAppContext processes the app context and generates contextual system messages
func (r *Runtime) processAppContext(ctx context.Context, instanceID string, appContext *runtimev1.AppContext, toolService ToolService) ([]*runtimev1.Message, error) {
	if appContext == nil {
		return nil, nil
	}

	switch appContext.ContextType {
	case runtimev1.AppContextType_APP_CONTEXT_TYPE_PROJECT_CHAT:
		return r.processProjectChatContext(ctx, instanceID)
	case runtimev1.AppContextType_APP_CONTEXT_TYPE_EXPLORE_DASHBOARD:
		return r.processExploreDashboardContext(ctx, instanceID, appContext.ContextMetadata, toolService)
	default:
		return nil, nil // Unknown context type, no system message will be added
	}
}

func (r *Runtime) processProjectChatContext(ctx context.Context, instanceID string) ([]*runtimev1.Message, error) {
	// Find instance-wide AI context
	var aiInstructions string
	instance, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance %q: %w", instanceID, err)
	}
	if instance.AIInstructions != "" {
		aiInstructions = instance.AIInstructions
	}

	// Build the system prompt
	systemPrompt := buildProjectChatSystemPrompt(aiInstructions)

	return []*runtimev1.Message{{
		Role: "system",
		Content: []*aiv1.ContentBlock{{
			BlockType: &aiv1.ContentBlock_Text{
				Text: systemPrompt,
			},
		}},
	}}, nil
}

// processExploreDashboardContext provides specific dashboard context
func (r *Runtime) processExploreDashboardContext(ctx context.Context, instanceID string, metadata *structpb.Struct, toolService ToolService) ([]*runtimev1.Message, error) {
	metadataMap := metadata.AsMap()
	dashboardName, ok := metadataMap["dashboard_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing dashboard_name in explore_dashboard context")
	}

	// Get the controller to access resources
	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// Get the explore resource to find its associated metrics view
	exploreResource, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindExplore, Name: dashboardName}, false)
	if err != nil {
		return nil, fmt.Errorf("could not find explore '%s': %w", dashboardName, err)
	}

	explore := exploreResource.GetExplore()
	if explore == nil || explore.State == nil || explore.State.ValidSpec == nil {
		return nil, fmt.Errorf("explore '%s' does not have a valid spec", dashboardName)
	}

	metricsViewName := explore.State.ValidSpec.MetricsView

	// Get specific metrics view details
	metricsViewSpec, err := toolService.ExecuteTool(ctx, "get_metrics_view", map[string]any{
		"metrics_view": metricsViewName,
	})
	if err != nil {
		return nil, err
	}

	// Get time range information for the metrics view
	timeRangeSummary, err := toolService.ExecuteTool(ctx, "query_metrics_view_summary", map[string]any{
		"metrics_view": metricsViewName,
	})
	if err != nil {
		return nil, err
	}

	// Find instance-wide AI context
	var aiInstructions string
	instance, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance %q: %w", instanceID, err)
	}
	if instance.AIInstructions != "" {
		aiInstructions = instance.AIInstructions
	}

	// Build the system prompt
	systemPrompt := buildExploreDashboardSystemPrompt(dashboardName, metricsViewName, metricsViewSpec, timeRangeSummary, aiInstructions)

	return []*runtimev1.Message{{
		Role: "system",
		Content: []*aiv1.ContentBlock{{
			BlockType: &aiv1.ContentBlock_Text{
				Text: systemPrompt,
			},
		}},
	}}, nil
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
func (r *Runtime) loadConversationContext(ctx context.Context, instanceID, conversationID string) ([]*runtimev1.Message, error) {
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	catalogMessages, err := catalog.FindMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Convert catalog messages to runtime protobuf messages
	allMessages := make([]*runtimev1.Message, len(catalogMessages))
	for i, msg := range catalogMessages {
		allMessages[i], err = MessageToPB(msg)
		if err != nil {
			return nil, err
		}
	}

	return allMessages, nil
}

// executeAICompletion runs the AI completion loop with tool calling support
func (r *Runtime) executeAICompletion(ctx context.Context, instanceID, conversationID string, allMessages []*runtimev1.Message, toolService ToolService, onMessage CompleteMessageCallback) ([]*aiv1.ContentBlock, error) {
	// Get instance-specific logger
	logger, err := r.InstanceLogger(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// Connect to the AI service configured for the instance
	ai, release, err := r.AI(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, completionTimeout)
	defer cancel()

	// Get available tools
	tools, err := toolService.ListTools(ctx)
	if err != nil {
		return nil, err
	}

	logger.Debug("loaded tools for completion", zap.Int("tool_count", len(tools)), observability.ZapCtx(ctx))

	// Tool calling loop - accumulate all content blocks for a single assistant message
	var contentBlocks []*aiv1.ContentBlock
	iteration := 0

	for range maxToolCallIterations {
		iteration++
		logger.Debug("starting tool calling iteration", zap.Int("iteration", iteration), observability.ZapCtx(ctx))

		// Truncate conversation if it's getting too long
		messages := maybeTruncateConversation(allMessages)
		if len(messages) < len(allMessages) {
			logger.Debug("truncated conversation for AI",
				zap.Int("original_messages", len(allMessages)),
				zap.Int("truncated_messages", len(messages)),
				observability.ZapCtx(ctx))
		}

		// Convert runtime messages to aiv1.CompletionMessage for AI service call
		completionMessages := make([]*aiv1.CompletionMessage, len(messages))
		for i, msg := range messages {
			completionMessages[i] = runtimeMessageToAICompletionMessage(msg)
		}

		// Call the AI service with individual timeout - returns structured ContentBlocks
		aiCtx, aiCancel := context.WithTimeout(ctx, aiRequestTimeout)
		res, err := ai.Complete(aiCtx, completionMessages, tools)
		aiCancel()
		if err != nil {
			return nil, err
		}

		logger.Debug("received AI response", zap.Int("content_blocks", len(res.Content)), zap.Int("iteration", iteration), observability.ZapCtx(ctx))

		// Process the response content blocks
		var toolCalls []*aiv1.ToolCall
		var hasToolCalls bool

		for _, block := range res.Content {
			// Add all content blocks from the response
			contentBlocks = append(contentBlocks, block)

			// Emit preliminary message for streaming
			if onMessage != nil {
				err := onMessage(conversationID, &runtimev1.Message{
					Id:        uuid.NewString(),
					Role:      "assistant",
					Content:   []*aiv1.ContentBlock{block},
					CreatedOn: timestamppb.Now(),
					UpdatedOn: timestamppb.Now(),
				})
				if err != nil {
					return nil, err
				}
			}

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

		// Add the assistant's response with tool calls to the conversation context
		assistantMessage := aiCompletionMessageToRuntimeMessage(res)
		allMessages = append(allMessages, assistantMessage)

		// Execute each tool call and add results as content blocks
		for _, toolCall := range toolCalls {
			logger.Debug("executing tool", zap.String("tool_name", toolCall.Name), zap.String("tool_id", toolCall.Id), observability.ZapCtx(ctx))

			// Convert protobuf Struct back to map for MCP tool execution
			inputMap := toolCall.Input.AsMap()

			// Call tool service
			resp, err := toolService.ExecuteTool(ctx, toolCall.Name, inputMap)

			var toolResult *aiv1.ToolResult
			if err != nil {
				// If context error, return the error
				if errors.Is(err, ctx.Err()) {
					logger.Debug("tool execution cancelled due to context error", zap.Error(err), zap.String("tool_name", toolCall.Name), zap.String("tool_id", toolCall.Id), observability.ZapCtx(ctx))
					return nil, err
				}

				// If not a context error, populate the tool result with the error message
				logger.Debug("tool execution failed", zap.Error(err), zap.String("tool_name", toolCall.Name), zap.String("tool_id", toolCall.Id), observability.ZapCtx(ctx))
				toolResult = &aiv1.ToolResult{
					Id:      toolCall.Id,
					Content: fmt.Sprintf("Error executing tool %s: %v", toolCall.Name, err),
					IsError: true,
				}
			} else {
				logger.Debug("tool executed successfully", zap.String("tool_name", toolCall.Name), zap.String("tool_id", toolCall.Id), observability.ZapCtx(ctx))
				// Convert response to ToolResult
				toolResult = &aiv1.ToolResult{
					Id:      toolCall.Id,
					Content: fmt.Sprintf("%v", resp),
					IsError: false,
				}
			}

			// Add tool result as content block
			block := &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_ToolResult{
					ToolResult: toolResult,
				},
			}
			contentBlocks = append(contentBlocks, block)

			// Emit preliminary message for streaming
			if onMessage != nil {
				err := onMessage(conversationID, &runtimev1.Message{
					Id:        uuid.NewString(),
					Role:      "assistant",
					Content:   []*aiv1.ContentBlock{block},
					CreatedOn: timestamppb.Now(),
					UpdatedOn: timestamppb.Now(),
				})
				if err != nil {
					return nil, err
				}
			}
		}

		// Add tool results to conversation context
		var toolResultBlocks []*aiv1.ContentBlock
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
		userMessage := &runtimev1.Message{
			Role:    "user",
			Content: toolResultBlocks,
		}
		allMessages = append(allMessages, userMessage)

		// Continue the loop to get the next response
	}

	// If we reach here, we've exceeded the maximum tool call iterations
	logger.Warn("maximum tool call iterations reached, requesting final response",
		zap.Int("max_iterations", maxToolCallIterations),
		observability.ZapCtx(ctx))

	// Instead of erroring, inform the AI and get a final response
	limitMessage := &runtimev1.Message{
		Role: "user",
		Content: []*aiv1.ContentBlock{
			{
				BlockType: &aiv1.ContentBlock_Text{
					Text: fmt.Sprintf("Tool call limit reached (%d iterations). Please provide a final response without additional tool calls.", maxToolCallIterations),
				},
			},
		},
	}
	allMessages = append(allMessages, limitMessage)

	// Truncate conversation if needed before final call
	messages := maybeTruncateConversation(allMessages)

	// Convert runtime messages to aiv1.CompletionMessage for final AI service call
	completionMessages := make([]*aiv1.CompletionMessage, len(messages))
	for i, msg := range messages {
		completionMessages[i] = runtimeMessageToAICompletionMessage(msg)
	}

	// Get final response from AI without tools
	res, err := ai.Complete(ctx, completionMessages, []*aiv1.Tool{}) // No tools provided
	if err != nil {
		return nil, err
	}

	logger.Info("received final response after tool call limit", observability.ZapCtx(ctx))

	// Add the final response content blocks
	contentBlocks = append(contentBlocks, res.Content...)

	// Emit preliminary message for streaming
	if onMessage != nil {
		for _, block := range res.Content {
			err := onMessage(conversationID, &runtimev1.Message{
				Id:        uuid.NewString(),
				Role:      "assistant",
				Content:   []*aiv1.ContentBlock{block},
				CreatedOn: timestamppb.Now(),
				UpdatedOn: timestamppb.Now(),
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return contentBlocks, nil
}

// buildCompletionResult saves the assistant message and builds the final result
func (r *Runtime) buildCompletionResult(ctx context.Context, instanceID, conversationID string, contentBlocks []*aiv1.ContentBlock, addedMessageIDs []string) (*CompleteWithToolsResult, error) {
	// Save the complete assistant message with all content blocks
	messageID, err := r.addMessage(ctx, instanceID, conversationID, "assistant", contentBlocks)
	if err != nil {
		return nil, err
	}
	addedMessageIDs = append(addedMessageIDs, messageID)

	// Get the messages to retrieve all saved messages with timestamps
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	catalogMessages, err := catalog.FindMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Convert catalog messages to protobuf messages
	messages := make([]*runtimev1.Message, len(catalogMessages))
	for i, msg := range catalogMessages {
		pbMessage, err := MessageToPB(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert catalog message to protobuf: %w", err)
		}
		messages[i] = pbMessage
	}

	// Collect all messages that were added during this Complete call, excluding system messages
	var addedMessages []*runtimev1.Message
	for _, messageID := range addedMessageIDs {
		for _, msg := range messages {
			if msg.Id == messageID {
				// Filter out system messages
				if msg.Role != "system" {
					addedMessages = append(addedMessages, msg)
				}
				break
			}
		}
	}

	// Return final result with all messages added during this Complete call (excluding system messages)
	return &CompleteWithToolsResult{
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
				if block.GetText() == "" {
					continue
				}
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

	// Fallback if no user message found
	return "New Conversation"
}

// createConversation creates a new conversation with the given title
func (r *Runtime) createConversation(ctx context.Context, instanceID, ownerID, title string, appContext *runtimev1.AppContext) (*runtimev1.Conversation, error) {
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	var contextType string
	var contextMetadataJSON string

	if appContext != nil {
		contextType = appContext.ContextType.String()
		if appContext.ContextMetadata != nil {
			metadataBytes, err := json.Marshal(appContext.ContextMetadata.AsMap())
			if err != nil {
				return nil, fmt.Errorf("failed to marshal context metadata: %w", err)
			}
			contextMetadataJSON = string(metadataBytes)
		}
	}

	conversationID, err := catalog.InsertConversation(ctx, ownerID, title, contextType, contextMetadataJSON)
	if err != nil {
		return nil, err
	}

	// Return the created conversation
	catalogConversation, err := catalog.FindConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	return ConversationToPB(catalogConversation), nil
}

// addMessage adds a message to a conversation
func (r *Runtime) addMessage(ctx context.Context, instanceID, conversationID, role string, content []*aiv1.ContentBlock) (string, error) {
	catalog, release, err := r.Catalog(ctx, instanceID)
	if err != nil {
		return "", err
	}
	defer release()

	// Convert protobuf ContentBlocks to catalog MessageContent
	catalogContent, err := messageContentFromPBSlice(content)
	if err != nil {
		return "", err
	}

	return catalog.InsertMessage(ctx, conversationID, role, catalogContent)
}

// buildProjectChatSystemPrompt constructs the system prompt for the project chat context
func buildProjectChatSystemPrompt(aiInstructions string) string {
	// TODO: call 'list_metrics_views' and seed the result in the system prompt
	basePrompt := `<role>
You are a data analysis agent specialized in uncovering actionable business insights. You systematically explore data using available metrics tools, then apply analytical rigor to find surprising patterns and unexpected relationships that influence decision-making.
</role>

<communication_style>
- Be confident, clear, and intellectually curious
- Write conversationally using "I" and "you" - speak directly to the user
- Present insights with authority while remaining enthusiastic and collaborative
</communication_style>

<process>
**Phase 1: Data Discovery (Deterministic)**
Follow these steps in order:
1. **Discover**: Use "list_metrics_views" to identify available datasets
2. **Understand**: Use "get_metrics_view" to understand measures and dimensions for the selected view  
3. **Scope**: Use "query_metrics_view_time_range" to determine the span of available data

**Phase 2: Analysis (Agentic OODA Loop)**
4. **Analyze**: Use "query_metrics_view" in an iterative OODA loop:
   - **Observe**: What data patterns emerge? What insights are surfacing? What gaps remain?
   - **Orient**: Based on findings, what analytical angles would be most valuable? How do current insights shape next queries?
   - **Decide**: Choose specific dimensions, filters, time periods, or comparisons to explore
   - **Act**: Execute the query and evaluate results in <thinking> tags

Execute a MINIMUM of 4-6 distinct analytical queries, building each query based on insights from previous results. Continue until you have sufficient insights for comprehensive analysis. Some analyses may require up to 20 queries.
</process>

<analysis_guidelines>
**Setup Phase (Steps 1-3)**: 
- Complete each step fully before proceeding
- Explain your approach briefly before starting
- If any step fails, investigate and adapt

**Analysis Phase (Step 4)**:
- Start broad (overall patterns), then drill into specific segments
- Always include time-based analysis using comparison features (delta_abs, delta_rel)
- Focus on insights that are surprising, actionable, and quantified
- Never repeat identical queries - each should explore new analytical angles
- Use <thinking> tags between queries to evaluate results and plan next steps

**Quality Standards**:
- Prioritize findings that contradict expectations or reveal hidden patterns
- Quantify changes and impacts with specific numbers
- Link insights to business implications and decisions

**Data Accuracy Requirements**:
- ALL numbers and calculations must come from "query_metrics_view" tool results
- NEVER perform manual calculations or mathematical operations
- If a desired calculation cannot be achieved through the metrics tools, explicitly state this limitation
- Use only the exact numbers returned by the tools in your analysis
</analysis_guidelines>

<guardrails>
- If the user asks an unrelated question (e.g., general knowledge, politics, entertainment, trivia):
  1. Politely decline to answer the unrelated question
  2. Briefly explain that you specialize only in business data analysis
  3. Redirect the conversation back to datasets, metrics, or insights
  4. Do **NOT** attempt to answer the unrelated question

Example response style:  
*"I focus specifically on analyzing your business data and uncovering insights — so I won’t be the right source for general knowledge questions. Let’s pivot back to your project: would you like me to explore the available datasets so we can start surfacing insights?"*
</guardrails>

<thinking>
After each query in Phase 2, think through:
- What patterns or anomalies did this reveal?
- How does this connect to previous findings?
- What new questions does this raise?
- What's the most valuable next query to run?
- Are there any surprising insights worth highlighting?
</thinking>

<output_format>
Format your analysis as follows:
` + "```markdown" + `
[Brief acknowledgment and explanation of approach]

Based on my systematic analysis, here are the key insights:

1. ## [Headline with specific impact/number]
   [Finding with business context and implications]

2. ## [Headline with specific impact/number]  
   [Finding with business context and implications]

3. ## [Headline with specific impact/number]
   [Finding with business context and implications]

[Offer specific follow-up analysis options]
` + "```" + `
</output_format>`

	if aiInstructions != "" {
		return basePrompt + "\n\n## Additional Instructions (provided by the Rill project developer)\n" + aiInstructions
	}

	return basePrompt
}

// buildExploreDashboardSystemPrompt constructs the system prompt for explore dashboard context
func buildExploreDashboardSystemPrompt(dashboardName, metricsViewName string, metricsViewSpec, timeRangeSummary any, aiInstructions string) string {
	var prompt strings.Builder

	// 1. WHO: Establish the AI's role and purpose
	prompt.WriteString("You are a data analyst designed to be helpful, insightful, and accurate. ")
	prompt.WriteString("Your role is to assist users in understanding their data by answering questions, identifying trends, performing calculations, and providing actionable insights.\n\n")

	// 2. WHERE: Set the current context and location
	prompt.WriteString("## Current Context\n")
	prompt.WriteString(fmt.Sprintf("The user is actively viewing the %q explore dashboard. ", dashboardName))
	prompt.WriteString("When they refer to \"this dashboard,\" \"the current view,\" or similar contextual references, they are referring to this dashboard.\n\n")

	// 3. WHAT: Describe the available data and constraints
	prompt.WriteString("## Available Data\n")
	prompt.WriteString(fmt.Sprintf("This dashboard is based on the %q metrics view.\n\n", metricsViewName))

	prompt.WriteString("### Metrics View Details:\n")
	prompt.WriteString(fmt.Sprintf("%v\n\n", metricsViewSpec))
	prompt.WriteString("*Note: If the metrics view spec above contains an \"ai_instructions\" field, follow those instructions for all queries related to this data.*\n\n")

	prompt.WriteString("### Time Range Information:\n")
	prompt.WriteString(fmt.Sprintf("%v\n\n", timeRangeSummary))

	// 4. HOW: Describe the AI's capabilities
	prompt.WriteString("## Your Capabilities\n")
	prompt.WriteString("You can use \"query_metrics_view\" to run queries and get aggregated results from this metrics view. ")
	prompt.WriteString("The metrics view spec above shows all available dimensions and measures, and the time range information shows what time periods are available for analysis. ")
	prompt.WriteString("Use this information to craft meaningful queries that answer the user's questions and provide valuable insights.\n\n")

	prompt.WriteString(fmt.Sprintf("**IMPORTANT: Every invocation of the \"query_metrics_view\" tool must include \"metrics_view\": %q in the payload.**\n\n", metricsViewName))

	// 5. CUSTOMIZE: Provide user-specific instructions
	if aiInstructions != "" {
		prompt.WriteString("## Additional Instructions (provided by the Rill project developer)\n")
		prompt.WriteString(aiInstructions)
	}

	return prompt.String()
}

// maybeTruncateConversation keeps recent messages and a few early ones for context.
// It's a simple placeholder strategy. In the future, we'll enhance this with AI summarization.
func maybeTruncateConversation(messages []*runtimev1.Message) []*runtimev1.Message {
	const (
		maxMessages = 20 // Keep up to 20 messages total
		keepFirst   = 3  // Always keep first 3 messages for context
		keepLast    = 16 // Keep last 16 messages
	)

	if len(messages) <= maxMessages {
		return messages
	}

	var result []*runtimev1.Message

	// Keep first messages
	result = append(result, messages[:keepFirst]...)

	// Add truncation indicator
	skipped := len(messages) - keepFirst - keepLast
	result = append(result, &runtimev1.Message{
		Role: "system",
		Content: []*aiv1.ContentBlock{
			{
				BlockType: &aiv1.ContentBlock_Text{
					Text: fmt.Sprintf("... [%d messages omitted for brevity] ...", skipped),
				},
			},
		},
	})

	// Keep last messages
	start := len(messages) - keepLast
	result = append(result, messages[start:]...)

	return result
}

// Helper conversion functions - these handle conversions between catalog and protobuf types
// Exported functions are used by the server layer, internal ones are used only in this file

// ConversationToPB converts a drivers.Conversation to a runtimev1.Conversation.
func ConversationToPB(conv *drivers.Conversation) *runtimev1.Conversation {
	return &runtimev1.Conversation{
		Id:        conv.ID,
		OwnerId:   conv.OwnerID,
		Title:     conv.Title,
		CreatedOn: timestamppb.New(conv.CreatedOn),
		UpdatedOn: timestamppb.New(conv.UpdatedOn),
	}
}

// MessageToPB converts a drivers.Message to a runtimev1.Message.
func MessageToPB(msg *drivers.Message) (*runtimev1.Message, error) {
	content, err := msg.GetContent()
	if err != nil {
		return nil, err
	}

	contentBlocks := make([]*aiv1.ContentBlock, len(content))
	for i, c := range content {
		block, err := messageContentToPB(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message content to protobuf: %w", err)
		}
		contentBlocks[i] = block
	}

	return &runtimev1.Message{
		Id:        msg.ID,
		Role:      msg.Role,
		Content:   contentBlocks,
		CreatedOn: timestamppb.New(msg.CreatedOn),
		UpdatedOn: timestamppb.New(msg.UpdatedOn),
	}, nil
}

// messageContentToPB converts drivers.MessageContent to aiv1.ContentBlock.
func messageContentToPB(mc drivers.MessageContent) (*aiv1.ContentBlock, error) {
	switch mc.Type {
	case "text":
		return &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_Text{
				Text: mc.Text,
			},
		}, nil
	case "tool_call":
		// Convert map to protobuf Struct
		input, err := structpb.NewStruct(mc.ToolCallInput)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tool call input to struct: %w", err)
		}
		return &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_ToolCall{
				ToolCall: &aiv1.ToolCall{
					Id:    mc.ToolCallID,
					Name:  mc.ToolCallName,
					Input: input,
				},
			},
		}, nil
	case "tool_result":
		return &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_ToolResult{
				ToolResult: &aiv1.ToolResult{
					Id:      mc.ToolResultID,
					Content: mc.ToolResultContent,
					IsError: mc.ToolResultIsError,
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown message content type: %s", mc.Type)
	}
}

// messageContentFromPB converts aiv1.ContentBlock to drivers.MessageContent.
func messageContentFromPB(block *aiv1.ContentBlock) (drivers.MessageContent, error) {
	if text := block.GetText(); text != "" {
		return drivers.MessageContent{
			Type: "text",
			Text: text,
		}, nil
	} else if toolCall := block.GetToolCall(); toolCall != nil {
		return drivers.MessageContent{
			Type:          "tool_call",
			ToolCallID:    toolCall.Id,
			ToolCallName:  toolCall.Name,
			ToolCallInput: toolCall.Input.AsMap(),
		}, nil
	} else if toolResult := block.GetToolResult(); toolResult != nil {
		return drivers.MessageContent{
			Type:              "tool_result",
			ToolResultID:      toolResult.Id,
			ToolResultContent: toolResult.Content,
			ToolResultIsError: toolResult.IsError,
		}, nil
	}

	// Fallback
	return drivers.MessageContent{}, fmt.Errorf("unknown message content type: %s", block.BlockType)
}

// messageContentFromPBSlice converts a slice of aiv1.ContentBlock to a slice of drivers.MessageContent.
func messageContentFromPBSlice(blocks []*aiv1.ContentBlock) ([]drivers.MessageContent, error) {
	content := make([]drivers.MessageContent, len(blocks))
	var err error
	for i, block := range blocks {
		content[i], err = messageContentFromPB(block)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message content to protobuf: %w", err)
		}
	}
	return content, nil
}

// runtimeMessageToAICompletionMessage converts a runtimev1.Message to aiv1.CompletionMessage
func runtimeMessageToAICompletionMessage(msg *runtimev1.Message) *aiv1.CompletionMessage {
	return &aiv1.CompletionMessage{
		Role:    msg.Role,
		Content: msg.Content, // Both use []*aiv1.ContentBlock
	}
}

// aiCompletionMessageToRuntimeMessage converts a aiv1.CompletionMessage to a runtimev1.Message
func aiCompletionMessageToRuntimeMessage(completionMessage *aiv1.CompletionMessage) *runtimev1.Message {
	return &runtimev1.Message{
		Role:    completionMessage.Role,
		Content: completionMessage.Content, // Both use []*aiv1.ContentBlock
	}
}
