package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/types/known/structpb"
)

type openAI struct {
	client *openai.Client
	apiKey string

	opts *Options
}

var _ Client = (*openAI)(nil)

type Options struct {
	BaseURL    string
	APIType    openai.APIType
	APIVersion string

	Model       string
	Temperature float32
}

func (o *Options) getModel() string {
	if o.Model != "" {
		return o.Model
	}
	return openai.GPT4Dot1 // Default model if not specified
}

func (o *Options) getTemperature() float32 {
	if o.Temperature > 0 {
		return o.Temperature
	}
	return 0.2 // Default temperature if not specified
}

func NewOpenAI(apiKey string, opts *Options) (Client, error) {
	if opts == nil {
		return &openAI{
			client: openai.NewClient(apiKey),
			apiKey: apiKey,
			opts:   &Options{},
		}, nil
	}

	var def openai.ClientConfig
	if opts.APIType == openai.APITypeAzure || opts.APIType == openai.APITypeAzureAD {
		def = openai.DefaultAzureConfig(apiKey, opts.BaseURL)
	} else {
		def = openai.DefaultConfig(apiKey)
	}
	if opts.BaseURL != "" {
		def.BaseURL = opts.BaseURL
	}
	if opts.APIVersion != "" {
		def.APIVersion = opts.APIVersion
	}
	if opts.APIType != "" {
		def.APIType = opts.APIType
	}
	c := openai.NewClientWithConfig(def)

	return &openAI{
		client: c,
		apiKey: apiKey,
		opts:   opts,
	}, nil
}

// Complete sends a chat completion request to OpenAI and returns the response.
// It handles conversion between Rill's message format and OpenAI's message format.
func (c *openAI) Complete(ctx context.Context, opts *CompleteOptions) (*CompleteResult, error) {
	// Convert Rill messages to OpenAI's message format
	var reqMsgs []openai.ChatCompletionMessage
	for _, msg := range opts.Messages {
		openaiMsgs, err := convertRillMessageToOpenAIMessages(msg) // each Rill message may become multiple OpenAI messages
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}
		reqMsgs = append(reqMsgs, openaiMsgs...)
	}

	// Convert Rill tools to OpenAI's tool format
	var openaiTools []openai.Tool
	if len(opts.Tools) > 0 {
		openaiTools = make([]openai.Tool, len(opts.Tools))
		for i, tool := range opts.Tools {
			openaiTool, err := convertRillToolToOpenAITool(tool)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool: %w", err)
			}
			openaiTools[i] = openaiTool
		}
	}

	// Determine response format based on output schema
	var responseFormat *openai.ChatCompletionResponseFormat
	if opts.OutputSchema != nil {
		responseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "llm_completion_result",
				Schema: opts.OutputSchema,
			},
		}
	}

	// Prepare request parameters
	params := openai.ChatCompletionRequest{
		Model:          c.opts.getModel(),
		Messages:       reqMsgs,
		Temperature:    c.opts.getTemperature(),
		ResponseFormat: responseFormat,
	}
	if len(openaiTools) > 0 {
		params.Tools = openaiTools
		params.ToolChoice = "auto"
	}

	// Send request to OpenAI
	res, err := c.client.CreateChatCompletion(ctx, params)
	if err != nil {
		return nil, err
	}

	// Return error if no choices are returned
	if len(res.Choices) == 0 {
		return nil, errors.New("no choices returned")
	}

	// Convert OpenAI's response to Rill's message format
	return &CompleteResult{
		Message:      convertOpenAIMessageToRillMessage(res.Choices[0].Message),
		InputTokens:  res.Usage.PromptTokens,
		OutputTokens: res.Usage.CompletionTokens,
	}, nil
}

// convertRillMessageToOpenAIMessages converts a single Rill CompletionMessage to one or more OpenAI ChatCompletionMessages.
//
// This handles the asymmetric nature of OpenAI's tool calling pattern:
// - Tool calls: Multiple calls are grouped in ONE assistant message (how OpenAI sends them)
// - Tool results: Each result becomes a SEPARATE message with role="tool" (how OpenAI expects responses)
//
// Note: In practice, Rill messages have at most 1 text block (from OpenAI's single Content field),
// so we don't need to worry about concatenating multiple text blocks.
func convertRillMessageToOpenAIMessages(msg *aiv1.CompletionMessage) ([]openai.ChatCompletionMessage, error) {
	var result []openai.ChatCompletionMessage

	// Separate content into regular content vs tool results
	var regularContent string
	var toolCalls []openai.ToolCall
	var toolResults []*aiv1.ToolResult

	for _, block := range msg.Content {
		switch blockType := block.BlockType.(type) {
		case *aiv1.ContentBlock_Text:
			regularContent += blockType.Text

		case *aiv1.ContentBlock_ToolCall:
			openaiToolCall, err := convertRillToolCallToOpenAIToolCall(blockType.ToolCall)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool call: %w", err)
			}
			toolCalls = append(toolCalls, openaiToolCall)

		case *aiv1.ContentBlock_ToolResult:
			toolResults = append(toolResults, blockType.ToolResult)
		}
	}

	// Create main message for text content and tool calls
	// This preserves OpenAI's original structure: text + multiple tool calls in one message
	if regularContent != "" || len(toolCalls) > 0 {
		mainMsg := openai.ChatCompletionMessage{
			Role:      msg.Role,
			Content:   regularContent,
			ToolCalls: toolCalls, // Multiple tool calls grouped together (mirrors original OpenAI response)
		}
		result = append(result, mainMsg)
	}

	// Create separate messages for each tool result
	// This follows OpenAI's expected pattern: each tool result = separate message with role="tool"
	// The tool_call_id links each result back to the specific tool call that generated it
	for _, toolResult := range toolResults {
		toolMsg := openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    toolResult.Content,
			ToolCallID: toolResult.Id, // Links back to the original tool call
		}
		result = append(result, toolMsg)
	}

	return result, nil
}

