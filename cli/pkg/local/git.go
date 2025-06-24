package local

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	localv1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
)

func (s *Server) GitStatus(ctx context.Context, r *connect.Request[localv1.GitStatusRequest]) (*connect.Response[localv1.GitStatusResponse], error) {
	// try with native git configurations
	err := gitutil.GitFetch(ctx, s.app.ProjectPath, nil)
	if err == nil || !s.app.ch.IsAuthenticated() {
		// if native git fetch succeeds, return the status
		gs, err := gitutil.RunGitStatus(s.app.ProjectPath, "origin")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:        gs.Branch,
			GithubUrl:     gs.RemoteURL,
			LocalChanges:  gs.LocalChanges,
			LocalCommits:  gs.LocalCommits,
			RemoteCommits: gs.RemoteCommits,
		}), nil
	}

	// if native git fetch fails, try with ephemeral token - this may be a managed git project

	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		// if the user is not authenticated, we cannot fetch the project
		// return the best effort status
		gs, err := gitutil.RunGitStatus(s.app.ProjectPath, "origin")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:    gs.Branch,
			GithubUrl: gs.RemoteURL,
		}), nil
	}

	project, err := s.app.ch.LoadProject(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}
	if project == nil {
		// If the project is not found return the best effort status
		gs, err := gitutil.RunGitStatus(s.app.ProjectPath, "origin")
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&localv1.GitStatusResponse{
			Branch:    gs.Branch,
			GithubUrl: gs.RemoteURL,
		}), nil
	}

	// get ephemeral git credentials
	config, err := s.app.ch.GitHelper(project.OrgName, project.Name, s.app.ProjectPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	// set remote
	// usually not needed but the older flow did not set the remote by name `rill`
	err = gitutil.SetRemote(s.app.ProjectPath, config.Remote)
	if err != nil {
		return nil, err
	}
	err = gitutil.GitFetch(ctx, s.app.ProjectPath, config)
	if err != nil {
		return nil, err
	}
	gs, err := gitutil.RunGitStatus(s.app.ProjectPath, "rill")
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitStatusResponse{
		Branch:        gs.Branch,
		GithubUrl:     gs.RemoteURL,
		ManagedGit:    true,
		LocalChanges:  gs.LocalChanges,
		LocalCommits:  gs.LocalCommits,
		RemoteCommits: gs.RemoteCommits,
	}), nil
}

func (s *Server) GitPull(ctx context.Context, r *connect.Request[localv1.GitPullRequest]) (*connect.Response[localv1.GitPullResponse], error) {
	out, err := gitutil.GitPull(ctx, s.app.ProjectPath, r.Msg.DiscardLocal, "origin")
	if err == nil && !strings.Contains(out, "Repository not found") {
		return connect.NewResponse(&localv1.GitPullResponse{
			Output: out,
		}), nil
	}
	// if native git pull fails, try with ephemeral token - this may be a managed git project

	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	project, err := s.app.ch.LoadProject(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("git credentials not set and repo is not connected to a project")
	}

	config, err := s.app.ch.GitHelper(project.OrgName, project.Name, s.app.ProjectPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	err = gitutil.SetRemote(s.app.ProjectPath, config.Remote)
	if err != nil {
		return nil, err
	}

	remote, err := config.FullyQualifiedRemote()
	if err != nil {
		return nil, err
	}

	out, err = gitutil.GitPull(ctx, s.app.ProjectPath, r.Msg.DiscardLocal, remote)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&localv1.GitPullResponse{
		Output: out,
	}), nil
}

func (s *Server) GitPush(ctx context.Context, r *connect.Request[localv1.GitPushRequest]) (*connect.Response[localv1.GitPushResponse], error) {
	st, err := gitutil.RunGitStatus(s.app.ProjectPath, "origin")
	if err != nil {
		return nil, err
	}
	if st.RemoteCommits > 0 && !r.Msg.Force {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot push with remote commits present, please pull first"))
	}

	// generate git signature
	author, err := gitutil.NativeGitSignature(ctx, s.app.ProjectPath)
	if err == nil {
		err = gitutil.CommitAndForcePush(ctx, s.app.ProjectPath, &gitutil.Config{Remote: st.RemoteURL, DefaultBranch: st.Branch}, r.Msg.CommitMessage, author, true)
		if err == nil {
			return connect.NewResponse(&localv1.GitPushResponse{}), nil
		}
	}
	// if native git push fails, try with ephemeral token - this may be a managed git project

	// Get authenticated admin client
	if !s.app.ch.IsAuthenticated() {
		return nil, errors.New("must authenticate before performing this action")
	}

	project, err := s.app.ch.LoadProject(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("git credentials not set and repo is not connected to a project")
	}

	author, err = s.app.ch.GitSignature(ctx, s.app.ProjectPath)
	if err != nil {
		return nil, err
	}

	config, err := s.app.ch.GitHelper(project.OrgName, project.Name, s.app.ProjectPath).GitConfig(ctx)
	if err != nil {
		return nil, err
	}
	config.ManagedRepo = project.ManagedGitId != ""
	err = gitutil.SetRemote(s.app.ProjectPath, config.Remote)
	if err != nil {
		return nil, err
	}

	// fetch the status again
	gs, err := gitutil.RunGitStatus(s.app.ProjectPath, "rill")
	if err != nil {
		return nil, err
	}
	if gs.RemoteCommits > 0 && !r.Msg.Force {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot push with remote commits present, please pull first"))
	}

	err = gitutil.CommitAndForcePush(ctx, s.app.ProjectPath, config, r.Msg.CommitMessage, author, true)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.GitPushResponse{}), nil
}
