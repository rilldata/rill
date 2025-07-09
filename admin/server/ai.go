package server

import (
	"context"
	"errors"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
)

func (s *Server) Complete(ctx context.Context, req *adminv1.CompleteRequest) (*adminv1.CompleteResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.Int("args.messages_len", len(req.Messages)),
		attribute.Int("args.tools_len", len(req.Tools)),
	)

	// Pass messages and tools directly to the AI service
	msg, err := s.admin.AI.Complete(ctx, req.Messages, req.Tools)
	if err != nil {
		return nil, err
	}

	if len(msg.Content) == 0 {
		return nil, errors.New("the AI responded with an empty message")
	}

	// Any tool use response will be passed to the client (the runtime server) for execution.
	return &adminv1.CompleteResponse{Message: msg}, nil
}