// convertOpenAIMessageToRillMessage converts OpenAI ChatCompletionMessage to Rill CompletionMessage format.
func convertOpenAIMessageToRillMessage(message openai.ChatCompletionMessage) *aiv1.CompletionMessage {
	// Handle standard text responses (simple case)
	if len(message.ToolCalls) == 0 {
		contentBlocks := []*aiv1.ContentBlock{
			{
				BlockType: &aiv1.ContentBlock_Text{Text: message.Content},
			},
		}
		return &aiv1.CompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: contentBlocks,
		}
	}

	// Handle responses with tool calls (complex case)
	var contentBlocks []*aiv1.ContentBlock

	// Include any text content alongside tool calls
	if message.Content != "" {
		contentBlocks = append(contentBlocks, &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_Text{Text: message.Content},
		})
	}

	// Convert each tool call to a content block
	for _, toolCall := range message.ToolCalls {
		toolCallProto, err := convertOpenAIToolCallToRillToolCall(toolCall)
		if err != nil {
			// Log the error but continue processing other tool calls
			// This prevents one malformed tool call from breaking the entire response
			continue
		}

		contentBlocks = append(contentBlocks, &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_ToolCall{
				ToolCall: toolCallProto,
			},
		})
	}

	return &aiv1.CompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: contentBlocks,
	}
}

// convertRillToolToOpenAITool converts a single Rill Tool to OpenAI Tool format.
func convertRillToolToOpenAITool(tool *aiv1.Tool) (openai.Tool, error) {
	schemaMap, err := parseToolSchema(tool.InputSchema)
	if err != nil {
		return openai.Tool{}, fmt.Errorf("failed to convert tool %s: %w", tool.Name, err)
	}

	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  schemaMap,
		},
	}, nil
}

// parseToolSchema parses a JSON schema string and returns a map, with fallback to default schema.
func parseToolSchema(schemaJSON string) (map[string]interface{}, error) {
	if schemaJSON == "" {
		// Default schema when none is provided
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}, nil
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schemaMap); err != nil {
		return nil, fmt.Errorf("failed to parse tool schema JSON: %w", err)
	}

	return schemaMap, nil
}

// convertRillToolCallToOpenAIToolCall converts a Rill ToolCall to OpenAI ToolCall format
func convertRillToolCallToOpenAIToolCall(toolCall *aiv1.ToolCall) (openai.ToolCall, error) {
	arguments, err := marshalToolCallInput(toolCall.Input)
	if err != nil {
		return openai.ToolCall{}, fmt.Errorf("failed to marshal input for tool %s: %w", toolCall.Name, err)
	}

	return openai.ToolCall{
		ID:   toolCall.Id,
		Type: openai.ToolTypeFunction,
		Function: openai.FunctionCall{
			Name:      toolCall.Name,
			Arguments: arguments,
		},
	}, nil
}

// marshalToolCallInput converts tool call input to JSON string for OpenAI API
func marshalToolCallInput(input *structpb.Struct) (string, error) {
	if input == nil {
		return "{}", nil
	}

	inputJSON, err := json.Marshal(input.AsMap())
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool call input: %w", err)
	}
	return string(inputJSON), nil
}

// convertOpenAIToolCallToRillToolCall converts an OpenAI ToolCall to Rill ToolCall format.
func convertOpenAIToolCallToRillToolCall(toolCall openai.ToolCall) (*aiv1.ToolCall, error) {
	// Parse OpenAI ToolCall arguments
	var input map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool call arguments for %s: %w", toolCall.Function.Name, err)
	}

	// Convert input to protobuf Struct
	inputStruct, err := structpb.NewStruct(input)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tool call input to protobuf struct for %s: %w", toolCall.Function.Name, err)
	}

	// Create Rill ToolCall
	return &aiv1.ToolCall{
		Id:    toolCall.ID,
		Name:  toolCall.Function.Name,
		Input: inputStruct,
	}, nil
}
