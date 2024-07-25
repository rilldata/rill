package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/server/auth"
)

func (s *Server) ConnectToGithubRepo(ctx context.Context, req *runtimev1.ConnectToGithubRepoRequest) (*runtimev1.ConnectToGithubRepoResponse, error) {
	// TODO: telemetry

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditInstance) {
		return nil, ErrForbidden
	}

	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	projectPath := repo.Root()
	// check if project has a git repo
	remote, _, err := gitutil.ExtractGitRemote(projectPath, "", false)
	if err != nil && !errors.Is(err, git.ErrRepositoryNotExists) && !errors.Is(err, gitutil.ErrGitRemoteNotFound) {
		return nil, err
	}
	if remote != nil {
		return nil, errors.New("git repository is already initialized with a remote")
	}

	var ghRepo *git.Repository
	ghRepo, err = git.PlainInitWithOptions(projectPath, &git.PlainInitOptions{
		InitOptions: git.InitOptions{
			DefaultBranch: plumbing.NewBranchReferenceName(req.Branch),
		},
		Bare: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init git repo: %w", err)
	}

	wt, err := ghRepo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// git add .
	if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return nil, fmt.Errorf("failed to add files to git: %w", err)
	}

	// git commit -m
	_, err = wt.Commit("Auto committed by Rill", &git.CommitOptions{All: true})
	if err != nil {
		if !errors.Is(err, git.ErrEmptyCommit) {
			return nil, fmt.Errorf("failed to commit files to git: %w", err)
		}
	}

	// Create the remote
	_, err = ghRepo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{req.Repo}})
	if err != nil {
		return nil, fmt.Errorf("failed to create remote: %w", err)
	}

	if err := ghRepo.PushContext(ctx, &git.PushOptions{Auth: &githttp.BasicAuth{Username: "x-access-token", Password: req.GhAccessToken}}); err != nil {
		return nil, fmt.Errorf("failed to push to remote %q : %w", req.Repo, err)
	}

	return &runtimev1.ConnectToGithubRepoResponse{}, nil
}
