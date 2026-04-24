package server

import (
	"context"
	"errors"
	"fmt"
	"maps"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/server/auth"
)

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
	p, err := parser.Parse(ctx, repo, req.InstanceId, inst.Environment, inst.OLAPConnector, false)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}

	localPerEnv := p.GetDotEnvPerEnvironment()

	// Fetch existing cloud variables
	cfg, err := admin.GetConfig(ctx)
	if err != nil && !errors.Is(err, drivers.ErrNotAuthenticated) {
		return nil, fmt.Errorf("failed to get project variables: %w", err)
	}
	var cloudPerEnv map[string]map[string]string
	if cfg != nil {
		cloudPerEnv = cfg.Variables
	} else {
		cloudPerEnv = make(map[string]map[string]string)
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
