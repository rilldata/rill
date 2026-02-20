package server

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/server/auth"
)

func (s *Server) PullEnv(ctx context.Context, req *runtimev1.PullEnvRequest) (*runtimev1.PullEnvResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	admin, release, err := s.runtime.Admin(ctx, req.InstanceId)
	if err != nil {
		if errors.Is(err, runtime.ErrAdminNotConfigured) && s.adminOverride != nil {
			admin = s.adminOverride
			release = func() {}
		} else {
			return nil, err
		}
	}
	defer release()

	environment := req.Environment
	if environment == "" {
		environment = "dev"
	}

	// Fetch cloud variables
	cloudVars, err := admin.GetProjectVariables(ctx, environment)
	if err != nil && !errors.Is(err, drivers.ErrNotAuthenticated) {
		return nil, fmt.Errorf("failed to get project variables: %w", err)
	}

	// Parse local .env
	// Instance's project_variables contains variables from both rill.yaml and .env so can't be used here
	p, err := parser.Parse(ctx, repo, req.InstanceId, inst.Environment, inst.OLAPConnector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}

	localDotEnv := p.GetDotEnv()

	// Check if variables are already up to date
	if maps.Equal(cloudVars, localDotEnv) {
		return &runtimev1.PullEnvResponse{
			VariablesCount: int32(len(cloudVars)),
			Modified:       false,
		}, nil
	}

	// Merge: start with local, overlay cloud
	mergedVars := make(map[string]string)
	maps.Copy(mergedVars, localDotEnv)
	maps.Copy(mergedVars, cloudVars)

	// Write merged variables to .env file
	root, err := repo.Root(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo root: %w", err)
	}
	envPath := filepath.Join(root, ".env")
	err = godotenv.Write(mergedVars, envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to write .env file: %w", err)
	}

	// Ensure .env is in .gitignore
	_, _ = cmdutil.EnsureGitignoreHasDotenv(ctx, repo)

	return &runtimev1.PullEnvResponse{
		VariablesCount: int32(len(cloudVars)),
		Modified:       true,
	}, nil
}

func (s *Server) PushEnv(ctx context.Context, req *runtimev1.PushEnvRequest) (*runtimev1.PushEnvResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	admin, release, err := s.runtime.Admin(ctx, req.InstanceId)
	if err != nil {
		if errors.Is(err, runtime.ErrAdminNotConfigured) && s.adminOverride != nil {
			admin = s.adminOverride
			release = func() {}
		} else {
			return nil, err
		}
	}
	defer release()

	// Parse local .env
	p, err := parser.Parse(ctx, repo, req.InstanceId, inst.Environment, inst.OLAPConnector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}

	localDotEnv := p.GetDotEnv()

	// Fetch existing cloud variables
	cloudVars, err := admin.GetProjectVariables(ctx, req.Environment)
	if err != nil && !errors.Is(err, drivers.ErrNotAuthenticated) {
		return nil, fmt.Errorf("failed to get project variables: %w", err)
	}

	// Merge: start with cloud, overlay local
	mergedVars := make(map[string]string)
	for k, v := range cloudVars {
		mergedVars[k] = v
	}

	addedCount := int32(0)
	changedCount := int32(0)

	for k, v := range localDotEnv {
		if _, exists := cloudVars[k]; !exists {
			addedCount++
			mergedVars[k] = v
		} else if cloudVars[k] != v {
			changedCount++
			mergedVars[k] = v
		}
	}

	// No changes
	if addedCount == 0 && changedCount == 0 {
		return &runtimev1.PushEnvResponse{
			AddedCount:   0,
			ChangedCount: 0,
		}, nil
	}

	// Update cloud variables
	err = admin.UpdateProjectVariables(ctx, req.Environment, mergedVars)
	if err != nil {
		return nil, fmt.Errorf("failed to update project variables: %w", err)
	}

	return &runtimev1.PushEnvResponse{
		AddedCount:   addedCount,
		ChangedCount: changedCount,
	}, nil
}
