package local

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	localv1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
)

func (s *Server) GitStatus(ctx context.Context, r *connect.Request[localv1.GitStatusRequest]) (*connect.Response[localv1.GitStatusResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	project, err := s.app.ch.LoadProject(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	config, err := s.app.ch.GitHelper(project.OrgName, project.Name, s.app.ProjectPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	remote, err := config.FullyQualifiedRemote()
	if err != nil {
		return nil, err
	}

	err = gitutil.GitFetch(ctx, s.app.ProjectPath, remote)
	if err != nil {
		return nil, err
	}

	gs, err := gitutil.RunGitStatus(s.app.ProjectPath)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitStatusResponse{
		Branch:        gs.Branch,
		GithubUrl:     config.Remote,
		ManagedGit:    project.ManagedGitId != "",
		LocalChanges:  gs.LocalChanges,
		LocalCommits:  gs.LocalCommits,
		RemoteCommits: gs.RemoteCommits,
	}), nil
}

func (s *Server) GitPull(ctx context.Context, r *connect.Request[localv1.GitPullRequest]) (*connect.Response[localv1.GitPullResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	project, err := s.app.ch.LoadProject(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	config, err := s.app.ch.GitHelper(project.OrgName, project.Name, s.app.ProjectPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}

	_, err = gitutil.GitPull(ctx, s.app.ProjectPath, r.Msg.DiscardLocal, config)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitPullResponse{}), nil
}

func (s *Server) GitPush(ctx context.Context, r *connect.Request[localv1.GitPushRequest]) (*connect.Response[localv1.GitPushResponse], error) {
	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	project, err := s.app.ch.LoadProject(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	config, err := s.app.ch.GitHelper(project.OrgName, project.Name, s.app.ProjectPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}

	author, err := s.app.ch.GitSignature(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}
	err = gitutil.CommitAndForcePush(ctx, s.app.ProjectPath, config.Remote, config.Username, config.Password, config.DefaultBranch, author, false)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitPushResponse{}), nil
}
