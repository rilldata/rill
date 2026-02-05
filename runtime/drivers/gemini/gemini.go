package gemini

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	drivers.Register("gemini", driver{})
	drivers.RegisterAsConnector("gemini", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Gemini",
	Description: "Connect to Google's Gemini API for language models.",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "api_key",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "API Key",
			Description: "API key for connecting to Gemini.",
			Secret:      true,
		},
		{
			Key:         "model",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Model",
			Description: "The Gemini model to use (e.g., 'gemini-3-pro-preview').",
			Placeholder: "",
		},
		{
			Key:         "max_output_tokens",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Max Output Tokens",
			Description: "Maximum number of tokens in the response.",
		},
		{
			Key:         "temperature",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Temperature",
			Description: "Sampling temperature to use (0.0-2.0).",
		},
		{
			Key:         "top_p",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Top P",
			Description: "Nucleus sampling parameter.",
		},
		{
			Key:         "top_k",
			Type:        drivers.NumberPropertyType,
			Required:    false,
			DisplayName: "Top K",
			Description: "Top-K sampling parameter.",
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
		return nil, errors.New("api_key is required")
	}
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		Backend: genai.BackendGeminiAPI,
		APIKey:  conf.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

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
	APIKey          string  `mapstructure:"api_key"`
	Model           string  `mapstructure:"model"`
	MaxOutputTokens int     `mapstructure:"max_output_tokens"`
	Temperature     float64 `mapstructure:"temperature"`
	TopP            float64 `mapstructure:"top_p"`
	TopK            float64 `mapstructure:"top_k"`
}

func (c *configProperties) getModel() string {
	if c.Model != "" {
		return c.Model
	}
	return "gemini-3-pro-preview"
}

type handle struct {
	client *genai.Client
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
	return "gemini"
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
	// Convert Rill messages to Gemini format, extracting system instruction separately
	systemInstructions, contents, err := convertMessages(opts.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	// Build generation config.
	// Only set parameters if explicitly configured (let Gemini use its defaults)
	genConfig := &genai.GenerateContentConfig{
		MaxOutputTokens:   int32(h.config.MaxOutputTokens),
		SystemInstruction: systemInstructions,
	}
	if h.config.Temperature > 0 {
		genConfig.Temperature = genai.Ptr(float32(h.config.Temperature))
	}
	if h.config.TopP > 0 {
		genConfig.TopP = genai.Ptr(float32(h.config.TopP))
	}
	if h.config.TopK > 0 {
		genConfig.TopK = genai.Ptr(float32(h.config.TopK))
	}

	// Convert tools
	if len(opts.Tools) > 0 {
		geminiTools, err := convertTools(opts.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tools: %w", err)
		}
		genConfig.Tools = geminiTools
	}

	// Convert output schema if present
	if opts.OutputSchema != nil {
		genConfig.ResponseMIMEType = "application/json"
		genConfig.ResponseJsonSchema = opts.OutputSchema
	}

	// Call Gemini API
	res, err := h.client.Models.GenerateContent(ctx, h.config.getModel(), contents, genConfig)
	if err != nil {
		return nil, err
	}

	// Convert response to Rill message format
	resMsg, err := convertResponseToRillMessage(res)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	return &drivers.CompleteResult{
		Message:      resMsg,
		InputTokens:  int(res.UsageMetadata.PromptTokenCount),
		OutputTokens: int(res.UsageMetadata.CandidatesTokenCount),
	}, nil
}

// convertTools converts Rill tools to Gemini tool format.
func convertTools(tools []*aiv1.Tool) ([]*genai.Tool, error) {
	var res []*genai.Tool
	for _, tool := range tools {
		var inputSchema map[string]any
		if tool.InputSchema != "" {
			if err := json.Unmarshal([]byte(tool.InputSchema), &inputSchema); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input schema for tool %q: %w", tool.Name, err)
			}
		} else {
			inputSchema = map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}
		}

		var outputSchema map[string]any
		if tool.OutputSchema != "" {
			if err := json.Unmarshal([]byte(tool.OutputSchema), &outputSchema); err != nil {
				return nil, fmt.Errorf("failed to unmarshal output schema for tool %q: %w", tool.Name, err)
			}
		}
		// Output schema is optional, so unlike for input schema, we don't need a fallback.

		res = append(res, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:                 tool.Name,
					Description:          tool.Description,
					ParametersJsonSchema: inputSchema,
					ResponseJsonSchema:   outputSchema,
				},
			},
		})
	}
	return res, nil
}

// convertMessages converts Rill messages to Gemini format.
// It returns system parts separately because Gemini's API treats them differently.
// It also merges consecutive contents with the same role since Gemini requires role alternation.
func convertMessages(msgs []*aiv1.CompletionMessage) (*genai.Content, []*genai.Content, error) {
	var systemParts []*genai.Part
	var contents []*genai.Content
	callIDToName := make(map[string]string)
	for _, msg := range msgs {
		if msg.Role == "system" {
			parts, err := convertSystemMessage(msg)
			if err != nil {
				return nil, nil, err
			}
			systemParts = append(systemParts, parts...)
			continue
		}

		converted, err := convertMessage(msg, callIDToName)
		if err != nil {
			return nil, nil, err
		}
		contents = append(contents, converted...)
	}

	contents = normalizeContents(contents)

	var systemInstructions *genai.Content
	if len(systemParts) > 0 {
		systemInstructions = &genai.Content{Parts: systemParts}
	}

	return systemInstructions, contents, nil
}

