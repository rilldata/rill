package drivers

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Tool represents a tool that can be called by the AI - defined here to avoid import cycles
type Tool struct {
	Name        string
	Description string
	InputSchema string
}

// Option is a configuration option for completion requests
type Option func(*CompletionConfig)

// CompletionConfig holds configuration for completion requests
type CompletionConfig struct {
	Tools []Tool
	// Add other configuration options as needed
}

// BuildConfig creates a CompletionConfig by applying the provided options
func BuildConfig(opts ...Option) *CompletionConfig {
	config := &CompletionConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// CompletionMessage represents a message with rich content blocks
// Drivers are responsible for parsing their API responses into this structure
type CompletionMessage struct {
	Role    string
	Content []*runtimev1.ContentBlock
}

type AIService interface {
	Complete(ctx context.Context, msgs []*CompletionMessage, config *CompletionConfig) (*CompletionMessage, error)
}

// WithTools sets the available tools for the AI service to use in planning
func WithTools(tools []Tool) Option {
	return func(config *CompletionConfig) {
		config.Tools = tools
	}
}

// WithToolNames creates tools from just names - useful when you only need
// to reference tools by name and don't have full tool definitions available
func WithToolNames(toolNames []string) Option {
	return func(config *CompletionConfig) {
		tools := make([]Tool, len(toolNames))
		for i, name := range toolNames {
			tools[i] = Tool{Name: name}
		}
		config.Tools = tools
	}
}
