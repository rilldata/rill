package bedrock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/auth/bearer"
	"github.com/mitchellh/mapstructure"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/awsutil"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	drivers.Register("bedrock", driver{})
	drivers.RegisterAsConnector("bedrock", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Amazon Bedrock",
	Description: "Connect to language models hosted in your own AWS account with Amazon Bedrock.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connectors/services/bedrock",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "model",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Model",
			Description: "The Bedrock model ID, inference profile ID, or ARN to use.",
			Placeholder: "us.anthropic.claude-sonnet-4-5-20250929-v1:0",
		},
		{
			Key:         "region",
			Type:        drivers.StringPropertyType,
			DisplayName: "Region",
			Description: "AWS region to call Bedrock in.",
			Placeholder: "us-east-1",
		},
		{
			Key:         "api_key",
			Type:        drivers.StringPropertyType,
			DisplayName: "API key",
			Description: "Amazon Bedrock API key. Use either an API key or AWS access keys.",
			Secret:      true,
		},
		{
			Key:         "aws_access_key_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS access key ID",
			Description: "AWS access key ID for explicit credentials.",
			Placeholder: "Enter your AWS access key ID",
			Secret:      true,
		},
		{
			Key:         "aws_secret_access_key",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS secret access key",
			Description: "AWS secret access key for explicit credentials.",
			Placeholder: "Enter your AWS secret access key",
			Secret:      true,
		},
		{
			Key:         "aws_access_token",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS session token",
			Description: "AWS session token for temporary credentials.",
			Secret:      true,
		},
		{
			Key:         "aws_role_arn",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS role ARN",
			Description: "AWS role ARN to assume before calling Bedrock.",
			Secret:      true,
		},
		{
			Key:         "aws_role_session_name",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS role session name",
			Description: "Optional session name to use when assuming an AWS role.",
			Secret:      true,
		},
		{
			Key:         "aws_external_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS external ID",
			Description: "Optional external ID to use when assuming an AWS role for cross-account access.",
			Secret:      true,
		},
		{
			Key:         "max_tokens",
			Type:        drivers.NumberPropertyType,
			DisplayName: "Max Tokens",
			Description: "Maximum number of tokens in the response. Defaults to the model's maximum.",
		},
		{
			Key:         "temperature",
			Type:        drivers.NumberPropertyType,
			DisplayName: "Temperature",
			Description: "Sampling temperature to use. Defaults to the model's default. Some models, such as reasoning models, don't support it.",
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
func (d driver) Open(ctx context.Context, _, instanceID string, cfgMap map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	conf := &configProperties{}
	err := mapstructure.WeakDecode(cfgMap, conf)
	if err != nil {
		return nil, err
	}

	if conf.Model == "" {
		return nil, errors.New("model is required")
	}

	awsConfig, err := conf.awsConfig(ctx, logger)
	if err != nil {
		return nil, err
	}

	client := bedrockruntime.NewFromConfig(awsConfig, func(o *bedrockruntime.Options) {
		switch {
		case conf.APIKey != "":
			o.BearerAuthTokenProvider = bearer.StaticTokenProvider{Token: bearer.Token{Value: conf.APIKey}}
			o.AuthSchemePreference = []string{"httpBearerAuth"}
		case conf.hasAWSCredentials() || !conf.AllowHostAccess:
			// The SDK picks up a Bedrock API key from the host (e.g. the AWS_BEARER_TOKEN_BEDROCK env var) and prefers it over SigV4,
			// which would silently override the configured credentials.
			// So we only allow it when host access is allowed and no credentials are configured.
			o.BearerAuthTokenProvider = nil
			o.AuthSchemePreference = nil
		}
	})

	return &handle{
		client:    client,
		awsConfig: awsConfig,
		config:    conf,
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
	Model           string   `mapstructure:"model"`
	Region          string   `mapstructure:"region"`
	APIKey          string   `mapstructure:"api_key"`
	AccessKeyID     string   `mapstructure:"aws_access_key_id"`
	SecretAccessKey string   `mapstructure:"aws_secret_access_key"`
	SessionToken    string   `mapstructure:"aws_access_token"`
	RoleARN         string   `mapstructure:"aws_role_arn"`
	RoleSessionName string   `mapstructure:"aws_role_session_name"`
	ExternalID      string   `mapstructure:"aws_external_id"`
	MaxTokens       int      `mapstructure:"max_tokens"`
	Temperature     *float64 `mapstructure:"temperature"`
	AllowHostAccess bool     `mapstructure:"allow_host_access"`
}

// hasAWSCredentials returns true if AWS credentials or a role to assume are configured explicitly.
func (c *configProperties) hasAWSCredentials() bool {
	return c.AccessKeyID != "" || c.SecretAccessKey != "" || c.SessionToken != "" || c.RoleARN != ""
}

// awsConfig builds the AWS config for calling Bedrock.
// If an API key is provided, it is used instead of AWS credentials (see Open).
// Otherwise, it uses static credentials if provided, else falls back to the SDK's default credential chain if host access is allowed.
// If a role ARN is configured, it assumes that role on top of the base credentials.
func (c *configProperties) awsConfig(ctx context.Context, logger *zap.Logger) (aws.Config, error) {
	loadOptions := []func(*config.LoadOptions) error{
		config.WithDefaultRegion("us-east-1"),
		// An empty region is ignored, so the region from the default config is used if not set here.
		config.WithRegion(c.Region),
		config.WithLogger(awsutil.NewAWSLogger(logger)),
	}

	if c.APIKey != "" {
		if c.hasAWSCredentials() {
			return aws.Config{}, errors.New("api_key cannot be combined with AWS access keys or aws_role_arn")
		}
		// Prevent the SDK from loading AWS credentials from the host since they are not used.
		loadOptions = append(loadOptions, config.WithCredentialsProvider(aws.AnonymousCredentials{}))
	} else if c.AccessKeyID != "" || c.SecretAccessKey != "" || c.SessionToken != "" {
		p := credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, c.SessionToken)
		loadOptions = append(loadOptions, config.WithCredentialsProvider(p))
	} else if !c.AllowHostAccess {
		return aws.Config{}, errors.New("either api_key or aws_access_key_id and aws_secret_access_key are required")
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	if c.RoleARN != "" {
		provider := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(awsConfig), c.RoleARN, func(o *stscreds.AssumeRoleOptions) {
			if c.RoleSessionName != "" {
				o.RoleSessionName = c.RoleSessionName
			}
			if c.ExternalID != "" {
				o.ExternalID = aws.String(c.ExternalID)
			}
		})
		awsConfig.Credentials = aws.NewCredentialsCache(provider)
	}

	return awsConfig, nil
}

type handle struct {
	client    *bedrockruntime.Client
	awsConfig aws.Config
	config    *configProperties
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

// AsMCP implements drivers.Handle.
func (h *handle) AsMCP() (drivers.MCPService, bool) {
	return nil, false
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
	return "bedrock"
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
// It verifies the AWS credentials (including any assumed role) without invoking the model.
// API keys can't be verified without invoking the model, so they are not checked.
func (h *handle) Ping(ctx context.Context) error {
	if h.config.APIKey != "" {
		return nil
	}
	_, err := sts.NewFromConfig(h.awsConfig).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to verify AWS credentials: %w", err)
	}
	return nil
}

// Complete implements drivers.AIService.
func (h *handle) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	system, msgs, err := convertMessages(opts.Messages, len(opts.Tools) > 0)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	input := &bedrockruntime.ConverseInput{
		ModelId:  aws.String(h.config.Model),
		Messages: msgs,
		System:   system,
	}

	// Only set inference parameters if configured,
	// since supported parameters and limits vary by model (e.g. reasoning models reject temperature).
	if h.config.Temperature != nil || h.config.MaxTokens > 0 {
		input.InferenceConfig = &types.InferenceConfiguration{}
		if h.config.Temperature != nil {
			input.InferenceConfig.Temperature = aws.Float32(float32(*h.config.Temperature))
		}
		if h.config.MaxTokens > 0 {
			input.InferenceConfig.MaxTokens = aws.Int32(int32(h.config.MaxTokens))
		}
	}

	if len(opts.Tools) > 0 {
		tools, err := convertTools(opts.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tools: %w", err)
		}
		input.ToolConfig = &types.ToolConfiguration{
			Tools:      tools,
			ToolChoice: &types.ToolChoiceMemberAuto{Value: types.AutoToolChoice{}},
		}
	}

	if opts.OutputSchema != nil {
		schema, err := json.Marshal(opts.OutputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal output schema: %w", err)
		}
		input.OutputConfig = &types.OutputConfig{
			TextFormat: &types.OutputFormat{
				Type: types.OutputFormatTypeJsonSchema,
				Structure: &types.OutputFormatStructureMemberJsonSchema{
					Value: types.JsonSchemaDefinition{
						Name:   aws.String("llm_completion_result"),
						Schema: aws.String(string(schema)),
					},
				},
			},
		}
	}

	res, err := h.client.Converse(ctx, input)
	if err != nil {
		return nil, err
	}

	out, ok := res.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return nil, fmt.Errorf("unexpected Bedrock output type %T", res.Output)
	}

	resMsg, err := convertResponseMessage(&out.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response message: %w", err)
	}

	// Like Claude, Bedrock reports cache-read tokens separately from InputTokens.
	result := &drivers.CompleteResult{
		Message:  resMsg,
		Provider: "bedrock",
	}
	if res.Usage != nil {
		result.InputTokens = int(aws.ToInt32(res.Usage.InputTokens))
		result.CachedInputTokens = int(aws.ToInt32(res.Usage.CacheReadInputTokens))
		result.OutputTokens = int(aws.ToInt32(res.Usage.OutputTokens))
	}
	return result, nil
}

