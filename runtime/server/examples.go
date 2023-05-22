package server

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/examples"
	"github.com/rilldata/rill/runtime/server/auth"
)

// ListExamples returns a list of embedded examples
func (s *Server) ListExamples(ctx context.Context, req *runtimev1.ListExamplesRequest) (*runtimev1.ListExamplesResponse, error) {
	list, err := examples.List()
	if err != nil {
		return nil, err
	}

	resp := make([]*runtimev1.Example, len(list))
	for i, example := range list {
		resp[i] = &runtimev1.Example{
			Name:        example.Name,
			Title:       example.Title,
			Description: example.Description,
		}
	}

	return &runtimev1.ListExamplesResponse{
		Examples: resp,
	}, nil
}

func (s *Server) UnpackExample(ctx context.Context, req *runtimev1.UnpackExampleRequest) (*runtimev1.UnpackExampleResponse, error) {
	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	repo, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	entries, entryPaths, err := examples.Unpack(req.Name)
	if err != nil {
		return nil, err
	}

	for i, entry := range entries {
		err := repo.Put(ctx, req.InstanceId, entryPaths[i], entry)
		if err != nil {
			return nil, err
		}
	}

	if !req.Force {
		paths, err := repo.ListRecursive(ctx, req.InstanceId, "**")
		if err != nil {
			return nil, err
		}

		for _, path := range paths {
			for _, entryPath := range entryPaths {
				if path == entryPath {
					return nil, fmt.Errorf("file already exists for path %q", path)
				}
			}
		}
	}

	return &runtimev1.UnpackExampleResponse{}, nil
}
