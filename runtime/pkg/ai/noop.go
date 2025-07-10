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

func (noop) Complete(ctx context.Context, msgs []*aiv1.CompletionMessage, tools []*aiv1.Tool) (*aiv1.CompletionMessage, error) {
	return &aiv1.CompletionMessage{}, nil
}