// convertMessages converts Rill messages to Bedrock Converse messages.
// It returns system blocks separately because the Converse API takes them as a separate field.
// Since Converse requires strictly alternating user and assistant turns,
// consecutive blocks with the same role are merged into a single message.
//
// Converse rejects tool call and tool result blocks unless tools are also offered in the request.
// So if toolsOffered is false (e.g. on the final iteration of a tool call loop), past tool calls and results are rendered as text.
func convertMessages(msgs []*aiv1.CompletionMessage, toolsOffered bool) ([]types.SystemContentBlock, []types.Message, error) {
	var system []types.SystemContentBlock
	var res []types.Message
	toolNames := make(map[string]string) // Tool call ID to tool name, for rendering tool results as text

	add := func(role types.ConversationRole, block types.ContentBlock) {
		if n := len(res); n > 0 && res[n-1].Role == role {
			res[n-1].Content = append(res[n-1].Content, block)
			return
		}
		res = append(res, types.Message{Role: role, Content: []types.ContentBlock{block}})
	}

	for _, msg := range msgs {
		if msg.Role == "system" {
			for _, block := range msg.Content {
				b, ok := block.BlockType.(*aiv1.ContentBlock_Text)
				if !ok {
					return nil, nil, fmt.Errorf("unsupported system message block type: %T", block.BlockType)
				}
				if b.Text != "" {
					system = append(system, &types.SystemContentBlockMemberText{Value: b.Text})
				}
			}
			continue
		}

		role := types.ConversationRoleUser
		if msg.Role == "assistant" {
			role = types.ConversationRoleAssistant
		}

		for _, block := range msg.Content {
			switch b := block.BlockType.(type) {
			case *aiv1.ContentBlock_Text:
				// Converse rejects blank text blocks
				if b.Text == "" {
					continue
				}
				add(role, &types.ContentBlockMemberText{Value: b.Text})
			case *aiv1.ContentBlock_ToolCall:
				input := map[string]any{}
				if b.ToolCall.Input != nil {
					input = b.ToolCall.Input.AsMap()
				}
				if !toolsOffered {
					inputJSON, err := json.Marshal(input)
					if err != nil {
						return nil, nil, fmt.Errorf("failed to marshal input for tool call %q: %w", b.ToolCall.Name, err)
					}
					toolNames[b.ToolCall.Id] = b.ToolCall.Name
					add(types.ConversationRoleAssistant, &types.ContentBlockMemberText{Value: fmt.Sprintf("Called tool %q with input: %s", b.ToolCall.Name, inputJSON)})
					continue
				}
				add(types.ConversationRoleAssistant, &types.ContentBlockMemberToolUse{Value: types.ToolUseBlock{
					ToolUseId: aws.String(b.ToolCall.Id),
					Name:      aws.String(b.ToolCall.Name),
					Input:     document.NewLazyDocument(input),
				}})
			case *aiv1.ContentBlock_ToolResult:
				if !toolsOffered {
					kind := "Result"
					if b.ToolResult.IsError {
						kind = "Error"
					}
					add(types.ConversationRoleUser, &types.ContentBlockMemberText{Value: fmt.Sprintf("%s from tool %q:\n%s", kind, toolNames[b.ToolResult.Id], b.ToolResult.Content)})
					continue
				}
				tr := types.ToolResultBlock{
					ToolUseId: aws.String(b.ToolResult.Id),
					Content:   []types.ToolResultContentBlock{&types.ToolResultContentBlockMemberText{Value: b.ToolResult.Content}},
				}
				if b.ToolResult.IsError {
					tr.Status = types.ToolResultStatusError
				}
				add(types.ConversationRoleUser, &types.ContentBlockMemberToolResult{Value: tr})
			default:
				return nil, nil, fmt.Errorf("unsupported message block type: %T", block.BlockType)
			}
		}
	}

	return system, res, nil
}

