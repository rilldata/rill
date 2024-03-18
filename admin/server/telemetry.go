package server

import (
	"context"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

func (s *Server) RecordEvents(ctx context.Context, req *adminv1.RecordEventsRequest) (*adminv1.RecordEventsResponse, error) {
	for _, e := range req.Events {
		err := s.activity.RecordRaw(e.AsMap())
		if err != nil {
			return nil, err
		}
	}
	return &adminv1.RecordEventsResponse{}, nil
}
