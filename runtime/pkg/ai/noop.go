package ai

import (
	"context"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
)

type noop struct{}

var _ Client = noop{}

func NewNoop() Client {
	return noop{}
}

func (noop) Complete(ctx context.Context, opts *CompleteOptions) (*CompleteResult, error) {
	return &CompleteResult{
		Message: &aiv1.CompletionMessage{},
	}, nil
}
