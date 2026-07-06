package admin

import (
	"context"
	"encoding/json"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

// MaxInputTokens implements drivers.AIService.
// The admin service proxies to a Rill-managed provider whose configuration isn't known here, so
// this defaults to a conservative limit that fits all providers currently used for managed AI.
func (h *Handle) MaxInputTokens() int {
	if h.config.MaxInputTokens > 0 {
		return h.config.MaxInputTokens
	}
	return drivers.DefaultAIMaxInputTokens
}

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
		CacheKey:         opts.CacheKey,
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
