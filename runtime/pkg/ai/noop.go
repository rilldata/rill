package ai

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
)

type noop struct{}

var _ Client = noop{}

func NewNoop() Client {
	return noop{}
}

func (noop) Complete(ctx context.Context, msgs []*aiv1.CompletionMessage, tools []*aiv1.Tool, outputSchema *jsonschema.Schema) (*aiv1.CompletionMessage, error) {
	return &aiv1.CompletionMessage{}, nil
}
