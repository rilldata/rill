package drivers

import (
	"context"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
)

type AIService interface {
	Complete(ctx context.Context, msgs []*aiv1.CompletionMessage, tools []*aiv1.Tool) (*aiv1.CompletionMessage, error)
}
