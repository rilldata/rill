package admin

import (
	"context"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

func (h *Handle) Complete(ctx context.Context, msgs []*drivers.CompletionMessage) (*drivers.CompletionMessage, error) {
	reqMsgs := make([]*adminv1.CompletionMessage, len(msgs))
	for i, msg := range msgs {
		reqMsgs[i] = &adminv1.CompletionMessage{Role: msg.Role, Data: msg.Data}
	}

	res, err := h.admin.Complete(ctx, &adminv1.CompleteRequest{Messages: reqMsgs})
	if err != nil {
		return nil, err
	}

	return &drivers.CompletionMessage{Role: res.Message.Role, Data: res.Message.Data}, nil
}
