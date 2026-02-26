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

	// Fetch cloud variables for all environments
	cloudPerEnv, err := admin.GetProjectVariables(ctx, inst.Environment)
	if err != nil && !errors.Is(err, drivers.ErrNotAuthenticated) {
		return nil, fmt.Errorf("failed to get project variables: %w", err)
	}

	// Parse local .env files
	// Instance's project_variables contains variables from both rill.yaml and .env so can't be used here
	p, err := parser.Parse(ctx, repo, req.InstanceId, inst.Environment, inst.OLAPConnector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}

	localPerEnv := p.GetDotEnvPerEnvironment()

	// Check if all environments are already up to date
	equal := true
	for env, cloudVars := range cloudPerEnv {
		if !maps.Equal(cloudVars, localPerEnv[env]) {
			equal = false
			break
		}
	}
	if equal {
		for env, localVars := range localPerEnv {
			if !maps.Equal(localVars, cloudPerEnv[env]) {
				equal = false
				break
			}
		}
	}

	totalCount := int32(0)
	for _, vars := range cloudPerEnv {
		totalCount += int32(len(vars))
	}

	if equal {
		return &runtimev1.PullEnvResponse{
			VariablesCount: totalCount,
			Modified:       false,
		}, nil
	}

	// Write merged variables per environment
	root, err := repo.Root(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo root: %w", err)
	}

	for env, cloudVars := range cloudPerEnv {
		merged := make(map[string]string)
		maps.Copy(merged, localPerEnv[env])
		maps.Copy(merged, cloudVars)

		var envFileName string
		if env == "" {
			envFileName = ".env"
		} else {
			envFileName = fmt.Sprintf(".%s.env", env)
		}

		err = godotenv.Write(merged, filepath.Join(root, envFileName))
		if err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", envFileName, err)
		}

		_, _ = cmdutil.EnsureGitignoreHasDotenv(ctx, repo, envFileName)
	}

	return &runtimev1.PullEnvResponse{
		VariablesCount: totalCount,
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

	// Parse local .env files
	p, err := parser.Parse(ctx, repo, req.InstanceId, inst.Environment, inst.OLAPConnector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}

	localPerEnv := p.GetDotEnvPerEnvironment()

	// Fetch existing cloud variables for all environments
	cloudPerEnv, err := admin.GetProjectVariables(ctx, inst.Environment)
	if err != nil && !errors.Is(err, drivers.ErrNotAuthenticated) {
		return nil, fmt.Errorf("failed to get project variables: %w", err)
	}

	var addedCount, changedCount int32

	for env, local := range localPerEnv {
		cloud := cloudPerEnv[env]

		// Merge: start with cloud, overlay local; track what changed
		merged := make(map[string]string)
		maps.Copy(merged, cloud)
		var added, changed int32
		for k, v := range local {
			if _, exists := cloud[k]; !exists {
				added++
			} else if cloud[k] != v {
				changed++
			}
			merged[k] = v
		}

		if added == 0 && changed == 0 {
			continue
		}

		err = admin.UpdateProjectVariables(ctx, env, merged)
		if err != nil {
			return nil, fmt.Errorf("failed to update project variables for environment %q: %w", env, err)
		}

		addedCount += added
		changedCount += changed
	}

	return &runtimev1.PushEnvResponse{
		AddedCount:   addedCount,
		ChangedCount: changedCount,
	}, nil
}
