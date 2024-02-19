package ai

import (
	"context"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

type noop struct{}

var _ Client = noop{}

func NewNoop() Client {
	return noop{}
}

func (noop) Complete(ctx context.Context, msgs []*adminv1.CompletionMessage) (*adminv1.CompletionMessage, error) {
	return &adminv1.CompletionMessage{}, nil
}
