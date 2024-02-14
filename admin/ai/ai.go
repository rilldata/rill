package ai

import "context"

type Client interface {
	Complete(ctx context.Context, prompt string) (string, error)
}
