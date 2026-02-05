package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	openaidriver "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	drivers.Register("openai", driver{})
	drivers.RegisterAsConnector("openai", driver{})
}

var spec = drivers.Spec{
	DisplayName: "OpenAI",
	Description: "Connect to OpenAI's API for language models.",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "api_key",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "API Key",
			Description: "API key for connecting to OpenAI.",
			Secret:      true,
		},
		{
			Key:         "model",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Model",
			Description: "The OpenAI model to use (e.g., 'gpt-4o').",
			Placeholder: "",
		},
		{
			Key:         "base_url",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Base URL",
			Description: "The base URL for the OpenAI API (e.g., 'https://api.openai.com/v1').",
			Placeholder: "",
		},
		{
			Key:         "api_type",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "API Type",
			Description: "The type of OpenAI API to use (e.g., 'OPEN_AI, AZURE').",
			Placeholder: "",
		},
		{
			Key:         "api_version",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "API Version",
			Description: "The version of the OpenAI API to use (e.g., '2023-05-15'). Required when APIType is APITypeAzure or APITypeAzureAD",
			Placeholder: "",
		},
	},
	ImplementsAI: true,
}

type driver struct{}

var _ drivers.Driver = driver{}

// HasAnonymousSourceAccess implements drivers.Driver.
func (d driver) HasAnonymousSourceAccess(ctx context.Context, srcProps map[string]any, logger *zap.Logger) (bool, error) {
	return false, drivers.ErrNotImplemented
}

// Open implements drivers.Driver.
func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	client, err := newOpenAIClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	return &openai{
		client: client,
		config: conf,
	}, nil
}

// Spec implements drivers.Driver.
func (d driver) Spec() drivers.Spec {
	return spec
}

// TertiarySourceConnectors implements drivers.Driver.
func (d driver) TertiarySourceConnectors(ctx context.Context, srcProps map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, drivers.ErrNotImplemented
}

type configProperties struct {
	APIKey string `mapstructure:"api_key"`

	Model       string  `mapstructure:"model"`
	Temperature float32 `mapstructure:"temperature"`

	BaseURL    string `mapstructure:"base_url"`
	APIType    string `mapstructure:"api_type"`
	APIVersion string `mapstructure:"api_version"`
}

func (c *configProperties) getModel() string {
	if c.Model != "" {
		return c.Model
	}
	return "gpt-5.2" // openai.GPT5
}

func (c *configProperties) getTemperature() float32 {
	if c.Temperature > 0 {
		return c.Temperature
	}
	return 1 // Default temperature if not specified
}

type openai struct {
	client *openaidriver.Client
	config *configProperties
}

var _ drivers.AIService = (*openai)(nil)

// newOpenAIClient creates a new OpenAI client based on configuration.
func newOpenAIClient(conf *configProperties) (*openaidriver.Client, error) {
	if conf.APIKey == "" {
		return nil, errors.New("API key is required")
	}

	var clientConfig openaidriver.ClientConfig
	apiType := openaidriver.APIType(conf.APIType)
	if apiType == openaidriver.APITypeAzure || apiType == openaidriver.APITypeAzureAD {
		clientConfig = openaidriver.DefaultAzureConfig(conf.APIKey, conf.BaseURL)
	} else {
		clientConfig = openaidriver.DefaultConfig(conf.APIKey)
	}

	if conf.BaseURL != "" {
		clientConfig.BaseURL = conf.BaseURL
	}
	if conf.APIVersion != "" {
		clientConfig.APIVersion = conf.APIVersion
	}
	if conf.APIType != "" {
		clientConfig.APIType = apiType
	}

	return openaidriver.NewClientWithConfig(clientConfig), nil
}

// AsAI implements drivers.Handle.
func (o *openai) AsAI(instanceID string) (drivers.AIService, bool) {
	return o, true
}

// AsAdmin implements drivers.Handle.
func (o *openai) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (o *openai) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (o *openai) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsInformationSchema implements drivers.Handle.
func (o *openai) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (o *openai) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (o *openai) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// AsNotifier implements drivers.Handle.
func (o *openai) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// AsOLAP implements drivers.Handle.
func (o *openai) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (o *openai) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsRegistry implements drivers.Handle.
func (o *openai) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (o *openai) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (o *openai) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// Close implements drivers.Handle.
func (o *openai) Close() error {
	return nil
}

// Config implements drivers.Handle.
func (o *openai) Config() map[string]any {
	var configMap map[string]any
	_ = mapstructure.Decode(o.config, &configMap)
	return configMap
}

// Driver implements drivers.Handle.
func (o *openai) Driver() string {
	return "openai"
}

// Migrate implements drivers.Handle.
func (o *openai) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (o *openai) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Ping implements drivers.Handle.
func (o *openai) Ping(ctx context.Context) error {
	return nil
}

