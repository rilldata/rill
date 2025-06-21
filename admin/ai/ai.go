package ai

import (
	"context"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

type Client interface {
	Complete(ctx context.Context, msgs []*adminv1.CompletionMessage, tools []*adminv1.Tool) (*adminv1.CompletionMessage, error)
}
