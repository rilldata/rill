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

// CompletionMessage represents a message with rich content blocks
// Drivers are responsible for parsing their API responses into this structure
type CompletionMessage struct {
	Role    string
	Content []*runtimev1.ContentBlock
}

type AIService interface {
	Complete(ctx context.Context, msgs []*CompletionMessage, tools []Tool) (*CompletionMessage, error)
}
