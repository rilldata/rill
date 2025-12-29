package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListGitBranches(ctx context.Context, req *runtimev1.ListGitBranchesRequest) (*runtimev1.ListGitBranchesResponse, error) {
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	branches, err := repo.ListBranches(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	// Convert to GitBranch objects with is_current flag
	res := make([]*runtimev1.GitBranch, 0, len(branches))
	for _, branch := range branches {
		res = append(res, &runtimev1.GitBranch{
			Name:                 branch.Name,
			IsCurrent:            branch.IsCurrent,
			HasPreviewDeployment: branch.HasPreviewDeployment,
		})
	}

	return &runtimev1.ListGitBranchesResponse{
		Branches: res,
	}, nil
}

func (s *Server) SwitchBranch(ctx context.Context, req *runtimev1.SwitchBranchRequest) (*runtimev1.SwitchBranchResponse, error) {
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	err = repo.SwitchBranch(ctx, req.BranchName, req.CreateIfNotExists, req.IgnoreLocalChanges)
	if err != nil {
		if errors.Is(err, git.ErrBranchNotFound) {
			return nil, status.Errorf(codes.NotFound, "branch %s not found", req.BranchName)
		}
		return nil, fmt.Errorf("failed to switch git branch: %w", err)
	}
	return &runtimev1.SwitchBranchResponse{}, nil
}

// GitStatus implements RuntimeService.
func (s *Server) GitStatus(ctx context.Context, req *runtimev1.GitStatusRequest) (*runtimev1.GitStatusResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	gs, err := repo.Status(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get git status: %v", err)
	}
	if !gs.IsGitRepo {
		return nil, status.Error(codes.FailedPrecondition, "not a git repository")
	}
	return &runtimev1.GitStatusResponse{
		Branch:        gs.Branch,
		GithubUrl:     gs.RemoteURL,
		ManagedGit:    gs.ManagedRepo,
		LocalChanges:  gs.LocalChanges,
		LocalCommits:  gs.LocalCommits,
		RemoteCommits: gs.RemoteCommits,
	}, nil
}

// GitCommit implements RuntimeService.
func (s *Server) GitCommit(ctx context.Context, req *runtimev1.GitCommitRequest) (*runtimev1.GitCommitResponse, error) {
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	err = repo.Commit(ctx, req.CommitMessage)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit: %v", err)
	}
	return &runtimev1.GitCommitResponse{}, nil
}

// RestoreGitCommit implements RuntimeService.
func (s *Server) RestoreGitCommit(ctx context.Context, req *runtimev1.RestoreGitCommitRequest) (*runtimev1.RestoreGitCommitResponse, error) {
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	newCommitSHA, err := repo.RestoreCommit(ctx, req.CommitSha)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to restore commit: %v", err)
	}
	return &runtimev1.RestoreGitCommitResponse{
		NewCommitSha: newCommitSHA,
	}, nil
}

// GitPull implements RuntimeService.
func (s *Server) GitPull(ctx context.Context, req *runtimev1.GitPullRequest) (*runtimev1.GitPullResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	err = repo.Pull(ctx, &drivers.PullOptions{
		UserTriggered:  true,
		DiscardChanges: req.DiscardLocal,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to pull: %v", err)
	}
	return &runtimev1.GitPullResponse{}, nil
}

// GitPush implements RuntimeService.
func (s *Server) GitPush(ctx context.Context, req *runtimev1.GitPushRequest) (*runtimev1.GitPushResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	err = repo.CommitAndPush(ctx, req.CommitMessage, req.Force)
	if err != nil {
		if errors.Is(err, drivers.ErrRemoteAhead) {
			return nil, status.Error(codes.FailedPrecondition, "remote repository has changes that are not in local state, please pull first")
		}
		return nil, status.Errorf(codes.Internal, "failed to push: %v", err)
	}
	return &runtimev1.GitPushResponse{}, nil
}