// normalizeContents normalizes contents to align with the Gemini API's expectations.
// Specifically, it merges consecutive contents with the same role, and ensures the conversation starts with a user turn.
func normalizeContents(contents []*genai.Content) []*genai.Content {
	if len(contents) == 0 {
		return contents
	}

	// Iterate through contents and merge consecutive ones with the same role.
	var merged []*genai.Content
	current := contents[0]
	for i := 1; i < len(contents); i++ {
		if contents[i].Role == current.Role {
			// Same role: merge parts into current
			current.Parts = append(current.Parts, contents[i].Parts...)
		} else {
			// Different role: save current and start new
			merged = append(merged, current)
			current = contents[i]
		}
	}
	merged = append(merged, current)

	// Ensure it starts with a user turn.
	if merged[0].Role != genai.RoleUser {
		return append([]*genai.Content{{
			Role:  genai.RoleUser,
			Parts: []*genai.Part{genai.NewPartFromText("Please proceed.")},
		}}, merged...)
	}

	return merged
}

// extractSystemParts extracts text parts from a system message.
func convertSystemMessage(msg *aiv1.CompletionMessage) ([]*genai.Part, error) {
	var parts []*genai.Part
	for _, block := range msg.Content {
		switch b := block.BlockType.(type) {
		case *aiv1.ContentBlock_Text:
			parts = append(parts, genai.NewPartFromText(b.Text))
		default:
			return nil, fmt.Errorf("unsupported system message block type: %T", block.BlockType)
		}
	}
	return parts, nil
}

// convertMessage converts a single Rill message to Gemini Content(s).
// The callIDToName map is populated with observed tool call IDs. The same map should be used in all calls to this function for a completion.
func convertMessage(msg *aiv1.CompletionMessage, callIDToName map[string]string) ([]*genai.Content, error) {
	role := genai.RoleUser
	if msg.Role == "assistant" {
		role = genai.RoleModel
	}

	// Collect all parts from this message into a single Content.
	// Gemini expects all parts from a single turn to be grouped together.
	var parts []*genai.Part
	for _, block := range msg.Content {
		switch b := block.BlockType.(type) {
		case *aiv1.ContentBlock_Text:
			parts = append(parts, genai.NewPartFromText(b.Text))

		case *aiv1.ContentBlock_ToolCall:
			callIDToName[b.ToolCall.Id] = b.ToolCall.Name

			part := genai.NewPartFromFunctionCall(b.ToolCall.Name, b.ToolCall.Input.AsMap())
			part.FunctionCall.ID = b.ToolCall.Id
			part.ThoughtSignature = []byte("skip_thought_signature_validator") // Necessary since we don't preserve signatures, and sometimes inject deterministic tool calls.
			parts = append(parts, part)

		case *aiv1.ContentBlock_ToolResult:
			part, err := convertFunctionResponse(b.ToolResult, callIDToName[b.ToolResult.Id])
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool result: %w", err)
			}
			parts = append(parts, part)

		default:
			return nil, fmt.Errorf("unsupported message block type: %T", block.BlockType)
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}

	return []*genai.Content{{Role: role, Parts: parts}}, nil
}

// convertFunctionResponse converts a Rill ToolResult to a Gemini FunctionResponse part.
func convertFunctionResponse(tr *aiv1.ToolResult, toolName string) (*genai.Part, error) {
	// Check tool name
	if toolName == "" {
		return nil, fmt.Errorf("tool name not found for tool result ID: %s", tr.Id)
	}

	// Parse the content as JSON if possible, otherwise wrap in a map
	var response map[string]any
	if err := json.Unmarshal([]byte(tr.Content), &response); err != nil {
		// If not valid JSON, wrap the content using Gemini's expected keys
		if tr.IsError {
			response = map[string]any{
				"error": tr.Content,
			}
		} else {
			response = map[string]any{
				"output": tr.Content,
			}
		}
	}

	return &genai.Part{
		FunctionResponse: &genai.FunctionResponse{
			ID:       tr.Id,
			Name:     toolName,
			Response: response,
		},
	}, nil
}

// convertResponseToRillMessage converts a Gemini response to a Rill CompletionMessage.
func convertResponseToRillMessage(res *genai.GenerateContentResponse) (*aiv1.CompletionMessage, error) {
	if len(res.Candidates) == 0 {
		return nil, errors.New("no candidates in response")
	}

	candidate := res.Candidates[0]
	if candidate.Content == nil {
		return nil, errors.New("no content in candidate")
	}

	var blocks []*aiv1.ContentBlock
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			blocks = append(blocks, &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_Text{Text: part.Text},
			})
		}
		if part.FunctionCall != nil {
			id := part.FunctionCall.ID
			if id == "" {
				id = randomID()
			}

			inputStruct, err := structpb.NewStruct(part.FunctionCall.Args)
			if err != nil {
				return nil, fmt.Errorf("failed to convert function call args to struct: %w", err)
			}

			blocks = append(blocks, &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_ToolCall{ToolCall: &aiv1.ToolCall{
					Id:    id,
					Name:  part.FunctionCall.Name,
					Input: inputStruct,
				}},
			})
		}
	}

	return &aiv1.CompletionMessage{
		Role:    "assistant",
		Content: blocks,
	}, nil
}

func randomID() string {
	id := make([]byte, 8)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	return "call_" + hex.EncodeToString(id)
}
