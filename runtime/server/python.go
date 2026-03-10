package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/python"
	"github.com/rilldata/rill/runtime/server/auth"
)

func (s *Server) DetectPython(ctx context.Context, req *runtimev1.DetectPythonRequest) (*runtimev1.DetectPythonResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditInstance) {
		return nil, ErrForbidden
	}

	info, err := python.DetectPython(req.PythonPath)
	if err != nil {
		// Not finding Python is not an error; return found=false
		return &runtimev1.DetectPythonResponse{
			Found: false,
		}, nil
	}

	return &runtimev1.DetectPythonResponse{
		Found:   info.Found,
		Path:    info.Path,
		Version: info.Version,
	}, nil
}

func (s *Server) SetupPythonEnvironment(ctx context.Context, req *runtimev1.SetupPythonEnvironmentRequest) (*runtimev1.SetupPythonEnvironmentResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditInstance) {
		return nil, ErrForbidden
	}

	// Get the repo root for the instance
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	projectRoot, err := repo.Root(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get project root: %w", err)
	}

	// Run setup
	result, err := python.SetupEnvironment(ctx, &python.SetupOptions{
		ProjectRoot: projectRoot,
		Packages:    req.Packages,
		PythonPath:  req.PythonPath,
	})
	if err != nil {
		return nil, err
	}

	// Write the connector YAML file
	connectorYAML := fmt.Sprintf("type: connector\ndriver: python\npython_path: %s\n", result.PythonPath)
	err = repo.Put(ctx, "connectors/python.yaml", strings.NewReader(connectorYAML))
	if err != nil {
		return nil, fmt.Errorf("failed to write connector YAML: %w", err)
	}

	return &runtimev1.SetupPythonEnvironmentResponse{
		PythonPath:        result.PythonPath,
		VenvPath:          result.VenvPath,
		InstalledPackages: result.Installed,
	}, nil
}
