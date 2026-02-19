package claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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

const defaultTemperature = 0.1

func init() {
	drivers.Register("claude", driver{})
	drivers.RegisterAsConnector("claude", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Claude",
	Description: "Connect to Anthropic's Claude API for language models.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connectors/services/claude",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "api_key",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "API Key",
			Description: "API key for connecting to Claude.",
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
			Default:     "8192",
		},
		{
			Key:         "temperature",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Temperature",
			Description: "Sampling temperature to use.",
			Default:     "0.1",
		},
		{
			Key:         "base_url",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Base URL",
			Description: "Custom base URL for the Claude API.",
			Placeholder: "",
		},
	},
	ImplementsAI: true,
}

type driver struct{}

var _ drivers.Driver = driver{}

// Spec implements drivers.Driver.
func (d driver) Spec() drivers.Spec {
	return spec
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

	opts := []option.RequestOption{
		option.WithAPIKey(conf.APIKey),
	}
	if conf.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(conf.BaseURL))
	}

	client := anthropic.NewClient(opts...)

	return &handle{
		client: client,
		config: conf,
	}, nil
}

// HasAnonymousSourceAccess implements drivers.Driver.
func (d driver) HasAnonymousSourceAccess(ctx context.Context, srcProps map[string]any, logger *zap.Logger) (bool, error) {
	return false, drivers.ErrNotImplemented
}

// TertiarySourceConnectors implements drivers.Driver.
func (d driver) TertiarySourceConnectors(ctx context.Context, srcProps map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, drivers.ErrNotImplemented
}

type configProperties struct {
	APIKey      string   `mapstructure:"api_key"`
	Model       string   `mapstructure:"model"`
	MaxTokens   int      `mapstructure:"max_tokens"`
	Temperature *float64 `mapstructure:"temperature"`
	BaseURL     string   `mapstructure:"base_url"`
}

func (c *configProperties) getModel() string {
	if c.Model != "" {
		return c.Model
	}
	return string(anthropic.ModelClaudeOpus4_5_20251101)
}

func (c *configProperties) getMaxTokens() int {
	if c.MaxTokens > 0 {
		return c.MaxTokens
	}
	return 8192 // Default max tokens
}

func (c *configProperties) getTemperature() float64 {
	if c.Temperature != nil {
		return *c.Temperature
	}
	return defaultTemperature
}

type handle struct {
	client anthropic.Client
	config *configProperties
}

var _ drivers.AIService = (*handle)(nil)

// AsAI implements drivers.Handle.
func (h *handle) AsAI(instanceID string) (drivers.AIService, bool) {
	return h, true
}

