package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/azure"
	"github.com/openai/openai-go/v3/option"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
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

	if conf.APIKey == "" {
		return nil, errors.New("API key is required")
	}

	var opts []option.RequestOption
	switch strings.ToLower(conf.APIType) {
	case "azure", "azure_ad": // azure_ad for backwards compatibility
		if conf.BaseURL == "" {
			return nil, errors.New("base_url is required for Azure clients")
		}
		apiVersion := conf.APIVersion
		if apiVersion == "" {
			apiVersion = "2024-06-01"
		}
		opts = append(opts,
			azure.WithEndpoint(conf.BaseURL, apiVersion),
			azure.WithAPIKey(conf.APIKey),
		)
	default:
		opts = append(opts, option.WithAPIKey(conf.APIKey))
		if conf.BaseURL != "" {
			opts = append(opts, option.WithBaseURL(conf.BaseURL))
		}
	}

	client := openai.NewClient(opts...)

	return &openaiHandle{
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
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	Temperature float32 `mapstructure:"temperature"`
	BaseURL     string  `mapstructure:"base_url"`
	APIType     string  `mapstructure:"api_type"`
	APIVersion  string  `mapstructure:"api_version"`
}

func (c *configProperties) getModel() string {
	if c.Model != "" {
		return c.Model
	}
	return openai.ChatModelGPT5_2
}

type openaiHandle struct {
	client openai.Client
	config *configProperties
}

var _ drivers.AIService = (*openaiHandle)(nil)

// AsAI implements drivers.Handle.
func (o *openaiHandle) AsAI(instanceID string) (drivers.AIService, bool) {
	return o, true
}

// AsAdmin implements drivers.Handle.
func (o *openaiHandle) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (o *openaiHandle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (o *openaiHandle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsInformationSchema implements drivers.Handle.
func (o *openaiHandle) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (o *openaiHandle) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (o *openaiHandle) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// AsNotifier implements drivers.Handle.
func (o *openaiHandle) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// AsOLAP implements drivers.Handle.
func (o *openaiHandle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (o *openaiHandle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsRegistry implements drivers.Handle.
func (o *openaiHandle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (o *openaiHandle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (o *openaiHandle) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// Close implements drivers.Handle.
func (o *openaiHandle) Close() error {
	return nil
}

// Config implements drivers.Handle.
func (o *openaiHandle) Config() map[string]any {
	var configMap map[string]any
	_ = mapstructure.Decode(o.config, &configMap)
	return configMap
}

// Driver implements drivers.Handle.
func (o *openaiHandle) Driver() string {
	return "openai"
}

// Migrate implements drivers.Handle.
func (o *openaiHandle) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (o *openaiHandle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Ping implements drivers.Handle.
func (o *openaiHandle) Ping(ctx context.Context) error {
	return nil
}

// Complete implements drivers.AIService.
func (o *openaiHandle) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	// Convert Rill messages to OpenAI's message format
	var reqMsgs []openai.ChatCompletionMessageParamUnion
	for _, msg := range opts.Messages {
		openaiMsgs, err := messageToOpenAI(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}
		reqMsgs = append(reqMsgs, openaiMsgs...)
	}

	// Convert Rill tools to OpenAI's tool format
	var openaiTools []openai.ChatCompletionToolUnionParam
	for _, tool := range opts.Tools {
		openaiTool, err := toolToOpenAI(tool)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tool: %w", err)
		}
		openaiTools = append(openaiTools, openaiTool)
	}

	// Prepare request parameters
	params := openai.ChatCompletionNewParams{
		Model:    o.config.getModel(),
		Messages: reqMsgs,
		Tools:    openaiTools,
	}
	if o.config.Temperature > 0 {
		params.Temperature = openai.Float(float64(o.config.Temperature))
	}

	// Set response format based on output schema
	if opts.OutputSchema != nil {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "llm_completion_result",
					Schema: opts.OutputSchema,
				},
			},
		}
	}

	// Send request to OpenAI
	res, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, err
	}

	// Return error if no choices are returned
	if len(res.Choices) == 0 {
		return nil, errors.New("no choices returned")
	}

	// Convert OpenAI's response to Rill's message format
	resMsgs, err := messageFromOpenAI(res.Choices[0].Message)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response message: %w", err)
	}
	result := &drivers.CompleteResult{
		Message:      resMsgs,
		InputTokens:  int(res.Usage.PromptTokens),
		OutputTokens: int(res.Usage.CompletionTokens),
	}
	return result, nil
}

