package admin

import (
	"context"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
)

func (h *Handle) Complete(ctx context.Context, msgs []*aiv1.CompletionMessage, tools []*aiv1.Tool) (*aiv1.CompletionMessage, error) {
	res, err := h.admin.Complete(ctx, &adminv1.CompleteRequest{
		Messages: msgs,
		Tools:    tools,
	})
	if err != nil {
		return nil, err
	}

	return res.Message, nil
}
