package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListGitBranches(ctx context.Context, req *runtimev1.ListGitBranchesRequest) (*runtimev1.ListGitBranchesResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	branches, currentBranch, err := repo.ListBranches(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	// List all deployments
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

	deployments, err := admin.ListDeployments(ctx)
	if err != nil && !errors.Is(err, drivers.ErrNotAuthenticated) && !errors.Is(err, cmdutil.ErrNoMatchingProject) {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	// Map deployments to branches with deployments
	deploymentMap := make(map[string]*drivers.Deployment, len(deployments))
	for _, d := range deployments {
		deploymentMap[d.Branch] = d
	}

	res := make([]*runtimev1.GitBranch, 0, len(branches))
	for _, branch := range branches {
		b := &runtimev1.GitBranch{
			Name: branch,
		}
		deployment, ok := deploymentMap[branch]
		if ok {
			b.HasDeployment = true
			b.EditableDeployment = deployment.Editable
		}
		res = append(res, b)
	}

	return &runtimev1.ListGitBranchesResponse{
		CurrentBranch: currentBranch,
		Branches:      res,
	}, nil
}

func (s *Server) GitSwitchBranch(ctx context.Context, req *runtimev1.GitSwitchBranchRequest) (*runtimev1.GitSwitchBranchResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	err = repo.SwitchBranch(ctx, req.Branch, req.Create, req.IgnoreLocalChanges)
	if err != nil {
		if errors.Is(err, git.ErrBranchNotFound) {
			return nil, status.Errorf(codes.NotFound, "branch %s not found", req.Branch)
		}
		return nil, fmt.Errorf("failed to switch git branch: %w", err)
	}
	return &runtimev1.GitSwitchBranchResponse{}, nil
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

func (s *Server) ListGitCommits(ctx context.Context, req *runtimev1.ListGitCommitsRequest) (*runtimev1.ListGitCommitsResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	pageSize := pagination.ValidPageSize(req.PageSize, 20)
	commits, nextPageToken, err := repo.ListCommits(ctx, req.PageToken, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list git commits: %v", err)
	}

	res := make([]*runtimev1.GitCommit, 0, len(commits))
	for _, c := range commits {
		res = append(res, &runtimev1.GitCommit{
			CommitSha:   c.CommitSha,
			AuthorName:  c.AuthorName,
			AuthorEmail: c.AuthorEmail,
			Message:     c.CommitMessage,
			CommittedOn: c.CommittedOn,
		})
	}

	return &runtimev1.ListGitCommitsResponse{
		Commits:       res,
		NextPageToken: nextPageToken,
	}, nil
}

// GitCommit implements RuntimeService.
func (s *Server) GitCommit(ctx context.Context, req *runtimev1.GitCommitRequest) (*runtimev1.GitCommitResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	hash, err := repo.Commit(ctx, req.CommitMessage)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit: %v", err)
	}
	return &runtimev1.GitCommitResponse{
		CommitSha: hash,
	}, nil
}

// RestoreGitCommit implements RuntimeService.
func (s *Server) RestoreGitCommit(ctx context.Context, req *runtimev1.RestoreGitCommitRequest) (*runtimev1.RestoreGitCommitResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	newCommitSHA, err := repo.RestoreCommit(ctx, req.CommitSha, req.RevertAll)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to restore commit: %v", err)
	}
	return &runtimev1.RestoreGitCommitResponse{
		NewCommitSha: newCommitSHA,
	}, nil
}

func (s *Server) GitMergeToBranch(ctx context.Context, req *runtimev1.GitMergeToBranchRequest) (*runtimev1.GitMergeToBranchResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	err = repo.MergeToBranch(ctx, req.Branch, req.Force)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to merge: %v", err)
	}
	return &runtimev1.GitMergeToBranchResponse{}, nil
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
