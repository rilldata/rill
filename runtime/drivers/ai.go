package drivers

import (
	"context"
)

type AIService interface {
	Complete(ctx context.Context, msgs []*CompletionMessage) (*CompletionMessage, error)
}

type CompletionMessage struct {
	Role string
	Data string
}
