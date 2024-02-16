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
	)

	msg, err := s.admin.AI.Complete(ctx, req.Messages)
	if err != nil {
		return nil, err
	}

	if msg.Data == "" {
		return nil, errors.New("the AI responded with an empty message")
	}

	return &adminv1.CompleteResponse{Message: msg}, nil
}
