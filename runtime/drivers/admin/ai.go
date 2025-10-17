package admin

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
)

func (h *Handle) Complete(ctx context.Context, msgs []*aiv1.CompletionMessage, tools []*aiv1.Tool, outputSchema *jsonschema.Schema) (*aiv1.CompletionMessage, error) {
	var outputJSONSchema string
	if outputSchema != nil {
		schemaBytes, err := json.Marshal(outputSchema)
		if err != nil {
			return nil, err
		}
		outputJSONSchema = string(schemaBytes)
	}

	res, err := h.admin.Complete(ctx, &adminv1.CompleteRequest{
		Messages:         msgs,
		Tools:            tools,
		OutputJsonSchema: outputJSONSchema,
	})
	if err != nil {
		return nil, err
	}

	return res.Message, nil
}
