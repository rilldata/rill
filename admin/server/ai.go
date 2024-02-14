package server

import (
	"context"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
)

func (s *Server) Complete(ctx context.Context, req *adminv1.CompleteRequest) (*adminv1.CompleteResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.Int("args.prompt_len", len(req.Prompt)),
	)

	// TODO: Invoke OpenAI API to complete the prompt

	return &adminv1.CompleteResponse{}, nil
}