// convertTools converts Rill tools to Bedrock Converse tools.
//
// Some Bedrock models (such as DeepSeek) don't resolve $ref in tool schemas,
// and treat parameters that don't declare a type themselves as strings.
// So we copy the referenced definition's type onto top-level $ref parameters,
// e.g. {"$ref": "#/$defs/TimeRange"} becomes {"$ref": "#/$defs/TimeRange", "type": "object"}.
func convertTools(tools []*aiv1.Tool) ([]types.Tool, error) {
	res := make([]types.Tool, 0, len(tools))
	for _, tool := range tools {
		// Converse requires an input schema, so default to an empty object schema.
		schema := map[string]any{"type": "object", "properties": map[string]any{}}
		if tool.InputSchema != "" {
			if err := json.Unmarshal([]byte(tool.InputSchema), &schema); err != nil {
				return nil, fmt.Errorf("failed to parse schema for tool %q: %w", tool.Name, err)
			}
		}

		props, _ := schema["properties"].(map[string]any)
		for _, p := range props {
			prop, ok := p.(map[string]any)
			if !ok || prop["type"] != nil {
				continue
			}
			if typ := propertyType(schema, prop); typ != nil {
				prop["type"] = typ
			}
		}

		ts := types.ToolSpecification{
			Name:        aws.String(tool.Name),
			InputSchema: &types.ToolInputSchemaMemberJson{Value: document.NewLazyDocument(schema)},
		}
		if tool.Description != "" {
			ts.Description = aws.String(tool.Description)
		}
		res = append(res, &types.ToolMemberToolSpec{Value: ts})
	}
	return res, nil
}

