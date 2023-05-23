package server

import (
	"context"
	"fmt"
	"io/fs"

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

	exampleFS, err := examples.Unpack(req.Name)
	// check for not exist
	if err != nil {
		return nil, err
	}

	existingPaths := make(map[string]bool)
	if !req.Force {
		paths, err := repo.ListRecursive(ctx, req.InstanceId, "**")
		if err != nil {
			return nil, err
		}

		for _, path := range paths {
			existingPaths[path] = true
		}
	}

	paths := make([]string, 0)
	err = fs.WalkDir(exampleFS, "./", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if _, ok := existingPaths[path]; ok {
			return fmt.Errorf("path %q already exists", path)
		}

		paths = append(paths, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		err = func() error {
			file, err := exampleFS.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			return repo.Put(ctx, req.InstanceId, path, file)
		}()
		if err != nil {
			return nil, err
		}
	}

	return &runtimev1.UnpackExampleResponse{}, nil
}
