package ai

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
)

type CompleteOptions struct {
	Messages     []*aiv1.CompletionMessage
	Tools        []*aiv1.Tool
	OutputSchema *jsonschema.Schema
}

type CompleteResult struct {
	Message      *aiv1.CompletionMessage
	InputTokens  int
	OutputTokens int
}

type Client interface {
	Complete(ctx context.Context, opts *CompleteOptions) (*CompleteResult, error)
}