// convertResponseMessage converts a Bedrock Converse response message to a Rill completion message.
func convertResponseMessage(msg *types.Message) (*aiv1.CompletionMessage, error) {
	var blocks []*aiv1.ContentBlock
	for _, block := range msg.Content {
		switch b := block.(type) {
		case *types.ContentBlockMemberText:
			blocks = append(blocks, &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_Text{Text: b.Value},
			})
		case *types.ContentBlockMemberToolUse:
			// Round-trip through JSON since decoding a smithy document into a map yields numbers as document.Number,
			// which structpb does not support.
			input := map[string]any{}
			if b.Value.Input != nil {
				data, err := b.Value.Input.MarshalSmithyDocument()
				if err != nil {
					return nil, fmt.Errorf("failed to marshal input for tool call %q: %w", aws.ToString(b.Value.Name), err)
				}
				if err := json.Unmarshal(data, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input for tool call %q: %w", aws.ToString(b.Value.Name), err)
				}
			}
			inputStruct, err := structpb.NewStruct(input)
			if err != nil {
				return nil, fmt.Errorf("failed to convert input for tool call %q: %w", aws.ToString(b.Value.Name), err)
			}
			blocks = append(blocks, &aiv1.ContentBlock{
				BlockType: &aiv1.ContentBlock_ToolCall{
					ToolCall: &aiv1.ToolCall{
						Id:    aws.ToString(b.Value.ToolUseId),
						Name:  aws.ToString(b.Value.Name),
						Input: inputStruct,
					},
				},
			})
		default:
			// We ignore other blocks (such as reasoning content) for now
		}
	}

	return &aiv1.CompletionMessage{
		Role:    "assistant",
		Content: blocks,
	}, nil
}

// propertyType returns the JSON schema type of a top-level property in a tool's input schema.
// If the property doesn't declare a type itself, it follows a local $ref (e.g. "#/$defs/TimeRange") to the definition's type.
// It returns nil if the type can't be determined.
func propertyType(schema, prop map[string]any) any {
	if typ, ok := prop["type"]; ok {
		return typ
	}
	ref, _ := prop["$ref"].(string)
	for key, prefix := range map[string]string{"$defs": "#/$defs/", "definitions": "#/definitions/"} {
		if name, ok := strings.CutPrefix(ref, prefix); ok {
			defs, _ := schema[key].(map[string]any)
			def, _ := defs[name].(map[string]any)
			return def["type"]
		}
	}
	return nil
}
