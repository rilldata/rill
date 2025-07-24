package ai

import (
	"context"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
)

type CompletionOptions struct {
	Model string
}

type Client interface {
	Complete(ctx context.Context, msgs []*aiv1.CompletionMessage, tools []*aiv1.Tool, opts CompletionOptions) (*aiv1.CompletionMessage, error)
}
