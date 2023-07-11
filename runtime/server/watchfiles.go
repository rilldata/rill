package server

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

const (
	WatchFilesMethod = "WatchFiles"
)

func (s *Server) WatchFiles(req *runtimev1.WatchFilesRequest, stream runtimev1.RuntimeService_WatchFilesServer) error {
	repo, err := s.runtime.Repo(stream.Context(), req.InstanceId)
	if err != nil {
		return err
	}

	return repo.Watch(stream.Context(), req.Replay, func(event drivers.WatchEvent) error {
		if !event.Dir {
			err := stream.Send(&runtimev1.WatchFilesResponse{
				Event: event.Type,
				Path:  event.Path,
			})

			return err
		}
		return nil
	})
}