// Complete implements drivers.AIService.
// It sends a chat completion request to OpenAI and returns the response.
func (o *openai) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	// Convert Rill messages to OpenAI's message format
	var reqMsgs []openaidriver.ChatCompletionMessage
	for _, msg := range opts.Messages {
		openaiMsgs, err := convertRillMessageToOpenAIMessages(msg) // each Rill message may become multiple OpenAI messages
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}
		reqMsgs = append(reqMsgs, openaiMsgs...)
	}

	// Convert Rill tools to OpenAI's tool format
	var openaiTools []openaidriver.Tool
	if len(opts.Tools) > 0 {
		openaiTools = make([]openaidriver.Tool, len(opts.Tools))
		for i, tool := range opts.Tools {
			openaiTool, err := convertRillToolToOpenAITool(tool)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool: %w", err)
			}
			openaiTools[i] = openaiTool
		}
	}

	// Determine response format based on output schema
	var responseFormat *openaidriver.ChatCompletionResponseFormat
	if opts.OutputSchema != nil {
		responseFormat = &openaidriver.ChatCompletionResponseFormat{
			Type: openaidriver.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openaidriver.ChatCompletionResponseFormatJSONSchema{
				Name:   "llm_completion_result",
				Schema: opts.OutputSchema,
			},
		}
	}

	// Prepare request parameters
	params := openaidriver.ChatCompletionRequest{
		Model:          o.config.getModel(),
		Messages:       reqMsgs,
		Temperature:    o.config.getTemperature(),
		ResponseFormat: responseFormat,
	}
	if len(openaiTools) > 0 {
		params.Tools = openaiTools
		params.ToolChoice = "auto"
	}

	// Send request to OpenAI
	res, err := o.client.CreateChatCompletion(ctx, params)
	if err != nil {
		return nil, err
	}

	// Return error if no choices are returned
	if len(res.Choices) == 0 {
		return nil, errors.New("no choices returned")
	}

	// Convert OpenAI's response to Rill's message format
	return &drivers.CompleteResult{
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
func convertRillMessageToOpenAIMessages(msg *aiv1.CompletionMessage) ([]openaidriver.ChatCompletionMessage, error) {
	var result []openaidriver.ChatCompletionMessage

	// Separate content into regular content vs tool results
	var regularContent string
	var toolCalls []openaidriver.ToolCall
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
		mainMsg := openaidriver.ChatCompletionMessage{
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
		toolMsg := openaidriver.ChatCompletionMessage{
			Role:       openaidriver.ChatMessageRoleTool,
			Content:    toolResult.Content,
			ToolCallID: toolResult.Id, // Links back to the original tool call
		}
		result = append(result, toolMsg)
	}

	return result, nil
}

// convertOpenAIMessageToRillMessage converts OpenAI ChatCompletionMessage to Rill CompletionMessage format.
func convertOpenAIMessageToRillMessage(message openaidriver.ChatCompletionMessage) *aiv1.CompletionMessage {
	// Handle standard text responses (simple case)
	if len(message.ToolCalls) == 0 {
		contentBlocks := []*aiv1.ContentBlock{
			{
				BlockType: &aiv1.ContentBlock_Text{Text: message.Content},
			},
		}
		return &aiv1.CompletionMessage{
			Role:    openaidriver.ChatMessageRoleAssistant,
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
		Role:    openaidriver.ChatMessageRoleAssistant,
		Content: contentBlocks,
	}
}

// convertRillToolToOpenAITool converts a single Rill Tool to OpenAI Tool format.
func convertRillToolToOpenAITool(tool *aiv1.Tool) (openaidriver.Tool, error) {
	schemaMap, err := parseToolSchema(tool.InputSchema)
	if err != nil {
		return openaidriver.Tool{}, fmt.Errorf("failed to convert tool %s: %w", tool.Name, err)
	}

	return openaidriver.Tool{
		Type: openaidriver.ToolTypeFunction,
		Function: &openaidriver.FunctionDefinition{
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
func convertRillToolCallToOpenAIToolCall(toolCall *aiv1.ToolCall) (openaidriver.ToolCall, error) {
	arguments, err := marshalToolCallInput(toolCall.Input)
	if err != nil {
		return openaidriver.ToolCall{}, fmt.Errorf("failed to marshal input for tool %s: %w", toolCall.Name, err)
	}

	return openaidriver.ToolCall{
		ID:   toolCall.Id,
		Type: openaidriver.ToolTypeFunction,
		Function: openaidriver.FunctionCall{
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
func convertOpenAIToolCallToRillToolCall(toolCall openaidriver.ToolCall) (*aiv1.ToolCall, error) {
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
