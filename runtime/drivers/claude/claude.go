package claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/mitchellh/mapstructure"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	drivers.Register("claude", driver{})
	drivers.RegisterAsConnector("claude", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Claude",
	Description: "Connect to Anthropic's Claude API for language models.",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "api_key",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "API Key",
			Description: "API key for connecting to Anthropic.",
			Secret:      true,
		},
		{
			Key:         "model",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Model",
			Description: "The Claude model to use (e.g., 'claude-opus-4-5-20251101').",
			Placeholder: "",
		},
		{
			Key:         "max_tokens",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Max Tokens",
			Description: "Maximum number of tokens in the response.",
			Default:     "4096",
		},
		{
			Key:         "base_url",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Base URL",
			Description: "Custom base URL for the Anthropic API.",
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

	client, err := newClaudeClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create Claude client: %w", err)
	}

	return &claude{
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
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`

	BaseURL string `mapstructure:"base_url"`
}

func (c *configProperties) getModel() string {
	if c.Model != "" {
		return c.Model
	}
	return "claude-opus-4-5-20251101"
}

func (c *configProperties) getMaxTokens() int {
	if c.MaxTokens > 0 {
		return c.MaxTokens
	}
	return 4096 // Default max tokens
}

func (c *configProperties) getTemperature() float64 {
	if c.Temperature > 0 {
		return c.Temperature
	}
	return 1.0 // Default temperature
}

type claude struct {
	client anthropic.Client
	config *configProperties
}

var _ drivers.AIService = (*claude)(nil)

// newClaudeClient creates a new Anthropic Claude client based on configuration.
func newClaudeClient(conf *configProperties) (anthropic.Client, error) {
	if conf.APIKey == "" {
		return anthropic.Client{}, errors.New("API key is required")
	}

	opts := []option.RequestOption{
		option.WithAPIKey(conf.APIKey),
	}

	if conf.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(conf.BaseURL))
	}

	return anthropic.NewClient(opts...), nil
}

// AsAI implements drivers.Handle.
func (c *claude) AsAI(instanceID string) (drivers.AIService, bool) {
	return c, true
}

// AsAdmin implements drivers.Handle.
func (c *claude) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (c *claude) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (c *claude) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsInformationSchema implements drivers.Handle.
func (c *claude) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *claude) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *claude) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// AsNotifier implements drivers.Handle.
func (c *claude) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// AsOLAP implements drivers.Handle.
func (c *claude) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (c *claude) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsRegistry implements drivers.Handle.
func (c *claude) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (c *claude) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *claude) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// Close implements drivers.Handle.
func (c *claude) Close() error {
	return nil
}

// Config implements drivers.Handle.
func (c *claude) Config() map[string]any {
	var configMap map[string]any
	_ = mapstructure.Decode(c.config, &configMap)
	return configMap
}

// Driver implements drivers.Handle.
func (c *claude) Driver() string {
	return "claude"
}

// Migrate implements drivers.Handle.
func (c *claude) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *claude) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Ping implements drivers.Handle.
func (c *claude) Ping(ctx context.Context) error {
	return nil
}

// Complete implements drivers.AIService.
// It sends a chat completion request to Claude and returns the response.
func (c *claude) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	// Extract system messages and convert remaining messages to Claude format
	systemPrompt, nonSystemMsgs := extractSystemMessages(opts.Messages)
	claudeMsgs, err := convertRillMessagesToClaudeMessages(nonSystemMsgs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	// Convert Rill tools to Claude's tool format
	var claudeTools []anthropic.ToolUnionParam
	if len(opts.Tools) > 0 {
		claudeTools = make([]anthropic.ToolUnionParam, len(opts.Tools))
		for i, tool := range opts.Tools {
			claudeTool, err := convertRillToolToClaudeTool(tool)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool: %w", err)
			}
			claudeTools[i] = claudeTool
		}
	}

	// Build system prompt blocks if present
	var systemBlocks []anthropic.TextBlockParam
	if systemPrompt != "" {
		systemBlocks = []anthropic.TextBlockParam{
			{Text: systemPrompt},
		}
	}

	// Handle structured outputs (beta feature) vs regular messages
	if opts.OutputSchema != nil {
		return c.completeWithStructuredOutput(ctx, claudeMsgs, claudeTools, systemBlocks, opts.OutputSchema)
	}

	return c.completeRegular(ctx, claudeMsgs, claudeTools, systemBlocks)
}

// completeRegular handles regular message completions without structured output.
func (c *claude) completeRegular(ctx context.Context, messages []anthropic.MessageParam, tools []anthropic.ToolUnionParam, systemBlocks []anthropic.TextBlockParam) (*drivers.CompleteResult, error) {
	params := anthropic.MessageNewParams{
		Model:       anthropic.Model(c.config.getModel()),
		MaxTokens:   int64(c.config.getMaxTokens()),
		Messages:    messages,
		Temperature: anthropic.Float(c.config.getTemperature()),
	}

	if len(systemBlocks) > 0 {
		params.System = systemBlocks
	}

	if len(tools) > 0 {
		params.Tools = tools
		params.ToolChoice = anthropic.ToolChoiceUnionParam{
			OfAuto: &anthropic.ToolChoiceAutoParam{},
		}
	}

	res, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return nil, err
	}

	return &drivers.CompleteResult{
		Message:      convertClaudeMessageToRillMessage(res),
		InputTokens:  int(res.Usage.InputTokens),
		OutputTokens: int(res.Usage.OutputTokens),
	}, nil
}

// completeWithStructuredOutput handles completions with structured JSON output using the beta API.
func (c *claude) completeWithStructuredOutput(ctx context.Context, messages []anthropic.MessageParam, tools []anthropic.ToolUnionParam, systemBlocks []anthropic.TextBlockParam, outputSchema any) (*drivers.CompleteResult, error) {
	// Convert messages to beta format
	betaMessages := convertToBetaMessages(messages)

	// Convert system blocks to beta format
	var betaSystemBlocks []anthropic.BetaTextBlockParam
	for _, block := range systemBlocks {
		betaSystemBlocks = append(betaSystemBlocks, anthropic.BetaTextBlockParam{
			Text: block.Text,
		})
	}

	// Convert tools to beta format
	var betaTools []anthropic.BetaToolUnionParam
	for _, tool := range tools {
		if tool.OfTool != nil {
			// Convert the input schema
			inputSchema := anthropic.BetaToolInputSchemaParam{
				Properties: tool.OfTool.InputSchema.Properties,
				Required:   tool.OfTool.InputSchema.Required,
			}
			betaTool := anthropic.BetaToolUnionParamOfTool(inputSchema, tool.OfTool.Name)
			if tool.OfTool.Description.Value != "" {
				betaTool.OfTool.Description = anthropic.String(tool.OfTool.Description.Value)
			}
			betaTools = append(betaTools, betaTool)
		}
	}

	// Convert outputSchema to the format expected by the beta API
	schemaMap, ok := outputSchema.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("output schema must be a map[string]any, got %T", outputSchema)
	}

	params := anthropic.BetaMessageNewParams{
		Model:        anthropic.Model(c.config.getModel()),
		MaxTokens:    int64(c.config.getMaxTokens()),
		Messages:     betaMessages,
		Betas:        []anthropic.AnthropicBeta{"structured-outputs-2025-11-13"},
		OutputFormat: anthropic.BetaJSONSchemaOutputFormat(schemaMap),
	}

	if len(betaSystemBlocks) > 0 {
		params.System = betaSystemBlocks
	}

	if len(betaTools) > 0 {
		params.Tools = betaTools
		params.ToolChoice = anthropic.BetaToolChoiceUnionParam{
			OfAuto: &anthropic.BetaToolChoiceAutoParam{},
		}
	}

	res, err := c.client.Beta.Messages.New(ctx, params)
	if err != nil {
		return nil, err
	}

	return &drivers.CompleteResult{
		Message:      convertBetaClaudeMessageToRillMessage(res),
		InputTokens:  int(res.Usage.InputTokens),
		OutputTokens: int(res.Usage.OutputTokens),
	}, nil
}

// extractSystemMessages extracts all system messages and concatenates their text content.
// Returns the concatenated system prompt and the remaining non-system messages.
func extractSystemMessages(msgs []*aiv1.CompletionMessage) (string, []*aiv1.CompletionMessage) {
	var systemTexts []string
	var nonSystemMessages []*aiv1.CompletionMessage

	for _, msg := range msgs {
		if msg.Role == "system" {
			// Extract text from content blocks
			for _, block := range msg.Content {
				if text := block.GetText(); text != "" {
					systemTexts = append(systemTexts, text)
				}
			}
		} else {
			nonSystemMessages = append(nonSystemMessages, msg)
		}
	}

	return strings.Join(systemTexts, "\n"), nonSystemMessages
}

// convertRillMessagesToClaudeMessages converts Rill messages to Claude message format.
func convertRillMessagesToClaudeMessages(msgs []*aiv1.CompletionMessage) ([]anthropic.MessageParam, error) {
	var result []anthropic.MessageParam

	for _, msg := range msgs {
		claudeMsgs, err := convertRillMessageToClaudeMessages(msg)
		if err != nil {
			return nil, err
		}
		result = append(result, claudeMsgs...)
	}

	return result, nil
}

// convertRillMessageToClaudeMessages converts a single Rill CompletionMessage to one or more Claude MessageParams.
//
// This handles the asymmetric nature of Claude's tool calling pattern:
// - Tool calls: Multiple calls are grouped in ONE assistant message
// - Tool results: Each result becomes content in a user message with role="user"
func convertRillMessageToClaudeMessages(msg *aiv1.CompletionMessage) ([]anthropic.MessageParam, error) {
	var result []anthropic.MessageParam

	// Separate content into regular content vs tool results
	var contentBlocks []anthropic.ContentBlockParamUnion
	var toolResults []anthropic.ToolResultBlockParam

	for _, block := range msg.Content {
		switch blockType := block.BlockType.(type) {
		case *aiv1.ContentBlock_Text:
			if blockType.Text != "" {
				contentBlocks = append(contentBlocks, anthropic.NewTextBlock(blockType.Text))
			}

		case *aiv1.ContentBlock_ToolCall:
			toolUseBlock := convertRillToolCallToClaudeToolUse(blockType.ToolCall)
			contentBlocks = append(contentBlocks, toolUseBlock)

		case *aiv1.ContentBlock_ToolResult:
			toolResults = append(toolResults, anthropic.ToolResultBlockParam{
				ToolUseID: blockType.ToolResult.Id,
				Content: []anthropic.ToolResultBlockParamContentUnion{
					{OfText: &anthropic.TextBlockParam{Text: blockType.ToolResult.Content}},
				},
				IsError: anthropic.Bool(blockType.ToolResult.IsError),
			})
		}
	}

	// Create main message for text content and tool calls
	if len(contentBlocks) > 0 {
		switch msg.Role {
		case "user":
			result = append(result, anthropic.NewUserMessage(contentBlocks...))
		case "assistant":
			result = append(result, anthropic.NewAssistantMessage(contentBlocks...))
		default:
			// For unknown roles, default to user
			result = append(result, anthropic.NewUserMessage(contentBlocks...))
		}
	}

	// Tool results are sent as user messages in Claude's API
	if len(toolResults) > 0 {
		toolResultBlocks := make([]anthropic.ContentBlockParamUnion, len(toolResults))
		for i, tr := range toolResults {
			trCopy := tr // Create a copy to avoid pointer issues
			toolResultBlocks[i] = anthropic.ContentBlockParamUnion{
				OfToolResult: &trCopy,
			}
		}
		result = append(result, anthropic.NewUserMessage(toolResultBlocks...))
	}

	return result, nil
}

// convertRillToolCallToClaudeToolUse converts a Rill ToolCall to Claude ToolUseBlock format.
func convertRillToolCallToClaudeToolUse(toolCall *aiv1.ToolCall) anthropic.ContentBlockParamUnion {
	var input map[string]any
	if toolCall.Input != nil {
		input = toolCall.Input.AsMap()
	} else {
		input = make(map[string]any)
	}

	return anthropic.ContentBlockParamUnion{
		OfToolUse: &anthropic.ToolUseBlockParam{
			ID:    toolCall.Id,
			Name:  toolCall.Name,
			Input: input,
		},
	}
}

// convertClaudeMessageToRillMessage converts Claude Message to Rill CompletionMessage format.
func convertClaudeMessageToRillMessage(message *anthropic.Message) *aiv1.CompletionMessage {
	var contentBlocks []*aiv1.ContentBlock

	for _, block := range message.Content {
		switch block.Type {
		case "text":
			contentBlocks = append(contentBlocks, &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_Text{Text: block.Text},
			})

		case "tool_use":
			// Unmarshal input from json.RawMessage to map[string]any
			var inputMap map[string]any
			if len(block.Input) > 0 {
				if err := json.Unmarshal(block.Input, &inputMap); err != nil {
					// If unmarshal fails, skip this tool call
					continue
				}
			} else {
				inputMap = make(map[string]any)
			}

			inputStruct, err := structpb.NewStruct(inputMap)
			if err != nil {
				// If conversion fails, skip this tool call
				continue
			}

			contentBlocks = append(contentBlocks, &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_ToolCall{
					ToolCall: &aiv1.ToolCall{
						Id:    block.ID,
						Name:  block.Name,
						Input: inputStruct,
					},
				},
			})
		}
	}

	return &aiv1.CompletionMessage{
		Role:    "assistant",
		Content: contentBlocks,
	}
}

// convertBetaClaudeMessageToRillMessage converts Beta Claude Message to Rill CompletionMessage format.
func convertBetaClaudeMessageToRillMessage(message *anthropic.BetaMessage) *aiv1.CompletionMessage {
	var contentBlocks []*aiv1.ContentBlock

	for _, block := range message.Content {
		switch block.Type {
		case "text":
			contentBlocks = append(contentBlocks, &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_Text{Text: block.Text},
			})

		case "tool_use":
			// Unmarshal input from json.RawMessage to map[string]any
			var inputMap map[string]any
			if len(block.Input) > 0 {
				if err := json.Unmarshal(block.Input, &inputMap); err != nil {
					// If unmarshal fails, skip this tool call
					continue
				}
			} else {
				inputMap = make(map[string]any)
			}

			inputStruct, err := structpb.NewStruct(inputMap)
			if err != nil {
				// If conversion fails, skip this tool call
				continue
			}

			contentBlocks = append(contentBlocks, &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_ToolCall{
					ToolCall: &aiv1.ToolCall{
						Id:    block.ID,
						Name:  block.Name,
						Input: inputStruct,
					},
				},
			})
		}
	}

	return &aiv1.CompletionMessage{
		Role:    "assistant",
		Content: contentBlocks,
	}
}

// convertRillToolToClaudeTool converts a Rill Tool to Claude Tool format.
func convertRillToolToClaudeTool(tool *aiv1.Tool) (anthropic.ToolUnionParam, error) {
	schemaMap, err := parseToolSchema(tool.InputSchema)
	if err != nil {
		return anthropic.ToolUnionParam{}, fmt.Errorf("failed to convert tool %s: %w", tool.Name, err)
	}

	// Build the input schema from the parsed map
	inputSchema := anthropic.ToolInputSchemaParam{}
	if props, ok := schemaMap["properties"]; ok {
		inputSchema.Properties = props
	}
	if req, ok := schemaMap["required"].([]any); ok {
		required := make([]string, len(req))
		for i, r := range req {
			if s, ok := r.(string); ok {
				required[i] = s
			}
		}
		inputSchema.Required = required
	}

	toolParam := anthropic.ToolUnionParamOfTool(inputSchema, tool.Name)
	if tool.Description != "" {
		toolParam.OfTool.Description = anthropic.String(tool.Description)
	}

	return toolParam, nil
}

// parseToolSchema parses a JSON schema string and returns a map, with fallback to default schema.
func parseToolSchema(schemaJSON string) (map[string]any, error) {
	if schemaJSON == "" {
		// Default schema when none is provided
		return map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}, nil
	}

	var schemaMap map[string]any
	if err := json.Unmarshal([]byte(schemaJSON), &schemaMap); err != nil {
		return nil, fmt.Errorf("failed to parse tool schema JSON: %w", err)
	}

	return schemaMap, nil
}

// convertToBetaMessages converts regular messages to beta message format.
func convertToBetaMessages(messages []anthropic.MessageParam) []anthropic.BetaMessageParam {
	betaMessages := make([]anthropic.BetaMessageParam, len(messages))
	for i, msg := range messages {
		betaMessages[i] = anthropic.BetaMessageParam{
			Role:    anthropic.BetaMessageParamRole(msg.Role),
			Content: convertToBetaContent(msg.Content),
		}
	}
	return betaMessages
}

// convertToBetaContent converts content blocks to beta content format.
func convertToBetaContent(content []anthropic.ContentBlockParamUnion) []anthropic.BetaContentBlockParamUnion {
	betaContent := make([]anthropic.BetaContentBlockParamUnion, len(content))
	for i, block := range content {
		switch {
		case block.OfText != nil:
			betaContent[i] = anthropic.NewBetaTextBlock(block.OfText.Text)
		case block.OfToolUse != nil:
			betaContent[i] = anthropic.NewBetaToolUseBlock(block.OfToolUse.ID, block.OfToolUse.Input, block.OfToolUse.Name)
		case block.OfToolResult != nil:
			betaContent[i] = anthropic.NewBetaToolResultBlock(block.OfToolResult.ToolUseID)
			// Note: NewBetaToolResultBlock creates a basic block; for content we need to set it
			if betaContent[i].OfToolResult != nil {
				betaContent[i].OfToolResult.Content = convertToBetaToolResultContent(block.OfToolResult.Content)
				betaContent[i].OfToolResult.IsError = block.OfToolResult.IsError
			}
		}
	}
	return betaContent
}

// convertToBetaToolResultContent converts tool result content to beta format.
func convertToBetaToolResultContent(content []anthropic.ToolResultBlockParamContentUnion) []anthropic.BetaToolResultBlockParamContentUnion {
	betaContent := make([]anthropic.BetaToolResultBlockParamContentUnion, len(content))
	for i, c := range content {
		if c.OfText != nil {
			betaContent[i] = anthropic.BetaToolResultBlockParamContentUnion{
				OfText: &anthropic.BetaTextBlockParam{
					Text: c.OfText.Text,
				},
			}
		}
	}
	return betaContent
}
