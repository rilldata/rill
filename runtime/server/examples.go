package server

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/bufbuild/connect-go"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/pkg/examples"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

// ListExamples returns a list of embedded examples
func (s *Server) ListExamples(ctx context.Context, req *connect.Request[runtimev1.ListExamplesRequest]) (*connect.Response[runtimev1.ListExamplesResponse], error) {
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

	return connect.NewResponse(&runtimev1.ListExamplesResponse{
		Examples: resp,
	}), nil
}

func (s *Server) UnpackExample(ctx context.Context, req *connect.Request[runtimev1.UnpackExampleRequest]) (*connect.Response[runtimev1.UnpackExampleResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.name", req.Msg.Name),
		attribute.Bool("args.force", req.Msg.Force),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	repo, err := s.runtime.Repo(ctx, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}

	exampleFS, err := examples.Get(req.Msg.Name)
	if err != nil {
		return nil, err
	}

	existingPaths := make(map[string]bool)
	if !req.Msg.Force {
		paths, err := repo.ListRecursive(ctx, req.Msg.InstanceId, "**")
		if err != nil {
			return nil, err
		}

		for _, path := range paths {
			existingPaths[path] = true
		}
	}

	paths := make([]string, 0)
	err = fs.WalkDir(exampleFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if _, ok := existingPaths[filepath.Join("/", path)]; ok {
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

			return repo.Put(ctx, req.Msg.InstanceId, path, file)
		}()
		if err != nil {
			return nil, err
		}
	}

	return connect.NewResponse(&runtimev1.UnpackExampleResponse{}), nil
}

func (s *Server) UnpackEmpty(ctx context.Context, req *connect.Request[runtimev1.UnpackEmptyRequest]) (*connect.Response[runtimev1.UnpackEmptyResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.title", req.Msg.Title),
		attribute.Bool("args.force", req.Msg.Force),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	repo, err := s.runtime.Repo(ctx, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}

	c := rillv1beta.New(repo, req.Msg.InstanceId)
	if c.IsInit(ctx) && !req.Msg.Force {
		return nil, fmt.Errorf("a Rill project already exists")
	}

	// Init empty project
	err = c.InitEmpty(ctx, req.Msg.Title)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.UnpackEmptyResponse{}), nil
}