// messageToOpenAI converts a single Rill CompletionMessage to one or more OpenAI ChatCompletionMessages.
//
// This handles the asymmetric nature of OpenAI's tool calling pattern:
// - Tool calls: Multiple calls are grouped in ONE assistant message (how OpenAI sends them)
// - Tool results: Each result becomes a SEPARATE message with role="tool" (how OpenAI expects responses)
//
// Note: In practice, Rill messages have at most 1 text block (from OpenAI's single Content field),
// so we don't need to worry about concatenating multiple text blocks.
func messageToOpenAI(msg *aiv1.CompletionMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	var result []openai.ChatCompletionMessageParamUnion
	var regularContent string
	var toolCalls []openai.ChatCompletionMessageToolCallUnionParam
	var toolResults []*aiv1.ToolResult

	for _, block := range msg.Content {
		switch blockType := block.BlockType.(type) {
		case *aiv1.ContentBlock_Text:
			regularContent += blockType.Text
		case *aiv1.ContentBlock_ToolCall:
			openaiToolCall, err := toolCallToOpenAI(blockType.ToolCall)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool call: %w", err)
			}
			toolCalls = append(toolCalls, openaiToolCall)
		case *aiv1.ContentBlock_ToolResult:
			toolResults = append(toolResults, blockType.ToolResult)
		}
	}

	// Create main message for text content and tool calls based on role
	if regularContent != "" || len(toolCalls) > 0 {
		switch msg.Role {
		case "user":
			result = append(result, openai.UserMessage(regularContent))
		case "system":
			result = append(result, openai.SystemMessage(regularContent))
		case "assistant":
			assistantMsg := openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: openai.String(regularContent),
				},
			}
			if len(toolCalls) > 0 {
				assistantMsg.ToolCalls = toolCalls
			}
			result = append(result, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &assistantMsg,
			})
		default:
			result = append(result, openai.UserMessage(regularContent))
		}
	}

	// Create separate messages for each tool result
	for _, toolResult := range toolResults {
		result = append(result, openai.ToolMessage(toolResult.Content, toolResult.Id))
	}

	return result, nil
}

func messageFromOpenAI(message openai.ChatCompletionMessage) (*aiv1.CompletionMessage, error) {
	var contentBlocks []*aiv1.ContentBlock

	if message.Content != "" {
		contentBlocks = append(contentBlocks, &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_Text{Text: message.Content},
		})
	}

	for i := range message.ToolCalls {
		toolCall := message.ToolCalls[i]

		// Parse tool call arguments: if malformed, include raw arguments as fallback
		var input map[string]any
		err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool call arguments for %s: %w", toolCall.Function.Name, err)
		}
		inputStruct, err := structpb.NewStruct(input)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tool call arguments to struct for %s: %w", toolCall.Function.Name, err)
		}

		contentBlocks = append(contentBlocks, &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_ToolCall{
				ToolCall: &aiv1.ToolCall{
					Id:    toolCall.ID,
					Name:  toolCall.Function.Name,
					Input: inputStruct,
				},
			},
		})
	}

	return &aiv1.CompletionMessage{
		Role:    "assistant",
		Content: contentBlocks,
	}, nil
}

func toolCallToOpenAI(toolCall *aiv1.ToolCall) (openai.ChatCompletionMessageToolCallUnionParam, error) {
	arguments := "{}"
	if toolCall.Input != nil {
		inputJSON, err := json.Marshal(toolCall.Input.AsMap())
		if err != nil {
			return openai.ChatCompletionMessageToolCallUnionParam{}, fmt.Errorf("failed to marshal input for tool %s: %w", toolCall.Name, err)
		}
		arguments = string(inputJSON)
	}

	return openai.ChatCompletionMessageToolCallUnionParam{
		OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
			ID: toolCall.Id,
			Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
				Name:      toolCall.Name,
				Arguments: arguments,
			},
		},
	}, nil
}

func toolToOpenAI(tool *aiv1.Tool) (openai.ChatCompletionToolUnionParam, error) {
	var schemaMap map[string]any
	if tool.InputSchema != "" {
		if err := json.Unmarshal([]byte(tool.InputSchema), &schemaMap); err != nil {
			return openai.ChatCompletionToolUnionParam{}, fmt.Errorf("failed to parse tool schema for %s: %w", tool.Name, err)
		}
	}

	return openai.ChatCompletionToolUnionParam{
		OfFunction: &openai.ChatCompletionFunctionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
				Parameters:  openai.FunctionParameters(schemaMap),
			},
		},
	}, nil
}