// AsAdmin implements drivers.Handle.
func (h *handle) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (h *handle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (h *handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsInformationSchema implements drivers.Handle.
func (h *handle) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (h *handle) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (h *handle) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// AsNotifier implements drivers.Handle.
func (h *handle) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// AsOLAP implements drivers.Handle.
func (h *handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (h *handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsRegistry implements drivers.Handle.
func (h *handle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (h *handle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (h *handle) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// Close implements drivers.Handle.
func (h *handle) Close() error {
	return nil
}

// Config implements drivers.Handle.
func (h *handle) Config() map[string]any {
	var configMap map[string]any
	_ = mapstructure.Decode(h.config, &configMap)
	return configMap
}

// Driver implements drivers.Handle.
func (h *handle) Driver() string {
	return "claude"
}

// Migrate implements drivers.Handle.
func (h *handle) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (h *handle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Ping implements drivers.Handle.
func (h *handle) Ping(ctx context.Context) error {
	return nil
}

// Complete implements drivers.AIService.
func (h *handle) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	system, msgs, err := convertMessages(opts.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	betaTools, err := convertTools(opts.Tools)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tools: %w", err)
	}

	params := anthropic.BetaMessageNewParams{
		Model:       anthropic.Model(h.config.getModel()),
		MaxTokens:   int64(h.config.getMaxTokens()),
		Temperature: anthropic.Float(h.config.getTemperature()),
		Messages:    msgs,
		System:      system,
	}

	if len(betaTools) > 0 {
		params.Tools = betaTools
		params.ToolChoice = anthropic.BetaToolChoiceUnionParam{
			OfAuto: &anthropic.BetaToolChoiceAutoParam{},
		}
	}

	if opts.OutputSchema != nil {
		schemaBytes, err := json.Marshal(opts.OutputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal output schema: %w", err)
		}
		var schemaMap map[string]any
		if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal output schema: %w", err)
		}
		params.Betas = []anthropic.AnthropicBeta{"structured-outputs-2025-11-13"}
		params.OutputFormat = anthropic.BetaJSONSchemaOutputFormat(schemaMap)
	}

	res, err := h.client.Beta.Messages.New(ctx, params)
	if err != nil {
		return nil, err
	}

	resMsgs, err := convertResponseMessage(res)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response message: %w", err)
	}

	return &drivers.CompleteResult{
		Message:      resMsgs,
		InputTokens:  int(res.Usage.InputTokens),
		OutputTokens: int(res.Usage.OutputTokens),
	}, nil
}

// convertMessages converts Rill messages to Claude beta message format.
// It returns system blocks separately because Claude's API treats them differently.
func convertMessages(msgs []*aiv1.CompletionMessage) ([]anthropic.BetaTextBlockParam, []anthropic.BetaMessageParam, error) {
	var system []anthropic.BetaTextBlockParam
	var other []anthropic.BetaMessageParam

	for _, msg := range msgs {
		if msg.Role != "system" {
			converted, err := convertMessage(msg)
			if err != nil {
				return nil, nil, err
			}
			other = append(other, converted...)
			continue
		}

		for _, block := range msg.Content {
			switch block := block.BlockType.(type) {
			case *aiv1.ContentBlock_Text:
				system = append(system, anthropic.BetaTextBlockParam{Text: block.Text})
			default:
				return nil, nil, fmt.Errorf("unsupported system message block type: %T", block)
			}
		}
	}

	return system, other, nil
}

// convertMessage converts a single Rill message to Claude beta messages.
// Tool results become separate user messages per Claude's API requirements.
func convertMessage(msg *aiv1.CompletionMessage) ([]anthropic.BetaMessageParam, error) {
	var result []anthropic.BetaMessageParam

	role := anthropic.BetaMessageParamRoleUser
	if msg.Role == "assistant" {
		role = anthropic.BetaMessageParamRoleAssistant
	}

	for _, block := range msg.Content {
		switch b := block.BlockType.(type) {
		case *aiv1.ContentBlock_Text:
			result = append(result, anthropic.BetaMessageParam{
				Role: role,
				Content: []anthropic.BetaContentBlockParamUnion{
					anthropic.NewBetaTextBlock(b.Text),
				},
			})
		case *aiv1.ContentBlock_ToolCall:
			result = append(result, anthropic.BetaMessageParam{
				Role: anthropic.BetaMessageParamRoleAssistant, // NOTE: Hard-coded as assistant
				Content: []anthropic.BetaContentBlockParamUnion{
					convertToolCall(b.ToolCall),
				},
			})
		case *aiv1.ContentBlock_ToolResult:
			result = append(result, anthropic.BetaMessageParam{
				Role: anthropic.BetaMessageParamRoleUser, // NOTE: Hard-coded as user
				Content: []anthropic.BetaContentBlockParamUnion{
					convertToolResult(b.ToolResult),
				},
			})
		default:
			return nil, fmt.Errorf("unsupported message block type: %T", block)
		}
	}

	return result, nil
}

// convertToolCall converts a Rill tool call to a Claude beta tool use block.
func convertToolCall(tc *aiv1.ToolCall) anthropic.BetaContentBlockParamUnion {
	input := make(map[string]any)
	if tc.Input != nil {
		input = tc.Input.AsMap()
	}
	return anthropic.NewBetaToolUseBlock(tc.Id, input, tc.Name)
}

// convertToolResult converts a Rill tool result to a Claude beta tool result block.
func convertToolResult(tr *aiv1.ToolResult) anthropic.BetaContentBlockParamUnion {
	block := anthropic.NewBetaToolResultBlock(tr.Id)
	block.OfToolResult.Content = []anthropic.BetaToolResultBlockParamContentUnion{
		{OfText: &anthropic.BetaTextBlockParam{Text: tr.Content}},
	}
	block.OfToolResult.IsError = anthropic.Bool(tr.IsError)
	return block
}

// convertTools converts Rill tools to Claude beta tool union params.
func convertTools(tools []*aiv1.Tool) ([]anthropic.BetaToolUnionParam, error) {
	if len(tools) == 0 {
		return nil, nil
	}

	result := make([]anthropic.BetaToolUnionParam, 0, len(tools))
	for _, tool := range tools {
		converted, err := convertTool(tool)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

// convertTool converts a single Rill tool to a Claude beta tool union param.
func convertTool(tool *aiv1.Tool) (anthropic.BetaToolUnionParam, error) {
	inputSchema := anthropic.BetaToolInputSchemaParam{}
	if tool.InputSchema == "" {
		// Default schema when none is provided
		inputSchema.Type = "object"
		inputSchema.Properties = map[string]any{}
	} else {
		if err := json.Unmarshal([]byte(tool.InputSchema), &inputSchema); err != nil {
			return anthropic.BetaToolUnionParam{}, fmt.Errorf("failed to parse schema for tool %q: %w", tool.Name, err)
		}
	}

	result := anthropic.BetaToolUnionParamOfTool(inputSchema, tool.Name)
	if tool.Description != "" {
		result.OfTool.Description = anthropic.String(tool.Description)
	}
	return result, nil
}

// convertResponseMessage converts a Claude beta message to a Rill completion message.
func convertResponseMessage(msg *anthropic.BetaMessage) (*aiv1.CompletionMessage, error) {
	var blocks []*aiv1.ContentBlock

	for idx := range msg.Content {
		block := msg.Content[idx] // Note: Separate line to avoid lint error about copying large structs
		switch block.Type {
		case "text":
			blocks = append(blocks, &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_Text{Text: block.Text},
			})
		case "thinking":
			blocks = append(blocks, &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_Text{Text: block.Thinking},
			})
		case "tool_use":
			cb, err := convertResponseToolUse(block)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, cb)
		default:
			// We ignore other blocks for now
		}
	}

	return &aiv1.CompletionMessage{
		Role:    "assistant",
		Content: blocks,
	}, nil
}

// convertResponseToolUse converts a Claude beta tool use block to a Rill content block.
func convertResponseToolUse(block anthropic.BetaContentBlockUnion) (*aiv1.ContentBlock, error) {
	inputMap := make(map[string]any)
	if len(block.Input) > 0 {
		if err := json.Unmarshal(block.Input, &inputMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input for tool call %q: %w", block.Name, err)
		}
	}

	inputStruct, err := structpb.NewStruct(inputMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input map to struct for tool call %q: %w", block.Name, err)
	}

	return &aiv1.ContentBlock{
		BlockType: &aiv1.ContentBlock_ToolCall{
			ToolCall: &aiv1.ToolCall{
				Id:    block.ID,
				Name:  block.Name,
				Input: inputStruct,
			},
		},
	}, nil
}
