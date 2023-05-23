package server

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/examples"
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
	repo, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	entries, entryPaths, err := examples.Unpack(req.Name)
	if err != nil {
		return nil, err
	}

	if !req.Force {
		paths, err := repo.ListRecursive(ctx, req.InstanceId, "{sources,models}/*.{yaml,yml,sql}")
		if err != nil {
			return nil, err
		}

		// Should we check content or filenames ??
		if len(paths) != 0 {
			return nil, fmt.Errorf("repo is not empty %s", paths)
		}
	}

	for i, entry := range entries {
		stat, err := entry.Stat()
		if err != nil {
			return nil, err
		}

		if stat.IsDir() {
			continue
		}

		err = repo.Put(ctx, req.InstanceId, entryPaths[i], entry)
		if err != nil {
			return nil, err
		}
	}

	return &runtimev1.UnpackExampleResponse{}, nil
}
