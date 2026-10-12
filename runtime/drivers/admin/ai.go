package admin

import (
	"context"
	"encoding/json"
	"errors"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

func (h *Handle) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	var outputJSONSchema string
	if opts.OutputSchema != nil {
		schemaBytes, err := json.Marshal(opts.OutputSchema)
		if err != nil {
			return nil, err
		}
		outputJSONSchema = string(schemaBytes)
	}

	res, err := h.admin.Complete(ctx, &adminv1.CompleteRequest{
		Messages:         opts.Messages,
		Tools:            opts.Tools,
		OutputJsonSchema: outputJSONSchema,
	})
	if err != nil {
		return nil, err
	}

	return &drivers.CompleteResult{
		Message:           res.Message,
		Provider:          res.GetProvider(),
		InputTokens:       int(res.InputTokens),
		OutputTokens:      int(res.OutputTokens),
		CachedInputTokens: int(res.GetCachedInputTokens()),
	}, nil
}

func (h *Handle) Evaluate(ctx context.Context, req *aiv1.EvaluateRequest) (*aiv1.EvaluateResponse, error) {
	return nil, errors.New("not implemented")
}
