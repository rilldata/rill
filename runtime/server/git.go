package server

import (
	"context"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GitStatus implements RuntimeService.
func (s *Server) GitStatus(ctx context.Context, req *runtimev1.GitStatusRequest) (*runtimev1.GitStatusResponse, error) {
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

// GitPull implements RuntimeService.
func (s *Server) GitPull(ctx context.Context, req *runtimev1.GitPullRequest) (*runtimev1.GitPullResponse, error) {
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
