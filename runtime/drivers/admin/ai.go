package admin

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

func (h *Handle) GenerateMetricsViewYAML(ctx context.Context, baseTable, sqlDialect string, schema *runtimev1.StructType) (string, error) {
	// TODO: Construct prompt and send to admin service for inference:
	// res, err := h.admin.Complete(ctx, &adminv1.CompleteRequest{
	// 	Prompt: "...",
	// })
	// if err != nil {
	// 	return "", err
	// }
	panic("not implemented")
}
