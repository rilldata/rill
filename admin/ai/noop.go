package ai

import "context"

type noop struct{}

var _ Client = noop{}

func NewNoop() Client {
	return noop{}
}

func (noop) Complete(ctx context.Context, prompt string) (string, error) {
	return "", nil
}
