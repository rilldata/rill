package ai

import (
	"context"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	drivers.Register("mock_ai", driver{})
	drivers.RegisterAsConnector("mock_ai", driver{})
}

type driver struct{}

var _ drivers.Driver = driver{}

func (d driver) Spec() drivers.Spec {
	return drivers.Spec{
		DisplayName: "Mock AI",
		Description: "Mock AI service for testing",
		ConfigProperties: []*drivers.PropertySpec{
			{
				Key:         "enable_tool_calling",
				Type:        drivers.BooleanPropertyType,
				Required:    false,
				DisplayName: "Enable Tool Calling",
				Description: "If true, returns mock tool calls. If false (default), echoes user messages.",
			},
		},
		SourceProperties: []*drivers.PropertySpec{},
		ImplementsAI:     true,
	}
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, srcProps map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, srcProps map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	toolCallingEnabled, _ := config["enable_tool_calling"].(bool)

	return &connection{
		config:          config,
		logger:          logger,
		toolCallingMode: toolCallingEnabled,
	}, nil
}

type connection struct {
	config          map[string]any
	logger          *zap.Logger
	toolCallingMode bool
}

var _ drivers.Handle = &connection{}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	return nil
}

// Driver implements drivers.Handle.
func (c *connection) Driver() string {
	return "mock_ai"
}

// Config implements drivers.Handle.
func (c *connection) Config() map[string]any {
	return c.config
}

// Close implements drivers.Handle.
func (c *connection) Close() error {
	return nil
}

// Migrate implements drivers.Handle.
func (c *connection) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsRegistry implements drivers.Handle.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return c, true
}

// AsOLAP implements drivers.Handle.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsInformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Handle.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// Complete implements drivers.AIService.
func (c *connection) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	if c.toolCallingMode {
		return &drivers.CompleteResult{
			Message:      c.handleToolCalling(),
			InputTokens:  10,
			OutputTokens: 20,
		}, nil
	}
	return &drivers.CompleteResult{
		Message:      c.echoUserMessage(opts.Messages),
		InputTokens:  10,
		OutputTokens: 20,
	}, nil
}

// handleToolCalling returns a simple mock tool call for testing
func (c *connection) handleToolCalling() *aiv1.CompletionMessage {
	inputStruct, _ := structpb.NewStruct(map[string]interface{}{})
	return &aiv1.CompletionMessage{
		Role: "assistant",
		Content: []*aiv1.ContentBlock{
			{
				BlockType: &aiv1.ContentBlock_ToolCall{
					ToolCall: &aiv1.ToolCall{
						Id:    "tool_call_123",
						Name:  "list_metrics_views",
						Input: inputStruct,
					},
				},
			},
		},
	}
}

// echoUserMessage finds the last user message and echoes it back
func (c *connection) echoUserMessage(msgs []*aiv1.CompletionMessage) *aiv1.CompletionMessage {
	text := c.findLastUserText(msgs)
	if text == "" {
		text = "No user message found"
	} else {
		text = "Echo: " + text
	}

	return &aiv1.CompletionMessage{
		Role: "assistant",
		Content: []*aiv1.ContentBlock{
			{
				BlockType: &aiv1.ContentBlock_Text{
					Text: text,
				},
			},
		},
	}
}

// findLastUserText extracts text from the most recent user message
func (c *connection) findLastUserText(msgs []*aiv1.CompletionMessage) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			for _, block := range msgs[i].Content {
				if text := block.GetText(); text != "" {
					return text
				}
			}
		}
	}
	return ""
}
