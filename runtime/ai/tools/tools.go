package tools

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// ServerTools defines the interface for server tool operations
type ServerTools interface {
	GenerateMetricsViewFile(ctx context.Context, req *runtimev1.GenerateMetricsViewFileRequest) (*runtimev1.GenerateMetricsViewFileResponse, error)
}

func newToolResult(result any, err error) map[string]any {
	if err != nil {
		return map[string]any{
			"error": err.Error(),
		}
	}
	return map[string]any{
		"result": result,
	}
}
