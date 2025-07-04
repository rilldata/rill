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
}

var _ Client = (*openAI)(nil)

func NewOpenAI(apiKey string) (Client, error) {
	c := openai.NewClient(apiKey)

	return &openAI{
		client: c,
		apiKey: apiKey,
	}, nil
}

// Complete sends a chat completion request to OpenAI and returns the response.
// It handles conversion between Rill's message format and OpenAI's message format.
func (c *openAI) Complete(ctx context.Context, msgs []*aiv1.CompletionMessage, tools []*aiv1.Tool) (*aiv1.CompletionMessage, error) {
	// Convert input to OpenAI format
	reqMsgs := make([]openai.ChatCompletionMessage, len(msgs))
	for i, msg := range msgs {
		reqMsgs[i] = convertRillMessageToOpenAIMessage(msg)
	}

	var openaiTools []openai.Tool
	if len(tools) > 0 {
		openaiTools = make([]openai.Tool, len(tools))
		for i, tool := range tools {
			openaiTool, err := convertRillToolToOpenAITool(tool)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool: %w", err)
			}
			openaiTools[i] = openaiTool
		}
	}

	// Prepare request parameters
	params := openai.ChatCompletionRequest{
		Model:       openai.GPT4o,
		Messages:    reqMsgs,
		Temperature: 0.2,
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

	// Convert OpenAI response to Rill format
	return convertOpenAIMessageToRillMessage(res.Choices[0].Message)
}

// convertRillMessageToOpenAIMessage converts a single Rill CompletionMessage to OpenAI ChatCompletionMessage format.
func convertRillMessageToOpenAIMessage(msg *aiv1.CompletionMessage) openai.ChatCompletionMessage {
	var content string

	// Process each content block in the message
	for _, block := range msg.Content {
		if text := block.GetText(); text != "" {
			content += text
		} else if toolCall := block.GetToolCall(); toolCall != nil {
			// Convert tool calls to JSON format for OpenAI
			toolCallJSON, err := json.Marshal(map[string]interface{}{
				"type":  "tool_use",
				"id":    toolCall.Id,
				"name":  toolCall.Name,
				"input": toolCall.Input.AsMap(),
			})
			if err == nil {
				content += string(toolCallJSON)
			}
		} else if toolResult := block.GetToolResult(); toolResult != nil {
			// Add tool results directly as text content
			content += toolResult.Content
		}
	}

	return openai.ChatCompletionMessage{
		Role:    msg.Role,
		Content: content,
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

// convertOpenAIMessageToRillMessage converts OpenAI ChatCompletionMessage to Rill CompletionMessage format.
func convertOpenAIMessageToRillMessage(message openai.ChatCompletionMessage) (*aiv1.CompletionMessage, error) {
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
		}, nil
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
	}, nil
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
