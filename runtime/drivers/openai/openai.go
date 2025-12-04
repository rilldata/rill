package openai

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ai"
	"github.com/rilldata/rill/runtime/storage"
	openaidriver "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
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
	if instanceID == "" {
		return nil, drivers.ErrNotImplemented
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	opts := &ai.Options{
		BaseURL:     conf.BaseURL,
		APIType:     openaidriver.APIType(conf.APIType),
		APIVersion:  conf.APIVersion,
		Model:       conf.Model,
		Temperature: conf.Temperature,
	}
	aiClient, err := ai.NewOpenAI(conf.APIKey, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	return &openai{
		aiClient: aiClient,
		config:   conf,
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

type openai struct {
	aiClient ai.Client
	config   *configProperties
}

var _ drivers.AIService = (*openai)(nil)

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
func (o *openai) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
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
func (o *openai) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	res, err := o.aiClient.Complete(ctx, &ai.CompleteOptions{
		Messages:     opts.Messages,
		Tools:        opts.Tools,
		OutputSchema: opts.OutputSchema,
	})
	if err != nil {
		return nil, err
	}

	return &drivers.CompleteResult{
		Message:      res.Message,
		InputTokens:  res.InputTokens,
		OutputTokens: res.OutputTokens,
	}, nil
}
