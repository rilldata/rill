package admin

import (
	"context"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
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

func (h *Handle) LLMProvider() model.Provider {
	return &provider{}
}

// Implement provider using the Complete method of the admin Handle.
type provider struct{}

var _ model.Provider = (*provider)(nil)

// GetModel implements model.Provider.
func (p *provider) GetModel(modelName string) (model.Model, error) {
	return nil, drivers.ErrNotImplemented
}
